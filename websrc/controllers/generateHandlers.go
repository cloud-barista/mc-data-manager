package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-barista/cm-data-mold/websrc/models"
	"github.com/labstack/echo/v4"
)

type GenerateLinuxPostHandlerResponseBody struct {
	models.BasicResponse
}

// GenerateLinuxPostHandler godoc
// @Summary Generate test data on on-premise Linux
// @Description Generate test data on on-premise Linux.
// @Tags [On-premise] Test Data Generation
// @Accept  json
// @Produce  json
// @Param RequestBody body GenDataParams true "Parameters required to generate test data"
// @Param CredentialGCP formData file true "Parameters required to generate test data"
// @Success 200 {object} GenerateLinuxPostHandlerResponseBody "Successfully generated test data"
// @Failure 400 {object} GenerateLinuxPostHandlerResponseBody "Invalid Request"
// @Failure 500 {object} GenerateLinuxPostHandlerResponseBody "Internal Server Error"
// @Router /generate/linux [post]
func GenerateLinuxPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genlinux", "Create dummy data in linux", start)

	if !osCheck(logger, start, "linux") {
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	params := getData("gen", ctx).(GenDataParams)

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully creating a dummy with Linux", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateWindowsPostHandler(ctx echo.Context) error {

	start := time.Now()

	logger, logstrings := pageLogInit("genwindows", "Create dummy data in windows", start)

	if !osCheck(logger, start, "windows") {
		fmt.Println("test")
		return ctx.JSON(http.StatusBadRequest, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	params := getData("gen", ctx).(GenDataParams)

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	jobEnd(logger, "Successfully creating a dummy with Windows", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateS3PostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genS3", "Create dummy data and import to s3", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	awsOSC := getS3OSC(logger, start, "gen", params)
	if awsOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "s3", awsOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	jobEnd(logger, "Dummy creation and import successful with s3", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateGCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genGCP", "Create dummy data and import to gcp", start)

	params := getData("gen", ctx).(GenDataParams)

	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)
	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	gcpOSC := getGCPCOSC(logger, start, "gen", params, credFileName)
	if gcpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "gcp", gcpOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with gcp", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateNCPPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genNCP", "Create dummy data and import to ncp objectstorage", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpOSC := getS3COSC(logger, start, "gen", params)
	if ncpOSC == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	if !oscImport(logger, start, "ncp", ncpOSC, params.DummyPath) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Create dummy data and import to ncp objectstorage", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateMySQLPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genmysql", "Create dummy data and import to mysql", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerSQL = "on"
	params.SizeServerSQL = "5"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	rdbc := getMysqlRDBC(logger, start, "gen", params)
	if rdbc == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	sqlList := []string{}
	if !walk(logger, start, &sqlList, params.DummyPath, ".sql") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	logger.Info("Start Import with mysql")
	for _, sql := range sqlList {
		logger.Infof("Read sql file : %s", sql)
		data, err := os.ReadFile(sql)
		if err != nil {
			end := time.Now()
			logger.Errorf("os ReadFile failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})

		}

		logger.Infof("Put start : %s", filepath.Base(sql))
		if err := rdbc.Put(string(data)); err != nil {
			end := time.Now()
			logger.Errorf("RDBController import failed : %v", err)
			logger.Infof("end time : %s", end.Format("2006-01-02T15:04:05-07:00"))
			logger.Infof("Elapsed time : %s", end.Sub(start).String())
			return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
				Result: logstrings.String(),
				Error:  nil,
			})
		}
		logger.Infof("sql put success : %s", filepath.Base(sql))
	}

	jobEnd(logger, "Dummy creation and import successful with mysql", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateDynamoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("gendynamodb", "Create dummy data and import to dynamoDB", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	awsNRDB := getDynamoNRDBC(logger, start, "gen", params)
	if awsNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "DynamoDB", awsNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with dynamoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateFirestorePostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genfirestore", "Create dummy data and import to firestoreDB", start)

	params := GenDataParams{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credTmpDir, credFileName, ok := gcpCreateCredFile(logger, start, ctx)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(credTmpDir)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	gcpNRDB := getFirestoreNRDBC(logger, start, "gen", params, credFileName)
	if gcpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "FirestoreDB", gcpNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with firestoreDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

func GenerateMongoDBPostHandler(ctx echo.Context) error {
	start := time.Now()

	logger, logstrings := pageLogInit("genmongodb", "Create dummy data and import to mongoDB", start)

	params := getData("gen", ctx).(GenDataParams)

	tmpDir, ok := createDummyTemp(logger, start)
	if !ok {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}
	defer os.RemoveAll(tmpDir)

	params.DummyPath = tmpDir
	params.CheckServerJSON = "on"
	params.SizeServerJSON = "1"

	if !dummyCreate(logger, start, params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jsonList := []string{}
	if !walk(logger, start, &jsonList, params.DummyPath, ".json") {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})

	}

	ncpNRDB := getMongoNRDBC(logger, start, "gen", params)
	if ncpNRDB == nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	if !nrdbPutWorker(logger, start, "MongoDB", ncpNRDB, jsonList) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Dummy creation and import successful with mongoDB", start)
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}
