package diagnostics

import (
	"database/sql"
)

const BufferCacheQuery string = `
SELECT
   ROUND(
		(SUM(IF(variable_name='Innodb_buffer_pool_read_requests', variable_value,  0))
		/ (SUM(IF(variable_name IN ('Innodb_buffer_pool_read_requests', 'Innodb_buffer_pool_reads'), variable_value, 0)))
	) * 100, 2) AS buffer_pool_hit_ratio_pct
FROM information_schema.GLOBAL_STATUS;
`

// FROM information_schema.GLOBAL_STATUS; // MariaDB
// FROM performance_schema.global_status; // MySQL;

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
	var hitRatio float64
	err := b.DB.QueryRow(BufferCacheQuery).Scan(&hitRatio)
	if err != nil {
		return DatabaseBufferStat{}, err
	}

	stat := DatabaseBufferStat{BufferPoolHitRatio: hitRatio}
	return stat, nil
}
