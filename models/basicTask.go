package models

type Task struct {
	TaskID      string `json:"taskId"`
	TaskName    string `json:"taskName"`
	Description string `json:"description"`
	Status      string `json:"status"`   // active, inactive, etc.
	Schedule    string `json:"schedule"` // Cron format
}

type Flow struct {
	FlowID   string `json:"flowId"`
	FlowName string `json:"flowName"`
	Tasks    []Task `json:"tasks"`  // List of tasks in the flow
	Status   string `json:"status"` // active, inactive, etc.
}

type Schedule struct {
	ScheduleID string `json:"scheduleId"`
	FlowID     string `json:"flowId,omitempty"` // Optional, if scheduling a flow
	TaskID     string `json:"taskId,omitempty"` // Optional, if scheduling a task
	Cron       string `json:"cron"`             // Cron expression for scheduling
	Status     string `json:"status"`           // active, inactive, etc.
}
