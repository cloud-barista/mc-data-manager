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
	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/labstack/echo/v4"
)

// TaskRoutes initializes the routes for the Task entity.
func ServiceRoutes(g *echo.Group, scheduleManager *task.FileScheduleManager) {
	ServiceRoot(g, scheduleManager)
}

// TaskRoot defines the root routes for Task related operations.
func ServiceRoot(g *echo.Group, scheduleManager *task.FileScheduleManager) {
	serviceController := controllers.TaskController{
		TaskService: scheduleManager,
	}

	// Route to clear all services and tasks
	g.DELETE("/clearAll", serviceController.DeleteServiceAndTaskAllHandler)

	// Route to apply resources using apply.sh
	g.POST("/apply", serviceController.ApplyResourceHandler)

	// Route to destroy resources using destroy.sh
	g.DELETE("/destroy", serviceController.DestroyResourceHandler)

}
