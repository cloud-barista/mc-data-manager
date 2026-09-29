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

import "time"

// ExecutionLog maps to table `tbexecutionlog`.
// WRITE-ONCE: a row is inserted only when a result is confirmed. In-progress
// transactions live in an in-memory map (not this table), so every row here is
// a terminal result.
//
// Size/count/throughput fields are pointers so that "not measured" (NULL) is
// distinguished from "measured 0" (0). Do NOT change them to value types.
type ExecutionLog struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TraceID     string `gorm:"column:trace_id;size:64" json:"traceId,omitempty"`
	NamespaceID string `gorm:"column:namespace_id;size:100" json:"namespaceId,omitempty"`

	ServiceType string `gorm:"column:service_type;size:20;not null" json:"serviceType"` // objectstorage | rdbms | nrdbms
	Operation   string `gorm:"column:operation;size:20;not null" json:"operation"`      // generate | migrate | backup | restore | delete
	Result      string `gorm:"column:result;size:20;not null" json:"result"`            // success | failed | canceled | timeout

	// Primary endpoint (the resource acted on / destination). Full source->target
	// detail (migrate/backup/restore) is written into Message.
	Provider string `gorm:"column:provider;size:20;not null" json:"provider"`
	Region   string `gorm:"column:region;size:50" json:"region,omitempty"`
	Resource string `gorm:"column:resource;size:255" json:"resource,omitempty"` // bucket / database / collection

	RequestedAt time.Time `gorm:"column:requested_at;type:datetime;not null" json:"requestedAt"`
	FinishedAt  time.Time `gorm:"column:finished_at;type:datetime;not null" json:"finishedAt"`
	DurationMs  int64     `gorm:"column:duration_ms;not null" json:"durationMs"`

	DataSizeBytes *int64   `gorm:"column:data_size_bytes" json:"dataSizeBytes,omitempty"` // not measured = nil
	DataCount     *int64   `gorm:"column:data_count" json:"dataCount,omitempty"`         // not measured = nil
	ThroughputBps *float64 `gorm:"column:throughput_bps" json:"throughputBps,omitempty"` // N/A = nil

	Message string `gorm:"column:message;type:text" json:"message,omitempty"`
}

// TableName maps the struct to the `tbexecutionlog` table.
func (ExecutionLog) TableName() string {
	return "tbexecutionlog"
}

// ExecutionLogListResponse is the paginated response for GET /history.
type ExecutionLogListResponse struct {
	Content    []ExecutionLog `json:"content"`
	Page       int            `json:"page"`
	Size       int            `json:"size"`
	TotalCount int64          `json:"totalCount"`
	TotalPages int            `json:"totalPages"`
}
