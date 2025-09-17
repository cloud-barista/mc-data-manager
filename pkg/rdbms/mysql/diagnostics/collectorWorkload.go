package diagnostics

import (
	"context"
	"database/sql"
	"time"
)

// 특정 스키마만 합산
const qWorkloadBySchema = `
SELECT
    SCHEMA_NAME,
    SUM(COUNT_STAR)         AS total_queries,
    SUM(SUM_TIMER_WAIT)     AS total_latency_ps,
    MAX(MAX_TIMER_WAIT)     AS max_latency_ps
 FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME NOT IN ('performance_schema', 'mysql')
  AND DIGEST_TEXT NOT LIKE 'COMMIT%'
  AND DIGEST_TEXT NOT LIKE 'ROLLBACK%'
  AND DIGEST_TEXT NOT LIKE 'START TRANSACTION%'
  AND DIGEST_TEXT NOT LIKE 'SELECT @@%'
  AND DIGEST_TEXT NOT LIKE 'SET%'
  AND DIGEST_TEXT NOT LIKE 'SHOW WARNINGS%'
  AND DIGEST_TEXT NOT LIKE 'BEGIN%'
GROUP BY SCHEMA_NAME ;
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

func (c *WorkloadCollector) Snapshot(ctx context.Context) (Snapshot[WorkloadStat], error) {
	rows, err := c.DB.QueryContext(ctx, qWorkloadBySchema)
	if err != nil {
		return Snapshot[WorkloadStat]{}, err
	}
	defer rows.Close()

	snap := Snapshot[WorkloadStat]{
		TakenAt: time.Now(),
		Items:   make(map[string]WorkloadStat), // 키를 스키마 단위로
	}

	for rows.Next() {
		var w WorkloadStat
		if err := rows.Scan(&w.Schema, &w.TotalQueries, &w.TotalLatencyPS, &w.MaxLatencyPS); err != nil {
			return Snapshot[WorkloadStat]{}, err
		}
		// 결과에 나온 스키마명을 키로 사용 (GROUP BY일 때 여러 개)
		snap.Items[w.Schema] = w
	}
	return snap, rows.Err()
}

func DiffWorkload(before, after Snapshot[WorkloadStat]) ([]WorkloadDelta, time.Duration) {
	elapsed := after.TakenAt.Sub(before.TakenAt)
	out := make([]WorkloadDelta, 0, len(after.Items))
	for schema, a := range after.Items {
		b, ok := before.Items[schema]

		var dq, dps uint64
		if ok {
			if a.TotalQueries >= b.TotalQueries {
				dq = a.TotalQueries - b.TotalQueries
			}
			if a.TotalLatencyPS >= b.TotalLatencyPS {
				dps = a.TotalLatencyPS - b.TotalLatencyPS
			}
		} else {
			// 새로 등장한 스키마로 간주 → 그 구간의 절대치로 취급
			dq = a.TotalQueries
			dps = a.TotalLatencyPS
		}

		// 평균 지연(정확): ΔSUM / ΔCOUNT
		var avgPS uint64
		if dq > 0 {
			avgPS = dps / dq
		}

		// (옵션) 구간 최대 지연 근사: 누적 최대가 증가했을 때만 after 값을 사용
		var maxPS uint64
		if ok && a.MaxLatencyPS > b.MaxLatencyPS {
			maxPS = a.MaxLatencyPS
		} else if !ok {
			// 처음 관측된 스키마라면 보수적으로 해당 값 채택하거나 0으로 둘지 정책 선택
			// 여기선 관측값 사용
			maxPS = a.MaxLatencyPS
		} else {
			maxPS = 0 // 증가 없으면 구간 내 최대를 확정하기 어려움
		}

		out = append(out, WorkloadDelta{
			Schema:         schema,
			TotalQueries:   dq,
			TotalLatencyPS: dps,
			AvgLatencyPS:   avgPS,
			MaxLatencyPS:   maxPS,
			Elapsed:        elapsed,
		})
	}

	return out, elapsed
}
