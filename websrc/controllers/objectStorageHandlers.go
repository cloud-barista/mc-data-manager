package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cloud-barista/mc-data-manager/internal/auth"
	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/cloud-barista/mc-data-manager/pkg/objectstorage/filtering"
	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/cloud-barista/mc-data-manager/service/osc"
	"github.com/cloud-barista/mc-data-manager/service/task"
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
		log.Error().Msgf("BucketList error (nsId=%s, provider=%s): %v", utils.GetNsId(), params.TargetPoint.Provider, err)
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
//	@Router			/objectstorage/buckets [put]
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
	// direct 경로(handleTask 우회) 이므로 실행 이력을 tracker 로 명시적 기록
	createErr := task.GetExecTracker().Track(task.ExecMeta{
		TraceID:     traceIDFromCtx(ctx),
		NamespaceID: nsId,
		ServiceType: "objectstorage",
		Operation:   "create",
		Provider:    params.TargetPoint.Provider,
		Region:      params.TargetPoint.Region,
		Resource:    bucket,
	}, func() error {
		_, e := utils.RequestTumblebug(createPath, http.MethodPut, connName, createBody)
		return e
	})
	if createErr != nil {
		log.Error().Msgf("CreateBucket error: %v", createErr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully created bucket: "+bucket, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}

// ObjectstorageDeleteBucketHandler godoc
//
//	@ID			ObjectstorageDeleteBucketHandler
//	@Summary	Delete a bucket and all its objects
//	@Description	Deletes the bucket and all its contents via Tumblebug with option=force.
//	@Tags			[ObjectStorage]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		models.DataTask			true	"Provider credentials, connection info, and bucket name"
//	@Success		200			{object}	models.BasicResponse	"Bucket deleted successfully"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/objectstorage/buckets [delete]
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

	// option=force: 내용물 포함 강제 삭제 (Spider force=true).
	// option=empty는 오브젝트만 비우고 CSP 버킷 삭제가 완료되지 않은 채 204를 반환하는 사례가 있어 사용하지 않음.
	bucketPath := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s?option=force", nsId, bucket)
	// direct 경로(handleTask 우회) 이므로 실행 이력을 tracker 로 명시적 기록
	delErr := task.GetExecTracker().Track(task.ExecMeta{
		TraceID:     traceIDFromCtx(ctx),
		NamespaceID: nsId,
		ServiceType: "objectstorage",
		Operation:   "delete",
		Provider:    params.TargetPoint.Provider,
		Region:      params.TargetPoint.Region,
		Resource:    bucket,
	}, func() error {
		_, e := utils.RequestTumblebug(bucketPath, http.MethodDelete, connName, nil)
		return e
	})
	if delErr != nil {
		log.Error().Msgf("DeleteBucket error: %v", delErr)
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

	connName := fmt.Sprintf("%s-%s", params.TargetPoint.Provider, params.TargetPoint.Region)
	nsId := utils.GetNsId()
	// objectKey는 슬래시를 포함하므로 반드시 %2F로 인코딩해야 Tumblebug 라우트에 매칭됨
	path := fmt.Sprintf("/tumblebug/ns/%s/resources/objectStorage/%s/object/%s",
		nsId, params.TargetPoint.Bucket, url.PathEscape(params.ObjectKey))

	if _, err := utils.RequestTumblebug(path, http.MethodDelete, connName, nil); err != nil {
		log.Error().Msgf("DeleteObject error: %v", err)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{Result: logstrings.String(), Error: nil})
	}

	jobEnd(logger, "Successfully deleted object: "+params.ObjectKey, start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{Result: logstrings.String(), Error: nil})
}
