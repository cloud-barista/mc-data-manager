package models

type Provider string

const (
	AWS Provider = "aws"
	GCP Provider = "gcp"
	NCP Provider = "ncp"
	OPM Provider = "on-premise"
)
