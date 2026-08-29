package httpx

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"golol/internal/champions"
	"golol/internal/items"
)

type frame struct {
	Version  string
	Title    string
	BrandSub string
	Active   string
}

type page struct {
	frame
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

type abilityView struct {
	Key      string
	Name     string
	IconURL  string
	Cooldown string
	Cost     string
	Range    string
	HTML     template.HTML
	Scaling  []champions.ScaleRow
	Forms    []champions.AbilityForm
}

type champPage struct {
	frame
	SelectedRoles map[champions.Role]bool
	Roles         []champions.RoleOption
	Champions     []champions.Champion
	Count         int
	Champion      champions.Champion
	Abilities     []abilityView
}

type buildPage struct {
	frame
	Champions  []champions.Champion
	ChampCount int
	Groups     []items.Group
	ItemCount  int
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if !s.ready.Load() {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}
	cat := s.catalog.Load()
	champs := s.champions.Load()
	if cat == nil || len(cat.Items) == 0 || champs == nil || len(champs.Champions) == 0 {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"version":   cat.Version,
		"items":     len(cat.Items),
		"champions": len(champs.Champions),
	})
}

func (s *Server) items(w http.ResponseWriter, r *http.Request) {
	cat := s.catalogOrEmpty()
	q := parseQuery(r)
	filtered := items.Filter(cat.Items, q)
	p := page{
		frame:         itemsFrame(cat.Version),
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

func (s *Server) championsPage(w http.ResponseWriter, r *http.Request) {
	cat := s.championsOrEmpty()
	roles := champions.ParseRoles(r.URL.Query()["role"])
	filtered := champions.Filter(cat.Champions, roles)
	p := champPage{
		frame:         champsFrame(cat.Version),
		SelectedRoles: selectedRoles(roles),
		Roles:         champions.AllRoles,
		Champions:     filtered,
		Count:         len(filtered),
	}
	name := "champions"
	if r.Header.Get("HX-Request") == "true" {
		name = "champion-grid"
	}
	s.render(w, name, p)
}

func (s *Server) buildsPage(w http.ResponseWriter, r *http.Request) {
	itemsCat := s.catalogOrEmpty()
	champsCat := s.championsOrEmpty()
	version := itemsCat.Version
	if version == "" {
		version = champsCat.Version
	}
	p := buildPage{
		frame:      buildsFrame(version),
		Champions:  champsCat.Champions,
		ChampCount: len(champsCat.Champions),
		Groups:     items.GroupByTier(itemsCat.Items),
		ItemCount:  len(itemsCat.Items),
	}
	s.render(w, "builds", p)
}

func (s *Server) championDetail(w http.ResponseWriter, r *http.Request) {
	cat := s.championsOrEmpty()
	id := strings.TrimSpace(r.PathValue("id"))
	ch, ok := cat.Get(id)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		s.render(w, "champion-detail-missing", champPage{})
		return
	}
	p := champPage{
		frame:     champsFrame(cat.Version),
		Champion:  ch,
		Abilities: s.abilityViews(ch),
	}
	s.render(w, "champion-detail", p)
}

func (s *Server) abilityViews(ch champions.Champion) []abilityView {
	out := make([]abilityView, 0, 1+len(ch.Spells))
	if ch.Passive.Name != "" || ch.Passive.Description != "" {
		out = append(out, s.abilityView(ch.Passive))
	}
	for _, sp := range ch.Spells {
		out = append(out, s.abilityView(sp))
	}
	return out
}

func (s *Server) abilityView(a champions.Ability) abilityView {
	return abilityView{
		Key:      a.Key,
		Name:     a.Name,
		IconURL:  a.IconURL,
		Cooldown: a.Cooldown,
		Cost:     a.Cost,
		Range:    a.Range,
		HTML:     template.HTML(sanitizeDescription(s.policy, a.Description)),
		Scaling:  a.Scaling,
		Forms:    a.Forms,
	}
}

func (s *Server) render(w http.ResponseWriter, name string, p any) {
	var buf strings.Builder
	if err := s.templates.ExecuteTemplate(&buf, name, p); err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(buf.String()))
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

func selectedRoles(roles []champions.Role) map[champions.Role]bool {
	out := make(map[champions.Role]bool, len(roles))
	for _, r := range roles {
		out[r] = true
	}
	return out
}

func itemsFrame(version string) frame {
	return frame{
		Version:  version,
		Title:    "golol — Objetos",
		BrandSub: "tienda de objetos",
		Active:   "items",
	}
}

func champsFrame(version string) frame {
	return frame{
		Version:  version,
		Title:    "golol — Campeones",
		BrandSub: "campeones",
		Active:   "champions",
	}
}

func buildsFrame(version string) frame {
	return frame{
		Version:  version,
		Title:    "golol — Builds",
		BrandSub: "creador de builds",
		Active:   "builds",
	}
}
