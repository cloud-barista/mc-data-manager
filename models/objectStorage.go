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

// ObjectStorageListResponse is the response body for GET /ns/{nsId}/resources/objectStorage
type ObjectStorageListResponse struct {
	ObjectStorage []ObjectStorage `json:"objectStorage"`
}

// ObjectStorage represents a single object storage resource returned by CB-Tumblebug
type ObjectStorage struct {
	ResourceType     string           `json:"resourceType"`
	ID               string           `json:"id"`
	UID              string           `json:"uid"`
	CspResourceName  string           `json:"cspResourceName"`
	CspResourceId    string           `json:"cspResourceId"`
	ConnectionName   string           `json:"connectionName"`
	ConnectionConfig ConnectionConfig `json:"connectionConfig"`
	Description      string           `json:"description"`
	Status           string           `json:"status"`
	SystemMessage    string           `json:"systemMessage"`
	Conditions       []Condition      `json:"conditions"`
	Name             string           `json:"name"`
	CreationDate     string           `json:"creationDate"`
	MaxKeys          int              `json:"maxKeys"`
	IsTruncated      bool             `json:"isTruncated"`
	Marker           string           `json:"marker"`
	Prefix           string           `json:"prefix"`
	Contents         []Content        `json:"contents"`
}

// Condition represents a status condition of an object storage resource
type Condition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	Message            string `json:"message"`
	Reason             string `json:"reason"`
	LastTransitionTime string `json:"lastTransitionTime"`
}

// Content represents a single object within an object storage bucket
type Content struct {
	ETag         string    `json:"eTag"`
	Key          string    `json:"key"`
	LastModified time.Time `json:"lastModified"`
	Size         int64     `json:"size"`
	StorageClass string    `json:"storageClass"`
}

type Bucket struct {
	Name         string `json:"Name"`
	CreationDate string `json:"CreationDate"`
}

type DeleteRequest struct {
	XMLName xml.Name   `xml:"Delete"`
	XMLNS   string     `xml:"xmlns,attr"`
	Objects []S3Object `xml:"Object"`
}

type S3Object struct {
	Key string `xml:"Key"`
}
