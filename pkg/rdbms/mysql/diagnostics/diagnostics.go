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
package diagnostics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Snapshot[T any] struct {
	TakenAt time.Time
	Items   map[string]T // key: "schema.table"
}

type Collector struct {
	Lock   *LockWaitsCollector
	IO     *TableIOCollector
	Buffer *DatabaseBufferCollector
	Work   *WorkloadCollector
	Thread *DatabaseThreadCollector
}

func NewCollector(db *sql.DB) *Collector {
	return &Collector{
		Lock:   NewLockWaitsCollector(db),
		IO:     NewTableIOCollector(db),
		Buffer: NewDatabaseBufferCollector(db),
		Work:   NewWorkloadCollector(db),
		Thread: NewDatabaseThreadCollector(db),
	}
}

type TimedResult struct {
	Lock    []TableLockDelta
	IO      []TableIODelta
	Buffer  DatabaseBufferStat
	Work    []WorkloadDelta
	Thread  DatabaseThreadStat
	Elapsed time.Duration

	// 수집 단계에서 발생한 에러 기록
	Errors map[string]error
}

func (c *Collector) RunTimed(ctx context.Context, d time.Duration) (TimedResult, error) {
	res := TimedResult{Errors: make(map[string]error)}

	// 시작 스냅샷
	lockBefore, err := c.Lock.Snapshot(ctx)
	if err != nil {
		res.Errors["lock_before"] = err
	}
	ioBefore, err := c.IO.Snapshot(ctx)
	if err != nil {
		res.Errors["io_before"] = err
	}
	workBefore, err := c.Work.Snapshot(ctx)
	if err != nil {
		res.Errors["work_before"] = err
	}

	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return TimedResult{}, ctx.Err()
	case <-t.C:
	}

	// 종료 스냅샷
	lockAfter, err := c.Lock.Snapshot(ctx)
	if err != nil {
		res.Errors["lock_after"] = err
	}
	ioAfter, err := c.IO.Snapshot(ctx)
	if err != nil {
		res.Errors["io_after"] = err
	}
	workAfter, err := c.Work.Snapshot(ctx)
	if err != nil {
		res.Errors["work_after"] = err
	}

	var (
		lDelta   []TableLockDelta
		lElapsed time.Duration
		iDelta   []TableIODelta
		iElapsed time.Duration
		wDelta   []WorkloadDelta
		wElapsed time.Duration
	)

	if res.Errors["lock_before"] == nil && res.Errors["lock_after"] == nil {
		lDelta, lElapsed = DiffLock(lockBefore, lockAfter)
	}
	if res.Errors["io_before"] == nil && res.Errors["io_after"] == nil {
		iDelta, iElapsed = DiffIO(ioBefore, ioAfter)
	}
	if res.Errors["work_before"] == nil && res.Errors["work_after"] == nil {
		wDelta, wElapsed = DiffWorkload(workBefore, workAfter)
	}

	// Buffer / Thread는 수집 실패해도 빈값
	if bStat, berr := c.Buffer.Collect(); berr != nil {
		res.Errors["buffer_collect"] = berr
	} else {
		res.Buffer = bStat
	}
	if t, terr := c.Thread.Collect(); terr != nil {
		res.Errors["thread_collect"] = terr
	} else {
		res.Thread = t
	}

	// Elapsed: 사용 가능한 것들 중 최댓값
	res.Elapsed = maxDuration(lElapsed, iElapsed, wElapsed)

	// 결과 채우기
	res.Lock = lDelta
	res.IO = iDelta
	res.Work = wDelta

	return res, nil
}

