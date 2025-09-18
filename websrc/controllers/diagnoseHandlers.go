package controllers

import (
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/rdbms/mysql/diagnostics"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type DiagnoseHandler struct {
}

func NewDiagnoseHandler() *DiagnoseHandler {
	return &DiagnoseHandler{}
}

func (d *DiagnoseHandler) PostDiagnose(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Diagnose-task", "Diagnose MySQL", start)
	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.DiagnoseResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	rdb, err := auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		errStr := "Invalid request data"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.DiagnoseResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	result, err := rdb.Client.Diagnose()
	if err != nil {
		errStr := "failed to diagnose"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusBadRequest, models.DiagnoseResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	logAll(logger, result)

	return ctx.JSON(http.StatusOK, models.DiagnoseResponse{
		Result:      logstrings.String(),
		Diagnostics: result,
		Error:       nil,
	})
}

func logAll(logger *zerolog.Logger, result diagnostics.TimedResult) {
	logger.Info().Msg(diagnostics.PrintBufferReport(result.Buffer))
	logger.Info().Msg(diagnostics.PrintThreadReport(result.Thread))
	logger.Info().Msg(diagnostics.PrintLockReport(result.Lock, result.Elapsed))
	logger.Info().Msg(diagnostics.PrintIOReport(result.IO, result.Elapsed))
	logger.Info().Msg(diagnostics.PrintWorkloadReport(result.Work, result.Elapsed))
}
