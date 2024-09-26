package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	"github.com/rs/zerolog/log"
)

var (
	managerInstance *FileScheduleManager
	once            sync.Once
)

// FileScheduleManager manages task schedules, flows, and tasks.
type FileScheduleManager struct {
	tasks      []models.BasicDataTask
	flows      []models.Flow
	schedules  []models.Schedule
	mu         sync.Mutex
	filename   string
	schedulers map[string]*gocron.Scheduler // Map of time zone to its scheduler
}

// InitFileScheduleManager initializes the singleton instance of FileScheduleManager.
func InitFileScheduleManager() *FileScheduleManager {
	once.Do(func() {
		filename := "./data/var/run/data-manager/task/task.json"

		managerInstance = &FileScheduleManager{
			tasks:      make([]models.BasicDataTask, 0),
			flows:      make([]models.Flow, 0),
			schedules:  make([]models.Schedule, 0),
			filename:   filename,
			schedulers: make(map[string]*gocron.Scheduler),
		}

		if err := managerInstance.loadFromFile(); err != nil {
			log.Error().Err(err).Msg("Failed to load tasks from file")
			managerInstance = nil
			return
		}

		managerInstance.StartSchedulers()
	})

	if managerInstance == nil {
		log.Error().Msg("FileScheduleManager initialization failed")
	}
	return managerInstance
}

// StartSchedulers starts all schedulers asynchronously.
func (m *FileScheduleManager) StartSchedulers() {
	for tz, scheduler := range m.schedulers {
		log.Info().Str("time_zone", tz).Msg("Starting scheduler")
		go scheduler.StartAsync()
	}
}

// StopSchedulers stops all schedulers.
func (m *FileScheduleManager) StopSchedulers() {
	for tz, scheduler := range m.schedulers {
		log.Info().Str("time_zone", tz).Msg("Stopping scheduler")
		scheduler.Stop()
	}
}

// loadFromFile loads the schedules from the specified file.
func (m *FileScheduleManager) loadFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Str("filename", m.filename).Msg("Task file does not exist, skipping load")
			return nil
		}
		return fmt.Errorf("failed to open task file %s: %w", m.filename, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	data := struct {
		Tasks     []models.BasicDataTask `json:"tasks"`
		Flows     []models.Flow          `json:"flows"`
		Schedules []models.Schedule      `json:"schedules"`
	}{}

	err = decoder.Decode(&data)
	if err != nil {
		log.Error().Err(err).Str("filename", m.filename).Msg("Failed to decode task file. Saving corrupted file as task_err.json and skipping load.")

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
		err := m.registerSchedule(schedule)
		if err != nil {
			return fmt.Errorf("failed to schedule tasks for schedule %s: %w", schedule.ScheduleID, err)
		}
	}

	log.Info().Int("schedules", len(m.schedules)).Str("filename", m.filename).Msg("Successfully loaded and scheduled tasks")
	return nil
}

// backupAndRemoveCorruptedFile creates a backup of a corrupted file and removes the original.
func backupAndRemoveCorruptedFile(srcFilename string) error {
	backupFilename := filepath.Join(filepath.Dir(srcFilename), "task_err.json")

	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcFilename, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(backupFilename)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", backupFilename, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data from %s to %s: %w", srcFilename, backupFilename, err)
	}

	err = os.Remove(srcFilename)
	if err != nil {
		return fmt.Errorf("failed to remove the original file %s: %w", srcFilename, err)
	}

	return nil
}

