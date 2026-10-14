package models

import "time"

type BackupRecord struct {
	ID           string    `gorm:"primaryKey;size:36"`
	NamespaceID  string    `gorm:"column:namespace_id;size:100;not null;uniqueIndex:idx_ns_type_name"`
	ServiceType  string    `gorm:"column:service_type;size:20;not null;uniqueIndex:idx_ns_type_name"`
	BackupName   string    `gorm:"column:backup_name;size:255;not null;uniqueIndex:idx_ns_type_name"`
	Provider     string    `gorm:"column:provider;size:20;not null"`
	Region       string    `gorm:"column:region;size:50;not null"`
	InstanceId   string    `gorm:"column:instance_id;size:255;not null"`
	InstanceName string    `gorm:"column:instance_name;size:255;not null"`
	DatabaseName *string   `gorm:"column:database_name;size:255"` // nullable, only for rdbms/nrdbms
	Path         string    `gorm:"column:path;size:500;not null"`
	Status       string    `gorm:"column:status;size:20;not null"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (BackupRecord) TableName() string {
	return "tbbackup"
}

type BackupListResponse struct {
	ID           string    `json:"id"`
	BackupName   string    `json:"backupName"`
	Provider     string    `json:"provider"`
	Region       string    `json:"region"`
	InstanceId   string    `json:"instanceId"`
	InstanceName string    `json:"instanceName"`
	DatabaseName *string   `json:"databaseName"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type BackupSourceCommon struct {
	Provider     string `json:"provider"`
	Region       string `json:"region"`
	InstanceId   string `json:"instanceId"`
	InstanceName string `json:"instanceName"`
}

type BackupTargetPoint struct {
	BackupName string `json:"backupName"`
}

type BackupObjectStorageSourcePoint struct {
	BackupSourceCommon
	ObjectStorageParams
}

type BackupObjectStorageRequest struct {
	SourcePoint  BackupObjectStorageSourcePoint `json:"sourcePoint"`
	TargetPoint  BackupTargetPoint              `json:"targetPoint"`
	SourceFilter *ObjectFilterParams            `json:"sourceFilter,omitempty"`
}

type BackupRDBSourcePoint struct {
	BackupSourceCommon
	MySQLParams
}

type BackupRDBRequest struct {
	SourcePoint BackupRDBSourcePoint `json:"sourcePoint"`
	TargetPoint BackupTargetPoint    `json:"targetPoint"`
}

type BackupNRDBSourcePoint struct {
	BackupSourceCommon
	MySQLParams
	DatabaseID string `json:"databaseId,omitempty"` // gcp (Firestore) only
}

type BackupNRDBRequest struct {
	SourcePoint BackupNRDBSourcePoint `json:"sourcePoint"`
	TargetPoint BackupTargetPoint     `json:"targetPoint"`
}
