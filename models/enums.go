/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package models

type Provider string

const (
	AWS Provider = "aws"
	GCP Provider = "gcp"
	NCP Provider = "ncp"
	OPM Provider = "on-premise"
)

// Service type
type CloudServiceType string

const (
	ComputeService CloudServiceType = "compute"
	ObejectStorage CloudServiceType = "objectStorage"
	RDBMS          CloudServiceType = "rdbms"
	NRDBMS         CloudServiceType = "nrdbms"
)

// Status type
type Status string

const (
	StatusActive    Status = "active"
	StatusInactive  Status = "inactive"
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Task type
type TaskType string

const (
	Generate TaskType = "generate"
	Migrate  TaskType = "migrate"
	Backup   TaskType = "backup"
	Restore  TaskType = "restore"
)
