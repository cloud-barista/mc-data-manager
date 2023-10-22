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
	g.POST("/linux/submit", controllers.GenerateLinuxSubmitPostHandler())

	g.GET("/windows", controllers.GenerateWindowsGetHandler())
	g.POST("/windows/submit", controllers.GenerateWindowsSubmitPostHandler())

	g.GET("/s3", controllers.GenerateS3GetHandler())
	g.POST("/s3/submit", controllers.GenerateS3SubmitPostHandler())

	g.GET("/gcs", controllers.GenerateGCSGetHandler())
	g.POST("/gcs/submit", controllers.GenerateGCSSubmitPostHandler())

	g.GET("/ncs", controllers.GenerateNCSGetHandler())
	g.POST("/ncs/submit", controllers.GenerateNCSSubmitPostHandler())

	g.GET("/mysql", controllers.GenerateMySQLGetHandler())
	g.POST("/mysql/submit", controllers.GenerateMySQLSubmitPostHandler())

	g.GET("/dynamodb", controllers.GenerateDynamoDBGetHandler())
	g.POST("/dynamodb/submit", controllers.GenerateDynamoDBSubmitPostHandler())

	g.GET("/firestore", controllers.GenerateFirestoreGetHandler())
	g.POST("/firestore/submit", controllers.GenerateFirestoreSubmitPostHandler())

	g.GET("/mongodb", controllers.GenerateMongoDBGetHandler())
	g.POST("/mongodb/submit", controllers.GenerateMongoDBSubmitPostHandler())
}
