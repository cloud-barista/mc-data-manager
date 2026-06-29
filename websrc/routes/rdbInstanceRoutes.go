package routes

import (
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/labstack/echo/v4"
)

// RDBInstanceRoutes registers RDB (database) instance infrastructure endpoints
// under the /db group.
func RDBInstanceRoutes(g *echo.Group) {
	g.POST("/rdbms", controllers.ListRDBInstancesHandler)
	g.PUT("/rdbms", controllers.CreateRDBInstanceHandler)
	g.POST("/rdbms/engine-versions", controllers.ListRDBEngineVersionsHandler)
	g.POST("/rdbms/instance-class", controllers.ListRDBInstanceClassesHandler)
	g.POST("/rdbms/databases", controllers.ListRDBDatabasesHandler)
}
