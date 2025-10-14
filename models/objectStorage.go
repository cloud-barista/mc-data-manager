package models

import "encoding/xml"

type ObjectFilterParams struct {
	Path              string   `json:"path"`
	Contains          []string `json:"contains"`
	Suffixes          []string `json:"suffixes"`
	Exact             []string `json:"exact"`
	MinSize           *float64 `json:"minSize"`
	MaxSize           *float64 `json:"maxSize"`
	ModifiedAfter     *string  `json:"modifiedAfter"`
	ModifiedBefore    *string  `json:"modifiedBefore"`
	SizeFilteringUnit string   `json:"sizeFilteringUnit"`
}

type BucketListResponse struct {
	Buckets []Bucket `json:"buckets"`
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name `xml:"ListAllMyBucketsResult"`
	XMLNS   string   `xml:"xmlns,attr"`
	Owner   Owner    `xml:"Owner"`
	Buckets struct {
		Bucket []Bucket `xml:"Bucket"`
	} `xml:"Buckets"`
}

// 변환 후 구조
type SimpleBuckets struct {
	Buckets []Bucket `json:"Buckets"`
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type Bucket struct {
	Name         string `xml:"Name" json:"name"`
	CreationDate string `xml:"CreationDate" json:"creationDate"`
}
