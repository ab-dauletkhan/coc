package http

import (
	"bytes"
	_ "embed"
	"net/http"
	"os"
	"strings"

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
		c.Data(http.StatusOK, "application/yaml", rewriteServers(openapiYAML, currentPublicURL(), currentEnvDescription()))
	})
	r.GET("/docs/openapi.yaml", func(c *gin.Context) {
		setCORS(c)
		c.Data(http.StatusOK, "application/yaml", rewriteServers(openapiYAML, currentPublicURL(), currentEnvDescription()))
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

// currentPublicURL returns the externally reachable base URL for this service.
// Controls:
// - APP_PUBLIC_URL: explicit full URL (e.g. https://coc-7tqg.onrender.com)
// - otherwise defaults to http://localhost:8080
func currentPublicURL() string {
	if v := os.Getenv("APP_PUBLIC_URL"); strings.TrimSpace(v) != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://localhost:8080"
}

// currentEnvDescription provides a human label for the server entry.
// Controls:
// - NODE_ENV=production → "Production"
// - else → "Local"
func currentEnvDescription() string {
	if strings.EqualFold(os.Getenv("NODE_ENV"), "production") {
		return "Production"
	}
	return "Local"
}

// rewriteServers performs a minimal, line-based replacement of the first server entry
// in the embedded OpenAPI YAML to inject the runtime public URL and description.
func rewriteServers(spec []byte, serverURL, desc string) []byte {
	lines := bytes.Split(spec, []byte("\n"))
	inServers := false
	replacedURL := false
	replacedDesc := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		// Detect start of servers section
		if bytes.HasPrefix(line, []byte("servers:")) {
			inServers = true
			continue
		}
		if inServers {
			// Replace first "- url:" under servers
			if !replacedURL && bytes.Contains(line, []byte("- url:")) {
				indent := line[:bytes.Index(line, []byte("- url:"))]
				lines[i] = append([]byte{}, append(indent, []byte("- url: ")...)...)
				lines[i] = append(lines[i], []byte(serverURL)...)
				replacedURL = true
				continue
			}
			// Replace the following description line once
			if replacedURL && !replacedDesc && bytes.Contains(line, []byte("description:")) {
				indent := line[:bytes.Index(line, []byte("description:"))]
				lines[i] = append([]byte{}, append(indent, []byte("description: ")...)...)
				lines[i] = append(lines[i], []byte(desc)...)
				replacedDesc = true
				// Stop after first server block updated
				break
			}
			// Exit servers section if another top-level tag begins
			if len(bytes.TrimSpace(line)) > 0 && !bytes.HasPrefix(line, []byte("  ")) && replacedURL {
				break
			}
		}
	}
	return bytes.Join(lines, []byte("\n"))
}
