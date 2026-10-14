package history

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/repository"
	"gorm.io/gorm"
)

const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 100

	dateLayout = "2006-01-02"
)

// ErrInvalidQuery wraps every validation failure so callers (the HTTP
// layer) can distinguish "bad request" from "internal/db error" via
// errors.Is.
var ErrInvalidQuery = errors.New("invalid query")

var allowedServiceTypes = map[string]bool{
	"objectstorage": true,
	"rdbms":         true,
	"nrdbms":        true,
}

// create/delete are literal Operation strings recorded by direct
// ExecTracker.Track calls (e.g. websrc/controllers/objectStorageHandlers.go
// bucket create/delete) that bypass handleTask/models.TaskType.
var allowedOperations = map[string]bool{
	"generate": true,
	"migrate":  true,
	"backup":   true,
	"restore":  true,
	"create":   true,
	"delete":   true,
}

// ExecutionHistoryQuery is the raw (still-string) query the HTTP layer
// parses from request query params.
type ExecutionHistoryQuery struct {
	Page        int
	Size        int
	ServiceType string
	Operation   string
	From        string // date, YYYY-MM-DD, optional (start of that day)
	To          string // date, YYYY-MM-DD, optional (end of that day)
	Sort        string // "asc" | "desc", optional (default: desc)
}

type ExecutionHistoryService struct {
	repo *repository.ExecutionLogRepository
}

func NewExecutionHistoryService(db *gorm.DB) *ExecutionHistoryService {
	return &ExecutionHistoryService{repo: repository.NewExecutionLogRepository(db)}
}

func (s *ExecutionHistoryService) ListExecutionLogs(q ExecutionHistoryQuery) (*models.ExecutionLogListResponse, error) {
	page, size := normalizePaging(q.Page, q.Size)

	if q.ServiceType != "" && !allowedServiceTypes[q.ServiceType] {
		return nil, fmt.Errorf("%w: serviceType %q is not one of objectstorage/rdbms/nrdbms", ErrInvalidQuery, q.ServiceType)
	}
	if q.Operation != "" && !allowedOperations[q.Operation] {
		return nil, fmt.Errorf("%w: operation %q is not one of generate/migrate/backup/restore/create/delete", ErrInvalidQuery, q.Operation)
	}

	sort := strings.ToLower(q.Sort)
	if sort != "" && sort != "asc" && sort != "desc" {
		return nil, fmt.Errorf("%w: sort %q is not one of asc/desc", ErrInvalidQuery, q.Sort)
	}
	desc := sort != "asc" // default desc; only "asc" flips it

	var from, to *time.Time
	if q.From != "" {
		t, err := time.Parse(dateLayout, q.From)
		if err != nil {
			return nil, fmt.Errorf("%w: from must be YYYY-MM-DD: %v", ErrInvalidQuery, err)
		}
		from = &t
	}
	if q.To != "" {
		t, err := time.Parse(dateLayout, q.To)
		if err != nil {
			return nil, fmt.Errorf("%w: to must be YYYY-MM-DD: %v", ErrInvalidQuery, err)
		}
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		to = &endOfDay
	}

	rows, total, err := s.repo.List(repository.ExecutionLogFilter{
		NamespaceID: utils.GetNsId(),
		ServiceType: q.ServiceType,
		Operation:   q.Operation,
		From:        from,
		To:          to,
		Page:        page,
		Size:        size,
		Desc:        desc,
	})
	if err != nil {
		return nil, err
	}

	return &models.ExecutionLogListResponse{
		Content:    rows,
		Page:       page,
		Size:       size,
		TotalCount: total,
		TotalPages: totalPages(total, size),
	}, nil
}

func normalizePaging(page, size int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if size < 1 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size
}

func totalPages(total int64, size int) int {
	if size <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(size)))
}
