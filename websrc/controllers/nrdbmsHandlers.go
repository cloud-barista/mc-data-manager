package controllers

import (
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	nrdbinstancepkg "github.com/cloud-barista/mc-data-manager/pkg/nrdbinstance"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/service/nrdbinstance"
	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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
//	@Router			/db/nrdbms/tables [post]
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

// NRDBMSListDatabasesHandler godoc
//
//	@ID			NRDBMSListDatabasesHandler
//	@Summary	List databases in a NRDBMS (MongoDB) instance
//	@Description	Connects to the MongoDB server using the target connection info
//	@Description	and returns the names of the databases on the server.
//	@Description	Supported providers: ncp (MongoDB), alibaba (MongoDB).
//	@Tags			[NRDBMS]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask		true	"Target connection info (host, port, username, password)"
//	@Success		200			{array}		string				"Database names"
//	@Failure		500			{object}	map[string]string	"Internal Server Error"
//	@Router			/db/nrdbms/databases [post]
func NRDBMSListDatabasesHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "nrdbms", "list databases in instance", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "invalid request body"})
	}

	NRDBC, err := auth.GetNRDMS(&params.TargetPoint)
	if err != nil {
		logger.Error().Err(err).Msg("NRDBController creation failed")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	dbList, err := NRDBC.ListDatabases()
	if err != nil {
		logger.Error().Err(err).Msg("ListDatabases failed")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	jobEnd(logger, "Successfully listed databases", start)
	return ctx.JSON(http.StatusOK, dbList)
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
//	@Router			/db/nrdbms/tables [put]
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
//	@Router			/db/nrdbms/tables/data [post]
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
//	@Router			/db/nrdbms/tables [delete]
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

// ListNRDBEngineVersionsHandler godoc
//
//	@ID			ListNRDBEngineVersionsHandler
//	@Summary	List available NRDB engine versions
//	@Description	Returns available NoSQL engine versions for the requested CSP and region.
//	@Description	Alibaba: MongoDB versions (4.4, 5.0, 6.0, 7.0). GCP Firestore returns empty.
//	@Tags			[NRDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBEngineVersionsRequest	true	"Provider and region"
//	@Success		200			{array}		models.NRDBEngineVersion			"Available engine versions"
//	@Failure		400			{object}	map[string]string					"Invalid Request"
//	@Failure		500			{object}	map[string]string					"Internal Server Error"
//	@Router			/db/nrdbms/engine-versions [post]
func ListNRDBEngineVersionsHandler(c echo.Context) error {
	var req models.NRDBEngineVersionsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}

	versions, err := nrdbinstance.ListEngineVersions(c.Request().Context(), req.Provider, req.Region)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("region", req.Region).Msg("list NRDB engine versions failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, versions)
}

// ListNRDBInstanceClassesHandler godoc
//
//	@ID			ListNRDBInstanceClassesHandler
//	@Summary	List orderable NRDB instance classes
//	@Description	Returns the instance classes orderable for the given engine version.
//	@Description	Alibaba: MongoDB instance classes (e.g. dds.mongo.mid). GCP Firestore returns empty.
//	@Tags			[NRDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBInstanceClassRequest	true	"Provider, region, engineVersion"
//	@Success		200			{array}		string							"Available instance class names"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/nrdbms/instance-class [post]
func ListNRDBInstanceClassesHandler(c echo.Context) error {
	var req models.NRDBInstanceClassRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}
	if req.EngineVersion == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "engineVersion is required"})
	}

	classes, err := nrdbinstance.ListInstanceClasses(c.Request().Context(), req.Provider, req.Region, req.EngineVersion)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("engineVersion", req.EngineVersion).Msg("list NRDB instance classes failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, classes)
}

// ListNRDBInstancesHandler godoc
//
//	@ID			ListNRDBInstancesHandler
//	@Summary	List NRDB (NoSQL) instances for a given provider
//	@Description	Returns managed NoSQL database instances for the requested CSP and region.
//	@Description	GCP: Firestore Databases. Credentials are resolved by provider.
//	@Tags			[NRDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBInstanceListRequest	true	"Provider and region"
//	@Success		200			{array}		models.NRDBInstance				"List of NRDB instances"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/nrdbms [post]
func ListNRDBInstancesHandler(c echo.Context) error {
	var req models.NRDBInstanceListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}

	instances, err := nrdbinstance.ListInstances(c.Request().Context(), req.Provider, req.Region)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("region", req.Region).Msg("list NRDB instances failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, instances)
}

