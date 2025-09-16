package models

type ObjectFilterParams struct {
	Path              string   `json:"path"`
	Contains          []string `json:"contains"`
	Suffixes          []string `json:"suffixes"`
	Exact             []string `json:"exact"`
	Regex             string   `json:"regex"`
	MinSize           *float64 `json:"minSize"`
	MaxSize           *float64 `json:"maxSize"`
	ModifiedAfter     *string  `json:"modifiedAfter"`
	ModifiedBefore    *string  `json:"modifiedBefore"`
	SizeFilteringUnit string   `json:"sizeFilteringUnit"`
}
