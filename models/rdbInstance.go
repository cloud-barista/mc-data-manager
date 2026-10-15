package models

type RDBInstanceListRequest struct {
	BaseParams
}

type RDBInstanceCreateRequest struct {
	BaseParams
	InstanceID       string `json:"instanceId"`
	InstanceClass    string `json:"instanceClass"`
	Engine           string `json:"engine"`
	EngineVersion    string `json:"engineVersion"`
	MasterUsername   string `json:"masterUsername"`
	MasterPassword   string `json:"masterPassword"`
	AllocatedStorage int32  `json:"allocatedStorage"`
}

type RDBInstanceDeleteRequest struct {
	BaseParams
	InstanceID string `json:"instanceId"`
}

type RDBEngineVersionsRequest struct {
	BaseParams
}

type RDBInstanceClassRequest struct {
	BaseParams
	Engine        string `json:"engine"`
	EngineVersion string `json:"engineVersion"`
}

type RDBAccountCreateRequest struct {
	BaseParams
	InstanceID     string `json:"instanceId"`
	MasterUsername string `json:"masterUsername"`
	MasterPassword string `json:"masterPassword"`
}

// DBEngineVersion is a CSP-agnostic available DB engine version.
type DBEngineVersion struct {
	Engine        string `json:"engine"`
	EngineVersion string `json:"engineVersion"`
}

// DBInstance is a CSP-agnostic representation of a managed database instance.
// AWS RDS is the first provider mapped onto this shape; other CSPs reuse it.
type DBInstance struct {
	Provider      string `json:"provider"`
	InstanceID    string `json:"instanceId"`
	Name          string `json:"name"`
	Engine        string `json:"engine"`
	EngineVersion string `json:"engineVersion"`
	Status        string `json:"status"`
	Endpoint      string `json:"endpoint"`
	Port          int32  `json:"port"`
	Region        string `json:"region"`
	InstanceClass string `json:"instanceClass"`
}

// RDBInstanceRecord is the persisted record of a created RDB (database) instance.
type RDBInstanceRecord struct {
	ID                  uint64 `gorm:"primaryKey;autoIncrement"`
	Provider            string `gorm:"column:provider;size:20;not null;uniqueIndex:idx_provider_region_instance"`
	Region              string `gorm:"column:region;size:50;not null;uniqueIndex:idx_provider_region_instance"`
	InstanceID          string `gorm:"column:instance_id;size:255;not null;uniqueIndex:idx_provider_region_instance"`
	InstanceName        string `gorm:"column:instance_name;size:255;not null"`
	CspInstanceName     string `gorm:"column:csp_instance_name;size:255;not null"`
	PublicNetPending    bool   `gorm:"column:public_net_pending;not null;default:false"`    // only for alibaba
	AccountCreateFailed bool   `gorm:"column:account_create_failed;not null;default:false"` // only for alibaba
	NamespaceID         string `gorm:"column:namespace_id;size:100;not null"`
}

func (RDBInstanceRecord) TableName() string {
	return "tbrdbinstance"
}
