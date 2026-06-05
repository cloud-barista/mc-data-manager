package routes

import (
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/labstack/echo/v4"
)

func NamespaceRoutes(g *echo.Group) {
	g.POST("", controllers.SetNsIdHandler)
	g.GET("", controllers.GetNsIdHandler)
}
