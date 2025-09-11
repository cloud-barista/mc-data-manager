package models

type ObjectFilterParams struct {
	Prefix         string   `json:"prefix"`
    Contains       []string `json:"contains"`
    Suffixes       []string `json:"suffixes"`
    Exact          []string `json:"exact"`
    Regex          string   `json:"regex"`
    MinSize        *int64   `json:"minSize"`
    MaxSize        *int64   `json:"maxSize"`
    ModifiedAfter  *string  `json:"modifiedAfter"`
    ModifiedBefore *string  `json:"modifiedBefore"`
}