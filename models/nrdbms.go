package models

// NRDBTableRequest is the request body for NRDBMS table create/delete endpoints.
type NRDBTableRequest struct {
	TargetPoint ProviderConfig `json:"targetPoint"`
	TableName   string         `json:"tableName"`
}

// NRDBTableListResponse is the response body for the NRDBMS table listing endpoint.
type NRDBTableListResponse struct {
	Tables []string `json:"tables"`
}

// NRDBTableGetResponse is the response body for the NRDBMS table get (export) endpoint.
type NRDBTableGetResponse struct {
	Data  []map[string]interface{} `json:"data"`
	Error *string                  `json:"error"`
}

// NRDBInstance is a CSP-agnostic representation of a managed NoSQL database instance.
// GCP: Firestore Database. NCP/Alibaba: MongoDB instance.
type NRDBInstance struct {
	Provider      string `json:"provider"`
	InstanceID    string `json:"instanceId"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Region        string `json:"region"`
	EngineVersion string `json:"engineVersion,omitempty"`
	Endpoint      string `json:"endpoint,omitempty"`
	Port          int32  `json:"port,omitempty"`
}

// NRDBInstanceListRequest is the POST body for listing NRDB instances.
type NRDBInstanceListRequest struct {
	BaseParams
}

// NRDBInstanceCreateRequest is the PUT body for creating an NRDB instance.
type NRDBInstanceCreateRequest struct {
	BaseParams
	InstanceID       string `json:"instanceId"`
	EngineVersion    string `json:"engineVersion"`
	InstanceClass    string `json:"instanceClass"`
	AllocatedStorage int32  `json:"allocatedStorage"`
	MasterUsername   string `json:"masterUsername"`
	MasterPassword   string `json:"masterPassword"`
}

// NRDBInstanceDeleteRequest is the DELETE body for deleting an NRDB instance.
type NRDBInstanceDeleteRequest struct {
	BaseParams
	InstanceID string `json:"instanceId"`
}

// NRDBEngineVersion is a CSP-agnostic available NoSQL engine version.
type NRDBEngineVersion struct {
	EngineVersion string `json:"engineVersion"`
}

// NRDBEngineVersionsRequest is the POST body for listing available NRDB engine versions.
type NRDBEngineVersionsRequest struct {
	BaseParams
}

// NRDBInstanceClassRequest is the POST body for listing orderable NRDB instance classes.
type NRDBInstanceClassRequest struct {
	BaseParams
	EngineVersion string `json:"engineVersion"`
}

// NRDBInstanceRecord is the persisted record of a created NRDB (NoSQL) instance.
type NRDBInstanceRecord struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	Provider        string `gorm:"column:provider;size:20;not null;uniqueIndex:idx_provider_region_instance"`
	Region          string `gorm:"column:region;size:50;not null;uniqueIndex:idx_provider_region_instance"`
	InstanceID      string `gorm:"column:instance_id;size:255;not null;uniqueIndex:idx_provider_region_instance"`
	InstanceName    string `gorm:"column:instance_name;size:255;not null"`
	CspInstanceName string `gorm:"column:csp_instance_name;size:255;not null"`
	NamespaceID     string `gorm:"column:namespace_id;size:100;not null"`
}

func (NRDBInstanceRecord) TableName() string {
	return "tbnrdbinstance"
}
