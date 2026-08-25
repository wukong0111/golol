package httpx

import (
	"html/template"
	"io/fs"
	"net/http"
	"sync/atomic"

	"github.com/microcosm-cc/bluemonday"

	"golol/internal/items"
	"golol/web"
)

// Server serves the /items shop.
type Server struct {
	catalog   atomic.Pointer[items.Catalog]
	ready     atomic.Bool
	templates *template.Template
	policy    *bluemonday.Policy
}

// New constructs the HTTP handler with an initial catalog.
func New(cat *items.Catalog) (*Server, error) {
	tmpl, err := template.ParseFS(web.Templates, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	s := &Server{
		templates: tmpl,
		policy:    descriptionPolicy(),
	}
	s.SetCatalog(cat)
	s.SetReady(true)
	return s, nil
}

// SetCatalog swaps the live inventory (used on Data Dragon refresh).
func (s *Server) SetCatalog(cat *items.Catalog) {
	if cat != nil {
		s.catalog.Store(cat)
	}
}

// SetReady toggles the health endpoint. False during graceful shutdown so
// Railway (or any proxy) stops sending new work to this instance.
func (s *Server) SetReady(ok bool) {
	s.ready.Store(ok)
}

func (s *Server) catalogOrEmpty() *items.Catalog {
	if cat := s.catalog.Load(); cat != nil {
		return cat
	}
	return &items.Catalog{ByID: map[string]items.Item{}}
}

// Handler returns the mux for ListenAndServe.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.redirectRoot)
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /items", s.items)
	mux.HandleFunc("GET /items/{id}", s.detail)

	static, err := fs.Sub(web.Static, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	return mux
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusFound)
}
