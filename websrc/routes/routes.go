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
