package diagnostics

import (
	"context"
	"database/sql"
	"time"
)

const LockQuery string = `
SELECT
    OBJECT_SCHEMA,
    OBJECT_NAME,
    COUNT_STAR       AS lock_wait_count,
    SUM_TIMER_WAIT   AS total_wait_time_ps
 FROM performance_schema.table_lock_waits_summary_by_table
WHERE OBJECT_SCHEMA NOT IN ('performance_schema', 'mysql');
`

type TableLockStat struct {
	Schema        string
	Table         string
	LockWaitCount uint64
	TotalWaitPS   uint64
}

type TableLockDelta struct {
	Schema        string
	Table         string
	LockWaitCount uint64
	TotalWaitPS   uint64
	Elapsed       time.Duration
}

func (d TableLockDelta) TotalWait() time.Duration {
	// ps → ns로 내림 변환 (1ns = 1000ps)
	return time.Duration(d.TotalWaitPS/1000) * time.Nanosecond
}

func (d TableLockDelta) WaitsPerSec() float64 {
	if d.Elapsed <= 0 {
		return 0
	}
	return float64(d.LockWaitCount) / d.Elapsed.Seconds()
}

func (d TableLockDelta) WaitTimePerSec() time.Duration {
	if d.Elapsed <= 0 {
		return 0
	}
	// 초당 누적 대기시간(= 대기시간/경과초)
	psPerSec := float64(d.TotalWaitPS) / d.Elapsed.Seconds()
	nsPerSec := psPerSec / 1000.0
	return time.Duration(nsPerSec) * time.Nanosecond
}

type LockWaitsCollector struct {
	DB *sql.DB
}

func NewLockWaitsCollector(db *sql.DB) *LockWaitsCollector {
	return &LockWaitsCollector{
		DB: db,
	}
}

func (c *LockWaitsCollector) Snapshot(ctx context.Context) (Snapshot[TableLockStat], error) {
	rows, err := c.DB.QueryContext(ctx, LockQuery)
	if err != nil {
		return Snapshot[TableLockStat]{}, err
	}
	defer rows.Close()

	snap := Snapshot[TableLockStat]{TakenAt: time.Now(), Items: make(map[string]TableLockStat)}
	for rows.Next() {
		var s TableLockStat
		if err := rows.Scan(&s.Schema, &s.Table, &s.LockWaitCount, &s.TotalWaitPS); err != nil {
			return Snapshot[TableLockStat]{}, err
		}
		snap.Items[key(s.Schema, s.Table)] = s
	}
	return snap, rows.Err()
}

func DiffLock(before, after Snapshot[TableLockStat]) ([]TableLockDelta, time.Duration) {
	elapsed := after.TakenAt.Sub(before.TakenAt)
	out := make([]TableLockDelta, 0, len(after.Items))
	for k, a := range after.Items {
		b, ok := before.Items[k]
		var dCount, dPS uint64
		if ok {
			if a.LockWaitCount >= b.LockWaitCount {
				dCount = a.LockWaitCount - b.LockWaitCount
			}
			if a.TotalWaitPS >= b.TotalWaitPS {
				dPS = a.TotalWaitPS - b.TotalWaitPS
			}
		} else {
			dCount, dPS = a.LockWaitCount, a.TotalWaitPS
		}
		out = append(out, TableLockDelta{
			Schema:        a.Schema,
			Table:         a.Table,
			LockWaitCount: dCount,
			TotalWaitPS:   dPS,
			Elapsed:       elapsed,
		})
	}
	return out, elapsed
}
