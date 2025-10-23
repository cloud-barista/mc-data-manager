package models

import (
	"encoding/xml"
	"time"
)

type ObjectFilterParams struct {
	Path              string   `json:"path"`
	PathExcludeYn     string   `json:"pathExcludeYn"`
	Contains          []string `json:"contains"`
	ContainExcludeYn  string   `json:"containExcludeYn"`
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
	Owner   Owner `json:"Owner"`
	Buckets struct {
		Bucket []Bucket `json:"Bucket"`
	} `json:"Buckets"`
}

// 변환 후 구조
type SimpleBuckets struct {
	Buckets []Bucket `json:"Buckets"`
}

type Owner struct {
	ID          string `json:"ID"`
	DisplayName string `json:"DisplayName"`
}

type Bucket struct {
	Name         string `json:"Name"`
	CreationDate string `json:"CreationDate"`
}

type ListBucketResult struct {
	Name        string     `json:"Name"`
	Prefix      string     `json:"Prefix"`
	Marker      string     `json:"Marker"`
	MaxKeys     int        `json:"MaxKeys"`
	IsTruncated bool       `json:"IsTruncated"`
	Contents    []Contents `json:"Contents"`
}

type Contents struct {
	Key          string    `json:"Key"`
	LastModified time.Time `json:"LastModified"`
	ETag         string    `json:"ETag"`
	Size         int64     `json:"Size"`
	StorageClass string    `json:"StorageClass"`
}

type DeleteRequest struct {
	XMLName xml.Name   `xml:"Delete"`
	XMLNS   string     `xml:"xmlns,attr"`
	Objects []S3Object `xml:"Object"`
}

type S3Object struct {
	Key string `xml:"Key"`
}
