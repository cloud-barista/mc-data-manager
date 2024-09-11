package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/internal/execfunc"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/service/nrdbc"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/cloud-barista/mc-data-manager/service/rdbc"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

var (
	managerInstance *FileScheduleManager
	once            sync.Once
)

// FileScheduleManager manages task schedules, flows, and tasks.
type FileScheduleManager struct {
	tasks     []models.DataTask
	flows     []models.Flow
	schedules []models.Schedule
	mu        sync.Mutex
	filename  string
	scheduler *gocron.Scheduler
}

// InitFileScheduleManager initializes the singleton instance of FileScheduleManager.
func InitFileScheduleManager() *FileScheduleManager {
	once.Do(func() {
		filename := "./data/var/run/data-manager/task/task.json"

		managerInstance = &FileScheduleManager{
			tasks:     make([]models.DataTask, 0),
			flows:     make([]models.Flow, 0),
			schedules: make([]models.Schedule, 0),
			filename:  filename,
			scheduler: gocron.NewScheduler(time.UTC),
		}

		if err := managerInstance.loadFromFile(); err != nil {
			logrus.Errorf("Failed to load tasks from file: %v", err)
			managerInstance = nil
			return
		}

		managerInstance.StartScheduler()
	})

	if managerInstance == nil {
		logrus.Error("FileScheduleManager initialization failed")
	}
	return managerInstance
}

// StartScheduler starts the gocron scheduler.
func (m *FileScheduleManager) StartScheduler() {
	m.scheduler.StartAsync()
}

// StopScheduler stops the gocron scheduler.
func (m *FileScheduleManager) StopScheduler() {
	m.scheduler.Stop()
}

// loadFromFile loads the schedules from the specified file.
func (m *FileScheduleManager) loadFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.filename)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("Task file %s does not exist, skipping load", m.filename)
			return nil
		}
		return fmt.Errorf("failed to open task file %s: %w", m.filename, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	data := struct {
		Tasks     []models.DataTask `json:"tasks"`
		Flows     []models.Flow     `json:"flows"`
		Schedules []models.Schedule `json:"schedules"`
	}{}

	err = decoder.Decode(&data)
	if err != nil {
		logrus.Errorf("Failed to decode task file %s: %v. Saving corrupted file as task_err.json and skipping load.", m.filename, err)

		// Create a backup of the corrupted file as task_err.json
		err = backupAndRemoveCorruptedFile(m.filename)
		if err != nil {
			return fmt.Errorf("failed to backup and remove corrupted file: %w", err)
		}

		return nil
	}

	m.tasks = data.Tasks
	m.flows = data.Flows
	m.schedules = data.Schedules

	for _, schedule := range m.schedules {
		_, err := m.scheduler.Cron(schedule.Cron).Tag(schedule.ScheduleID).Do(m.RunTasks, schedule.Tasks)
		if err != nil {
			return fmt.Errorf("failed to schedule tasks for schedule %s: %w", schedule.ScheduleID, err)
		}
	}

	logrus.Infof("Successfully loaded and scheduled %d tasks from %s", len(m.schedules), m.filename)
	return nil
}

func backupAndRemoveCorruptedFile(srcFilename string) error {
	// Define the backup filename
	backupFilename := filepath.Join(filepath.Dir(srcFilename), "task_err.json")

	// Open the source file
	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcFilename, err)
	}
	defer srcFile.Close()

	// Create the destination backup file
	destFile, err := os.Create(backupFilename)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", backupFilename, err)
	}
	defer destFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", srcFilename, backupFilename, err)
	}

	// Close files before removing the original file
	srcFile.Close()
	destFile.Close()

	// Remove the original corrupted file
	err = os.Remove(srcFilename)
	if err != nil {
		return fmt.Errorf("failed to remove the original file %s: %w", srcFilename, err)
	}

	return nil
}

