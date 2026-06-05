package controllers

import (
	"net/http"

	"github.com/cloud-barista/mc-data-manager/pkg/utils"
	"github.com/labstack/echo/v4"
)

type setNsIdRequest struct {
	NsId string `json:"nsId"`
}

// SetNsIdHandler godoc
//
//	@ID			SetNsIdHandler
//	@Summary	Set the active namespace ID
//	@Description	Sets the runtime nsId received from the parent page via postMessage
//	@Tags			[Namespace]
//	@Accept			json
//	@Produce		json
//	@Param			RequestBody	body		setNsIdRequest		true	"Namespace ID to set"
//	@Success		200			{object}	map[string]string	"Active namespace ID"
//	@Failure		400			{object}	map[string]string	"Invalid Request"
//	@Router			/namespace [post]
func SetNsIdHandler(c echo.Context) error {
	var req setNsIdRequest
	if err := c.Bind(&req); err != nil || req.NsId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "nsId is required"})
	}

	utils.SetNsId(req.NsId)
	return c.JSON(http.StatusOK, map[string]string{"nsId": req.NsId})
}

// GetNsIdHandler godoc
//
//	@ID			GetNsIdHandler
//	@Summary	Get the active namespace ID
//	@Description	Returns the currently active nsId (runtime value or env default)
//	@Tags			[Namespace]
//	@Produce		json
//	@Success		200	{object}	map[string]string	"Active namespace ID"
//	@Router			/namespace [get]
func GetNsIdHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"nsId": utils.GetNsId()})
}
