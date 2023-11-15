package routes

import (
	"github.com/cloud-barista/cm-data-mold/websrc/controllers"
	"github.com/gin-gonic/gin"
)

func MigrationRoutes(g *gin.RouterGroup) {
	// Migration From On-premise (Linux, Windows) to Object Storage
	MigrationFromOnpremiseToObjectStorage(g)

	// Migration MySQL to MySQL
	MigrationMySQL(g)

	// Migration From Object Storage to Other Object Storage
	MigrationFromS3Routes(g)
	MigrationFromGCSRoutes(g)
	MigrationFromNCSRoutes(g)

	// Migration No-SQL to the other No-SQL
	MigrationNoSQLRoutes(g)
}

func MigrationFromOnpremiseToObjectStorage(g *gin.RouterGroup) {
	g.GET("/linux/s3", controllers.MigrationLinuxToS3GetHandler())
	g.POST("/linux/s3", controllers.MigrationLinuxToS3PostHandler())

	g.GET("/linux/gcs", controllers.MigrationLinuxToGCSGetHandler())
	g.POST("/linux/gcs", controllers.MigrationLinuxToGCSPostHandler())

	g.GET("/linux/ncs", controllers.MigrationLinuxToNCSGetHandler())
	g.POST("/linux/ncs", controllers.MigrationLinuxToNCSPostHandler())

	g.GET("/windows/s3", controllers.MigrationWindowsToS3GetHandler())
	g.POST("/windows/s3", controllers.MigrationWindowsToS3PostHandler())

	g.GET("/windows/gcs", controllers.MigrationWindowsToGCSGetHandler())
	g.POST("/windows/gcs", controllers.MigrationWindowsToGCSPostHandler())

	g.GET("/windows/ncs", controllers.MigrationWindowsToNCSGetHandler())
	g.POST("/windows/ncs", controllers.MigrationWindowsToNCSPostHandler())
}

func MigrationMySQL(g *gin.RouterGroup) {
	g.GET("/mysql", controllers.MigrationMySQLGetHandler())
	g.POST("/mysql", controllers.MigrationMySQLPostHandler())
}

func MigrationFromS3Routes(g *gin.RouterGroup) {
	g.GET("/s3/linux", controllers.MigrationS3ToLinuxGetHandler())
	g.POST("/s3/linux", controllers.MigrationS3ToLinuxPostHandler())

	g.GET("/s3/windows", controllers.MigrationS3ToWindowsGetHandler())
	g.POST("/s3/windows", controllers.MigrationS3ToWindowsPostHandler())

	g.GET("/s3/gcs", controllers.MigrationS3ToGCSGetHandler())
	g.POST("/s3/gcs", controllers.MigrationS3ToGCSPostHandler())

	g.GET("/s3/ncs", controllers.MigrationS3ToNCSGetHandler())
	g.POST("/s3/ncs", controllers.MigrationS3ToNCSPostHandler())
}

func MigrationFromGCSRoutes(g *gin.RouterGroup) {
	g.GET("/gcs/linux", controllers.MigrationGCSToLinuxGetHandler())
	g.POST("/gcs/linux", controllers.MigrationGCSToLinuxPostHandler())

	g.GET("/gcs/windows", controllers.MigrationGCSToWindowsGetHandler())
	g.POST("/gcs/windows", controllers.MigrationGCSToWindowsPostHandler())

	g.GET("/gcs/s3", controllers.MigrationGCSToS3GetHandler())
	g.POST("/gcs/s3", controllers.MigrationGCSToS3PostHandler())

	g.GET("/gcs/ncs", controllers.MigrationGCSToNCSGetHandler())
	g.POST("/gcs/ncs", controllers.MigrationGCSToNCSPostHandler())
}

func MigrationFromNCSRoutes(g *gin.RouterGroup) {
	g.GET("/ncs/linux", controllers.MigrationNCSToLinuxGetHandler())
	g.POST("/ncs/linux", controllers.MigrationNCSToLinuxPostHandler())

	g.GET("/ncs/windows", controllers.MigrationNCSToWindowsGetHandler())
	g.POST("/ncs/windows", controllers.MigrationNCSToWindowsPostHandler())

	g.GET("/ncs/s3", controllers.MigrationNCSToS3GetHandler())
	g.POST("/ncs/s3", controllers.MigrationNCSToS3PostHandler())

	g.GET("/ncs/gcs", controllers.MigrationNCSToGCSGetHandler())
	g.POST("/ncs/gcs", controllers.MigrationNCSToGCSPostHandler())
}

func MigrationNoSQLRoutes(g *gin.RouterGroup) {
	g.GET("/dynamodb/firestore", controllers.MigrationDynamoDBToFirestoreGetHandler())
	g.POST("/dynamodb/firestore", controllers.MigrationDynamoDBToFirestorePostHandler())

	g.GET("/dynamodb/mongodb", controllers.MigrationDynamoDBToMongoDBGetHandler())
	g.POST("/dynamodb/mongodb", controllers.MigrationDynamoDBToMongoDBPostHandler())

	g.GET("/firestore/dynamodb", controllers.MigrationFirestoreToDynamoDBGetHandler())
	g.POST("/firestore/dynamodb", controllers.MigrationFirestoreToDynamoDBPostHandler())

	g.GET("/firestore/mongodb", controllers.MigrationFirestoreToMongoDBGetHandler())
	g.POST("/firestore/mongodb", controllers.MigrationFirestoreToMongoDBPostHandler())

	g.GET("/mongodb/dynamodb", controllers.MigrationMongoDBToDynamoDBGetHandler())
	g.POST("/mongodb/dynamodb", controllers.MigrationMongoDBToDynamoDBPostHandler())

	g.GET("/mongodb/firestore", controllers.MigrationMongoDBToFirestoreGetHandler())
	g.POST("/mongodb/firestore", controllers.MigrationMongoDBToFirestorePostHandler())
}
