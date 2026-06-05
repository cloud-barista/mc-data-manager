package controllers

import (
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// ObjectstorageBucketsHandler godoc
//
//	@ID			ObjectstorageBucketsHandler
//	@Summary	List available buckets for a given provider
//	@Description	Returns the list of buckets accessible with the given credentials. Optionally filters by a tag key/value pair.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			filterKey	query		string					false	"Tag key to filter buckets by"
//	@Param			filterVal	query		string					false	"Tag value to filter buckets by (used with filterKey)"
//	@Param			RequestBody	body		models.DataTask			true	"Provider credentials and connection info"
//	@Success		200			{object}	models.ObjectStorageListResponse	"List of accessible buckets"
//	@Failure		500			{object}	models.ObjectStorageListResponse	"Internal Server Error"
//	@Router			/buckets [post]
func ObjectstorageBucketsHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "object storage", "get object storage bucket list", start)

	filterKey := ctx.QueryParam("filterKey")
	filterVal := ctx.QueryParam("filterVal")

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.ObjectStorageListResponse{
			ObjectStorage: []models.ObjectStorage{},
		})
	}

	var OSC *osc.OSController
	var err error
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.ObjectStorageListResponse{
			ObjectStorage: []models.ObjectStorage{},
		})
	}

	objectStorages, err := OSC.BucketList(filterKey, filterVal)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.ObjectStorageListResponse{
			ObjectStorage: []models.ObjectStorage{},
		})
	}

	return ctx.JSON(http.StatusOK, models.ObjectStorageListResponse{
		ObjectStorage: objectStorages,
	})
}
