package controllers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
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
//	@Router			/objectstorage/buckets [post]
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

// ObjectstorageObjectListHandler godoc
//
//	@ID			ObjectstorageObjectListHandler
//	@Summary	List objects in a bucket
//	@Description	Returns all objects stored in the bucket specified by the target connection. Supports optional filter parameters.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask				true	"Provider credentials, connection info, and optional sourceFilter"
//	@Success		200			{object}	models.ObjectListResponse	"List of objects in the bucket"
//	@Failure		500			{object}	models.ObjectListResponse	"Internal Server Error"
//	@Router			/objectstorage/objects [post]
func ObjectstorageObjectListHandler(ctx echo.Context) error {
	start := time.Now()

	logger, _ := pageLogInit(ctx, "object storage", "list objects in bucket", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.ObjectListResponse{Objects: []*models.ObjectInfo{}})
	}

	connName := fmt.Sprintf("%s-%s", params.TargetPoint.Provider, params.TargetPoint.Region)
	nsId := utils.GetNsId()
	path := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s", nsId, params.TargetPoint.Bucket)

	body, err := utils.RequestTumblebug(path, http.MethodGet, connName, nil)
	if err != nil {
		log.Error().Msgf("RequestTumblebug error listing objects: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.ObjectListResponse{Objects: []*models.ObjectInfo{}})
	}

	var osResp models.ObjectStorage
	if err := json.Unmarshal(body, &osResp); err != nil {
		log.Error().Msgf("Unmarshal error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.ObjectListResponse{Objects: []*models.ObjectInfo{}})
	}

	var flt *filtering.ObjectFilter
	if params.SourceFilter != nil {
		flt, err = filtering.FromParams(params.SourceFilter)
		if err != nil {
			log.Error().Msgf("ObjectFilter parse error: %v", err)
			return ctx.JSON(http.StatusInternalServerError, models.ObjectListResponse{Objects: []*models.ObjectInfo{}})
		}
	}

	result := make([]*models.ObjectInfo, 0, len(osResp.Contents))
	for _, o := range osResp.Contents {
		c := filtering.Candidate{Key: o.Key, Size: o.Size, LastModified: o.LastModified}
		if filtering.MatchCandidate(flt, c) {
			result = append(result, &models.ObjectInfo{
				Key:          o.Key,
				Size:         o.Size,
				LastModified: o.LastModified,
				ETag:         o.ETag,
				StorageClass: o.StorageClass,
			})
		}
	}

	jobEnd(logger, fmt.Sprintf("Listed %d objects", len(result)), start)
	return ctx.JSON(http.StatusOK, models.ObjectListResponse{Objects: result})
}

// ObjectstorageDeleteObjectHandler godoc
//
//	@ID			ObjectstorageDeleteObjectHandler
//	@Summary	Delete a single object from a bucket
//	@Description	Deletes the object identified by objectKey from the bucket specified by the target connection.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.ObjectDeleteRequest	true	"Target connection info and key of the object to delete"
//	@Success		200			{object}	models.BasicResponse		"Object deleted successfully"
//	@Failure		400			{object}	models.BasicResponse		"Bad Request — objectKey is empty"
//	@Failure		500			{object}	models.BasicResponse		"Internal Server Error"
//	@Router			/objectstorage/object [delete]
func ObjectstorageDeleteObjectHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "object storage", "delete object", start)

	params := models.ObjectDeleteRequest{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	if params.ObjectKey == "" {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{Result: "objectKey is required", Error: nil})
	}

	deleteReq := models.DeleteRequest{
		XMLNS:   "http://s3.amazonaws.com/doc/2006-03-01/",
		Objects: []models.S3Object{{Key: params.ObjectKey}},
	}
	xmlOutput, err := xml.MarshalIndent(deleteReq, "", "    ")
	if err != nil {
		log.Error().Msgf("XML marshal error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	connName := fmt.Sprintf("%s-%s", params.TargetPoint.Provider, params.TargetPoint.Region)
	nsId := utils.GetNsId()
	path := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s?delete=true", nsId, params.TargetPoint.Bucket)

	if _, err := utils.RequestTumblebug(path, http.MethodPost, connName, []byte(xml.Header+string(xmlOutput))); err != nil {
		log.Error().Msgf("DeleteObject error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully deleted object: "+params.ObjectKey, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}
