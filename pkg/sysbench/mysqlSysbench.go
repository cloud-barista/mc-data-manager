package sysbench

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MySQLSysbenchParsed struct {
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

func ParseMySQLSysbenchOutput(out []byte) (MySQLSysbenchParsed, error) {
	sp := MySQLSysbenchParsed{Raw: string(out)}
	sc := bufio.NewScanner(bytes.NewReader(out))

	// 자주 쓰는 정규식
	reUint := regexp.MustCompile(`([0-9]+)`)
	// reFloat := regexp.MustCompile(`:\s*([0-9]*\.?[0-9]+)\s*$`)
	reFloat := regexp.MustCompile("(?:(?::\\s*)|^)\\s*([0-9]*\\.?[0-9]+)\\s*$")
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
	left = parts[1]
	// right: "0.00"
	right = parts[len(parts)-1]
	return strings.TrimSpace(left), strings.TrimSpace(right)
}

func (s MySQLSysbenchParsed) FormatMySQLSysbenchLike() string {
	return fmt.Sprintf(
		`SQL statistics:
    queries performed:
        read:                            %d
        write:                           %d
        other:                           %d
        total:                           %d
    transactions:                        %d   (%.2f per sec.)
    queries:                             %d  (%.2f per sec.)
    ignored errors:                      %d      (%.2f per sec.)
    reconnects:                          %d      (%.2f per sec.)

General statistics:
    total time:                          %s
    total number of events:              %d

Latency (ms):
         min:                                    %.2f
         avg:                                    %.2f
         max:                                    %.2f
         95th percentile:                        %.2f
         sum:                                 %.2f

Threads fairness:
    events (avg/stddev):           %.4f/%.2f
    execution time (avg/stddev):   %.4f/%.2f
`,
		// SQL stats
		s.Read, s.Write, s.Other, s.Total,
		s.Transactions, s.TPS,
		s.Queries, s.QPS,
		s.IgnoredErr, s.IgnoredErrPS,
		s.Reconnects, s.ReconnectsPS,
		// General
		formatDurationLikeSysbench(s.TotalTime),
		s.TotalEvents,
		// Latency
		s.LatencyMin, s.LatencyAvg, s.LatencyMax, s.LatencyP95, s.LatencySum,
		// Threads fairness
		s.EventsAvg, s.EventsStddev, s.ExecTimeAvgSec, s.ExecTimeStddevSec,
	)
}

// sysbench는 보통 소수점까지 "10.0048s" 형태로 보여줘서,
// time.Duration을 그 느낌에 맞춰 문자열로 바꿔줍니다.
func formatDurationLikeSysbench(d time.Duration) string {
	if d <= 0 {
		return "0.0000s"
	}
	sec := float64(d) / float64(time.Second)
	return fmt.Sprintf("%.4fs", sec)
}
