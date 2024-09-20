/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/labstack/echo/v4"
)

func MainGetHandler(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "main",
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to Dashboard data

func DashBoardHandler(ctx echo.Context) error {
	logger := getLogger("dashboard")
	logger.Info().Msg("dashboard get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "dashboard",
		AWSRegions: GetAWSRegions(),
		GCPRegions: GetGCPRegions(),
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		Error:      nil,
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to generate data

func GenerateLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("genlinux")
	logger.Info().Msg("genlinux get page accessed")

	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-Linux",
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func GenerateWindowsGetHandler(ctx echo.Context) error {

	// tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("genwindows")
	logger.Info().Msg("genwindows get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-Windows",
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func GenerateS3GetHandler(ctx echo.Context) error {

	logger := getLogger("genS3")
	logger.Info().Msg("genS3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-S3",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetAWSRegions(),
	})
}

func GenerateGCPGetHandler(ctx echo.Context) error {
	logger := getLogger("genGCP")
	logger.Info().Msg("genGCP get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-GCP",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetGCPRegions(),
	})
}

func GenerateNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("genNCP")
	logger.Info().Msg("genNCP get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-NCP",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetNCPRegions(),
	})
}

func GenerateMySQLGetHandler(ctx echo.Context) error {

	logger := getLogger("genmysql")
	logger.Info().Msg("genmysql get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-MySQL",
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func GenerateDynamoDBGetHandler(ctx echo.Context) error {
	logger := getLogger("gendynamodb")
	logger.Info().Msg("gendynamodb get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-DynamoDB",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetAWSRegions(),
	})
}

func GenerateFirestoreGetHandler(ctx echo.Context) error {
	logger := getLogger("genfirestore")
	logger.Info().Msg("genfirestore get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-Firestore",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetGCPRegions(),
	})
}

func GenerateMongoDBGetHandler(ctx echo.Context) error {
	logger := getLogger("genfirestore")
	logger.Info().Msg("genmongodb get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Generate-MongoDB",
		OS:      runtime.GOOS,
		Error:   nil,
		Regions: GetNCPRegions(),
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to backup data

func BackupHandler(ctx echo.Context) error {
	logger := getLogger("backup")
	logger.Info().Msg("backup get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Backup",
		AWSRegions: GetAWSRegions(),
		GCPRegions: GetGCPRegions(),
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		Error:      nil,
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to backup data

func RestoreHandler(ctx echo.Context) error {
	logger := getLogger("restore")
	logger.Info().Msg("restore get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Restore",
		AWSRegions: GetAWSRegions(),
		GCPRegions: GetGCPRegions(),
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		Error:      nil,
	})
}

///////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to migration data

// linux to object storage

func MigrationLinuxToS3GetHandler(ctx echo.Context) error {
	logger := getLogger("miglins3")
	logger.Info().Msg("miglinux get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Linux-S3",
		Regions: GetAWSRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func MigrationLinuxToGCPGetHandler(ctx echo.Context) error {
	logger := getLogger("miglingcp")
	logger.Info().Msg("miglingcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Linux-GCP",
		Regions: GetGCPRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func MigrationLinuxToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("miglinncp")
	logger.Info().Msg("miglinncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Linux-NCP",
		Regions: GetNCPRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

// windows to object storage

func MigrationWindowsToS3GetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	logger := getLogger("migwins3")
	logger.Info().Msg("migwins3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Windows-S3",
		Regions: GetAWSRegions(),
		OS:      runtime.GOOS,
		TmpPath: tmpPath,
		Error:   nil,
	})

}

func MigrationWindowsToGCPGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")
	logger := getLogger("migwingcp")
	logger.Info().Msg("migwingcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Windows-GCP",
		Regions: GetGCPRegions(),
		OS:      runtime.GOOS,
		TmpPath: tmpPath,
		Error:   nil,
	})
}

func MigrationWindowsToNCPGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migwinncp")
	logger.Info().Msg("migwinncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Windows-NCP",
		Regions: GetNCPRegions(),
		OS:      runtime.GOOS,
		TmpPath: tmpPath,
		Error:   nil,
	})
}

// mysql migration page

func MigrationMySQLGetHandler(ctx echo.Context) error {

	logger := getLogger("migmysql")
	logger.Info().Msg("migmysql get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-MySQL",
		Error:   nil,
		OS:      runtime.GOOS,
	})
}

// Object Storage
// AWS to others

func MigrationS3ToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3lin")
	logger.Info().Msg("migs3lin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-S3-Linux",
		Regions: GetAWSRegions(),
		Error:   nil,
		OS:      runtime.GOOS,
	})
}

