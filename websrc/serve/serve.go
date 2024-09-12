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
package serve

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/cloud-barista/mc-data-manager/service/task"
	"github.com/cloud-barista/mc-data-manager/websrc/controllers"
	"github.com/cloud-barista/mc-data-manager/websrc/middlewares"
	"github.com/cloud-barista/mc-data-manager/websrc/routes"
	"github.com/rs/zerolog/log"

	// REST API (echo)
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// echo-swagger middleware
	_ "github.com/cloud-barista/mc-data-manager/websrc/docs"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const (
	infoColor    = "\033[1;34m%s\033[0m"
	noticeColor  = "\033[1;36m%s\033[0m"
	warningColor = "\033[1;33m%s\033[0m"
	errorColor   = "\033[1;31m%s\033[0m"
	debugColor   = "\033[0;36m%s\033[0m"
)

// TemplateRenderer is a custom html/template renderer for Echo framework
type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template document
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

// Custom middleware to check the list of trusted proxies
func TrustedProxiesMiddleware(trustedProxies []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			clientIP := net.ParseIP(c.RealIP()) // Parse the real IP of the client

			if clientIP == nil {
				return echo.NewHTTPError(http.StatusForbidden, "Invalid IP address")
			}

			for _, proxy := range trustedProxies {
				// Append /32 if no subnet mask is specified
				if !strings.Contains(proxy, "/") {
					proxy += "/32"
				}
				_, cidr, err := net.ParseCIDR(proxy)
				if err != nil {
					continue
				}
				if cidr.Contains(clientIP) {
					// Request is from a trusted proxy
					return next(c)
				}
			}

			// Handling requests from untrusted sources
			return echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
	}
}

// RunServer func start Rest API server

// @title MC-Data-Manager REST API
// @version latest
// @description MC-Data-Manager REST API

// @contact.name API Support
// @contact.url http://cloud-barista.github.io
// @contact.email contact-to-cloud-barista@googlegroups.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /
func InitServer(port string, addIP ...string) *echo.Echo {
	e := echo.New()

	e.HideBanner = true

	allowIP := []string{"127.0.0.1", "::1"}
	allowIP = append(allowIP, addIP...)

	// Middleware
	e.Use(TrustedProxiesMiddleware(allowIP))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// Custom middleware for tracing
	e.Use(middlewares.TracingMiddleware)

	e.Static("/res", "./web")
	e.File("/favicon.ico", "./web/assets/favicon.ico")
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("./web/templates/*.html")),
	}
	e.Renderer = renderer

	// go cron
	scheduleManager := task.InitFileScheduleManager()

	// Route for system management
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.GET("/", controllers.MainGetHandler)

	backupGroup := e.Group("/backup")
	routes.BackupRoutes(backupGroup)

	generateGroup := e.Group("/generate")
	routes.GenerateRoutes(generateGroup)

	migrationGroup := e.Group("/migration")
	routes.MigrationRoutes(migrationGroup)

	restoreGroup := e.Group("/restore")
	routes.RestoreRoutes(restoreGroup)

	taskGroup := e.Group("/task")
	routes.TaskRoutes(taskGroup, scheduleManager)

	selfEndpoint := "localhost" + ":" + port
	website := " http://" + selfEndpoint
	apidashboard := " http://" + selfEndpoint + "/swagger/index.html"

	log.Info().Msgf("Data Manager Web UI is available at")
	log.Info().Msgf(noticeColor, website)
	log.Info().Msgf("\n ")

	log.Info().Msgf("Swagger UI (REST API Document) is available at")
	log.Info().Msgf(noticeColor, apidashboard)
	log.Info().Msgf("\n ")

	return e
}

func Run(rt *echo.Echo, port string) {
	port = fmt.Sprintf(":%s", port)
	if err := rt.Start(port); err != nil && err != http.ErrServerClosed {
		rt.Logger.Error(err)
		rt.Logger.Panic("shuttig down the server")
	}
}

func RunTLS(rt *echo.Echo, port, cert, key string) {
	rt.StartTLS(":"+port, cert, key)
}
