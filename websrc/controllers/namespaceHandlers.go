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
// @Summary Set the active namespace ID
// @Description Sets the runtime nsId received from the parent page via postMessage
// @Tags namespace
// @Accept json
// @Produce json
// @Param body body setNsIdRequest true "namespace ID"
// @Success 200 {object} map[string]string
// @Router /namespace [post]
func SetNsIdHandler(c echo.Context) error {
	var req setNsIdRequest
	if err := c.Bind(&req); err != nil || req.NsId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "nsId is required"})
	}

	utils.SetNsId(req.NsId)
	return c.JSON(http.StatusOK, map[string]string{"nsId": req.NsId})
}

// GetNsIdHandler godoc
// @Summary Get the active namespace ID
// @Description Returns the currently active nsId (runtime value or env default)
// @Tags namespace
// @Produce json
// @Success 200 {object} map[string]string
// @Router /namespace [get]
func GetNsIdHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"nsId": utils.GetNsId()})
}
