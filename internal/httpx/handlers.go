package httpx

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"golol/internal/items"
)

type page struct {
	Version       string
	Role          items.Role
	SelectedStats map[string]bool
	Roles         []items.RoleOption
	Stats         []items.StatFilter
	Groups        []items.Group
	Count         int
	Item          items.Item
	HTML          template.HTML
	Components    []items.Item
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.ready.Load() {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}
	cat := s.catalog.Load()
	if cat == nil || len(cat.Items) == 0 {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"version": cat.Version,
		"items":   len(cat.Items),
	})
}

func (s *Server) items(w http.ResponseWriter, r *http.Request) {
	cat := s.catalogOrEmpty()
	q := parseQuery(r)
	filtered := items.Filter(cat.Items, q)
	p := page{
		Version:       cat.Version,
		Role:          q.Role,
		SelectedStats: selectedMap(q.Stats),
		Roles:         items.AllRoles,
		Stats:         items.AllStatFilters,
		Groups:        items.GroupByTier(filtered),
		Count:         len(filtered),
	}
	name := "items"
	if r.Header.Get("HX-Request") == "true" {
		name = "grid"
	}
	s.render(w, name, p)
}

func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	cat := s.catalogOrEmpty()
	id := strings.TrimSpace(r.PathValue("id"))
	it, ok := cat.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		s.render(w, "detail-missing", page{})
		return
	}
	p := page{
		Item:       it,
		HTML:       template.HTML(sanitizeDescription(s.policy, it.Description)),
		Components: cat.Components(it),
	}
	s.render(w, "detail", p)
}

func (s *Server) render(w http.ResponseWriter, name string, p page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, p); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

func parseQuery(r *http.Request) items.Query {
	return items.Query{
		Role:  items.ParseRole(r.URL.Query().Get("role")),
		Stats: items.ParseStats(r.URL.Query()["stat"]),
	}
}

func selectedMap(stats []items.StatFilter) map[string]bool {
	out := make(map[string]bool, len(stats))
	for _, s := range stats {
		out[s.ID] = true
	}
	return out
}