// saveToFile saves the schedules to the specified file.
func (m *FileScheduleManager) saveToFile() error {

	data := struct {
		Tasks     []models.DataTask `json:"tasks"`
		Flows     []models.Flow     `json:"flows"`
		Schedules []models.Schedule `json:"schedules"`
	}{
		Tasks:     m.tasks,
		Flows:     m.flows,
		Schedules: m.schedules,
	}
	// Ensure the directory exists
	dir := filepath.Dir(m.filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories %s: %w", dir, err)
	}

	file, err := os.Create(m.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(&data)
}

// CreateSchedule creates a new schedule, saves it to the file, and registers it with the scheduler.
func (m *FileScheduleManager) CreateSchedule(schedule models.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if schedule.OperationId == "" {
		return errors.New("OperationId is required")
	}

	schedule.ScheduleID = utils.GenerateScheduleID(schedule.OperationId)

	for i, task := range schedule.Tasks {
		task.TaskMeta.TaskID = utils.GenerateTaskID(schedule.OperationId, i)
		m.tasks = append(m.tasks, task)
	}

	m.schedules = append(m.schedules, schedule)

	// Register the schedule with gocron using the Cron expression
	if schedule.TimeZone != "" {
		loc, err := time.LoadLocation(schedule.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone: %v", err)
		}
		m.scheduler.ChangeLocation(loc)
	} else {
		m.scheduler.ChangeLocation(time.UTC) // Default to UTC if no time zone is specified
	}

	_, err := m.scheduler.Cron(schedule.Cron).Tag(schedule.ScheduleID).Do(m.RunTasks, schedule.Tasks)
	if err != nil {
		return fmt.Errorf("failed to schedule tasks: %v", err)
	}

	return m.saveToFile()
}

// GetSchedule retrieves a schedule by its ID or OperationID.
func (m *FileScheduleManager) GetSchedule(id string) (models.Schedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try to find by ScheduleID
	for _, schedule := range m.schedules {
		if schedule.ScheduleID == id || schedule.OperationId == id {
			return schedule, nil
		}
	}

	return models.Schedule{}, errors.New("schedule not found")
}

// GetScheduleList retrieves a list of all schedules.
func (m *FileScheduleManager) GetScheduleList() ([]models.Schedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.schedules, nil
}

// UpdateSchedule updates an existing schedule by ScheduleID or OperationID.
func (m *FileScheduleManager) UpdateSchedule(id string, updatedSchedule models.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, schedule := range m.schedules {
		if schedule.ScheduleID == id || schedule.OperationId == id {
			// Remove the existing schedule from gocron
			m.scheduler.RemoveByTag(schedule.ScheduleID)

			// Update the schedule details
			updatedSchedule.ScheduleID = schedule.ScheduleID
			m.schedules[i] = updatedSchedule

			// Clear the existing tasks associated with this schedule
			m.tasks = []models.DataTask{}

			// Iterate over the tasks and unmarshal them into DataTask objects
			for j, task := range updatedSchedule.Tasks {
				// Assign a new TaskID to each task
				task.TaskMeta.TaskID = utils.GenerateTaskID(schedule.ScheduleID, j)
				m.tasks = append(m.tasks, task)
			}

			// Re-register the updated schedule with gocron
			_, err := m.scheduler.Cron(updatedSchedule.Cron).Tag(updatedSchedule.ScheduleID).Do(m.RunTasks, updatedSchedule.Tasks)
			if err != nil {
				return fmt.Errorf("failed to schedule tasks: %v", err)
			}

			return m.saveToFile()
		}
	}

	return errors.New("schedule not found")
}

// DeleteSchedule deletes a schedule by ScheduleID or OperationID.
func (m *FileScheduleManager) DeleteSchedule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, schedule := range m.schedules {
		if schedule.ScheduleID == id || schedule.OperationId == id {
			// Remove the schedule from gocron
			m.scheduler.RemoveByTag(schedule.ScheduleID)

			// Delete the schedule from the internal lists
			m.schedules = append(m.schedules[:i], m.schedules[i+1:]...)

			for j, flow := range m.flows {
				if flow.OperationId == schedule.OperationId {
					m.flows = append(m.flows[:j], m.flows[j+1:]...)
					break
				}
			}

			for j, task := range m.tasks {
				if task.TaskMeta.TaskID == schedule.ScheduleID {
					m.tasks = append(m.tasks[:j], m.tasks[j+1:]...)
					break
				}
			}

			return m.saveToFile()
		}
	}

	return errors.New("schedule not found")
}

