package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golol/internal/ddragon"
	"golol/internal/items"
)

func testCatalog(t *testing.T) *items.Catalog {
	t.Helper()
	raw := []byte(`{
		"version":"16.16.1",
		"data":{
			"1029":{
				"name":"Armadura de tela",
				"description":"<stats><attention>15</attention> Armadura</stats>",
				"plaintext":"Un poco de armadura",
				"gold":{"purchasable":true,"total":300},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"1029.png"}
			},
			"1042":{
				"name":"Daga",
				"description":"<stats>AS</stats>",
				"plaintext":"Velocidad",
				"gold":{"purchasable":true,"total":250},
				"tags":["AttackSpeed"],
				"maps":{"11":true},
				"image":{"full":"1042.png"}
			},
			"3082":{
				"name":"Guardabrazos",
				"description":"<stats>armadura y as</stats>",
				"from":["1029"],
				"gold":{"purchasable":true,"total":800},
				"tags":["Armor","AttackSpeed"],
				"maps":{"11":true},
				"depth":2,
				"image":{"full":"3082.png"}
			}
		}
	}`)
	cat, err := items.Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	return cat
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	s, err := New(testCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	return s.Handler()
}

func TestRootRedirects(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/items" {
		t.Fatalf("location %s", loc)
	}
}

func TestItemsPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Armadura de tela", "Daga", "golol", "Parche 16.16.1", "Riot Games"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in page", want)
		}
	}
}

func TestItemsHTMXPartial(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items?stat=Armor", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatal("partial should not be a full page")
	}
	if !strings.Contains(body, "Armadura de tela") {
		t.Fatal("armor item missing")
	}
	if strings.Contains(body, "Daga") {
		t.Fatal("dagger should be filtered out")
	}
	if !strings.Contains(body, "2 objeto") && !strings.Contains(body, "2 objetos") {
		// cloth + wardens: both have Armor
		if !strings.Contains(body, "objeto") {
			t.Fatalf("count missing: %s", body)
		}
	}
}

func TestItemsStatAND(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items?stat=Armor&stat=AttackSpeed", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "Guardabrazos") {
		t.Fatal("intersection item missing")
	}
	if strings.Contains(body, "Armadura de tela") || strings.Contains(body, "Daga") {
		t.Fatalf("AND leaked extras: %s", body)
	}
}

func TestItemsPageChecksSelectedRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items?role=tank", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `value="tank" checked`) {
		t.Fatal("tank radio should be checked")
	}
	if strings.Contains(body, `value="" checked`) {
		t.Fatal("Todos should not stay checked when role=tank")
	}
}

func TestItemsRoleTank(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items?role=tank", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "Daga") {
		t.Fatal("dagger is not a tank item")
	}
	if !strings.Contains(body, "Armadura de tela") {
		t.Fatal("cloth armor should be tank")
	}
}

func TestItemDetail(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items/1029", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Armadura de tela") || !strings.Contains(body, "300 oro") {
		t.Fatalf("detail: %s", body)
	}
	if !strings.Contains(body, `class="attention"`) {
		t.Fatalf("riot stats not rewritten: %s", body)
	}
	if strings.Contains(body, "<script") {
		t.Fatal("script leaked into sanitized HTML")
	}
}

func TestItemDetailNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items/99999", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestStaticCSS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/static/css/app.css", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "--gold") {
		t.Fatal("css not served")
	}
}
