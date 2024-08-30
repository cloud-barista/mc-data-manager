package models

import (
	"errors"
	"sync"

	"github.com/cloud-barista/mc-data-manager/models"
)

type Task = models.Task
type Flow = models.Flow
type Schedule = models.Schedule

type TaskService struct {
	tasks     map[string]Task
	flows     map[string]Flow
	schedules map[string]Schedule
	mu        sync.Mutex
}

func NewTaskService() *TaskService {
	return &TaskService{
		tasks:     make(map[string]Task),
		flows:     make(map[string]Flow),
		schedules: make(map[string]Schedule),
	}
}

func (s *TaskService) CreateTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.TaskID] = task
}

func (s *TaskService) GetTask(taskID string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[taskID]
	if !exists {
		return Task{}, errors.New("task not found")
	}
	return task, nil
}

func (s *TaskService) GetTaskList() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	tasksList := make([]Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasksList = append(tasksList, task)
	}
	return tasksList
}

func (s *TaskService) UpdateTask(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.tasks[task.TaskID]
	if !exists {
		return errors.New("task not found")
	}
	s.tasks[task.TaskID] = task
	return nil
}

func (s *TaskService) DeleteTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.tasks[taskID]
	if !exists {
		return errors.New("task not found")
	}
	delete(s.tasks, taskID)
	return nil
}

func (s *TaskService) CreateFlow(flow Flow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flows[flow.FlowID] = flow
}

func (s *TaskService) GetFlow(flowID string) (Flow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	flow, exists := s.flows[flowID]
	if !exists {
		return Flow{}, errors.New("flow not found")
	}
	return flow, nil
}

func (s *TaskService) GetFlowList() []Flow {
	s.mu.Lock()
	defer s.mu.Unlock()
	flowList := make([]Flow, 0, len(s.flows))
	for _, flow := range s.flows {
		flowList = append(flowList, flow)
	}
	return flowList
}

func (s *TaskService) CreateSchedule(schedule Schedule) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.schedules[schedule.ScheduleID] = schedule
}

func (s *TaskService) GetSchedule(scheduleID string) (Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	schedule, exists := s.schedules[scheduleID]
	if !exists {
		return Schedule{}, errors.New("schedule not found")
	}
	return schedule, nil
}

func (s *TaskService) GetScheduleList() []Schedule {
	s.mu.Lock()
	defer s.mu.Unlock()
	scheduleList := make([]Schedule, 0, len(s.schedules))
	for _, item := range s.schedules {
		scheduleList = append(scheduleList, item)
	}
	return scheduleList
}
