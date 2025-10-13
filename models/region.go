package models

import "encoding/json"

type Region struct {
	RegionID    string          `json:"regionId"`
	RegionName  string          `json:"regionName"`
	Description string          `json:"description"`
	Location    json.RawMessage `json:"location"`
	Zones       []string        `json:"zones"`
}

type Regions struct {
	Regions []Region `json:"regions"`
}
