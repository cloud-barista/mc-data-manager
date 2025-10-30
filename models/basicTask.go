package models

import "time"

type OperationParams struct {
	OperationId string `json:"operationId" form:"operationId"`
}

type SysbenchParams struct {
	TargetType   string `json:"targetType"`
	TableCount   int64  `json:"tableCount"`
	TableSize    int64  `json:"tableSize"`
	ThreadsCount int64  `json:"threadsCount"`
	MySQLParams
}

type StatusParams struct {
	TargetPoint ProviderConfig `json:"targetPoint,omitempty"`
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
	ScheduleID   string          `json:"ScheduleID,omitempty"`
	ScheduleName string          `json:"ScheduleName"`
	Tasks        []BasicDataTask `json:"tasks"`
	Cron         string          `json:"cron,omitempty"`
	StartTime    *time.Time      `json:"startTime,omitempty"`
	TimeZone     string          `json:"tz,omitempty"`

	Status Status `json:"status"`
}

type Schedule struct {
	OperationParams
	TagParams
	BasicSchedule
}

type GenarateTask struct {
	Task
	Dummy       GenFileParams  `json:"dummy"`
	TargetPoint ProviderConfig `json:"targetPoint"`
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
	Directory    string              `json:"Directory,omitempty" swaggerignore:"true"`
	Dummy        GenFileParams       `json:"dummy"`
	SourcePoint  ProviderConfig      `json:"sourcePoint,omitempty"`
	TargetPoint  ProviderConfig      `json:"targetPoint,omitempty"`
	SourceFilter *ObjectFilterParams `json:"sourceFilter,omitempty"`
}
type DiagnosticTask struct {
	SysbenchParams
	StatusParams
	Time int64 `json:"time"`
}
type DataTask struct {
	OperationParams
	BasicDataTask
}
type MigrateTask struct {
	BasicTask
	Directory    string              `json:"Directory,omitempty" swaggerignore:"true"`
	SourcePoint  ProviderConfig      `json:"sourcePoint,omitempty"`
	TargetPoint  ProviderConfig      `json:"targetPoint,omitempty"`
	SourceFilter *ObjectFilterParams `json:"sourceFilter,omitempty"`
}

type BasicBackupTask struct {
	BasicTask
	SourcePoint ProviderConfig `json:"sourcePoint,omitempty"`
	TargetPoint ProviderConfig `json:"targetPoint,omitempty"`
}
type BackupTask struct {
	BasicTask
	Directory    string              `json:"Directory,omitempty" swaggerignore:"true"`
	SourcePoint  ProviderConfig      `json:"sourcePoint,omitempty"`
	TargetPoint  ProviderConfig      `json:"targetPoint,omitempty"`
	SourceFilter *ObjectFilterParams `json:"sourceFilter,omitempty"`
}

type RestoreTask struct {
	DataTask
}
