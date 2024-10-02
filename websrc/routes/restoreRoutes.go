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

func RestoreRoutes(g *echo.Group) {
	// RestoreURL
	RestoreRoot(g)
	// RestoreFrom On-premise (Linux, Windows) to Object Storage

	// RestoreOBJ storage to linux
	RestoreObjectStorage(g)
	// RestoreMySQL to linux
	RestoreRDB(g)
	RestoreNRDB(g)
}

func RestoreRoot(g *echo.Group) {

	g.GET("/register", controllers.RestoreHandler)
	g.GET("", controllers.GetAllRestoreHandler)         // Retrieve all tasks
	g.GET("/:id", controllers.GetRestoreHandler)        // Retrieve a single task by ID
	g.PUT("/:id", controllers.UpdateRestoreHandler)     // Update an existing task by ID
	g.DELETE("/:id", controllers.DeleteRestorekHandler) // Delete a task by ID

}

func RestoreObjectStorage(g *echo.Group) {
	// g.GET("/objectstorage", controllers.RestoreOSGetHandler)
	g.POST("/objectstorage", controllers.RestoreOSPostHandler)
}
func RestoreRDB(g *echo.Group) {
	// g.GET("/rdb", controllers.RestoreRDBGetHandler)
	g.POST("/rdb", controllers.RestoreRDBPostHandler)
}

func RestoreNRDB(g *echo.Group) {
	// g.GET("/nrdb", controllers.RestoreNRDBGetHandler)
	g.POST("/nrdb", controllers.RestoreNRDBPostHandler)
}
