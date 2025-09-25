package http

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed spec/openapi.yaml
var openapiYAML []byte

func RegisterSwagger(r *gin.Engine) {
	// Serve spec at both root and under /docs to work behind path-based proxies
	r.OPTIONS("/openapi.yaml", corsPreflight)
	r.OPTIONS("/docs/openapi.yaml", corsPreflight)
	r.GET("/openapi.yaml", func(c *gin.Context) {
		setCORS(c)
		c.Data(http.StatusOK, "application/yaml", openapiYAML)
	})
	r.GET("/docs/openapi.yaml", func(c *gin.Context) {
		setCORS(c)
		c.Data(http.StatusOK, "application/yaml", openapiYAML)
	})
	r.OPTIONS("/docs", corsPreflight)
	r.GET("/docs", func(c *gin.Context) {
		setCORS(c)
		html := `<!doctype html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script>
      (function(){
        var base = window.location.pathname.replace(/\/docs.*/, '/docs');
        var specUrl = base + '/openapi.yaml';
        window.ui = SwaggerUIBundle({
          url: specUrl,
          dom_id: '#swagger-ui',
          deepLinking: true,
        });
      })();
    </script>
  </body>
</html>`
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
}

func setCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Vary", "Origin")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Origin, X-Requested-With")
}

func corsPreflight(c *gin.Context) {
	setCORS(c)
	c.Status(http.StatusNoContent)
}
