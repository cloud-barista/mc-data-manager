package models

type OperationParams struct {
	OperationId string `json:"operationId" form:"operationId"`
}
type TagParams struct {
	Tag []string `json:"tag,omitempty"`
}
type TaskMeta struct {
	ServiceType CloudServiceType `json:"serviceType"`
	TaskType    TaskType         `json:"taskType" `
	TaskID      string           `json:"taskId,omitempty" `
	TaskName    string           `json:"taskName,omitempty" `
	Description string           `json:"description,omitempty"`
}

type BasicTask struct {
	TaskMeta `json:"meta,omitempty" swaggerignore:"true"`
	Status   `json:"status,omitempty" swaggerignore:"true"`
}

type Task struct {
	OperationParams
	TagParams
	BasicTask
}

type BasicFlow struct {
	FlowID   string     `json:"FlowID,omitempty"`
	FlowName string     `json:"FlowName"`
	Tasks    []DataTask `json:"tasks"`
	Status   Status     `json:"status"`
}

type Flow struct {
	OperationParams
	BasicFlow
}

type BasicSchedule struct {
	ScheduleID   string     `json:"ScheduleID,omitempty"`
	ScheduleName string     `json:"ScheduleName"`
	Tasks        []DataTask `json:"tasks"`
	Cron         string     `json:"cron"`
	TimeZone     string     `json:"tz"`

	Status Status `json:"status"`
}

type Schedule struct {
	OperationParams
	TagParams
	BasicSchedule
}

type GenarateTask struct {
	Task        `json:"inline"`
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
type BasicDataTask struct {
	BasicTask
	Directory   string         `json:"Directory,omitempty" swaggerignore:"true"`
	SourcePoint ProviderConfig `json:"sourcePoint,omitempty"`
	TargetPoint ProviderConfig `json:"targetPoint,omitempty"`
}
type DataTask struct {
	OperationParams
	BasicDataTask
}
type MigrateTask struct {
	DataTask
}

type BasicBackupTask struct {
	BasicTask
	SourcePoint ProviderConfig `json:"sourcePoint,omitempty"`
	TargetPoint ProviderConfig `json:"targetPoint,omitempty"`
}
type BackupTask struct {
	DataTask
}

type RestoreTask struct {
	DataTask
}