// runTasks executes the tasks associated with a schedule.
func (m *FileScheduleManager) RunTasks(tasks []models.DataTask) {
	for _, task := range tasks {
		// Call the handleTask function to process the task
		task.Status = handleTask(task.ServiceType, task.TaskType, task)
		m.updateTaskStatus(task)

	}
	err := m.saveToFile()
	if err != nil {
		fmt.Printf("Error saving tasks to file: %v\n", err)
	}
}

// handleTask is a function that processes a task based on its ServiceType and TaskType.
func handleTask(serviceType models.CloudServiceType, taskType models.TaskType, params models.DataTask) models.Status {

	var taskStatus models.Status

	switch serviceType {

	case "objectStorage":
		switch taskType {
		case "generate":
			taskStatus = handleGenTest(params)
		case "migrate":
			taskStatus = handleObjectStorageMigrateTask(params)
		case "backup":
			taskStatus = handleObjectStorageBackupTask(params)
		case "restore":
			taskStatus = handleObjectStorageRestoreTask(params)
		default:
			fmt.Printf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed
		}
	case "rdbms":
		switch taskType {
		case "generate":
			taskStatus = handleGenTest(params)
		case "migrate":
			taskStatus = handleRDBMSMigrateTask(params)
		case "backup":
			taskStatus = handleRDBMSBackupTask(params)
		case "restore":
			taskStatus = handleRDBMSRestoreTask(params)
		default:
			fmt.Printf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed

		}
	case "nrdbms":
		switch taskType {
		case "generate":
			taskStatus = handleGenTest(params)
		case "migrate":
			taskStatus = handleNRDBMSMigrateTask(params)
		case "backup":
			taskStatus = handleNRDBMSBackupTask(params)
		case "restore":
			taskStatus = handleNRDBMSRestoreTask(params)
		default:
			fmt.Printf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed

		}
	default:
		fmt.Printf("Error: Unknown ServiceType: %s\n", serviceType)
		taskStatus = models.StatusFailed

	}

	return taskStatus
}

func handleGenTest(params models.DataTask) models.Status {
	logrus.Infof("Handling object storage Gen task")
	_ = params
	var cParams models.CommandTask
	cParams.SizeServerSQL = "1"
	cParams.DummyPath = "./tmp/Schedule/dummy"
	execfunc.DummyCreate(cParams)
	return models.StatusCompleted
}

func handleObjectStorageMigrateTask(params models.DataTask) models.Status {
	fmt.Println("Handling object storage migrate task")

	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error

	logrus.Infof("Source Information")
	src, srcErr = auth.GetOS(&params.SourcePoint)
	if srcErr != nil {
		logrus.Errorf("OSController error migration into objectstorage : %v", srcErr)
		return models.StatusFailed
	}
	logrus.Infof("Target Information")
	dst, dstErr = auth.GetOS(&params.TargetPoint)
	if dstErr != nil {
		logrus.Errorf("OSController error migration into objectstorage : %v", dstErr)
		return models.StatusFailed
	}

	logrus.Info("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		logrus.Errorf("Copy error copying into objectstorage : %v", err)
		return models.StatusFailed
	}
	logrus.Info("successfully migrationed")
	return models.StatusCompleted
}

func handleObjectStorageBackupTask(params models.DataTask) models.Status {
	fmt.Println("Handling object storage backup task")
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return models.StatusFailed
	}

	logrus.Info("Launch OSController MGet")
	if err := OSC.MGet(params.Directory); err != nil {
		logrus.Errorf("MGet error exporting into objectstorage : %v", err)
		return models.StatusFailed
	}
	logrus.Infof("successfully backup : %s", params.Directory)
	return models.StatusCompleted
}

