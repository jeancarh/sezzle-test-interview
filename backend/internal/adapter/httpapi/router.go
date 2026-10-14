package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jeancarh/sezzle-test-interview/backend/internal/calculator"
)

type Deps struct {
	Calculator calculator.Calculator
	Assets     fs.FS
	Spec       []byte
}

func NewRouter(d Deps) http.Handler {
	a := &api{calc: d.Calculator}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("POST /api/v1/calculations", a.calculate)
	mux.HandleFunc("GET /api/v1/operations", a.listOperations)
	mux.HandleFunc("GET /api/openapi.yaml", serveSpec(d.Spec))
	mux.HandleFunc("GET /docs", serveDocs)
	mux.Handle("/", fallback(spa(d.Assets)))

	return withRequestID(withLogging(withRecover(mux)))
}

func serveSpec(spec []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec)
	}
}

const docsPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Calculator API docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({ url: "/api/openapi.yaml", dom_id: "#swagger" });</script>
</body>
</html>`

func serveDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docsPage))
}

func fallback(ui http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if isKnownAPIPath(r.URL.Path) {
				writeProblem(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, r.Method+" is not supported on "+r.URL.Path)
				return
			}
			writeProblem(w, http.StatusNotFound, codeNotFound, "no such endpoint")
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeProblem(w, http.StatusMethodNotAllowed, codeMethodNotAllowed, r.Method+" is not supported here")
			return
		}
		ui.ServeHTTP(w, r)
	})
}

func isKnownAPIPath(path string) bool {
	switch path {
	case "/api/v1/calculations", "/api/v1/operations", "/api/openapi.yaml":
		return true
	}
	return false
}

func spa(assets fs.FS) http.Handler {
	files := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(assets, name); err == nil && !info.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(assets, "index.html")
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "The UI is not built into this binary. Use `make dev` or `make up` instead.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}
