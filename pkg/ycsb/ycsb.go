package ycsb

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	"go.mongodb.org/mongo-driver/mongo"
)

func RunYCSB(ctx context.Context, phase, binding string, env []string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{phase, binding}, args...)

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "./bin/go-ycsb", cmdArgs...)
	cmd.Dir = filepath.Join(home, "go-ycsb")
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("go-ycsb failed: %w; output: %s", err, string(out))
	}
	return out, nil
}

var (
	statLineRe    = regexp.MustCompile(`^([A-Z_]+)\s+- Takes\(s\): ([\d.]+), Count: (\d+), OPS: ([\d.]+), Avg\(us\): (\d+), Min\(us\): (\d+), Max\(us\): (\d+), 50th\(us\): (\d+), 90th\(us\): (\d+), 95th\(us\): (\d+), 99th\(us\): (\d+), 99\.9th\(us\): (\d+), 99\.99th\(us\): (\d+)`)
	commandRe     = regexp.MustCompile(`^"command"="(load|run)"`)
	runFinishedRe = regexp.MustCompile(`^Run finished, takes`)
)

// phaseOps whitelists which operations' final summary lines are kept per
// phase; TOTAL and *_ERROR lines are excluded (still visible in Raw).
var phaseOps = map[string]map[string]bool{
	"load": {"INSERT": true},
	"run":  {"READ": true, "UPDATE": true},
}

func ParseYCSBOutput(out []byte) (models.YcsbParsed, error) {
	parsed := models.YcsbParsed{Raw: string(out)}

	phase := ""
	inSummary := false
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()

		if m := commandRe.FindStringSubmatch(line); m != nil {
			phase = m[1]
			inSummary = false
			continue
		}
		if runFinishedRe.MatchString(line) {
			inSummary = true
			continue
		}
		if !inSummary {
			continue
		}

		m := statLineRe.FindStringSubmatch(line)
		if m == nil {
			// The summary block ends at the first non-stat line
			// (e.g. trailing logs or the next phase's preamble).
			inSummary = false
			continue
		}
		if !phaseOps[phase][m[1]] {
			continue
		}

		parsed.Operations = append(parsed.Operations, models.YcsbOperationReport{
			Operation: m[1],
			TakesSec:  atof(m[2]),
			Count:     atoi(m[3]),
			OPS:       atof(m[4]),
			AvgMs:     usToMs(m[5]),
			MinMs:     usToMs(m[6]),
			MaxMs:     usToMs(m[7]),
			P50Ms:     usToMs(m[8]),
			P90Ms:     usToMs(m[9]),
			P95Ms:     usToMs(m[10]),
			P99Ms:     usToMs(m[11]),
			P999Ms:    usToMs(m[12]),
			P9999Ms:   usToMs(m[13]),
		})
	}
	if err := scanner.Err(); err != nil {
		return parsed, fmt.Errorf("failed to scan go-ycsb output: %w", err)
	}
	return parsed, nil
}

// atoi/atof convert regex-matched numeric groups; the patterns guarantee
// digits, so conversion errors are not possible.
func atoi(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func atof(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// usToMs converts a microsecond value matched from go-ycsb output (which
// reports latencies in us) to milliseconds for the response model.
func usToMs(s string) float64 {
	return atof(s) / 1000
}

// DropMongoCollection removes the collection created for a mongodb diagnose run
// (ncp/alibaba only — dynamodb cleans up itself via dynamodb.delete.after.run.stage).
// go-ycsb's mongodb binding always writes to the "ycsb" database, so only the
// collection named after the diagnose table is dropped.
func DropMongoCollection(ctx context.Context, provider, host, port, username, password, table string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", port, err)
	}

	var client *mongo.Client
	switch provider {
	case "ncp":
		client, err = config.NewNCPMongoDBClient(username, password, host, portNum)
	case "alibaba":
		client, err = config.NewAlibabaMongoDBClient(username, password, host, portNum)
	default:
		return fmt.Errorf("unsupported provider for mongodb cleanup: %s", provider)
	}
	if err != nil {
		return fmt.Errorf("mongodb connect failed: %w", err)
	}
	defer func() { _ = client.Disconnect(ctx) }()

	return client.Database("ycsb").Collection(table).Drop(ctx)
}
