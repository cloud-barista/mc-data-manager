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
package task

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/rs/zerolog/log"
)

// ExecMeta is the metadata captured for a single execution (transaction).
type ExecMeta struct {
	TraceID     string
	NamespaceID string
	ServiceType string
	Operation   string
	Provider    string
	Region      string
	Resource    string
	Message     string
}

// execInProgress is an in-memory entry for a running transaction.
type execInProgress struct {
	meta        ExecMeta
	requestedAt time.Time
}

// ExecTracker keeps in-progress transactions in an in-memory map and writes a
// WRITE-ONCE row to tbexecutionlog when a result is confirmed.
type ExecTracker struct {
	mu      sync.Mutex
	running map[uint64]*execInProgress
	seq     uint64
	// lastErr maps traceId -> last failure cause, consumed once by the HTTP
	// layer so the real cause can be surfaced in the response.
	lastErr sync.Map
}

var (
	execTracker     *ExecTracker
	execTrackerOnce sync.Once
)

// GetExecTracker returns the singleton tracker.
func GetExecTracker() *ExecTracker {
	execTrackerOnce.Do(func() {
		execTracker = &ExecTracker{running: make(map[uint64]*execInProgress)}
	})
	return execTracker
}

// track registers the transaction in the map, runs fn, and writes a write-once
// log row when the result is confirmed (success / failed / panic).
func (t *ExecTracker) track(meta ExecMeta, fn func() (models.Status, error)) (status models.Status) {
	key := atomic.AddUint64(&t.seq, 1)
	reqAt := time.Now()

	t.mu.Lock()
	t.running[key] = &execInProgress{meta: meta, requestedAt: reqAt}
	t.mu.Unlock()

	var taskErr error
	defer func() {
		result := "success"
		if r := recover(); r != nil {
			result = "failed"
			cause := fmt.Sprintf("panic: %v", r)
			meta.Message = appendMsg(meta.Message, cause)
			t.storeErr(meta.TraceID, cause)
			status = models.StatusFailed // swallow panic -> record as failed
		} else if status == models.StatusFailed {
			result = "failed"
			if taskErr != nil {
				meta.Message = appendMsg(meta.Message, taskErr.Error())
				t.storeErr(meta.TraceID, taskErr.Error())
			}
		}
		t.finish(key, meta, reqAt, result)
	}()

	status, taskErr = fn()
	return status
}

// Track records a WRITE-ONCE row for a DIRECT operation that bypasses handleTask
// (e.g., bucket delete). fn returns error: nil = success, non-nil = failed.
func (t *ExecTracker) Track(meta ExecMeta, fn func() error) (err error) {
	key := atomic.AddUint64(&t.seq, 1)
	reqAt := time.Now()

	t.mu.Lock()
	t.running[key] = &execInProgress{meta: meta, requestedAt: reqAt}
	t.mu.Unlock()

	defer func() {
		result := "success"
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			meta.Message = appendMsg(meta.Message, err.Error())
			t.storeErr(meta.TraceID, err.Error())
			result = "failed"
		} else if err != nil {
			meta.Message = appendMsg(meta.Message, err.Error())
			t.storeErr(meta.TraceID, err.Error())
			result = "failed"
		}
		t.finish(key, meta, reqAt, result)
	}()

	err = fn()
	return err
}

// finish removes the entry from the map and inserts one row into tbexecutionlog.
func (t *ExecTracker) finish(key uint64, meta ExecMeta, reqAt time.Time, result string) {
	t.mu.Lock()
	delete(t.running, key)
	t.mu.Unlock()

	if config.DB == nil {
		log.Warn().Msg("[execlog] config.DB is nil; skip tbexecutionlog insert")
		return
	}
	now := time.Now()
	row := &models.ExecutionLog{
		TraceID:     meta.TraceID,
		NamespaceID: meta.NamespaceID,
		ServiceType: meta.ServiceType,
		Operation:   meta.Operation,
		Result:      result,
		Provider:    meta.Provider,
		Region:      meta.Region,
		Resource:    meta.Resource,
		RequestedAt: reqAt,
		FinishedAt:  now,
		DurationMs:  now.Sub(reqAt).Milliseconds(),
		Message:     meta.Message,
	}
	if err := config.DB.Create(row).Error; err != nil {
		log.Error().Err(err).Msg("[execlog] failed to insert tbexecutionlog")
		return
	}
	log.Info().
		Str("traceId", meta.TraceID).
		Str("service", meta.ServiceType).
		Str("operation", meta.Operation).
		Str("result", result).
		Int64("durationMs", row.DurationMs).
		Msg("[execlog] recorded")
}

// storeErr records the failure cause for a traceId so the HTTP layer can surface it.
func (t *ExecTracker) storeErr(traceID, cause string) {
	if traceID != "" && cause != "" {
		t.lastErr.Store(traceID, cause)
	}
}

// LastError returns and consumes the last recorded failure message for a traceId.
// Returns "" if none. Used by HTTP handlers to surface the real cause.
func (t *ExecTracker) LastError(traceID string) string {
	if traceID == "" {
		return ""
	}
	if v, ok := t.lastErr.LoadAndDelete(traceID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// ListRunning returns a snapshot of in-progress transactions (for future 진행중 조회).
func (t *ExecTracker) ListRunning() []ExecMeta {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]ExecMeta, 0, len(t.running))
	for _, ip := range t.running {
		out = append(out, ip.meta)
	}
	return out
}

// metaFromTask extracts ExecMeta from a task's params (best-effort; refine later).
func metaFromTask(st models.CloudServiceType, tt models.TaskType, p models.BasicDataTask) ExecMeta {
	tgt := p.TargetPoint
	src := p.SourcePoint

	// primary endpoint = target, or source when target has no provider (backup).
	primary := tgt
	if primary.Provider == "" {
		primary = src
	}

	m := ExecMeta{
		NamespaceID: utils.GetNsId(),
		ServiceType: string(st),
		Operation:   string(tt),
		Provider:    primary.Provider,
		Region:      primary.Region,
		Resource:    resourceOf(st, primary),
	}
	// migrate: both endpoints are cloud -> record the flow in message
	if src.Provider != "" && tgt.Provider != "" {
		m.Message = fmt.Sprintf("source: %s/%s/%s -> target: %s/%s/%s",
			src.Provider, src.Region, resourceOf(st, src),
			tgt.Provider, tgt.Region, resourceOf(st, tgt))
	}
	return m
}

// resourceOf picks the resource name per service type: bucket / database / collection.
func resourceOf(st models.CloudServiceType, pc models.ProviderConfig) string {
	switch st {
	case models.ObejectStorage:
		return pc.Bucket
	case models.RDBMS:
		return pc.DatabaseName
	case models.NRDBMS:
		if pc.DatabaseID != "" {
			return pc.DatabaseID
		}
		return pc.DatabaseName
	default:
		return ""
	}
}

func appendMsg(base, add string) string {
	if base == "" {
		return add
	}
	return base + " | " + add
}