func handleObjectStorageRestoreTask(params models.DataTask) models.Status {
	fmt.Println("Handling object storage restore task")
	var OSC *osc.OSController
	var err error
	logrus.Infof("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("OSController error importing into objectstorage : %v", err)
		return models.StatusFailed
	}

	logrus.Info("Launch OSController MGet")
	if err := OSC.MPut(params.SourcePoint.Path); err != nil {
		logrus.Errorf("MPut error importing into objectstorage : %v", err)
		return models.StatusFailed
	}
	logrus.Infof("successfully restore : %s", params.Directory)
	return models.StatusCompleted
}

func handleRDBMSMigrateTask(params models.DataTask) models.Status {
	fmt.Println("Handling RDBMS migrate task")
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	logrus.Infof("Source Information")
	srcRDBC, srcErr = auth.GetRDMS(&params.SourcePoint)
	if srcErr != nil {
		logrus.Errorf("RDBController error migration into rdbms : %v", srcErr)
		return models.StatusFailed
	}
	logrus.Infof("Target Information")
	dstRDBC, dstErr = auth.GetRDMS(&params.TargetPoint)
	if dstErr != nil {
		logrus.Errorf("RDBController error migration into rdbms : %v", dstErr)
		return models.StatusFailed
	}

	logrus.Info("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		logrus.Errorf("Copy error copying into rdbms : %v", err)
		return models.StatusFailed
	}
	logrus.Info("successfully migrationed")
	return models.StatusCompleted

}

func handleRDBMSBackupTask(params models.DataTask) models.Status {
	fmt.Println("Handling RDBMS backup task")
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return models.StatusFailed
	}

	err = os.MkdirAll(params.Directory, 0755)
	if err != nil {
		logrus.Errorf("MkdirAll error : %v", err)
		return models.StatusFailed
	}

	dbList := []string{}
	if err := RDBC.ListDB(&dbList); err != nil {
		logrus.Errorf("ListDB error : %v", err)
		return models.StatusFailed
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		logrus.Infof("Export start: %s", db)
		if err := RDBC.Get(db, &sqlData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return models.StatusFailed
		}

		file, err := os.Create(filepath.Join(params.Directory, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return models.StatusFailed
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			logrus.Errorf("File write error : %v", err)
			return models.StatusFailed
		}
		logrus.Infof("successfully exported : %s", file.Name())
		file.Close()
	}
	logrus.Infof("successfully backup : %s", params.Directory)
	return models.StatusCompleted

}

func handleRDBMSRestoreTask(params models.DataTask) models.Status {
	fmt.Println("Handling RDBMS restore task")
	var RDBC *rdbc.RDBController
	var err error
	logrus.Infof("User Information")
	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("RDBController error importing into rdbms : %v", err)
		return models.StatusFailed
	}

	sqlList := []string{}
	err = filepath.Walk(params.Directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".sql" {
			sqlList = append(sqlList, path)
		}
		return nil
	})
	if err != nil {
		logrus.Errorf("Walk error : %v", err)
		return models.StatusFailed
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			logrus.Errorf("ReadFile error : %v", err)
			return models.StatusFailed
		}
		logrus.Infof("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			logrus.Error("Put error importing into rdbms")
			return models.StatusFailed
		}
		logrus.Infof("Import success: %s", sqlPath)
	}
	logrus.Infof("successfully restore : %s", params.Directory)
	return models.StatusCompleted

}

func handleNRDBMSMigrateTask(params models.DataTask) models.Status {
	fmt.Println("Handling NRDBMS migrate task")
	var srcNRDBC *nrdbc.NRDBController
	var srcErr error
	var dstNRDBC *nrdbc.NRDBController
	var dstErr error
	logrus.Infof("Source Information")
	srcNRDBC, srcErr = auth.GetNRDMS(&params.SourcePoint)
	if srcErr != nil {
		logrus.Errorf("NRDBController error migration into nrdbms : %v", srcErr)
		return models.StatusFailed
	}
	logrus.Infof("Target Information")
	dstNRDBC, dstErr = auth.GetNRDMS(&params.TargetPoint)
	if dstErr != nil {
		logrus.Errorf("NRDBController error migration into nrdbms : %v", dstErr)
		return models.StatusFailed
	}

	logrus.Info("Launch NRDBController Copy")
	if err := srcNRDBC.Copy(dstNRDBC); err != nil {
		logrus.Errorf("Copy error copying into nrdbms : %v", err)
		return models.StatusFailed
	}
	logrus.Info("successfully migrationed")
	return models.StatusCompleted

}

