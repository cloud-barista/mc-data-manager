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

// ScheduleRoutes initializes the routes for the Task entity.
func ScheduleRoutes(g *echo.Group, scheduleManager *task.FileScheduleManager) {
	ScheduleRoot(g, scheduleManager)
}

// ScheduleRoot defines the root routes for Schedule related operations.
func ScheduleRoot(g *echo.Group, scheduleManager *task.FileScheduleManager) {

	scheduleController := controllers.ScheduleController{
		ScheduleService: scheduleManager,
	}

	g.GET("", scheduleController.GetAllSchedulesHandler)       // Retrieve all schedules
	g.GET("/:id", scheduleController.GetScheduleHandler)       // Retrieve a single schedule by ID
	g.POST("", scheduleController.CreateScheduleHandler)       // Create a new schedule
	g.PUT("/:id", scheduleController.UpdateScheduleHandler)    // Update an existing schedule by ID
	g.DELETE("/:id", scheduleController.DeleteScheduleHandler) // Delete a schedule by ID
	g.GET("/register", controllers.TaskRegisterHandler)

}
