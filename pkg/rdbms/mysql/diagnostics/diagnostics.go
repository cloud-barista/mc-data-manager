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
	"strings"
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
	Digest *DigestCollector
}

func NewCollector(db *sql.DB) *Collector {
	return &Collector{
		Lock:   NewLockWaitsCollector(db),
		IO:     NewTableIOCollector(db),
		Buffer: NewDatabaseBufferCollector(db),
		Work:   NewWorkloadCollector(db),
		Thread: NewDatabaseThreadCollector(db),
		Digest: NewDigestCollector(db),
	}
}

type SysbenchDigestReport struct {
	Read  uint64
	Write uint64
	Other uint64
	Total uint64

	AvgMs float64 // ΔSUM/ΔCOUNT (ms)
	SumMs float64 // ΔSUM (ms)
	QPS   float64
	// TPS: 옵션 (아래 주석 참고)
	Elapsed time.Duration
}

type TimedResult struct {
	Lock    []TableLockDelta
	IO      []TableIODelta
	Buffer  DatabaseBufferStat
	Work    []WorkloadDelta
	Thread  DatabaseThreadStat
	Report  SysbenchDigestReport
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
	digestBefore, err := c.Digest.Snapshot(ctx, "")
	if err != nil {
		res.Errors["digest_before"] = err
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
	digestAfter, err := c.Digest.Snapshot(ctx, "")
	if err != nil {
		res.Errors["digest_after"] = err
	}

	var (
		lDelta   []TableLockDelta
		lElapsed time.Duration
		iDelta   []TableIODelta
		iElapsed time.Duration
		wDelta   []WorkloadDelta
		wElapsed time.Duration
		dDelta   []DigestDelta
		dElapsed time.Duration
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
	if res.Errors["digest_before"] == nil && res.Errors["digest_after"] == nil {
		dDelta, dElapsed, _ = DiffDigest(digestBefore, digestAfter)
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
	res.Elapsed = maxDuration(lElapsed, iElapsed, wElapsed, dElapsed)

	// 결과 채우기
	res.Lock = lDelta
	res.IO = iDelta
	res.Work = wDelta
	res.Report = BuildSysbenchDigestReport(wDelta, dDelta, res.Elapsed)

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
	digestBefore, err := c.Digest.Snapshot(ctx, "")
	if err != nil {
		res.Errors["digest_before"] = err
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
	digestAfter, err := c.Digest.Snapshot(ctx, "")
	if err != nil {
		res.Errors["digest_after"] = err
	}

	var (
		lDelta   []TableLockDelta
		lElapsed time.Duration
		iDelta   []TableIODelta
		iElapsed time.Duration
		wDelta   []WorkloadDelta
		wElapsed time.Duration
		dDelta   []DigestDelta
		dElapsed time.Duration
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
	if res.Errors["digest_before"] == nil && res.Errors["digest_after"] == nil {
		dDelta, dElapsed, _ = DiffDigest(digestBefore, digestAfter)
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
	res.Elapsed = maxDuration(lElapsed, iElapsed, wElapsed, dElapsed)

	// 결과 채우기
	res.Lock = lDelta
	res.IO = iDelta
	res.Work = wDelta
	res.Report = BuildSysbenchDigestReport(wDelta, dDelta, res.Elapsed)

	// 반환 error는 오직 fn(ctx)에서 난 것만
	return res, opErr
}

func BuildSysbenchDigestReport(work []WorkloadDelta, digests []DigestDelta, elapsed time.Duration) SysbenchDigestReport {
	var r SysbenchDigestReport
	r.Elapsed = elapsed

	// 1) read/write/other/total : DIGEST_TEXT 첫 토큰으로 분류
	for _, d := range digests {
		verb := firstVerb(d.Text) // SELECT / INSERT / UPDATE / DELETE / ...
		switch verb {
		case "SELECT":
			r.Read += d.Count
		case "INSERT", "UPDATE", "DELETE", "REPLACE", "LOAD":
			r.Write += d.Count
		default:
			r.Other += d.Count
		}
	}
	r.Total = r.Read + r.Write + r.Other

	// 2) avg / sum latency : WorkloadDelta 합산(스키마별일 수 있으니 모두 더함)
	var totalQueries uint64
	var totalLatencyPS uint64
	for _, w := range work {
		totalQueries += w.TotalQueries
		totalLatencyPS += w.TotalLatencyPS
	}
	if totalQueries > 0 {
		r.AvgMs = (float64(totalLatencyPS) / float64(totalQueries)) / 1e9
	}
	r.SumMs = float64(totalLatencyPS) / 1e9

	// 3) QPS
	sec := elapsed.Seconds()
	if sec > 0 {
		r.QPS = float64(r.Total) / sec
	}

	// 4) TPS (옵션)
	// 정확도를 높이려면 'COMMIT/ROLLBACK'을 관리성 제외에서 **한시적으로** 해제한
	// digest 또는 event_name summary로 Δ카운트를 구해 r.TPS = trxDelta / sec 로 채우세요.

	return r
}

func firstVerb(digestText string) string {
	// DIGEST_TEXT는 대부분 대문자 시작(정규화). 혹시 몰라 trim + 대문자화.
	s := strings.TrimSpace(digestText)
	if s == "" {
		return ""
	}
	// 첫 단어 추출
	i := strings.IndexAny(s, " \t\n\r(")
	if i < 0 {
		i = len(s)
	}
	return strings.ToUpper(s[:i])
}

/************** 4) 간단 리포트 출력(옵셔널) **************/

func PrintBufferReport(d DatabaseBufferStat) string {
	s := "[Buffer hit ratio]\n"
	s += fmt.Sprintf("%-35s %-10f\n", "buffer pool hit ratio percentage", d.BufferPoolHitRatio)

	return s
}

func PrintLockReport(d []TableLockDelta, elapsed time.Duration) string {
	s := fmt.Sprintf("[Lock waits] interval=%s\n", elapsed)
	s += fmt.Sprintf("%-30s %-10s %10s %18s %16s\n", "table", "schema", "waits", "wait_time", "waits/s")

	for _, x := range d {
		s += fmt.Sprintf("%-30s %-10s %10d %18s %16.2f\n",
			x.Table, x.Schema, x.LockWaitCount, x.TotalWait(), x.WaitsPerSec())
	}

	return s
}

func PrintIOReport(d []TableIODelta, elapsed time.Duration) string {
	s := fmt.Sprintf("[Table IO] interval=%s\n", elapsed)
	s += fmt.Sprintf("%-30s %-10s %12s %12s %12s %12s\n", "table", "schema", "read", "insert", "update", "delete")
	for _, x := range d {
		s += fmt.Sprintf("%-30s %-10s %12d %12d %12d %12d\n",
			x.Table, x.Schema, x.RowsRead, x.RowsInserted, x.RowsUpdated, x.RowsDeleted)
	}

	return s
}

func PrintWorkloadReport(w []WorkloadDelta, elapsed time.Duration) string {
	s := fmt.Sprintf("[Workload] interval=%s\n", elapsed)
	for _, a := range w {
		s += fmt.Sprintf("schema  : %s\n", a.Schema)
		s += fmt.Sprintf("queries : %d (QPS: %.2f)\n", a.TotalQueries, float64(a.TotalQueries)/a.Elapsed.Seconds())
		s += fmt.Sprintf("latency : total=%s avg/stmt=%s",
			(time.Duration(a.TotalLatencyPS/1000) * time.Nanosecond).String(),
			a.AvgLatency().String(),
		)
		if a.MaxLatencyPS > 0 {
			s += fmt.Sprintf(" max=%s (new max observed)\n", a.MaxLatency().String())
		} else {
			s += " max=? (no new record; exact max in interval unknown)\n"
		}
	}

	return s
}

func PrintThreadReport(d DatabaseThreadStat) string {
	s := "[Thread count]\n"
	s += fmt.Sprintf("%-10s %-10s\n", "threads_connected", "threads_running")
	s += fmt.Sprintf("%-17d %-10d\n", d.ThreadConnected, d.ThreadRunning)

	return s
}

// 보기 좋은 출력 포맷
func (r SysbenchDigestReport) String() string {
	return fmt.Sprintf(
		`SQL statistics (digest Δ-based):
    queries performed:
        read:                            %d
        write:                           %d
        other:                           %d
        total:                           %d
    queries:                             %d  (%.2f per sec.)

General statistics:
    total time:                          %s

Latency (avg only, ms):
         avg:                            %.2f
         sum:                            %.2f
`,
		r.Read, r.Write, r.Other, r.Total,
		r.Total, r.QPS,
		r.Elapsed,
		r.AvgMs, r.SumMs,
	)
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
