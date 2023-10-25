package serve

import (
	"github.com/cloud-barista/cm-data-mold/websrc/routes"
	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func InitServer() *gin.Engine {
	router = gin.New()
	router.Use(gin.Logger())
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.Static("/res", "./web")
	router.LoadHTMLGlob("./web/templates/*")
	router.StaticFile("/favicon.ico", "./web/assets/favicon.ico")

	mainGroup := router.Group("/")
	routes.MainRoutes(mainGroup)

	generateGroup := router.Group("/generate")
	routes.GenerateRoutes(generateGroup)

	migrationGroup := router.Group("/migration")
	routes.MigrationRoutes(migrationGroup)

	return router
}

func Run(rt *gin.Engine, port string) {
	rt.Run(":" + port)
}

func RunTLS(rt *gin.Engine, port, cert, key string) {
	rt.RunTLS(":"+port, cert, key)
}
