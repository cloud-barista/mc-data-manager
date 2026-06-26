package controllers

import (
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/labstack/echo/v4"
)

// NRDBMSListTablesHandler godoc
//
//	@ID			NRDBMSListTablesHandler
//	@Summary	List tables in a NRDBMS instance
//	@Description	Returns the list of tables (collections) accessible with the given credentials.
//	@Description	Supported providers: aws (DynamoDB), gcp (Firestore), ncp (MongoDB), alibaba (MongoDB).
//	@Tags			[NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask				true	"Provider credentials and connection info"
//	@Success		200			{object}	models.NRDBTableListResponse	"List of tables"
//	@Failure		500			{object}	models.NRDBTableListResponse	"Internal Server Error"
//	@Router			/db/nrdbms [post]
func NRDBMSListTablesHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "nrdbms", "list NRDBMS tables", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableListResponse{Tables: []string{}})
	}

	NRDBC, err := auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logger.Error().Err(err).Msg("NRDBController creation failed")
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableListResponse{Tables: []string{}})
	}

	tables, err := NRDBC.ListTables()
	if err != nil {
		logger.Error().Err(err).Msg("ListTables failed")
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableListResponse{Tables: []string{}})
	}

	jobEnd(logger, "Successfully listed tables", start)
	return ctx.JSON(http.StatusOK, models.NRDBTableListResponse{Tables: tables})
}

// NRDBMSCreateTableHandler godoc
//
//	@ID			NRDBMSCreateTableHandler
//	@Summary	Create a table in a NRDBMS instance
//	@Description	Creates a table (collection) with the given name. If the table already exists the request is a no-op.
//	@Description	Supported providers: aws (DynamoDB), gcp (Firestore), ncp (MongoDB), alibaba (MongoDB).
//	@Tags			[NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBTableRequest	true	"Provider credentials, connection info, and table name"
//	@Success		200			{object}	models.BasicResponse	"Table created successfully"
//	@Failure		400			{object}	models.BasicResponse	"Bad Request — tableName is empty"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/db/nrdbms [put]
func NRDBMSCreateTableHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "nrdbms", "create NRDBMS table", start)

	params := models.NRDBTableRequest{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	if params.TableName == "" {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{Result: "tableName is required", Error: nil})
	}

	NRDBC, err := auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logger.Error().Err(err).Msg("NRDBController creation failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	if err := NRDBC.CreateTable(params.TableName); err != nil {
		logger.Error().Err(err).Msgf("CreateTable failed: %s", params.TableName)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully created table: "+params.TableName, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// NRDBMSGetTableHandler godoc
//
//	@ID			NRDBMSGetTableHandler
//	@Summary	Export data from a NRDBMS table
//	@Description	Retrieves all items from the specified table (collection).
//	@Description	Supported providers: aws (DynamoDB), gcp (Firestore), ncp (MongoDB), alibaba (MongoDB).
//	@Tags			[NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBTableRequest		true	"Provider credentials, connection info, and table name"
//	@Success		200			{object}	models.NRDBTableGetResponse	"Table data exported successfully"
//	@Failure		400			{object}	models.NRDBTableGetResponse	"Bad Request — tableName is empty"
//	@Failure		500			{object}	models.NRDBTableGetResponse	"Internal Server Error"
//	@Router			/db/nrdbms/data [post]
func NRDBMSGetTableHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "nrdbms", "get NRDBMS table data", start)

	params := models.NRDBTableRequest{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableGetResponse{Data: []map[string]interface{}{}})
	}

	if params.TableName == "" {
		return ctx.JSON(http.StatusBadRequest, models.NRDBTableGetResponse{Data: []map[string]interface{}{}})
	}

	NRDBC, err := auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logger.Error().Err(err).Msg("NRDBController creation failed")
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableGetResponse{Data: []map[string]interface{}{}})
	}

	data := []map[string]interface{}{}
	if err := NRDBC.Get(params.TableName, &data); err != nil {
		logger.Error().Err(err).Msgf("Get table data failed: %s", params.TableName)
		return ctx.JSON(http.StatusInternalServerError, models.NRDBTableGetResponse{Data: []map[string]interface{}{}})
	}

	jobEnd(logger, "Successfully exported table data: "+params.TableName, start)
	return ctx.JSON(http.StatusOK, models.NRDBTableGetResponse{Data: data})
}

// NRDBMSDeleteTableHandler godoc
//
//	@ID			NRDBMSDeleteTableHandler
//	@Summary	Delete a table from a NRDBMS instance
//	@Description	Deletes the table (collection) with the given name and all its data.
//	@Description	Supported providers: aws (DynamoDB), gcp (Firestore), ncp (MongoDB), alibaba (MongoDB).
//	@Tags			[NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBTableRequest	true	"Provider credentials, connection info, and table name"
//	@Success		200			{object}	models.BasicResponse	"Table deleted successfully"
//	@Failure		400			{object}	models.BasicResponse	"Bad Request — tableName is empty"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/db/nrdbms [delete]
func NRDBMSDeleteTableHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "nrdbms", "delete NRDBMS table", start)

	params := models.NRDBTableRequest{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	if params.TableName == "" {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{Result: "tableName is required", Error: nil})
	}

	NRDBC, err := auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logger.Error().Err(err).Msg("NRDBController creation failed")
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	if err := NRDBC.DeleteTables(params.TableName); err != nil {
		logger.Error().Err(err).Msgf("DeleteTable failed: %s", params.TableName)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully deleted table: "+params.TableName, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}
