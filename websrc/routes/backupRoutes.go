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

	// Backup OBJ storage to linux
	BackupObjectStorage(g)
	// Backup MySQL to linux
	BackupRDB(g)
	BackupNRDB(g)

}

func BackupRoot(g *echo.Group) {
	g.GET("/register", controllers.BackupHandler)
	g.GET("", controllers.GetAllBackupHandler)         // Retrieve all tasks
	g.GET("/:id", controllers.GetBackupHandler)        // Retrieve a single task by ID
	g.PUT("/:id", controllers.UpdateBackupHandler)     // Update an existing task by ID
	g.DELETE("/:id", controllers.DeleteBackupkHandler) // Delete a task by ID

}

func BackupObjectStorage(g *echo.Group) {
	// g.GET("/objectstorage", controllers.BackupOSGetHandler)
	g.POST("/objectstorage", controllers.BackupOSPostHandler)
}
func BackupRDB(g *echo.Group) {
	// g.GET("/rdb", controllers.BackupRDBGetHandler)
	g.POST("/rdbms", controllers.BackupRDBPostHandler)
}

func BackupNRDB(g *echo.Group) {
	// g.GET("/nrdb", controllers.BackupNRDBGetHandler)
	g.POST("/nrdbms", controllers.BackupNRDBPostHandler)
}