// saveToFile saves the schedules to the specified file.
func (m *FileScheduleManager) saveToFile() error {
	data := struct {
		Tasks     []models.BasicDataTask `json:"tasks"`
		Flows     []models.Flow          `json:"flows"`
		Schedules []models.Schedule      `json:"schedules"`
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

// validateCronExpression checks if the provided cron expression is valid.
// It returns an error if the expression is invalid.
func validateCronExpression(cronExpr string) error {
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 {
		return errors.New("cron expression must have exactly 5 fields")
	}

	fieldPatterns := []string{
		`^(\*|([0-5]?\d)(-[0-5]?\d)?(\/\d+)?)$`, // Minute
		`^(\*|([01]?\d|2[0-3])(\/\d+)?)$`,       // Hour
		`^(\*|([01]?\d|2[0-9]|3[01])(\/\d+)?)$`, // Day of Month
		`^(\*|(1[0-2]|0?[1-9])(\/\d+)?)$`,       // Month
		`^(\*|(0|1|2|3|4|5|6)(\/\d+)?)$`,        // Day of Week
	}

	for i, field := range fields {
		matched, err := regexp.MatchString(fieldPatterns[i], field)
		if err != nil {
			return fmt.Errorf("error validating cron expression: %v", err)
		}
		if !matched {
			return fmt.Errorf("invalid cron expression in field %d", i+1)
		}
	}

	return nil
}

// registerSchedule registers a schedule with the appropriate scheduler based on its time zone.
func (m *FileScheduleManager) registerSchedule(schedule models.Schedule) error {
	// Determine the time zone
	var loc *time.Location
	if schedule.TimeZone != "" {
		var err error
		loc, err = time.LoadLocation(schedule.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone: %v", err)
		}
	} else {
		loc = time.UTC // Default to UTC if no time zone is specified
	}

	// Get or create the scheduler for the specified time zone
	tz := loc.String()
	scheduler, exists := m.schedulers[tz]
	if !exists {
		scheduler = gocron.NewScheduler(loc)
		m.schedulers[tz] = scheduler
		go scheduler.StartAsync()
	}

	// Validate and schedule the cron expression or one-time task
	if schedule.Cron != "" {
		if err := validateCronExpression(schedule.Cron); err != nil {
			return fmt.Errorf("invalid cron expression for schedule %s: %w", schedule.ScheduleID, err)
		}

		_, err := scheduler.Cron(schedule.Cron).Tag(schedule.ScheduleID).Do(m.RunTasks, schedule.Tasks)
		if err != nil {
			return fmt.Errorf("failed to schedule tasks for schedule %s: %w", schedule.ScheduleID, err)
		}
	} else if schedule.StartTime != nil {
		_, err := scheduler.
			Tag(schedule.ScheduleID).
			StartAt(*schedule.StartTime).
			LimitRunsTo(1).
			Do(m.RunTasks, schedule.Tasks)
		if err != nil {
			return fmt.Errorf("failed to schedule one-time task for schedule %s: %w", schedule.ScheduleID, err)
		}
	} else {
		// If neither Cron nor StartTime is specified, schedule to run immediately once
		_, err := scheduler.
			Every(1).
			Tag(schedule.ScheduleID).
			LimitRunsTo(1).
			Do(m.RunTasks, schedule.Tasks)
		if err != nil {
			return fmt.Errorf("failed to schedule immediate task for schedule %s: %w", schedule.ScheduleID, err)
		}

	}
	return nil
}

// CreateSchedule creates a new schedule, saves it to the file, and registers it with the scheduler.
// It handles multiple time zones by assigning schedules to their respective schedulers.
// If a schedule with the same ScheduleID already exists, it rejects the registration.
func (m *FileScheduleManager) CreateSchedule(schedule models.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if schedule.OperationId == "" {
		return errors.New("OperationId is required")
	}

	// Generate ScheduleID based on OperationId
	schedule.ScheduleID = utils.GenerateScheduleID(schedule.OperationId)

	// Check if a schedule with the same ScheduleID already exists
	for _, existingSchedule := range m.schedules {
		if existingSchedule.OperationId == schedule.OperationId {
			return fmt.Errorf("schedule with operation ID %s already exists", schedule.OperationId)
		}
	}

	// Initialize tasks and assign TaskIDs
	for i, task := range schedule.Tasks {
		task.TaskMeta.TaskID = utils.GenerateTaskID(schedule.OperationId, i)
		schedule.Tasks[i] = task
		m.tasks = append(m.tasks, task)
	}

	// Add the new schedule to the list
	m.schedules = append(m.schedules, schedule)

	// Register the schedule with the appropriate scheduler
	if err := m.registerSchedule(schedule); err != nil {
		return fmt.Errorf("failed to register schedule: %v", err)
	}

	return m.saveToFile()
}

// GetSchedule retrieves a schedule by its ID or OperationID.
func (m *FileScheduleManager) GetSchedule(id string) (models.Schedule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
// It handles time zone changes by moving the schedule to the appropriate scheduler.
func (m *FileScheduleManager) UpdateSchedule(id string, updatedSchedule models.Schedule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, schedule := range m.schedules {
		if schedule.ScheduleID == id || schedule.OperationId == id {
			// Remove the existing schedule from its scheduler
			if err := m.removeSchedule(schedule); err != nil {
				return fmt.Errorf("failed to remove existing schedule: %v", err)
			}

			// Preserve the ScheduleID
			updatedSchedule.ScheduleID = schedule.ScheduleID

			// Update tasks
			m.tasks = []models.BasicDataTask{}
			for j, task := range updatedSchedule.Tasks {
				task.TaskMeta.TaskID = utils.GenerateTaskID(schedule.ScheduleID, j)
				m.tasks = append(m.tasks, task)
			}

			// Update the schedule in the list
			m.schedules[i] = updatedSchedule

			// Register the updated schedule
			if err := m.registerSchedule(updatedSchedule); err != nil {
				return fmt.Errorf("failed to register updated schedule: %v", err)
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
			// Remove the schedule from its scheduler
			if err := m.removeSchedule(schedule); err != nil {
				return fmt.Errorf("failed to remove schedule: %v", err)
			}

			// Remove the schedule from the list
			m.schedules = append(m.schedules[:i], m.schedules[i+1:]...)

			// Optionally, remove associated flows and tasks if necessary
			// Remove flows
			for j, flow := range m.flows {
				if flow.OperationId == schedule.OperationId {
					m.flows = append(m.flows[:j], m.flows[j+1:]...)
					break
				}
			}

			// Remove tasks
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

// removeSchedule removes a schedule from its scheduler based on its time zone.
func (m *FileScheduleManager) removeSchedule(schedule models.Schedule) error {
	// Determine the time zone
	var loc *time.Location
	if schedule.TimeZone != "" {
		var err error
		loc, err = time.LoadLocation(schedule.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid time zone: %v", err)
		}
	} else {
		loc = time.UTC // Default to UTC if no time zone is specified
	}

	// Get the scheduler for the time zone
	tz := loc.String()
	scheduler, exists := m.schedulers[tz]
	if !exists {
		return fmt.Errorf("scheduler for time zone %s does not exist", tz)
	}

	// Remove the job by tag
	scheduler.RemoveByTag(schedule.ScheduleID)

	return nil
}

// RunTasks executes the tasks associated with a schedule.
func (m *FileScheduleManager) RunTasks(tasks []models.BasicDataTask) {
	for _, task := range tasks {
		// Call the handleTask function to process the task
		task.Status = handleTask(task.ServiceType, task.TaskType, task)
		log.Debug().Msgf("status : %v", task.Status)
		m.updateTaskStatus(task)
	}

	err := m.saveToFile()
	if err != nil {
		log.Error().Err(err).Msg("Error saving tasks to file")
	}
}

// handleTask is a function that processes a task based on its ServiceType and TaskType.
func handleTask(serviceType models.CloudServiceType, taskType models.TaskType, params models.BasicDataTask) models.Status {

	var taskStatus models.Status

	switch serviceType {

	case "objectstorage":
		switch taskType {
		case "generate":
			taskStatus = handleObjectStorageGenerateTask(params)
		case "migrate":
			taskStatus = handleObjectStorageMigrateTask(params)
		case "backup":
			taskStatus = handleObjectStorageBackupTask(params)
		case "restore":
			taskStatus = handleObjectStorageRestoreTask(params)
		case "delete":
			taskStatus = handleObjectStorageDeleteTask(params)
		default:
			log.Error().Msgf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed
		}
	case "rdbms":
		switch taskType {
		case "generate":
			taskStatus = handleRDBMSGenerateTask(params)
		case "migrate":
			taskStatus = handleRDBMSMigrateTask(params)
		case "backup":
			taskStatus = handleRDBMSBackupTask(params)
		case "restore":
			taskStatus = handleRDBMSRestoreTask(params)
		case "delete":
			taskStatus = handleRDBMSDeleteTask(params)
		default:
			log.Error().Msgf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed
		}
	case "nrdbms":
		switch taskType {
		case "generate":
			taskStatus = handleNRDBMSGenerateTask(params)
		case "migrate":
			taskStatus = handleNRDBMSMigrateTask(params)
		case "backup":
			taskStatus = handleNRDBMSBackupTask(params)
		case "restore":
			taskStatus = handleNRDBMSRestoreTask(params)
		case "delete":
			taskStatus = handleNRDBMSDeleteTask(params)
		default:
			log.Error().Msgf("Error: Unknown TaskType: %s for ServiceType: %s\n", taskType, serviceType)
			taskStatus = models.StatusFailed
		}
	default:
		log.Error().Msgf("Error: Unknown ServiceType: %s\n", serviceType)
		taskStatus = models.StatusFailed

	}

	return taskStatus
}

// Test func
func handleGenTest(params models.BasicDataTask) models.Status {

	log.Info().Msg("Handling object storage Gen task")
	_ = params
	var cParams models.CommandTask
	cParams.SizeServerSQL = "1"
	cParams.DummyPath = "./tmp/Schedule/dummy"
	execfunc.DummyCreate(cParams)
	return models.StatusCompleted
}

func handleObjectStorageGenerateTask(params models.BasicDataTask) models.Status {

	var OSC *osc.OSController
	var err error
	log.Info().Msgf("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return models.StatusFailed
	}

	log.Info().Msgf("Launch OSController MPut")
	if err := OSC.MPut(params.Directory); err != nil {
		log.Error().Msgf("MPut error importing into objectstorage")
		log.Info().Msgf("params : %+v", params.TargetPoint)

		return models.StatusFailed
	}
	log.Info().Msgf("successfully imported : %s", params.Directory)
	return models.StatusCompleted
}

func handleObjectStorageDeleteTask(params models.BasicDataTask) models.Status {

	var OSC *osc.OSController
	var err error
	log.Info().Msgf("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return models.StatusFailed
	}

	log.Info().Msgf("Launch OSController Delete")
	if err := OSC.DeleteBucket(); err != nil {
		log.Error().Msgf("Delete error deleting into objectstorage : %v", err)
		return models.StatusFailed
	}
	log.Info().Msgf("successfully deleted")

	return models.StatusCompleted
}

func handleObjectStorageMigrateTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling object storage migrate task")

	var src *osc.OSController
	var srcErr error
	var dst *osc.OSController
	var dstErr error

	log.Info().Msg("Source Information")
	src, srcErr = auth.GetOS(&params.SourcePoint)
	if srcErr != nil {
		log.Error().Err(srcErr).Msg("OSController error migration into object storage")
		return models.StatusFailed
	}
	log.Info().Msg("Target Information")
	dst, dstErr = auth.GetOS(&params.TargetPoint)
	if dstErr != nil {
		log.Error().Err(dstErr).Msg("OSController error migration into object storage")
		return models.StatusFailed
	}

	log.Info().Msg("Launch OSController Copy")
	if err := src.Copy(dst); err != nil {
		log.Error().Err(err).Msg("Copy error copying into object storage")
		return models.StatusFailed
	}
	log.Info().Msg("Successfully migrated")
	return models.StatusCompleted
}

func handleObjectStorageBackupTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling object storage backup task")
	var OSC *osc.OSController
	var err error
	log.Info().Msg("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("OSController error importing into objectstorage ")
		return models.StatusFailed
	}

	log.Info().Msg("Launch OSController MGet")
	if err := OSC.MGet(params.Directory); err != nil {
		log.Error().Err(err).Msg("MGet error exporting into objectstorage ")
		return models.StatusFailed
	}
	log.Info().Msgf("successfully backup : %s", params.Directory)
	return models.StatusCompleted
}

func handleObjectStorageRestoreTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling object storage restore task")
	var OSC *osc.OSController
	var err error
	log.Info().Msg("User Information")
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("OSController error importing into objectstorage ")
		return models.StatusFailed
	}

	log.Info().Msg("Launch OSController MGet")
	if err := OSC.MPut(params.SourcePoint.Path); err != nil {
		log.Error().Err(err).Msg("MPut error importing into objectstorage ")
		return models.StatusFailed
	}
	log.Info().Msgf("successfully restore : %s", params.Directory)
	return models.StatusCompleted
}

func handleRDBMSGenerateTask(params models.BasicDataTask) models.Status {
	var RDBC *rdbc.RDBController
	var err error
	log.Info().Msgf("User Information")
	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("RDBController error importing into rdbms : %v", err)
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
		log.Error().Msgf("Walk error : %v", err)
		return models.StatusFailed
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			log.Error().Msgf("ReadFile error : %v", err)
			return models.StatusFailed
		}
		log.Info().Msgf("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			log.Error().Msgf("Put error importing into rdbms")
			return models.StatusFailed
		}
		log.Info().Msgf("Import success: %s", sqlPath)
	}
	log.Info().Msgf("successfully imported : %s", params.Directory)
	return models.StatusCompleted
}

func handleRDBMSDeleteTask(params models.BasicDataTask) models.Status {
	var RDBC *rdbc.RDBController
	var err error
	RDBC, err = auth.GetRDMS(&params.TargetPoint)

	if err != nil {
		log.Error().Msgf("RDBController error deleting into rdbms : %v", err)
		return models.StatusFailed
	}

	var dbList []string
	if err := RDBC.ListDB(&dbList); err != nil {
		log.Error().Err(err).Msgf("ListDB error : %s", err)
		return models.StatusFailed
	}

	log.Info().Msgf("Launch RDBController Delete")
	if err := RDBC.DeleteDB(dbList...); err != nil {
		log.Error().Msgf("Delete error deleting into rdbms : %v", err)
		return models.StatusFailed
	}
	log.Info().Msgf("successfully deleted")
	return models.StatusCompleted
}

func handleRDBMSMigrateTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling RDBMS migrate task")
	var srcRDBC *rdbc.RDBController
	var srcErr error
	var dstRDBC *rdbc.RDBController
	var dstErr error
	log.Info().Msg("Source Information")
	srcRDBC, srcErr = auth.GetRDMS(&params.SourcePoint)
	if srcErr != nil {
		log.Error().Err(srcErr).Msg("RDBController error migration into rdbms ")
		return models.StatusFailed
	}
	log.Info().Msg("Target Information")
	dstRDBC, dstErr = auth.GetRDMS(&params.TargetPoint)
	if dstErr != nil {
		log.Error().Err(dstErr).Msg("RDBController error migration into rdbms ")
		return models.StatusFailed
	}

	log.Info().Msg("Launch RDBController Copy")
	if err := srcRDBC.Copy(dstRDBC); err != nil {
		log.Error().Err(err).Msg("Copy error copying into rdbms ")
		return models.StatusFailed
	}
	log.Info().Msg("successfully migrationed")
	return models.StatusCompleted

}

func handleRDBMSBackupTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling RDBMS backup task")
	var RDBC *rdbc.RDBController
	var err error
	log.Info().Msg("User Information")
	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("RDBController error importing into rdbms ")
		return models.StatusFailed
	}

	err = os.MkdirAll(params.Directory, 0755)
	if err != nil {
		log.Error().Err(err).Msg("MkdirAll error ")
		return models.StatusFailed
	}

	dbList := []string{}
	if err := RDBC.ListDB(&dbList); err != nil {
		log.Error().Err(err).Msg("ListDB error ")
		return models.StatusFailed
	}

	var sqlData string
	for _, db := range dbList {
		sqlData = ""
		log.Info().Msgf("Export start: %s", db)
		if err := RDBC.Get(db, &sqlData); err != nil {
			log.Error().Err(err).Msg("Get error ")
			return models.StatusFailed
		}

		file, err := os.Create(filepath.Join(params.Directory, fmt.Sprintf("%s.sql", db)))
		if err != nil {
			log.Error().Err(err).Msg("File create error ")
			return models.StatusFailed
		}
		defer file.Close()

		_, err = file.WriteString(sqlData)
		if err != nil {
			log.Error().Err(err).Msg("File write error ")
			return models.StatusFailed
		}
		log.Info().Msgf("successfully exported : %s", file.Name())
		file.Close()
	}
	log.Info().Msgf("successfully backup : %s", params.Directory)
	return models.StatusCompleted

}

func handleRDBMSRestoreTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling RDBMS restore task")
	var RDBC *rdbc.RDBController
	var err error
	log.Info().Msg("User Information")
	RDBC, err = auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("RDBController error importing into rdbms ")
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
		log.Error().Err(err).Msg("Walk error ")
		return models.StatusFailed
	}

	for _, sqlPath := range sqlList {
		data, err := os.ReadFile(sqlPath)
		if err != nil {
			log.Error().Err(err).Msg("ReadFile error ")
			return models.StatusFailed
		}
		log.Info().Msgf("Import start: %s", sqlPath)
		if err := RDBC.Put(string(data)); err != nil {
			log.Error().Msg("Put error importing into rdbms")
			return models.StatusFailed
		}
		log.Info().Msgf("Import success: %s", sqlPath)
	}
	log.Info().Msgf("successfully restore : %s", params.Directory)
	return models.StatusCompleted

}

func handleNRDBMSGenerateTask(params models.BasicDataTask) models.Status {

	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("NRDBController error importing into nrdbms : %v", err)
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
		log.Error().Msgf("Walk error : %v", err)
		return models.StatusFailed
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			log.Error().Msgf("file open error : %v", err)
			return models.StatusFailed
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			log.Error().Msgf("file decoding error : %v", err)
			return models.StatusFailed
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		log.Info().Msgf("Import start: %s", fileName)
		if err := NRDBC.Put(tableName, &srcData); err != nil {
			log.Error().Msgf("Put error importing into nrdbms")
			return models.StatusFailed
		}
		log.Info().Msgf("successfully imported : %s", params.Directory)
	}
	return models.StatusCompleted
}

