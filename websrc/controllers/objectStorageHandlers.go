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

func ObjectstorageBucketsHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "object storage", "get object storage bucket list", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BucketListResponse{
			Buckets: []models.Bucket{},
		})
	}

	var OSC *osc.OSController
	var err error
	OSC, err = auth.GetOS(&params.TargetPoint)
	if err != nil {
		log.Error().Msgf("OSController error importing into objectstorage : %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BucketListResponse{
			Buckets: []models.Bucket{},
		})
	}

	buckets, err := OSC.BucketList()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BucketListResponse{
			Buckets: []models.Bucket{},
		})
	}

	return ctx.JSON(http.StatusOK, models.BucketListResponse{
		Buckets: buckets,
	})
}
