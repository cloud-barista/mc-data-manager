package diagnostics

import "database/sql"

const ThreadQuery string = `
SELECT
    MAX(CASE WHEN VARIABLE_NAME = 'Threads_connected' THEN VARIABLE_VALUE END) AS threads_connected,
    MAX(CASE WHEN VARIABLE_NAME = 'Threads_running'   THEN VARIABLE_VALUE END) AS threads_running
 FROM information_schema.GLOBAL_STATUS
WHERE VARIABLE_NAME IN ('Threads_connected', 'Threads_running');
`

// FROM information_schema.GLOBAL_STATUS; // MariaDB
// FROM performance_schema.global_status; // MySQL;

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
	var threadStat DatabaseThreadStat
	err := b.DB.QueryRow(ThreadQuery).Scan(&threadStat.ThreadConnected, &threadStat.ThreadRunning)
	if err != nil {
		return DatabaseThreadStat{}, err
	}

	return threadStat, nil
}