func MigrationS3ToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migs3win")
	logger.Info().Msg("migs3win get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-S3-Windows",
		Regions: GetAWSRegions(),
		TmpPath: tmpPath,
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func MigrationS3ToGCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3gcp")
	logger.Info().Msg("migs3gcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-S3-GCP",
		AWSRegions: GetAWSRegions(),
		GCPRegions: GetGCPRegions(),
		OS:         runtime.GOOS,
		Error:      nil,
	})
}

func MigrationS3ToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migs3ncp")
	logger.Info().Msg("migs3ncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-S3-NCP",
		AWSRegions: GetAWSRegions(),
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		Error:      nil,
	})
}

// Object Storage
// GCP to others

func MigrationGCPToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("miggcplin")
	logger.Info().Msg("miggcplin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-GCP-Linux",
		OS:      runtime.GOOS,
		Regions: GetGCPRegions(),
		Error:   nil,
	})
}

func MigrationGCPToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("miggcpwin")
	logger.Info().Msg("miggcpwin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-GCP-Windows",
		OS:      runtime.GOOS,
		Regions: GetGCPRegions(),
		TmpPath: tmpPath,
		Error:   nil,
	})
}

func MigrationGCPToS3GetHandler(ctx echo.Context) error {

	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-GCP-S3",
		OS:         runtime.GOOS,
		GCPRegions: GetGCPRegions(),
		AWSRegions: GetAWSRegions(),
		Error:      nil,
	})
}

func MigrationGCPToNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("miggcpncp")
	logger.Info().Msg("miggcpncp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-GCP-NCP",
		OS:         runtime.GOOS,
		GCPRegions: GetGCPRegions(),
		NCPRegions: GetNCPRegions(),
		Error:      nil,
	})
}

// Object Storage
// NCP to others

func MigrationNCPToLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("migncplin")
	logger.Info().Msg("migncplin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-NCP-Linux",
		Regions: GetNCPRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func MigrationNCPToWindowsGetHandler(ctx echo.Context) error {
	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("migncpwin")
	logger.Info().Msg("migncpwin get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-NCP-Windows",
		Regions: GetNCPRegions(),
		OS:      runtime.GOOS,
		TmpPath: tmpPath,
		Error:   nil,
	})
}

func MigrationNCPToS3GetHandler(ctx echo.Context) error {

	logger := getLogger("migncps3")
	logger.Info().Msg("migncps3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-NCP-S3",
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		AWSRegions: GetAWSRegions(),
		Error:      nil,
	})
}

func MigrationNCPToGCPGetHandler(ctx echo.Context) error {

	logger := getLogger("migncpgcp")
	logger.Info().Msg("migncpgcp get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-NCP-GCP",
		NCPRegions: GetNCPRegions(),
		OS:         runtime.GOOS,
		GCPRegions: GetGCPRegions(),
		Error:      nil,
	})
}

// No-SQL
// AWS DynamoDB to others

func MigrationDynamoDBToFirestoreGetHandler(ctx echo.Context) error {

	logger := getLogger("migDNFS")
	logger.Info().Msg("migDNFS get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-DynamoDB-Firestore",
		AWSRegions: GetAWSRegions(),
		OS:         runtime.GOOS,
		GCPRegions: GetGCPRegions(),
		Error:      nil,
	})
}

func MigrationDynamoDBToMongoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migDNMG")
	logger.Info().Msg("migDNMG get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-DynamoDB-MongoDB",
		Regions: GetAWSRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

// No-SQL
// GCP Firestore to others

func MigrationFirestoreToDynamoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migFSDN")
	logger.Info().Msg("migFSDN get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content:    "Migration-Firestore-DynamoDB",
		AWSRegions: GetAWSRegions(),
		OS:         runtime.GOOS,
		GCPRegions: GetGCPRegions(),
		Error:      nil,
	})
}

func MigrationFirestoreToMongoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migFSMG")
	logger.Info().Msg("migFSMG get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-Firestore-MongoDB",
		Regions: GetGCPRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

// No-SQL
// NCP MongoDB to others

func MigrationMongoDBToDynamoDBGetHandler(ctx echo.Context) error {

	logger := getLogger("migMGDN")
	logger.Info().Msg("migMGDN get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-MongoDB-DynamoDB",
		Regions: GetAWSRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}

func MigrationMongoDBToFirestoreGetHandler(ctx echo.Context) error {

	logger := getLogger("migMGFS")
	logger.Info().Msg("migMGFS get page accessed")
	return ctx.Render(http.StatusOK, "index.html", models.BasicPageResponse{
		Content: "Migration-MongoDB-Firestore",
		Regions: GetGCPRegions(),
		OS:      runtime.GOOS,
		Error:   nil,
	})
}
