package serve

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/cloud-barista/cm-data-mold/websrc/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// var router *gin.Engine

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

func InitServer() *echo.Echo {
	// router = gin.New()
	// router.Use(gin.Logger())

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// router.ForwardedByClientIP = true
	// router.SetTrustedProxies([]string{"127.0.0.1"})

	// help needed

	// router.Static("/res", "./web")
	// router.LoadHTMLGlob("./web/templates/*")
	// router.StaticFile("/favicon.ico", "./web/assets/favicon.ico")

	e.Static("/res", "./web")
	e.File("/favicon.ico", "./web/assets/favicon.ico")
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("*.html")),
	}
	e.Renderer = renderer

	// mainGroup := router.Group("/")
	// routes.MainRoutes(mainGroup)
	mainGroup := e.Group("/")
	routes.MainRoutes(mainGroup)

	// generateGroup := router.Group("/generate")
	// routes.GenerateRoutes(generateGroup)
	generateGroup := e.Group("/generate")
	routes.GenerateRoutes(generateGroup)

	// migrationGroup := router.Group("/migration")
	// routes.MigrationRoutes(migrationGroup)
	migrationGroup := e.Group("/migration")
	routes.MigrationRoutes(migrationGroup)

	// return router
	return e
}

func Run(rt *echo.Echo, port string) {
	// rt.Run(":" + port)
	port = fmt.Sprintf(":%s", port)
	if err := rt.Start(port); err != nil && err != http.ErrServerClosed {
		rt.Logger.Panic("shuttig down the server")
	}
}

func RunTLS(rt *echo.Echo, port, cert, key string) {
	rt.StartTLS(":"+port, cert, key)
}
