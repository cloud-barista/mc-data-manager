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

// 모드 A: 지정 시간 대기 후 두 스냅샷 차분
type TimedResult struct {
	Lock   []TableLockDelta
	IO     []TableIODelta
	Buffer DatabaseBufferStat
	Work   WorkloadDelta
	Thread DatabaseThreadStat
	// 공통 경과시간: 두 수집기의 elapsed 가 다를 일은 거의 없음
	Elapsed time.Duration
}

func (c *Collector) RunTimed(ctx context.Context, schema string, d time.Duration) (TimedResult, error) {
	// 시작 스냅샷
	lockBefore, err := c.Lock.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	ioBefore, err := c.IO.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	workBefore, err := c.Work.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}

	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return TimedResult{}, ctx.Err()
	case <-t.C:
	}

	// 종료 스냅샷
	lockAfter, err := c.Lock.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	ioAfter, err := c.IO.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	workAfter, err := c.Work.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}

	lDelta, lElapsed := DiffLock(lockBefore, lockAfter)
	iDelta, iElapsed := DiffIO(ioBefore, ioAfter)
	bStat := c.Buffer.Collect()
	workDelta, wElapsed := DiffWorkload(workBefore, workAfter)
	tStat := c.Thread.Collect()

	// 두 elapsed 중 큰 값 사용(거의 동일)
	elapsed := lElapsed
	if iElapsed > elapsed {
		elapsed = iElapsed
	}
	if wElapsed > elapsed {
		elapsed = wElapsed
	}

	return TimedResult{
		Lock:    lDelta,
		IO:      iDelta,
		Buffer:  bStat,
		Work:    workDelta,
		Thread:  tStat,
		Elapsed: elapsed,
	}, nil
}

// 모드 B: 다른 기능 실행 전후로 스냅샷 차분
func (c *Collector) WithDiagnostic(ctx context.Context, schema string, fn func(context.Context) error) (TimedResult, error) {
	lockBefore, err := c.Lock.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	ioBefore, err := c.IO.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	workBefore, err := c.Work.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}

	if err := fn(ctx); err != nil {
		return TimedResult{}, err
	}

	lockAfter, err := c.Lock.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	ioAfter, err := c.IO.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}
	workAfter, err := c.Work.Snapshot(ctx, schema)
	if err != nil {
		return TimedResult{}, err
	}

	lDelta, lElapsed := DiffLock(lockBefore, lockAfter)
	iDelta, iElapsed := DiffIO(ioBefore, ioAfter)
	bStat := c.Buffer.Collect()
	workDelta, wElapsed := DiffWorkload(workBefore, workAfter)
	tStat := c.Thread.Collect()

	elapsed := lElapsed
	if iElapsed > elapsed {
		elapsed = iElapsed
	}
	if wElapsed > elapsed {
		elapsed = wElapsed
	}
	return TimedResult{
		Lock:    lDelta,
		IO:      iDelta,
		Buffer:  bStat,
		Work:    workDelta,
		Thread:  tStat,
		Elapsed: elapsed,
	}, nil
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

func PrintWorkloadReport(w WorkloadDelta) {
	fmt.Printf("[Workload] schema=%s interval=%s\n", w.Schema, w.Elapsed)
	fmt.Printf("  queries: %d (QPS: %.2f)\n", w.TotalQueries, float64(w.TotalQueries)/w.Elapsed.Seconds())
	fmt.Printf("  latency: total=%s avg/stmt=%s",
		(time.Duration(w.TotalLatencyPS/1000) * time.Nanosecond).String(),
		w.AvgLatency().String(),
	)
	if w.MaxLatencyPS > 0 {
		fmt.Printf(" max=%s (new max observed)\n", w.MaxLatency().String())
	} else {
		fmt.Printf(" max=? (no new record; exact max in interval unknown)\n")
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
