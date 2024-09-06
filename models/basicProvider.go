package models

type ProviderConfig struct {
	// common
	BaseParams
	// linux,win
	LinuxMigrationParams
	// osc
	ObjectStorageParams
	// RDB
	MySQLParams
	// NRDB
	NoSQLParams
}
