package diagnostics

import (
	"fmt"
	"database/sql"
)

const BufferCacheQueryTpl string = `
SELECT
   ROUND(
		(SUM(IF(variable_name='Innodb_buffer_pool_read_requests', variable_value,  0))
		/ (SUM(IF(variable_name IN ('Innodb_buffer_pool_read_requests', 'Innodb_buffer_pool_reads'), variable_value, 0)))
	) * 100, 2) AS buffer_pool_hit_ratio_pct
FROM %s
`

type DatabaseBufferStat struct {
	BufferPoolHitRatio float64
}

type DatabaseBufferCollector struct {
	DB *sql.DB
}

func NewDatabaseBufferCollector(db *sql.DB) *DatabaseBufferCollector {
	return &DatabaseBufferCollector{
		DB: db,
	}
}

func (b *DatabaseBufferCollector) Collect() (DatabaseBufferStat, error) {
	var fromCandidates = []string{
		"performance_schema.global_status", // MySQL
		"information_schema.GLOBAL_STATUS", // MariaDB
	}
    var stat DatabaseBufferStat
    var lastErr error

    for _, from := range fromCandidates {
        query := fmt.Sprintf(BufferCacheQueryTpl, from)

        var hitRatio float64
        err := b.DB.QueryRow(query).Scan(&hitRatio)
        if err == nil {
            stat.BufferPoolHitRatio = hitRatio
            return stat, nil
        }

        lastErr = fmt.Errorf("buffer cache query failed on %s: %w", from, err)
    }

    return DatabaseBufferStat{}, lastErr
}
