package diagnostics

import (
	"context"
	"database/sql"
	"sort"
	"time"
)

// Digest 요약 스냅샷 1행
type DigestStat struct {
	Schema    string
	Digest    string // 해시
	Text      string // 정규화된 SQL (DIGEST_TEXT)
	Count     uint64 // COUNT_STAR
	LatencyPS uint64 // SUM_TIMER_WAIT (ps)
}

// 두 스냅샷 차분
type DigestDelta struct {
	Schema    string
	Digest    string
	Text      string
	Count     uint64 // ΔCOUNT_STAR
	LatencyPS uint64 // ΔSUM_TIMER_WAIT (ps)
	Elapsed   time.Duration
}

type DigestCollector struct {
	DB *sql.DB
}

// 관리성 쿼리 필터 (raw string 안에 backtick은 문자열 결합으로 삽입)
const digestMgmtFilter = `
  AND DIGEST_TEXT NOT LIKE 'SET ` + "`autocommit`" + `%%'
  AND DIGEST_TEXT NOT LIKE 'SET NAMES%%'
  AND DIGEST_TEXT NOT LIKE 'COMMIT%%'
  AND DIGEST_TEXT NOT LIKE 'ROLLBACK%%'
  AND DIGEST_TEXT NOT LIKE 'START TRANSACTION%%'
  AND DIGEST_TEXT NOT LIKE 'SELECT @@%%'
  AND DIGEST_TEXT NOT LIKE 'BEGIN%%'
`

// 스키마 지정 버전 (schema != "" 인 경우 사용)
const qDigestBySchema = `
SELECT
  SCHEMA_NAME,
  DIGEST,
  DIGEST_TEXT,
  COUNT_STAR,
  SUM_TIMER_WAIT
 FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME = ?
` + digestMgmtFilter

// 전체 사용자 스키마 대상 (schema == ""인 경우 사용)
const qDigestAllUserSchemas = `
SELECT
  SCHEMA_NAME,
  DIGEST,
  DIGEST_TEXT,
  COUNT_STAR,
  SUM_TIMER_WAIT
 FROM performance_schema.events_statements_summary_by_digest
WHERE SCHEMA_NAME NOT IN ('performance_schema','mysql','information_schema','sys')
` + digestMgmtFilter

func NewDigestCollector(db *sql.DB) *DigestCollector { return &DigestCollector{DB: db} }

// schema: "" 이면 시스템 스키마 제외 전체, 아니면 해당 스키마만
func (c *DigestCollector) Snapshot(ctx context.Context, schema string) (Snapshot[DigestStat], error) {
	var rows *sql.Rows
	var err error
	if schema == "" {
		rows, err = c.DB.QueryContext(ctx, qDigestAllUserSchemas)
	} else {
		rows, err = c.DB.QueryContext(ctx, qDigestBySchema, schema)
	}
	if err != nil {
		return Snapshot[DigestStat]{}, err
	}
	defer rows.Close()

	snap := Snapshot[DigestStat]{TakenAt: time.Now(), Items: make(map[string]DigestStat)}
	for rows.Next() {
		var s DigestStat
		if err := rows.Scan(&s.Schema, &s.Digest, &s.Text, &s.Count, &s.LatencyPS); err != nil {
			return Snapshot[DigestStat]{}, err
		}
		// digest 해시를 키로 사용 (스키마+digest로 키 구성해도 OK)
		snap.Items[s.Digest] = s
	}
	if err := rows.Err(); err != nil {
		return Snapshot[DigestStat]{}, err
	}
	return snap, nil
}

// 차분 계산 + 총 지연(ps) 반환 (TopN 비중 계산 등에 유용)
func DiffDigest(before, after Snapshot[DigestStat]) ([]DigestDelta, time.Duration, uint64) {
	elapsed := after.TakenAt.Sub(before.TakenAt)
	out := make([]DigestDelta, 0, len(after.Items))
	var totalLatencyPS uint64

	for k, a := range after.Items {
		b, ok := before.Items[k]

		var dCnt, dLat uint64
		var text = a.Text
		var schema = a.Schema

		if ok {
			if a.Count >= b.Count {
				dCnt = a.Count - b.Count
			}
			if a.LatencyPS >= b.LatencyPS {
				dLat = a.LatencyPS - b.LatencyPS
			}
			if text == "" {
				text = b.Text
			}
			if schema == "" {
				schema = b.Schema
			}
		} else {
			// 새로 등장 → 구간 절대치로 간주
			dCnt = a.Count
			dLat = a.LatencyPS
		}

		if dCnt == 0 && dLat == 0 {
			continue
		}

		totalLatencyPS += dLat
		out = append(out, DigestDelta{
			Schema: schema, Digest: k, Text: text,
			Count: dCnt, LatencyPS: dLat, Elapsed: elapsed,
		})
	}

	// 지연시간 내림차순 정렬
	sort.Slice(out, func(i, j int) bool { return out[i].LatencyPS > out[j].LatencyPS })

	return out, elapsed, totalLatencyPS
}
