/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package routes

import (
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/labstack/echo/v4"
)

func BackupRoutes(g *echo.Group) {
	// Backup URL
	BackupRoot(g)
	// Backup From On-premise (Linux, Windows) to Object Storage
	BackupFromOnpremiseToObjectStorage(g)

	// Backup MySQL to MySQL
	BackupMySQL(g)

	// Backup From Object Storage to Other Object Storage
	BackupFromS3Routes(g)
	BackupFromGCPRoutes(g)
	BackupFromNCPRoutes(g)

	// Backup No-SQL to the other No-SQL
	BackupNoSQLRoutes(g)
}

func BackupRoot(g *echo.Group) {
	g.GET("", controllers.BackupHandler)
}

func BackupFromOnpremiseToObjectStorage(g *echo.Group) {
	g.GET("/linux/aws", controllers.MigrationLinuxToS3GetHandler)
	g.POST("/linux/aws", controllers.MigrationLinuxToS3PostHandler)

	g.GET("/linux/gcp", controllers.MigrationLinuxToGCPGetHandler)
	g.POST("/linux/gcp", controllers.MigrationLinuxToGCPPostHandler)

	g.GET("/linux/ncp", controllers.MigrationLinuxToNCPGetHandler)
	g.POST("/linux/ncp", controllers.MigrationLinuxToNCPPostHandler)

	g.GET("/windows/aws", controllers.MigrationWindowsToS3GetHandler)
	g.POST("/windows/aws", controllers.MigrationWindowsToS3PostHandler)

	g.GET("/windows/gcp", controllers.MigrationWindowsToGCPGetHandler)
	g.POST("/windows/gcp", controllers.MigrationWindowsToGCPPostHandler)

	g.GET("/windows/ncp", controllers.MigrationWindowsToNCPGetHandler)
	g.POST("/windows/ncp", controllers.MigrationWindowsToNCPPostHandler)
}

func BackupMySQL(g *echo.Group) {
	g.GET("/mysql", controllers.MigrationMySQLGetHandler)
	g.POST("/mysql", controllers.MigrationMySQLPostHandler)
}

func BackupFromS3Routes(g *echo.Group) {
	g.GET("/aws/linux", controllers.MigrationS3ToLinuxGetHandler)
	g.POST("/aws/linux", controllers.MigrationS3ToLinuxPostHandler)

	g.GET("/aws/windows", controllers.MigrationS3ToWindowsGetHandler)
	g.POST("/aws/windows", controllers.MigrationS3ToWindowsPostHandler)

	g.GET("/aws/gcp", controllers.MigrationS3ToGCPGetHandler)
	g.POST("/aws/gcp", controllers.MigrationS3ToGCPPostHandler)

	g.GET("/aws/ncp", controllers.MigrationS3ToNCPGetHandler)
	g.POST("/aws/ncp", controllers.MigrationS3ToNCPPostHandler)
}

func BackupFromGCPRoutes(g *echo.Group) {
	g.GET("/gcp/linux", controllers.MigrationGCPToLinuxGetHandler)
	g.POST("/gcp/linux", controllers.MigrationGCPToLinuxPostHandler)

	g.GET("/gcp/windows", controllers.MigrationGCPToWindowsGetHandler)
	g.POST("/gcp/windows", controllers.MigrationGCPToWindowsPostHandler)

	g.GET("/gcp/aws", controllers.MigrationGCPToS3GetHandler)
	g.POST("/gcp/aws", controllers.MigrationGCPToS3PostHandler)

	g.GET("/gcp/ncp", controllers.MigrationGCPToNCPGetHandler)
	g.POST("/gcp/ncp", controllers.MigrationGCPToNCPPostHandler)
}

func BackupFromNCPRoutes(g *echo.Group) {
	g.GET("/ncp/linux", controllers.MigrationNCPToLinuxGetHandler)
	g.POST("/ncp/linux", controllers.MigrationNCPToLinuxPostHandler)

	g.GET("/ncp/windows", controllers.MigrationNCPToWindowsGetHandler)
	g.POST("/ncp/windows", controllers.MigrationNCPToWindowsPostHandler)

	g.GET("/ncp/aws", controllers.MigrationNCPToS3GetHandler)
	g.POST("/ncp/aws", controllers.MigrationNCPToS3PostHandler)

	g.GET("/ncp/gcp", controllers.MigrationNCPToGCPGetHandler)
	g.POST("/ncp/gcp", controllers.MigrationNCPToGCPPostHandler)
}

func BackupNoSQLRoutes(g *echo.Group) {
	g.GET("/dynamodb/firestore", controllers.MigrationDynamoDBToFirestoreGetHandler)
	g.POST("/dynamodb/firestore", controllers.MigrationDynamoDBToFirestorePostHandler)

	g.GET("/dynamodb/mongodb", controllers.MigrationDynamoDBToMongoDBGetHandler)
	g.POST("/dynamodb/mongodb", controllers.MigrationDynamoDBToMongoDBPostHandler)

	g.GET("/firestore/dynamodb", controllers.MigrationFirestoreToDynamoDBGetHandler)
	g.POST("/firestore/dynamodb", controllers.MigrationFirestoreToDynamoDBPostHandler)

	g.GET("/firestore/mongodb", controllers.MigrationFirestoreToMongoDBGetHandler)
	g.POST("/firestore/mongodb", controllers.MigrationFirestoreToMongoDBPostHandler)

	g.GET("/mongodb/dynamodb", controllers.MigrationMongoDBToDynamoDBGetHandler)
	g.POST("/mongodb/dynamodb", controllers.MigrationMongoDBToDynamoDBPostHandler)

	g.GET("/mongodb/firestore", controllers.MigrationMongoDBToFirestoreGetHandler)
	g.POST("/mongodb/firestore", controllers.MigrationMongoDBToFirestorePostHandler)
}