// 모드 B: 다른 기능 실행 전후로 스냅샷 차분
func (c *Collector) WithDiagnostic(ctx context.Context, fn func(context.Context) error) (TimedResult, error) {
	res := TimedResult{Errors: make(map[string]error)}

	// ---------- BEFORE snapshots ----------
	lockBefore, err := c.Lock.Snapshot(ctx)
	if err != nil {
		res.Errors["lock_before"] = err
	}
	ioBefore, err := c.IO.Snapshot(ctx)
	if err != nil {
		res.Errors["io_before"] = err
	}
	workBefore, err := c.Work.Snapshot(ctx)
	if err != nil {
		res.Errors["work_before"] = err
	}

	// ---------- RUN the operation ----------
	opErr := fn(ctx) // 에러만 WithDiagnostic의 반환

	// ---------- AFTER snapshots ----------
	lockAfter, err := c.Lock.Snapshot(ctx)
	if err != nil {
		res.Errors["lock_after"] = err
	}
	ioAfter, err := c.IO.Snapshot(ctx)
	if err != nil {
		res.Errors["io_after"] = err
	}
	workAfter, err := c.Work.Snapshot(ctx)
	if err != nil {
		res.Errors["work_after"] = err
	}

	var (
		lDelta   []TableLockDelta
		lElapsed time.Duration
		iDelta   []TableIODelta
		iElapsed time.Duration
		wDelta   []WorkloadDelta
		wElapsed time.Duration
	)

	if res.Errors["lock_before"] == nil && res.Errors["lock_after"] == nil {
		lDelta, lElapsed = DiffLock(lockBefore, lockAfter)
	}
	if res.Errors["io_before"] == nil && res.Errors["io_after"] == nil {
		iDelta, iElapsed = DiffIO(ioBefore, ioAfter)
	}
	if res.Errors["work_before"] == nil && res.Errors["work_after"] == nil {
		wDelta, wElapsed = DiffWorkload(workBefore, workAfter)
	}

	// Buffer / Thread는 수집 실패해도 빈값
	if bStat, berr := c.Buffer.Collect(); berr != nil {
		res.Errors["buffer_collect"] = berr
	} else {
		res.Buffer = bStat
	}
	if t, terr := c.Thread.Collect(); terr != nil {
		res.Errors["thread_collect"] = terr
	} else {
		res.Thread = t
	}

	// ---------- Elapsed: 사용 가능한 것들 중 최댓값 ----------
	res.Elapsed = maxDuration(lElapsed, iElapsed, wElapsed)

	// 결과 채우기
	res.Lock = lDelta
	res.IO = iDelta
	res.Work = wDelta

	// 반환 error는 오직 fn(ctx)에서 난 것만
	return res, opErr
}

/************** 4) 간단 리포트 출력(옵셔널) **************/

func PrintBufferReport(d DatabaseBufferStat) {
	fmt.Printf("[Buffer hit ratio]\n")
	fmt.Printf("%-35s %-10f\n", "buffer pool hit ratio pecentage", d.BufferPoolHitRatio)
}

func PrintLockReport(d []TableLockDelta, elapsed time.Duration) {
	fmt.Printf("[Lock waits] interval=%s\n", elapsed)
	fmt.Printf("%-30s %-10s %10s %18s %16s\n", "table", "schema", "waits", "wait_time", "waits/s")
	for _, x := range d {
		fmt.Printf("%-30s %-10s %10d %18s %16.2f\n",
			x.Table, x.Schema, x.LockWaitCount, x.TotalWait(), x.WaitsPerSec())
	}
}

func PrintIOReport(d []TableIODelta, elapsed time.Duration) {
	fmt.Printf("[Table IO] interval=%s\n", elapsed)
	fmt.Printf("%-30s %-10s %12s %12s %12s %12s\n", "table", "schema", "read", "insert", "update", "delete")
	for _, x := range d {
		fmt.Printf("%-30s %-10s %12d %12d %12d %12d\n",
			x.Table, x.Schema, x.RowsRead, x.RowsInserted, x.RowsUpdated, x.RowsDeleted)
	}
}

func PrintWorkloadReport(w []WorkloadDelta, elapsed time.Duration) {
	fmt.Printf("[Workload] interval=%s\n", elapsed)
	for _, a := range w {
		fmt.Printf("  schema  : %s\n", a.Schema)
		fmt.Printf("  queries : %d (QPS: %.2f)\n", a.TotalQueries, float64(a.TotalQueries)/a.Elapsed.Seconds())
		fmt.Printf("  latency : total=%s avg/stmt=%s",
			(time.Duration(a.TotalLatencyPS/1000) * time.Nanosecond).String(),
			a.AvgLatency().String(),
		)
		if a.MaxLatencyPS > 0 {
			fmt.Printf(" max=%s (new max observed)\n", a.MaxLatency().String())
		} else {
			fmt.Printf(" max=? (no new record; exact max in interval unknown)\n")
		}
	}
}

func PrintThreadReport(d DatabaseThreadStat) {
	fmt.Printf("[Thread count]\n")
	fmt.Printf("%-10s %-10s\n", "threads_connected", "threads_running")
	fmt.Printf("%-17d %-10d\n", d.threadConnected, d.threadRunning)
}

func key(schema, table string) string {
	return fmt.Sprintf("%s.%s", schema, table)
}

func maxDuration(ds ...time.Duration) time.Duration {
	var m time.Duration
	for _, d := range ds {
		if d > m {
			m = d
		}
	}
	return m
}