func handleNRDBMSDeleteTask(params models.BasicDataTask) models.Status {

	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)

	if err != nil {
		log.Error().Msgf("NRDBController error deleting into nrdbms : %v", err)
		return models.StatusFailed
	}

	tbList, err := NRDBC.ListTables()
	if err != nil {
		log.Error().Err(err).Msgf("ListTable error : %s", err)
		return models.StatusFailed
	}

	log.Info().Msgf("Launch NRDBController Delete")
	if err := NRDBC.DeleteTables(tbList...); err != nil {
		log.Error().Msgf("Delete error deleting into nrdbms : %v", err)
		return models.StatusFailed
	}
	log.Info().Msgf("successfully deleted")

	return models.StatusCompleted
}

func handleNRDBMSMigrateTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling NRDBMS migrate task")
	var srcNRDBC *nrdbc.NRDBController
	var srcErr error
	var dstNRDBC *nrdbc.NRDBController
	var dstErr error
	log.Info().Msg("Source Information")
	srcNRDBC, srcErr = auth.GetNRDMS(&params.SourcePoint)
	if srcErr != nil {
		log.Error().Err(srcErr).Msg("NRDBController error migration into nrdbms ")
		return models.StatusFailed
	}
	log.Info().Msg("Target Information")
	dstNRDBC, dstErr = auth.GetNRDMS(&params.TargetPoint)
	if dstErr != nil {
		log.Error().Err(dstErr).Msg("NRDBController error migration into nrdbms ")
		return models.StatusFailed
	}

	log.Info().Msg("Launch NRDBController Copy")
	if err := srcNRDBC.Copy(dstNRDBC); err != nil {
		log.Error().Err(err).Msg("Copy error copying into nrdbms ")
		return models.StatusFailed
	}
	log.Info().Msg("successfully migrationed")
	return models.StatusCompleted

}

func handleNRDBMSBackupTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling NRDBMS backup task")
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("NRDBController error importing into nrdbms ")
		return models.StatusFailed
	}

	tableList, err := NRDBC.ListTables()
	if err != nil {
		log.Info().Msgf("ListTables error : %v", err)
		return models.StatusFailed
	}

	if !utils.FileExists(params.Directory) {
		log.Info().Msg("directory does not exist")
		log.Info().Msg("Make Directory")
		err = os.MkdirAll(params.Directory, 0755)
		if err != nil {
			log.Info().Msgf("Make Failed 0755 : %s", params.Directory)
			return models.StatusFailed
		}
	}

	var dstData []map[string]interface{}
	for _, table := range tableList {
		log.Info().Msgf("Export start: %s", table)
		dstData = []map[string]interface{}{}

		if err := NRDBC.Get(table, &dstData); err != nil {
			log.Error().Err(err).Msg("Get error ")
			return models.StatusFailed
		}

		file, err := os.Create(filepath.Join(params.Directory, fmt.Sprintf("%s.json", table)))
		if err != nil {
			log.Error().Err(err).Msg("File create error ")
			return models.StatusFailed
		}
		defer file.Close()

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "    ")
		if err := encoder.Encode(dstData); err != nil {
			log.Error().Err(err).Msg("data encoding error ")
			return models.StatusFailed
		}
		log.Info().Msgf("successfully create File : %s", file.Name())
	}
	log.Info().Msgf("successfully backup to : %s", params.Directory)
	return models.StatusCompleted

}

