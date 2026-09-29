/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cloud-barista/mc-data-manager/config"
	"github.com/cloud-barista/mc-data-manager/models"
	service "github.com/cloud-barista/mc-data-manager/service/history"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ExecutionLogHandler struct {
	historyService *service.ExecutionHistoryService
}

func NewExecutionLogHandler(db *gorm.DB) *ExecutionLogHandler {
	return &ExecutionLogHandler{historyService: service.NewExecutionHistoryService(db)}
}

// ListExecutionLogsHandler godoc
//
//	@ID				ListExecutionLogsHandler
//	@Summary		List execution history
//	@Description	Retrieve completed execution history (tbexecutionlog) for the current namespace. In-progress transactions are not included. Supports period (interval-overlap), serviceType, and operation filters, paginated and sorted by FinishedAt descending.
//	@Tags			[History]
//	@Produce		json
//	@Param			page		query	int		false	"Page number (default 1)"
//	@Param			size		query	int		false	"Page size (default 20, max 100)"
//	@Param			serviceType	query	string	false	"objectstorage | rdbms | nrdbms"
//	@Param			operation	query	string	false	"generate | migrate | backup | restore | create | delete"
//	@Param			from		query	string	false	"Period start date (YYYY-MM-DD)"
//	@Param			to			query	string	false	"Period end date (YYYY-MM-DD)"
//	@Param			sort		query	string	false	"asc | desc, sort by FinishedAt (default desc)"
//	@Success		200			{object}	models.ExecutionLogListResponse
//	@Failure		400			{object}	models.BasicResponse	"Invalid query parameter"
//	@Failure		503			{object}	models.BasicResponse	"Execution history DB is not configured"
//	@Router			/history [get]
func (h *ExecutionLogHandler) ListExecutionLogsHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "history", "list execution history", start)

	if config.DB == nil {
		return ctx.JSON(http.StatusServiceUnavailable, models.BasicResponse{
			Result: "execution history DB is not configured",
			Error:  nil,
		})
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	size, _ := strconv.Atoi(ctx.QueryParam("size"))

	query := service.ExecutionHistoryQuery{
		Page:        page,
		Size:        size,
		ServiceType: ctx.QueryParam("serviceType"),
		Operation:   ctx.QueryParam("operation"),
		From:        ctx.QueryParam("from"),
		To:          ctx.QueryParam("to"),
		Sort:        ctx.QueryParam("sort"),
	}

	resp, err := h.historyService.ListExecutionLogs(query)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidQuery) {
			status = http.StatusBadRequest
		}
		logstrings.WriteString(err.Error())
		return ctx.JSON(status, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully retrieved execution history", start)
	return ctx.JSON(http.StatusOK, resp)
}
