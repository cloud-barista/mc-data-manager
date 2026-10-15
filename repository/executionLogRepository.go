package repository

import (
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ExecutionLogFilter describes the query params accepted by List.
// ServiceType/Operation/From/To are optional; the zero value means "no filter".
// Desc controls the FinishedAt sort direction (true = DESC, false = ASC).
type ExecutionLogFilter struct {
	NamespaceID string
	ServiceType string
	Operation   string
	From        *time.Time
	To          *time.Time
	Page        int
	Size        int
	Desc        bool
}

type ExecutionLogRepository struct {
	db *gorm.DB
}

func NewExecutionLogRepository(db *gorm.DB) *ExecutionLogRepository {
	return &ExecutionLogRepository{db: db}
}

// buildFilterQuery applies the WHERE conditions shared by the count and page
// queries. Period filtering is an interval-overlap check against
// [RequestedAt, FinishedAt], not a FinishedAt-only range: a long-running
// execution that started before `From` but finished inside [From, To] must
// still be included.
func buildFilterQuery(db *gorm.DB, f ExecutionLogFilter) *gorm.DB {
	q := db.Model(&models.ExecutionLog{}).Where("namespace_id = ?", f.NamespaceID)

	if f.ServiceType != "" {
		q = q.Where("service_type = ?", f.ServiceType)
	}
	if f.Operation != "" {
		q = q.Where("operation = ?", f.Operation)
	}
	if f.From != nil {
		q = q.Where("finished_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("requested_at <= ?", *f.To)
	}
	return q
}

// List returns the page of rows matching f, plus the total count matching
// the filter (ignoring Page/Size, for pagination metadata).
func (r *ExecutionLogRepository) List(f ExecutionLogFilter) ([]models.ExecutionLog, int64, error) {
	base := buildFilterQuery(r.db, f)

	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.ExecutionLog
	offset := (f.Page - 1) * f.Size
	if err := base.Session(&gorm.Session{}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "finished_at"}, Desc: f.Desc}).
		Offset(offset).
		Limit(f.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}
