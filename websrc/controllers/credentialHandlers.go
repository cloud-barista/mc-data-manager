package controllers

import (
	"fmt"
	"net/http"
	"time"

	"strconv"

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

// CreateCredentialHandler godoc
//
//	@ID 			CreateCredentialHandler
//	@Summary		save encrypted credential.
//	@Description	save encrypted credential.
//	@Tags			[Credential]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody		body	models.CredentialCreateRequest	true	"Parameters required for Credential"
//	@Success		200			{object}	models.BasicResponse	"Successfully saved credential"
//	@Failure		400			{object}	models.BasicResponse	"Invalid Request"
//	@Failure		500			{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/credentials [post]
func (c *CredentialHandler) CreateCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "create credential", start)

	params := models.CredentialCreateRequest{}
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

	jobEnd(logger, "Successfully saved credential", start)
	return ctx.JSON(http.StatusOK, credential)
}

// ListCredentialsHandler godoc
//
//	@ID 			ListCredentialsHandler
//	@Summary		Get all credentials
//	@Description	Retrieve a list of all credentials in the system.
//	@Tags			[Credential]
//	@Produce		json
//	@Success		200		{array}		models.Task	"Successfully retrieved all credentials"
//	@Failure		500		{object}	models.BasicResponse	"Internal Server Error"
//	@Router			/credentials [get]
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

	jobEnd(logger, "Successfully retrieved all credentials", start)
	return ctx.JSON(http.StatusOK, credential)
}

// GetCredentialHandler godoc
//
//	@ID 			GetCredentialHandler
//	@Summary		Get a Credential by ID
//	@Description	Get the details of a Credential using its ID.
//	@Tags			[Credential]
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string	true	"Credential ID"
//	@Success		200		{object}	models.Task	"Successfully retrieved credential"
//	@Failure		404		{object}	models.BasicResponse	"Task not found"
//	@Router			/credentials/{id} [get]
func (c *CredentialHandler) GetCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %w", err)
	}
	credential, err := c.credetialService.GetCredentialById(id)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully retrieved credential", start)
	return ctx.JSON(http.StatusOK, credential)
}

// PUT /credentials/:id
func (c *CredentialHandler) UpdateCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)

	credentialId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %w", err)
	}
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

// DeleteCredentialHandler godoc
//
//	@ID 			DeleteCredentialHandler
//	@Summary		Delete a credential
//	@Description	Delete an existing credential using its ID.
//	@Tags			[Credential]
//	@Produce		json
//	@Param			id		path	string	true	"Credential ID"
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted Credential"
//	@Failure		404		{object}	models.BasicResponse	"Credential not found"
//	@Router			/credentials/{id} [delete]
func (c *CredentialHandler) DeleteCredentialHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "credential", "get credential", start)

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid number: %w", err)
	}

	if err := c.credetialService.DeleteCredential(id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  nil,
		})
	}

	jobEnd(logger, "Successfully deleted Credential", start)
	return ctx.JSON(http.StatusOK, map[string]uint64{"deleted": id})
}
