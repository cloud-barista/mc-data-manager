package diagnostics

import (
	"fmt"
	"database/sql"
)

const ThreadQueryTpl = `
SELECT
    MAX(CASE WHEN VARIABLE_NAME = 'Threads_connected' THEN VARIABLE_VALUE END) AS threads_connected,
    MAX(CASE WHEN VARIABLE_NAME = 'Threads_running'   THEN VARIABLE_VALUE END) AS threads_running
FROM %s
WHERE VARIABLE_NAME IN ('Threads_connected', 'Threads_running');
`

type DatabaseThreadStat struct {
	ThreadConnected int64
	ThreadRunning   int64
}

type DatabaseThreadCollector struct {
	DB *sql.DB
}

func NewDatabaseThreadCollector(db *sql.DB) *DatabaseThreadCollector {
	return &DatabaseThreadCollector{
		DB: db,
	}
}

func (b *DatabaseThreadCollector) Collect() (DatabaseThreadStat, error) {
	var fromCandidates = []string{
		"performance_schema.global_status", // MySQL
		"information_schema.GLOBAL_STATUS", // MariaDB
	}
    var threadStat DatabaseThreadStat
    var lastErr error

    for _, from := range fromCandidates {
        query := fmt.Sprintf(ThreadQueryTpl, from)

        err := b.DB.QueryRow(query).Scan(&threadStat.ThreadConnected, &threadStat.ThreadRunning)
        if err == nil {
            // 성공
            return threadStat, nil
        }

        // 실패하면 다음 candidate 시도
        lastErr = fmt.Errorf("query failed on %s: %w", from, err)
    }

    // 모든 candidate 실패 시 에러 반환
    return DatabaseThreadStat{}, lastErr
}
