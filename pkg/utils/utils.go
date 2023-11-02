package utils

import (
	"time"
)

type Object struct {
	ChecksumAlgorithm []string
	ETag              string
	Key               string
	LastModified      time.Time
	Size              int64
	StorageClass      string
}

type Provider string

const (
	AWS Provider = "aws"
	GCP Provider = "gcp"
	NCP Provider = "ncp"
)
