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
func TaskRoutes(g *echo.Group, scheduleManager *task.FileScheduleManager) {
	TaskRoot(g, scheduleManager)
}

// TaskRoot defines the root routes for Task related operations.
func TaskRoot(g *echo.Group, scheduleManager *task.FileScheduleManager) {
	taskController := controllers.TaskController{
		TaskService: scheduleManager,
	}

	g.GET("", taskController.GetAllTasksHandler)       // Retrieve all tasks
	g.GET("/:id", taskController.GetTaskHandler)       // Retrieve a single task by ID or OperationID
	g.POST("", taskController.CreateTaskHandler)       // Create a new task
	g.PUT("/:id", taskController.UpdateTaskHandler)    // Update an existing task by ID or OperationID
	g.DELETE("/:id", taskController.DeleteTaskHandler) // Delete a task by ID or OperationID
}