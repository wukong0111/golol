package httpx

import (
	"html/template"
	"io/fs"
	"net/http"
	"sync/atomic"

	"github.com/microcosm-cc/bluemonday"

	"golol/internal/champions"
	"golol/internal/items"
	"golol/web"
)

// Server serves the /items shop and the /champions roster.
type Server struct {
	catalog   atomic.Pointer[items.Catalog]
	champions atomic.Pointer[champions.Catalog]
	ready     atomic.Bool
	templates *template.Template
	policy    *bluemonday.Policy
}

// New constructs the HTTP handler with initial catalogs.
func New(cat *items.Catalog, champs *champions.Catalog) (*Server, error) {
	tmpl, err := template.ParseFS(web.Templates, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	s := &Server{
		templates: tmpl,
		policy:    descriptionPolicy(),
	}
	s.SetCatalog(cat)
	s.SetChampions(champs)
	s.SetReady(true)
	return s, nil
}

// SetCatalog swaps the live inventory (used on Data Dragon refresh).
func (s *Server) SetCatalog(cat *items.Catalog) {
	if cat != nil {
		s.catalog.Store(cat)
	}
}

// SetChampions swaps the live roster (used on Data Dragon refresh).
func (s *Server) SetChampions(cat *champions.Catalog) {
	if cat != nil {
		s.champions.Store(cat)
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

func (s *Server) championsOrEmpty() *champions.Catalog {
	if cat := s.champions.Load(); cat != nil {
		return cat
	}
	return &champions.Catalog{ByID: map[string]champions.Champion{}}
}

// Handler returns the mux for ListenAndServe.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.redirectRoot)
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /items", s.items)
	mux.HandleFunc("GET /items/{id}", s.detail)
	mux.HandleFunc("GET /champions", s.championsPage)
	mux.HandleFunc("GET /champions/{id}", s.championDetail)

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
