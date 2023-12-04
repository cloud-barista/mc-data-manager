package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/labstack/echo/v4"
)

func MainGetHandler(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "main",
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Page handlers related to generate data

func GenerateLinuxGetHandler(ctx echo.Context) error {

	logger := getLogger("genlinux")
	logger.Info("genlinux get page accessed")

	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-Linux",
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func GenerateWindowsGetHandler(ctx echo.Context) error {

	tmpPath := filepath.Join(os.TempDir(), "dummy")

	logger := getLogger("genwindows")
	logger.Info("genwindows get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-Windows",
		"os":      runtime.GOOS,
		"tmpPath": tmpPath,
		"error":   nil,
	})
}

func GenerateS3GetHandler(ctx echo.Context) error {

	logger := getLogger("genS3")
	logger.Info("genS3 get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-S3",
		"os":      runtime.GOOS,
		"Regions": GetAWSRegions(),
		"Error":   nil,
	})
}

func GenerateGCPGetHandler(ctx echo.Context) error {
	logger := getLogger("genGCP")
	logger.Info("genGCP get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-GCP",
		"os":      runtime.GOOS,
		"Regions": GetGCPRegions(),
		"error":   nil,
	})
}

func GenerateNCPGetHandler(ctx echo.Context) error {

	logger := getLogger("genNCP")
	logger.Info("genNCP get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-NCP",
		"os":      runtime.GOOS,
		"Regions": GetNCPRegions(),
		"error":   nil,
	})
}

func GenerateMySQLGetHandler(ctx echo.Context) error {

	logger := getLogger("genmysql")
	logger.Info("genmysql get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-MySQL",
		"os":      runtime.GOOS,
		"error":   nil,
	})
}

func GenerateDynamoDBGetHandler(ctx echo.Context) error {
	logger := getLogger("gendynamodb")
	logger.Info("gendynamodb get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-DynamoDB",
		"os":      runtime.GOOS,
		"Regions": GetAWSRegions(),
		"error":   nil,
	})
}

func GenerateFirestoreGetHandler(ctx echo.Context) error {
	logger := getLogger("genfirestore")
	logger.Info("genfirestore get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-Firestore",
		"os":      runtime.GOOS,
		"Regions": GetGCPRegions(),
		"error":   nil,
	})
}

func GenerateMongoDBGetHandler(ctx echo.Context) error {
	logger := getLogger("genfirestore")
	logger.Info("genmongodb get page accessed")
	return ctx.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Content": "Generate-MongoDB",
		"os":      runtime.GOOS,
		"error":   nil,
	})
}
