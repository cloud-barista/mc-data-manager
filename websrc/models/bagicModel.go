package models

type BasicPageResponse struct {
	Content string  `json:"Content"`
	Error   *string `json:"Error"`
	OS      string  `json:"OS"`
	TmpPath string  `json:"TmpPath"`

	Regions    []string `json:"Regions"`
	AWSRegions []string `json:"AWSRegions"`
	GCPRegions []string `json:"GCPRegions"`
	NCPRegions []string `json:"NCPRegions"`
}

type BasicResponse struct {
	Result string  `json:"Result"`
	Error  *string `json:"Error"`
}