func handleNRDBMSBackupTask(params models.DataTask) models.Status {
	fmt.Println("Handling NRDBMS backup task")
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("NRDBController error importing into nrdbms : %v", err)
		return models.StatusFailed
	}

	tableList, err := NRDBC.ListTables()
	if err != nil {
		logrus.Infof("ListTables error : %v", err)
		return models.StatusFailed
	}

	if !utils.FileExists(params.Directory) {
		logrus.Infof("directory does not exist")
		logrus.Infof("Make Directory")
		err = os.MkdirAll(params.Directory, 0755)
		if err != nil {
			logrus.Infof("Make Failed 0755 : %s", params.Directory)
			return models.StatusFailed
		}
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		logrus.Infof("Export start: %s", table)
		dstData = []map[string]interface{}{}

		if err := NRDBC.Get(table, &dstData); err != nil {
			logrus.Errorf("Get error : %v", err)
			return models.StatusFailed
		}

		file, err := os.Create(filepath.Join(params.Directory, fmt.Sprintf("%s.json", table)))
		if err != nil {
			logrus.Errorf("File create error : %v", err)
			return models.StatusFailed
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			logrus.Errorf("data encoding error : %v", err)
			return models.StatusFailed
		}
		logrus.Infof("successfully create File : %s", file.Name())
	}
	logrus.Infof("successfully backup to : %s", params.Directory)
	return models.StatusCompleted

}

func handleNRDBMSRestoreTask(params models.DataTask) models.Status {
	fmt.Println("Handling NRDBMS restore task")
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logrus.Errorf("NRDBController error importing into nrdbms : %v", err)
		return models.StatusFailed
	}

	jsonList := []string{}
	err = filepath.Walk(params.Directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".json" {
			jsonList = append(jsonList, path)
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("Walk error : %v", err)
		return models.StatusFailed
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			logrus.Errorf("file open error : %v", err)
			return models.StatusFailed
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			logrus.Errorf("file decoding error : %v", err)
			return models.StatusFailed
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		logrus.Infof("Import start: %s", fileName)
		if err := NRDBC.Put(tableName, &srcData); err != nil {
			logrus.Error("Put error importing into nrdbms")
			return models.StatusFailed
		}
		logrus.Infof("successfully Restore : %s", params.Directory)
	}
	return models.StatusCompleted

}

// Facade function to create a new schedule and manage it.
func (m *FileScheduleManager) CreateAndStartSchedule(schedule models.Schedule) error {
	m.StopScheduler()
	defer m.StartScheduler()

	if err := m.CreateSchedule(schedule); err != nil {
		return err
	}

	return nil
}

// Facade function to update a schedule.
func (m *FileScheduleManager) UpdateAndRestartSchedule(scheduleID string, updatedSchedule models.Schedule) error {
	m.StopScheduler()
	defer m.StartScheduler()

	if err := m.UpdateSchedule(scheduleID, updatedSchedule); err != nil {
		return err
	}

	return nil
}

// Facade function to delete a schedule.
func (m *FileScheduleManager) DeleteAndRestartScheduler(scheduleID string) error {
	m.StopScheduler()
	defer m.StartScheduler()

	if err := m.DeleteSchedule(scheduleID); err != nil {
		return err
	}

	return nil
}

// updateTaskStatus updates the status of the task in the internal data structure
func (m *FileScheduleManager) updateTaskStatus(task models.DataTask) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Iterate over the slice to find the task by TaskID
	for i, existingTask := range m.tasks {
		if existingTask.TaskMeta.TaskID == task.TaskMeta.TaskID {
			// Update the status of the task
			m.tasks[i].Status = task.Status
			return
		}
	}
}
