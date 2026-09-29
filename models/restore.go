package models

type RestoreSourcePoint struct {
	BackupId string `json:"backupId"`
}

type RestoreObjectStorageTargetPoint struct {
	Provider string `json:"provider"`
	Region   string `json:"region"`
	ObjectStorageParams
}

type RestoreObjectStorageRequest struct {
	SourcePoint RestoreSourcePoint              `json:"sourcePoint"`
	TargetPoint RestoreObjectStorageTargetPoint `json:"targetPoint"`
}

type RestoreRDBTargetPoint struct {
	Provider string `json:"provider"`
	Region   string `json:"region"`
	MySQLParams
}

type RestoreRDBRequest struct {
	SourcePoint RestoreSourcePoint    `json:"sourcePoint"`
	TargetPoint RestoreRDBTargetPoint `json:"targetPoint"`
}

type RestoreNRDBTargetPoint struct {
	Provider string `json:"provider"`
	Region   string `json:"region"`
	MySQLParams
	DatabaseID string `json:"databaseId,omitempty"`
}

type RestoreNRDBRequest struct {
	SourcePoint RestoreSourcePoint     `json:"sourcePoint"`
	TargetPoint RestoreNRDBTargetPoint `json:"targetPoint"`
}
