package serve

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/cloud-barista/cm-data-mold/websrc/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
			clientIP := c.RealIP() // Echo gets the real IP of the client

			for _, proxy := range trustedProxies {
				if strings.HasPrefix(clientIP, proxy) {
					// Request is from a trusted proxy
					return next(c)
				}
			}

			// Handling requests from untrusted sources
			return echo.NewHTTPError(http.StatusForbidden, "Access denied")
		}
	}
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
	e.Use(TrustedProxiesMiddleware([]string{"127.0.0.1"}))

	// router.Static("/res", "./web")
	// router.LoadHTMLGlob("./web/templates/*")
	// router.StaticFile("/favicon.ico", "./web/assets/favicon.ico")
	e.Static("/res", "./web")
	e.File("/favicon.ico", "./web/assets/favicon.ico")
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("./web/templates/*.html")),
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
		rt.Logger.Error(err)
		rt.Logger.Panic("shuttig down the server")
	}
}

func RunTLS(rt *echo.Echo, port, cert, key string) {
	rt.StartTLS(":"+port, cert, key)
}
