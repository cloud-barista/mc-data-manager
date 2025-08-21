package controllers

import (
	"net/http"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	service "github.com/cloud-barista/mc-data-manager/service/credential"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CredentialHandler struct {
	credetialService *service.CredentialService
}

func NewCredentialHandler(db *gorm.DB) *CredentialHandler {
	credentialService := service.NewCredentialService(db)

	return &CredentialHandler{
		credetialService: credentialService,
	}
}

// POST /credentials
func (c *CredentialHandler) CreateCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "create credential", start)

	params := models.Credential{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	// 서비스 호출
	credential, err := c.credetialService.CreateCredential(params)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully created credential", start)
	return ctx.JSON(http.StatusOK, credential)
}

// GET /credentials
func (c *CredentialHandler) ListCredentialsHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get all credentials", start)

	// 서비스 호출
	credential, err := c.credetialService.ListCredentials()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully get all credentials", start)
	return ctx.JSON(http.StatusOK, credential)
}

// GET /credentials/:id
func (c *CredentialHandler) GetCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)

	id := ctx.Param("id")
	credential, err := c.credetialService.GetCredentialById(id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully get credential", start)
	return ctx.JSON(http.StatusOK, credential)
}

// PUT /credentials/:id
func (c *CredentialHandler) UpdateCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)

	credentialId := ctx.Param("id")
	params := models.Credential{}
	if !getDataWithBind(logger, start, ctx, &params) {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	credential, err := c.credetialService.UpdateCredential(credentialId, params)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully updated credential", start)
	return ctx.JSON(http.StatusOK, credential)
}

// DELETE /credentials/:id
func (c *CredentialHandler) DeleteCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)

	id := ctx.Param("id")
	if err := c.credetialService.DeleteCredential(id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully deleted credential", start)
	return ctx.JSON(http.StatusOK, map[string]string{"deleted": id})
}