// CreateNRDBInstanceHandler godoc
//
//	@ID			CreateNRDBInstanceHandler
//	@Summary	Create an NRDB (NoSQL) instance
//	@Description	Provisions a new managed NoSQL database instance for the requested CSP.
//	@Description	GCP: creates a Firestore Database. type defaults to "FIRESTORE_NATIVE" if omitted.
//	@Tags			[NRDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBInstanceCreateRequest	true	"Instance specification"
//	@Success		200			{object}	models.NRDBInstance					"Created instance (status: creating)"
//	@Failure		400			{object}	map[string]string					"Invalid Request"
//	@Failure		500			{object}	map[string]string					"Internal Server Error"
//	@Router			/db/nrdbms [put]
func CreateNRDBInstanceHandler(c echo.Context) error {
	var req models.NRDBInstanceCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	log.Info().
		Str("provider", req.Provider).
		Str("region", req.Region).
		Str("instanceId", req.InstanceID).
		Str("engineVersion", req.EngineVersion).
		Str("instanceClass", req.InstanceClass).
		Int32("allocatedStorage", req.AllocatedStorage).
		Str("masterUsername", req.MasterUsername).
		Msg("CreateNRDBInstance request body")

	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}
	if req.InstanceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "instanceId is required"})
	}

	if req.Provider == "alibaba" || req.Provider == "ncp" {
		missing := map[string]string{
			"engineVersion":  req.EngineVersion,
			"instanceClass":  req.InstanceClass,
			"masterUsername": req.MasterUsername,
			"masterPassword": req.MasterPassword,
		}
		for field, val := range missing {
			if val == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
			}
		}
		if req.AllocatedStorage <= 0 && req.Provider != "ncp" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "allocatedStorage must be greater than 0"})
		}
	}

	spec := nrdbinstancepkg.CreateSpec{
		InstanceID:       req.InstanceID,
		EngineVersion:    req.EngineVersion,
		InstanceClass:    req.InstanceClass,
		AllocatedStorage: req.AllocatedStorage,
		MasterUsername:   req.MasterUsername,
		MasterPassword:   req.MasterPassword,
	}

	var instance models.NRDBInstance
	createErr := task.GetExecTracker().Track(task.ExecMeta{
		TraceID:     traceIDFromCtx(c),
		NamespaceID: utils.GetNsId(),
		ServiceType: "nrdbms",
		Operation:   "create",
		Provider:    req.Provider,
		Region:      req.Region,
		Resource:    req.InstanceID,
		Message:     "create request accepted; async provisioning completion is not tracked",
	}, func() error {
		var err error
		instance, err = nrdbinstance.CreateInstance(c.Request().Context(), req.Provider, req.Region, spec)
		return err
	})
	if createErr != nil {
		log.Error().Err(createErr).Str("provider", req.Provider).Str("instanceId", req.InstanceID).Msg("create NRDB instance failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": createErr.Error()})
	}

	return c.JSON(http.StatusOK, instance)
}

// DeleteNRDBInstanceHandler godoc
//
//	@ID			DeleteNRDBInstanceHandler
//	@Summary	Delete an NRDB (NoSQL) instance
//	@Description	Deletes the NoSQL database instance identified by instanceId.
//	@Tags			[NRDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.NRDBInstanceDeleteRequest	true	"Provider, region, instanceId"
//	@Success		200			{object}	models.NRDBInstance					"Deleted instance (status: deleting)"
//	@Failure		400			{object}	map[string]string					"Invalid Request"
//	@Failure		500			{object}	map[string]string					"Internal Server Error"
//	@Router			/db/nrdbms [delete]
func DeleteNRDBInstanceHandler(c echo.Context) error {
	var req models.NRDBInstanceDeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}
	if req.InstanceID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "instanceId is required"})
	}

	var instance models.NRDBInstance
	deleteErr := task.GetExecTracker().Track(task.ExecMeta{
		TraceID:     traceIDFromCtx(c),
		NamespaceID: utils.GetNsId(),
		ServiceType: "nrdbms",
		Operation:   "delete",
		Provider:    req.Provider,
		Region:      req.Region,
		Resource:    req.InstanceID,
		Message:     "delete request accepted; async completion is not tracked",
	}, func() error {
		var err error
		instance, err = nrdbinstance.DeleteInstance(c.Request().Context(), req.Provider, req.Region, req.InstanceID)
		return err
	})
	if deleteErr != nil {
		log.Error().Err(deleteErr).Str("provider", req.Provider).Str("instanceId", req.InstanceID).Msg("delete NRDB instance failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": deleteErr.Error()})
	}

	return c.JSON(http.StatusOK, instance)
}
