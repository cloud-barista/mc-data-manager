package models

// 전체 루트 구조
type ConnectionConfigList struct {
	ConnectionConfig []ConnectionConfig `json:"connectionconfig"`
}

// 각 ConnectionConfig 항목
type ConnectionConfig struct {
	ConfigName           string         `json:"configName"`
	ProviderName         string         `json:"providerName"`
	DriverName           string         `json:"driverName"`
	CredentialName       string         `json:"credentialName"`
	CredentialHolder     string         `json:"credentialHolder"`
	RegionZoneInfoName   string         `json:"regionZoneInfoName"`
	RegionZoneInfo       RegionZoneInfo `json:"regionZoneInfo"`
	RegionDetail         RegionDetail   `json:"regionDetail"`
	RegionRepresentative bool           `json:"regionRepresentative"`
	Verified             bool           `json:"verified"`
}

// regionZoneInfo 구조
type RegionZoneInfo struct {
	AssignedRegion string `json:"assignedRegion"`
	AssignedZone   string `json:"assignedZone"`
}

// regionDetail 구조
type RegionDetail struct {
	RegionID    string   `json:"regionId"`
	RegionName  string   `json:"regionName"`
	Description string   `json:"description"`
	Location    Location `json:"location"`
	Zones       []string `json:"zones"`
}

// location 구조
type Location struct {
	Display   string  `json:"display"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
