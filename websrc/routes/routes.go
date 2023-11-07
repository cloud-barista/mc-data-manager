package routes

import (
	"github.com/cloud-barista/cm-data-mold/websrc/controllers"
	"github.com/gin-gonic/gin"
)

func MainRoutes(g *gin.RouterGroup) {
	g.GET("/", controllers.MainGetHandler())
}

func GenerateRoutes(g *gin.RouterGroup) {
	g.GET("/linux", controllers.GenerateLinuxGetHandler())
	g.POST("/linux", controllers.GenerateLinuxPostHandler())

	g.GET("/windows", controllers.GenerateWindowsGetHandler())
	g.POST("/windows", controllers.GenerateWindowsPostHandler())

	g.GET("/s3", controllers.GenerateS3GetHandler())
	g.POST("/s3", controllers.GenerateS3PostHandler())

	g.GET("/gcs", controllers.GenerateGCSGetHandler())
	g.POST("/gcs", controllers.GenerateGCSPostHandler())

	g.GET("/ncs", controllers.GenerateNCSGetHandler())
	g.POST("/ncs", controllers.GenerateNCSPostHandler())

	g.GET("/mysql", controllers.GenerateMySQLGetHandler())
	g.POST("/mysql", controllers.GenerateMySQLPostHandler())

	g.GET("/dynamodb", controllers.GenerateDynamoDBGetHandler())
	g.POST("/dynamodb", controllers.GenerateDynamoDBPostHandler())

	g.GET("/firestore", controllers.GenerateFirestoreGetHandler())
	g.POST("/firestore", controllers.GenerateFirestorePostHandler())

	g.GET("/mongodb", controllers.GenerateMongoDBGetHandler())
	g.POST("/mongodb", controllers.GenerateMongoDBPostHandler())
}