func handleNRDBMSRestoreTask(params models.BasicDataTask) models.Status {
	log.Info().Msg("Handling NRDBMS restore task")
	var NRDBC *nrdbc.NRDBController
	var err error
	NRDBC, err = auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("NRDBController error importing into nrdbms ")
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
		log.Error().Err(err).Msg("Walk error ")
		return models.StatusFailed
	}

	var srcData []map[string]interface{}
	for _, jsonFile := range jsonList {
		srcData = []map[string]interface{}{}

		file, err := os.Open(jsonFile)
		if err != nil {
			log.Error().Err(err).Msg("file open error ")
			return models.StatusFailed
		}
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&srcData); err != nil {
			log.Error().Err(err).Msg("file decoding error ")
			return models.StatusFailed
		}

		fileName := filepath.Base(jsonFile)
		tableName := fileName[:len(fileName)-len(filepath.Ext(fileName))]

		log.Info().Msgf("Import start: %s", fileName)
		if err := NRDBC.Put(tableName, &srcData); err != nil {
			log.Error().Msg("Put error importing into nrdbms")
			return models.StatusFailed
		}
		log.Info().Msgf("successfully Restore : %s", params.Directory)
	}
	return models.StatusCompleted

}

// CreateAndStartSchedule creates a new schedule and registers it without stopping any schedulers.
func (m *FileScheduleManager) CreateAndStartSchedule(schedule models.Schedule) error {
	if err := m.CreateSchedule(schedule); err != nil {
		return err
	}

	return nil
}

// UpdateAndRestartSchedule updates an existing schedule without stopping any schedulers.
func (m *FileScheduleManager) UpdateAndRestartSchedule(scheduleID string, updatedSchedule models.Schedule) error {
	if err := m.UpdateSchedule(scheduleID, updatedSchedule); err != nil {
		return err
	}

	return nil
}

// DeleteAndRestartScheduler deletes a schedule without stopping any schedulers.
func (m *FileScheduleManager) DeleteAndRestartScheduler(scheduleID string) error {
	if err := m.DeleteSchedule(scheduleID); err != nil {
		return err
	}

	return nil
}

// updateTaskStatus updates the status of the task in the internal data structure
func (m *FileScheduleManager) updateTaskStatus(task models.BasicDataTask) {
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
