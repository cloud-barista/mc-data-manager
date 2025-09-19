package diagnostics_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	// sqlDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", "root", "N@mutech7^^7", "10.10.12.131", "3306"))
	// if err != nil {
	// 	panic(err)
	// }

	// ctx := context.Background()
	// collector := diagnostics.NewCollector(sqlDB)
	// // res2, err := collector.RunTimed(ctx, "mcmp", 30*time.Second)
	// res2, err := collector.WithDiagnostic(ctx, func(ctx context.Context) error {
	// 	// 여기서 실제 기능 실행: 예) 배치 작업, 마이그레이션, 대량 업데이트 등
	// 	count()

	// 	return nil
	// })
	// if err != nil { /* handle */
	// 	panic(err)
	// }

	// fmt.Println(diagnostics.PrintBufferReport(res2.Buffer))
	// fmt.Println(diagnostics.PrintWorkloadReport(res2.Work, res2.Elapsed))
	// fmt.Println(diagnostics.PrintThreadReport(res2.Thread))
	// fmt.Println(diagnostics.PrintLockReport(res2.Lock, res2.Elapsed))
	// fmt.Println(diagnostics.PrintIOReport(res2.IO, res2.Elapsed))
	// fmt.Println(res2.Report.String())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, prepareErr := RunSysbench(ctx,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host=10.10.12.140",
		"--mysql-port=3307",
		"--mysql-user=root",
		"--mysql-password=mcmp",
		"--mysql-db=mcmp",
		"--tables=1",
		"--table-size=1000",
		"--threads=8",
		"--time=10",
		"prepare",
	)
	if prepareErr != nil {
		fmt.Println("sysbench error:", prepareErr)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	// 예: sysbench oltp_read_write --threads=8 --time=10 --report-interval=1 run
	res, err := RunSysbench(ctx,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host=10.10.12.140",
		"--mysql-port=3307",
		"--mysql-user=root",
		"--mysql-password=mcmp",
		"--mysql-db=mcmp",
		"--tables=1",
		"--table-size=1000",
		"--threads=8",
		"--time=10",
		"run",
	)
	if err != nil {
		fmt.Println("sysbench error:", err)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	_, cleanErr := RunSysbench(ctx,
		"oltp_read_write",
		"--db-driver=mysql",
		"--mysql-host=10.10.12.140",
		"--mysql-port=3307",
		"--mysql-user=root",
		"--mysql-password=mcmp",
		"--mysql-db=mcmp",
		"--tables=1",
		"--table-size=1000",
		"--threads=8",
		"--time=10",
		"cleanup",
	)
	if cleanErr != nil {
		fmt.Println("sysbench error:", cleanErr)
		// 그래도 res.Raw 안에 원문이 있으니 디버깅 가능
	}

	fmt.Printf("QPS=%.2f, TPS=%.2f, p95=%.2fms, total=%d\n",
		res.QPS, res.TPS, res.LatencyP95, res.Total)
}

// func count() {
// 	for i := 1; i <= 20; i++ {
// 		fmt.Printf("%d \n", i)
// 		time.Sleep(time.Second * 1)
// 	}
// }

type SysbenchParsed struct {
	// SQL statistics
	Read         uint64  `json:"read"`
	Write        uint64  `json:"write"`
	Other        uint64  `json:"other"`
	Total        uint64  `json:"total"`
	Transactions uint64  `json:"transactions"`
	TPS          float64 `json:"tps"`
	Queries      uint64  `json:"queries"`
	QPS          float64 `json:"qps"`
	IgnoredErr   uint64  `json:"ignoredErrors"`
	IgnoredErrPS float64 `json:"ignoredErrorsPerSec"`
	Reconnects   uint64  `json:"reconnects"`
	ReconnectsPS float64 `json:"reconnectsPerSec"`

	// General
	TotalTime   time.Duration `json:"totalTime"`
	TotalEvents uint64        `json:"totalEvents"`

	// Latency (ms)
	LatencyMin float64 `json:"latencyMinMs"`
	LatencyAvg float64 `json:"latencyAvgMs"`
	LatencyMax float64 `json:"latencyMaxMs"`
	LatencyP95 float64 `json:"latencyP95Ms"`
	LatencySum float64 `json:"latencySumMs"`

	// Threads fairness
	EventsAvg         float64 `json:"eventsAvg"`
	EventsStddev      float64 `json:"eventsStddev"`
	ExecTimeAvgSec    float64 `json:"execTimeAvgSec"`
	ExecTimeStddevSec float64 `json:"execTimeStddevSec"`

	// Raw output (원문 보존)
	Raw string `json:"raw"`
}

/***** 실행 + 파싱 *****/

// RunSysbench 실행: ctx로 타임아웃/취소 제어, sysbench args를 그대로 전달
func RunSysbench(ctx context.Context, args ...string) (SysbenchParsed, error) {
	cmd := exec.CommandContext(ctx, "sysbench", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// sysbench 실행 실패(커맨드 자체 실패)와 벤치 실패를 구분하기 어렵기 때문에
		// 원문 출력과 함께 에러 반환
		return SysbenchParsed{Raw: string(out)}, fmt.Errorf("sysbench failed: %w; output: %s", err, string(out))
	}
	return ParseSysbenchOutput(out)
}

/***** 파서 본체 *****/

func ParseSysbenchOutput(out []byte) (SysbenchParsed, error) {
	sp := SysbenchParsed{Raw: string(out)}
	sc := bufio.NewScanner(bytes.NewReader(out))

	// 자주 쓰는 정규식
	reUint := regexp.MustCompile(`([0-9]+)`)
	reFloat := regexp.MustCompile(`([0-9]*\.?[0-9]+)`)
	// 패턴: "X: <num>  (<num> per sec.)"
	reValueAndRate := regexp.MustCompile(`:\s*([0-9]+)\s*\(\s*([0-9]*\.?[0-9]+)\s*per sec\.\s*\)`)
	// 패턴: "X: <num>  (<num> per sec.)" (0.00도 포함)
	reValueAndRateZeroOk := reValueAndRate

	section := ""
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case strings.HasPrefix(line, "SQL statistics"):
			section = "sql"
			continue
		case strings.HasPrefix(line, "General statistics"):
			section = "general"
			continue
		case strings.HasPrefix(line, "Latency (ms)"):
			section = "latency"
			continue
		case strings.HasPrefix(line, "Threads fairness"):
			section = "fairness"
			continue
		}

		switch section {
		case "sql":
			switch {
			case strings.HasPrefix(line, "read:"):
				sp.Read = parseUintFromTail(reUint, line)
			case strings.HasPrefix(line, "write:"):
				sp.Write = parseUintFromTail(reUint, line)
			case strings.HasPrefix(line, "other:"):
				sp.Other = parseUintFromTail(reUint, line)
			case strings.HasPrefix(line, "total:"):
				sp.Total = parseUintFromTail(reUint, line)
			case strings.HasPrefix(line, "transactions:"):
				// "transactions: 2170   (216.84 per sec.)"
				val, rate := parseValueAndRate(reValueAndRate, line)
				sp.Transactions = val
				sp.TPS = rate
			case strings.HasPrefix(line, "queries:"):
				val, rate := parseValueAndRate(reValueAndRate, line)
				sp.Queries = val
				sp.QPS = rate
			case strings.HasPrefix(line, "ignored errors:"):
				val, rate := parseValueAndRate(reValueAndRateZeroOk, line)
				sp.IgnoredErr = val
				sp.IgnoredErrPS = rate
			case strings.HasPrefix(line, "reconnects:"):
				val, rate := parseValueAndRate(reValueAndRateZeroOk, line)
				sp.Reconnects = val
				sp.ReconnectsPS = rate
			}

		case "general":
			if strings.HasPrefix(line, "total time:") {
				// e.g. "total time: 10.0048s"
				txt := strings.TrimSpace(strings.TrimPrefix(line, "total time:"))
				dur, err := time.ParseDuration(strings.ReplaceAll(txt, " ", ""))
				if err == nil {
					sp.TotalTime = dur
				}
			} else if strings.HasPrefix(line, "total number of events:") {
				sp.TotalEvents = parseUintFromTail(reUint, line)
			}

		case "latency":
			switch {
			case strings.HasPrefix(line, "min:"):
				sp.LatencyMin = parseFloatFromTail(reFloat, line)
			case strings.HasPrefix(line, "avg:"):
				sp.LatencyAvg = parseFloatFromTail(reFloat, line)
			case strings.HasPrefix(line, "max:"):
				sp.LatencyMax = parseFloatFromTail(reFloat, line)
			case strings.HasPrefix(line, "95th percentile:"):
				sp.LatencyP95 = parseFloatFromTail(reFloat, line)
			case strings.HasPrefix(line, "sum:"):
				sp.LatencySum = parseFloatFromTail(reFloat, line)
			}

		case "fairness":
			// "events (avg/stddev):           2170.0000/0.00"
			if strings.HasPrefix(line, "events (avg/stddev):") {
				l, r := splitBySlash(line)
				sp.EventsAvg = parseFloatFromTail(reFloat, l)
				sp.EventsStddev = parseFloatFromTail(reFloat, r)
			}
			// "execution time (avg/stddev):   9.9989/0.00"
			if strings.HasPrefix(line, "execution time (avg/stddev):") {
				l, r := splitBySlash(line)
				sp.ExecTimeAvgSec = parseFloatFromTail(reFloat, l)
				sp.ExecTimeStddevSec = parseFloatFromTail(reFloat, r)
			}
		}
	}

	if err := sc.Err(); err != nil {
		return sp, err
	}

	// 간단 유효성 체크
	if sp.Total == 0 && sp.Queries == 0 && sp.Transactions == 0 && sp.TotalEvents == 0 {
		return sp, errors.New("parse failed: no recognizable metrics in sysbench output")
	}
	return sp, nil
}

/***** 유틸 *****/

func parseUintFromTail(re *regexp.Regexp, line string) uint64 {
	m := re.FindStringSubmatch(line)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(m[1], 10, 64)
	return v
}

func parseFloatFromTail(re *regexp.Regexp, line string) float64 {
	m := re.FindStringSubmatch(line)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.ParseFloat(m[1], 64)
	return v
}

// line like: "transactions: 2170   (216.84 per sec.)"
func parseValueAndRate(re *regexp.Regexp, line string) (uint64, float64) {
	m := re.FindStringSubmatch(line)
	if len(m) < 3 {
		return 0, 0
	}
	val, _ := strconv.ParseUint(m[1], 10, 64)
	rate, _ := strconv.ParseFloat(m[2], 64)
	return val, rate
}

func splitBySlash(line string) (left string, right string) {
	parts := strings.Split(line, "/")
	if len(parts) < 2 {
		return line, ""
	}
	// left: "events (avg/stddev):  2170.0000"
	left = parts[0]
	// right: "0.00"
	right = parts[len(parts)-1]
	return strings.TrimSpace(left), strings.TrimSpace(right)
}
