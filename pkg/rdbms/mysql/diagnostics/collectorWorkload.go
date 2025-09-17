package diagnostics

import (
	"context"
	"database/sql"
	"time"
)

// 특정 스키마만 합산
const qWorkloadBySchema = `
SELECT
    ?                       AS schema_name,
    SUM(COUNT_STAR)         AS total_queries,
    SUM(SUM_TIMER_WAIT)     AS total_latency_ps,
    MAX(MAX_TIMER_WAIT)     AS max_latency_ps
FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME = ?
  AND DIGEST_TEXT NOT LIKE 'SET ` + "`autocommit`" + `%'
  AND DIGEST_TEXT NOT LIKE 'SET NAMES%'
  AND DIGEST_TEXT NOT LIKE 'COMMIT%'
  AND DIGEST_TEXT NOT LIKE 'ROLLBACK%'
  AND DIGEST_TEXT NOT LIKE 'START TRANSACTION%'
  AND DIGEST_TEXT NOT LIKE 'SELECT @@%'
  AND DIGEST_TEXT NOT LIKE 'BEGIN%';
`

type WorkloadStat struct {
	Schema         string
	TotalQueries   uint64 // SUM(COUNT_STAR)
	TotalLatencyPS uint64 // SUM(SUM_TIMER_WAIT)
	MaxLatencyPS   uint64 // MAX(MAX_TIMER_WAIT)  ← 추가
}

type WorkloadDelta struct {
	Schema         string
	TotalQueries   uint64
	TotalLatencyPS uint64
	// 아래 2개는 파생값
	AvgLatencyPS uint64 // ΔSUM/ΔCOUNT 를 ps로 표현 (보기용 메서드도 제공)
	MaxLatencyPS uint64 // "해당 구간의 최대" (근사: 아래 설명)
	Elapsed      time.Duration
}

// 편의: ns 단위로 변환
func (d WorkloadDelta) TotalLatency() time.Duration {
	return time.Duration(d.TotalLatencyPS/1000) * time.Nanosecond
}

func (d WorkloadDelta) AvgLatency() time.Duration {
	if d.TotalQueries == 0 {
		return 0
	}
	// ps -> ns
	psPerStmt := d.TotalLatencyPS / d.TotalQueries
	return time.Duration(psPerStmt/1000) * time.Nanosecond
}

func (d WorkloadDelta) MaxLatency() time.Duration {
	return time.Duration(d.MaxLatencyPS/1000) * time.Nanosecond
}

// 초당 처리 건수 / 초당 지연시간
func (d WorkloadDelta) QPS() float64 {
	if d.Elapsed <= 0 {
		return 0
	}
	return float64(d.TotalQueries) / d.Elapsed.Seconds()
}

func (d WorkloadDelta) LatencyPerSec() time.Duration {
	if d.Elapsed <= 0 {
		return 0
	}
	psPerSec := float64(d.TotalLatencyPS) / d.Elapsed.Seconds()
	nsPerSec := psPerSec / 1000.0
	return time.Duration(nsPerSec) * time.Nanosecond
}

type WorkloadCollector struct {
	DB *sql.DB
}

func NewWorkloadCollector(db *sql.DB) *WorkloadCollector {
	return &WorkloadCollector{
		DB: db,
	}
}

func (c *WorkloadCollector) Snapshot(ctx context.Context, schema string) (Snapshot[WorkloadStat], error) {
	row := c.DB.QueryRowContext(ctx, qWorkloadBySchema, schema, schema)

	var w WorkloadStat
	if err := row.Scan(&w.Schema, &w.TotalQueries, &w.TotalLatencyPS, &w.MaxLatencyPS); err != nil {
		return Snapshot[WorkloadStat]{}, err
	}

	snap := Snapshot[WorkloadStat]{
		TakenAt: time.Now(),
		Items:   map[string]WorkloadStat{schema: w}, // 키를 스키마 단위로
	}
	return snap, nil
}

func DiffWorkload(before, after Snapshot[WorkloadStat]) (WorkloadDelta, time.Duration) {
	elapsed := after.TakenAt.Sub(before.TakenAt)
	b := one(before.Items)
	a := one(after.Items)

	dq, dps := uint64(0), uint64(0)
	if a.TotalQueries >= b.TotalQueries {
		dq = a.TotalQueries - b.TotalQueries
	}
	if a.TotalLatencyPS >= b.TotalLatencyPS {
		dps = a.TotalLatencyPS - b.TotalLatencyPS
	}

	// 평균 지연(정확): ΔSUM / ΔCOUNT
	avgPS := uint64(0)
	if dq > 0 {
		avgPS = dps / dq
	}

	// 최댓값(근사, 아래 설명 참고)
	// MAX_TIMER_WAIT은 “서버 재시작 이후 관측된 최대”라 누적 지표입니다.
	// 구간 내 새로운 최대가 발생했다면 after.Max > before.Max 가 됩니다.
	// 이 경우, “그 구간에서 관측된 최대”를 after.Max 로 간주(보수적 근사)합니다.
	var maxPS uint64
	if a.MaxLatencyPS > b.MaxLatencyPS {
		maxPS = a.MaxLatencyPS
	} else {
		maxPS = 0 // 새 최대가 없다면 구간 내 최대를 확정할 수 없어 0으로 두고, 설명을 함께 표기
	}

	return WorkloadDelta{
		Schema:         a.Schema,
		TotalQueries:   dq,
		TotalLatencyPS: dps,
		AvgLatencyPS:   avgPS,
		MaxLatencyPS:   maxPS,
		Elapsed:        elapsed,
	}, elapsed
}

func one[M any](m map[string]M) M {
	for _, v := range m {
		return v
	}
	var z M
	return z
}
