package swagger

import (
	"embed"
	"net/http"
)

//go:embed openapi.json
var fs embed.FS

func OpenAPIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := fs.ReadFile("openapi.json")
		if err != nil {
			http.Error(w, "openapi not available", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})
}

func SwaggerUIHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swagger/" && r.URL.Path != "/swagger/index.html" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerHTML))
	})
}

const swaggerHTML = `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        SwaggerUIBundle({ url: "/openapi.json", dom_id: "#swagger-ui" });
      };
    </script>
  </body>
</html>`
