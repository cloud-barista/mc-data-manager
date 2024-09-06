package models

type OperationParams struct {
	OperationId string `json:"operationId" form:"operationId"`
}

type TaskMeta struct {
	ServiceType CloudServiceType `json:"serviceType"`
	TaskType    TaskType         `json:"taskType" `
	TaskID      string           `json:"taskId,omitempty" `
	TaskName    string           `json:"taskName,omitempty" `
	Description string           `json:"description,omitempty"`
}

type Task struct {
	OperationParams
	TaskMeta `json:"meta,omitempty" swaggerignore:"true"`
	Status   `json:"status,omitempty" swaggerignore:"true"` // active, inactive, etc.
}

type Flow struct {
	OperationParams
	FlowID   string        `json:"flowId,omitempty"`
	FlowName string        `json:"flowName"`
	Tasks    []interface{} `json:"tasks"`  // List of tasks in the flow
	Status   Status        `json:"status"` // active, inactive, etc.
}

type Schedule struct {
	OperationParams
	ScheduleID string `json:"scheduleId,omitempty"`
	FlowID     string `json:"flowId,omitempty"` // Optional, if scheduling a flow
	TaskID     string `json:"taskId,omitempty"` // Optional, if scheduling a task
	Cron       string `json:"cron"`             // Cron expression for scheduling
	Status     Status `json:"status"`           // active, inactive, etc.
}

type GenarateTask struct {
	Task
	TargetPoint GenTaskTarget `json:"targetPoint"`
}

type CommandTask struct {
	Task
	TaskFilePath string
	GenFileParams
	SourcePoint     ProviderConfig `json:"sourcePoint,omitempty"`
	TargetPoint     ProviderConfig `json:"targetPoint,omitempty"`
	DeleteDBList    []string
	DeleteTableList []string
}
type GenTaskTarget struct {
	ProviderConfig
	GenFileParams
}

type DataTask struct {
	Task
	SourcePoint ProviderConfig `json:"sourcePoint,omitempty"`
	TargetPoint ProviderConfig `json:"targetPoint,omitempty"`
}
type MigrateTask struct {
	DataTask
}

type BackupTask struct {
	DataTask
}

type RestoreTask struct {
	DataTask
}
