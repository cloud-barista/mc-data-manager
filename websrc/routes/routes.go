package routes

import (
	"github.com/cloud-barista/cm-data-mold/websrc/controllers"
	"github.com/labstack/echo/v4"
)

func GenerateRoutes(g *echo.Group) {
	g.GET("/linux", controllers.GenerateLinuxGetHandler)
	g.POST("/linux", controllers.GenerateLinuxPostHandler)

	g.GET("/windows", controllers.GenerateWindowsGetHandler)
	g.POST("/windows", controllers.GenerateWindowsPostHandler)

	g.GET("/s3", controllers.GenerateS3GetHandler)
	g.POST("/s3", controllers.GenerateS3PostHandler)

	g.GET("/gcp", controllers.GenerateGCPGetHandler)
	g.POST("/gcp", controllers.GenerateGCPPostHandler)

	g.GET("/ncp", controllers.GenerateNCPGetHandler)
	g.POST("/ncp", controllers.GenerateNCPPostHandler)

	g.GET("/mysql", controllers.GenerateMySQLGetHandler)
	g.POST("/mysql", controllers.GenerateMySQLPostHandler)

	g.GET("/dynamodb", controllers.GenerateDynamoDBGetHandler)
	g.POST("/dynamodb", controllers.GenerateDynamoDBPostHandler)

	g.GET("/firestore", controllers.GenerateFirestoreGetHandler)
	g.POST("/firestore", controllers.GenerateFirestorePostHandler)

	g.GET("/mongodb", controllers.GenerateMongoDBGetHandler)
	g.POST("/mongodb", controllers.GenerateMongoDBPostHandler)
}
