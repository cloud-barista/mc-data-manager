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

// ObjectstorageCreateBucketHandler godoc
//
//	@ID			ObjectstorageCreateBucketHandler
//	@Summary	Create a bucket
//	@Description	Creates a bucket for the given provider. If the bucket already exists, the request is a no-op.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask			true	"Provider credentials, connection info, and bucket name"
//	@Success		200			{object}	models.BasicResponse	"Bucket created successfully"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/objectstorage/bucket [put]
func ObjectstorageCreateBucketHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "object storage", "create bucket", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	connName := fmt.Sprintf("%s-%s", params.TargetPoint.Provider, params.TargetPoint.Region)
	nsId := utils.GetNsId()
	bucket := params.TargetPoint.Bucket

	headPath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s", nsId, bucket)
	if _, err := utils.RequestTumblebug(headPath, http.MethodHead, connName, nil); err == nil {
		jobEnd(logger, "Bucket already exists: "+bucket, start)
		return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	createBody := []byte(fmt.Sprintf(`{"bucketName":%q,"connectionName":%q}`, bucket, connName))
	createPath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage", nsId)
	if _, err := utils.RequestTumblebug(createPath, http.MethodPut, connName, createBody); err != nil {
		log.Error().Msgf("CreateBucket error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully created bucket: "+bucket, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// ObjectstorageDeleteBucketHandler godoc
//
//	@ID			ObjectstorageDeleteBucketHandler
//	@Summary	Delete a bucket and all its objects
//	@Description	Empties the bucket by deleting all objects (in batches of 1000), then deletes the bucket itself.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask			true	"Provider credentials, connection info, and bucket name"
//	@Success		200			{object}	models.BasicResponse	"Bucket deleted successfully"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/objectstorage/bucket [delete]
func ObjectstorageDeleteBucketHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit(ctx, "object storage", "delete bucket", start)

	params := models.DataTask{}
	if !getDataWithReBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	connName := fmt.Sprintf("%s-%s", params.TargetPoint.Provider, params.TargetPoint.Region)
	nsId := utils.GetNsId()
	bucket := params.TargetPoint.Bucket

	// 1. 오브젝트 목록 조회
	listPath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s", nsId, bucket)
	body, err := utils.RequestTumblebug(listPath, http.MethodGet, connName, nil)
	if err != nil {
		log.Error().Msgf("ObjectList error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	var osResp models.ObjectStorage
	if err := json.Unmarshal(body, &osResp); err != nil {
		log.Error().Msgf("Unmarshal error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	// 2. 오브젝트 일괄 삭제 (1000개 단위)
	const batchSize = 1000
	deletePath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s?delete=true", nsId, bucket)
	keys := make([]string, 0, len(osResp.Contents))
	for _, o := range osResp.Contents {
		keys = append(keys, o.Key)
	}
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		deleteReq := models.DeleteRequest{
			XMLNS: "http://s3.amazonaws.com/doc/2006-03-01/",
		}
		for _, k := range keys[i:end] {
			deleteReq.Objects = append(deleteReq.Objects, models.S3Object{Key: k})
		}
		xmlOutput, err := xml.MarshalIndent(deleteReq, "", "    ")
		if err != nil {
			log.Error().Msgf("XML marshal error: %v", err)
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
		}
		if _, err := utils.RequestTumblebug(deletePath, http.MethodPost, connName, []byte(xml.Header+string(xmlOutput))); err != nil {
			log.Error().Msgf("DeleteObjects error: %v", err)
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
		}
	}

	// 3. 버킷 삭제
	bucketPath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s", nsId, bucket)
	if _, err := utils.RequestTumblebug(bucketPath, http.MethodDelete, connName, nil); err != nil {
		log.Error().Msgf("DeleteBucket error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully deleted bucket: "+bucket, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
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
//	@Router			/objectstorage/buckets/objects [post]
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
//	@Router			/objectstorage/buckets/object [delete]
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
