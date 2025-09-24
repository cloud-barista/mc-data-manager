package diagnostics

import (
	"context"
	"database/sql"
	"time"
)

const IOQueryBySchema string = `
SELECT 
	OBJECT_SCHEMA, 
	OBJECT_NAME,
	SUM(COUNT_READ + COUNT_FETCH) AS rows_read,
	SUM(COUNT_INSERT) AS rows_inserted,
	SUM(COUNT_UPDATE) AS rows_updated,
	SUM(COUNT_DELETE) AS rows_deleted
 FROM performance_schema.table_io_waits_summary_by_table
WHERE OBJECT_SCHEMA = ?
GROUP BY OBJECT_SCHEMA, OBJECT_NAME;
`

const IOQueryAllUserSchemas string = `
SELECT 
	OBJECT_SCHEMA, 
	OBJECT_NAME,
	SUM(COUNT_READ + COUNT_FETCH) AS rows_read,
	SUM(COUNT_INSERT) AS rows_inserted,
	SUM(COUNT_UPDATE) AS rows_updated,
	SUM(COUNT_DELETE) AS rows_deleted
 FROM performance_schema.table_io_waits_summary_by_table
WHERE OBJECT_SCHEMA NOT IN ('performance_schema', 'mysql')
GROUP BY OBJECT_SCHEMA, OBJECT_NAME;
`

type TableIOStat struct {
	Schema       string
	Table        string
	RowsRead     uint64
	RowsInserted uint64
	RowsUpdated  uint64
	RowsDeleted  uint64
}

type TableIODelta struct {
	Schema       string
	Table        string
	RowsRead     uint64
	RowsInserted uint64
	RowsUpdated  uint64
	RowsDeleted  uint64
	Elapsed      time.Duration
}

type TableIOCollector struct {
	DB *sql.DB
}

func NewTableIOCollector(db *sql.DB) *TableIOCollector {
	return &TableIOCollector{
		DB: db,
	}
}

func (c *TableIOCollector) Snapshot(ctx context.Context, schema string) (Snapshot[TableIOStat], error) {
	var rows *sql.Rows
	var err error
	if schema == "" {
		rows, err = c.DB.QueryContext(ctx, IOQueryAllUserSchemas)
	} else {
		rows, err = c.DB.QueryContext(ctx, IOQueryBySchema, schema)
	}
	if err != nil {
		return Snapshot[TableIOStat]{}, err
	}
	defer rows.Close()

	snap := Snapshot[TableIOStat]{
		TakenAt: time.Now(),
		Items:   make(map[string]TableIOStat),
	}
	for rows.Next() {
		var s TableIOStat
		if err := rows.Scan(&s.Schema, &s.Table, &s.RowsRead, &s.RowsInserted, &s.RowsUpdated, &s.RowsDeleted); err != nil {
			return Snapshot[TableIOStat]{}, err
		}
		snap.Items[key(s.Schema, s.Table)] = s
	}
	return snap, rows.Err()
}

func DiffIO(before, after Snapshot[TableIOStat]) ([]TableIODelta, time.Duration) {
	elapsed := after.TakenAt.Sub(before.TakenAt)
	out := make([]TableIODelta, 0, len(after.Items))
	for k, a := range after.Items {
		b, ok := before.Items[k]
		var rd, ins, upd, del uint64
		if ok {
			if a.RowsRead >= b.RowsRead {
				rd = a.RowsRead - b.RowsRead
			}
			if a.RowsInserted >= b.RowsInserted {
				ins = a.RowsInserted - b.RowsInserted
			}
			if a.RowsUpdated >= b.RowsUpdated {
				upd = a.RowsUpdated - b.RowsUpdated
			}
			if a.RowsDeleted >= b.RowsDeleted {
				del = a.RowsDeleted - b.RowsDeleted
			}
		} else {
			rd, ins, upd, del = a.RowsRead, a.RowsInserted, a.RowsUpdated, a.RowsDeleted
		}
		out = append(out, TableIODelta{
			Schema:       a.Schema,
			Table:        a.Table,
			RowsRead:     rd,
			RowsInserted: ins,
			RowsUpdated:  upd,
			RowsDeleted:  del,
			Elapsed:      elapsed,
		})
	}
	return out, elapsed
}
