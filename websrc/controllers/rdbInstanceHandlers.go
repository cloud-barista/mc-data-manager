package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	rdbinstancepkg "github.com/cloud-barista/mc-data-manager/pkg/rdbinstance"
	"github.com/cloud-barista/mc-data-manager/service/rdbinstance"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ListRDBInstancesHandler godoc
//
//	@ID			ListRDBInstancesHandler
//	@Summary	List RDB (database) instances for a given provider
//	@Description	Returns managed database instances for the requested CSP and region.
//	@Description	Credentials are resolved by provider (one credential per CSP). Only AWS is supported for now.
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.RDBInstanceListRequest	true	"Provider and region"
//	@Success		200			{array}		models.DBInstance				"List of database instances"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/rdbms [post]
func ListRDBInstancesHandler(c echo.Context) error {
	var req models.RDBInstanceListRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}

	instances, err := rdbinstance.ListInstances(c.Request().Context(), req.Provider, req.Region)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("region", req.Region).Msg("list RDB instances failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, instances)
}

// supportedCreateEngines is the whitelist of DB engines allowed for creation.
var supportedCreateEngines = map[string]bool{
	"mysql":   true,
	"mariadb": true,
}

// CreateRDBInstanceHandler godoc
//
//	@ID			CreateRDBInstanceHandler
//	@Summary	Create an RDB (database) instance
//	@Description	Provisions a new managed database instance for the requested CSP.
//	@Description	Only AWS with mysql/mariadb engines is supported. The instance is created publicly accessible.
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.RDBInstanceCreateRequest	true	"Instance specification"
//	@Success		200			{object}	models.DBInstance				"Created instance (status: creating)"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/rdbms [put]
func CreateRDBInstanceHandler(c echo.Context) error {
	var req models.RDBInstanceCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	missing := map[string]string{
		"provider":       req.Provider,
		"region":         req.Region,
		"instanceId":     req.InstanceID,
		"instanceClass":  req.InstanceClass,
		"engine":         req.Engine,
		"engineVersion":  req.EngineVersion,
		"masterUsername": req.MasterUsername,
		"masterPassword": req.MasterPassword,
	}
	for field, val := range missing {
		if val == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": field + " is required"})
		}
	}
	if req.AllocatedStorage <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "allocatedStorage must be greater than 0"})
	}
	if !supportedCreateEngines[strings.ToLower(req.Engine)] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported engine: only mysql and mariadb are allowed"})
	}

	spec := rdbinstancepkg.CreateSpec{
		InstanceID:       req.InstanceID,
		InstanceClass:    req.InstanceClass,
		Engine:           strings.ToLower(req.Engine),
		EngineVersion:    req.EngineVersion,
		MasterUsername:   req.MasterUsername,
		MasterPassword:   req.MasterPassword,
		AllocatedStorage: req.AllocatedStorage,
	}

	instance, err := rdbinstance.CreateInstance(c.Request().Context(), req.Provider, req.Region, spec)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("instanceId", req.InstanceID).Msg("create RDB instance failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, instance)
}

// DeleteRDBInstanceHandler godoc
//
//	@ID			DeleteRDBInstanceHandler
//	@Summary	Delete an RDB (database) instance
//	@Description	Deletes the database instance identified by instanceId. The final snapshot is skipped.
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.RDBInstanceDeleteRequest	true	"Provider, region, instanceId"
//	@Success		200			{object}	models.DBInstance				"Deleted instance (status: deleting)"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/rdbms [delete]
func DeleteRDBInstanceHandler(c echo.Context) error {
	var req models.RDBInstanceDeleteRequest
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

	instance, err := rdbinstance.DeleteInstance(c.Request().Context(), req.Provider, req.Region, req.InstanceID)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("instanceId", req.InstanceID).Msg("delete RDB instance failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, instance)
}

// ListRDBEngineVersionsHandler godoc
//
//	@ID			ListRDBEngineVersionsHandler
//	@Summary	List available RDB engine versions
//	@Description	Returns available mysql and mariadb engine versions for the requested CSP and region.
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.RDBEngineVersionsRequest	true	"Provider and region"
//	@Success		200			{array}		models.DBEngineVersion			"Available engine versions"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/rdbms/engine-versions [post]
func ListRDBEngineVersionsHandler(c echo.Context) error {
	var req models.RDBEngineVersionsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}

	versions, err := rdbinstance.ListEngineVersions(c.Request().Context(), req.Provider, req.Region)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("region", req.Region).Msg("list engine versions failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, versions)
}

// ListRDBInstanceClassesHandler godoc
//
//	@ID			ListRDBInstanceClassesHandler
//	@Summary	List orderable RDB instance classes
//	@Description	Returns the instance classes orderable for the given engine and version.
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.RDBInstanceClassRequest	true	"Provider, region, engine, engineVersion"
//	@Success		200			{array}		string							"Available instance class names"
//	@Failure		400			{object}	map[string]string				"Invalid Request"
//	@Failure		500			{object}	map[string]string				"Internal Server Error"
//	@Router			/db/rdbms/instance-class [post]
func ListRDBInstanceClassesHandler(c echo.Context) error {
	var req models.RDBInstanceClassRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Provider == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "provider is required"})
	}
	if req.Region == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "region is required"})
	}
	if req.Engine == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "engine is required"})
	}
	if !supportedCreateEngines[strings.ToLower(req.Engine)] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "unsupported engine: only mysql and mariadb are allowed"})
	}
	if req.EngineVersion == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "engineVersion is required"})
	}

	classes, err := rdbinstance.ListInstanceClasses(c.Request().Context(), req.Provider, req.Region, strings.ToLower(req.Engine), req.EngineVersion)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Str("engine", req.Engine).Str("engineVersion", req.EngineVersion).Msg("list instance classes failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, classes)
}

// ListRDBDatabasesHandler godoc
//
//	@ID			ListRDBDatabasesHandler
//	@Summary	List databases inside an RDB instance
//	@Description	Connects directly to the database instance using the target connection
//	@Description	info and returns the names of the databases it contains (SHOW DATABASES).
//	@Tags			[RDB Instance]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask		true	"Target connection info (host, port, username, password)"
//	@Success		200			{array}		string				"Database names"
//	@Failure		500			{object}	map[string]string	"Internal Server Error"
//	@Router			/db/rdbms/databases [post]
func ListRDBDatabasesHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "rdb database", "list databases in instance", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "invalid request body"})
	}

	rdbCtrl, err := auth.GetRDMS(&params.TargetPoint)
	if err != nil {
		log.Error().Err(err).Msg("GetRDMS error listing databases")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	dbList := []string{}
	if err := rdbCtrl.ListDB(&dbList); err != nil {
		log.Error().Err(err).Msg("ListDB error")
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, dbList)
}
