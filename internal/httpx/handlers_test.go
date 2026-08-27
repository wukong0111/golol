package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golol/internal/champions"
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

func testChampCatalog(t *testing.T) *champions.Catalog {
	t.Helper()
	raw := []byte(`{
		"version":"16.16.1",
		"data":{
			"Aatrox":{
				"id":"Aatrox","name":"Aatrox","title":"la Espada de los Oscuros",
				"tags":["Fighter","Tank"],
				"image":{"full":"Aatrox.png"},
				"passive":{"name":"Aspecto de la muerte","description":"<healing>cura</healing>","image":{"full":"Aatrox_Passive.png"}},
				"spells":[
					{"id":"AatroxQ","name":"La Espada de los Oscuros","description":"<physicalDamage>golpe</physicalDamage>","cooldownBurn":"14/12/10/8/6","image":{"full":"AatroxQ.png"}},
					{"id":"AatroxW","name":"Cadenas infernales","description":"cadena","cooldownBurn":"20","image":{"full":"AatroxW.png"}},
					{"id":"AatroxE","name":"Deslizamiento sombrío","description":"dash","cooldownBurn":"9","image":{"full":"AatroxE.png"}},
					{"id":"AatroxR","name":"El Aniquilador de mundos","description":"<script>alert(1)</script>ulti","cooldownBurn":"120/100/80","image":{"full":"AatroxR.png"}}
				]
			},
			"Ahri":{
				"id":"Ahri","name":"Ahri","title":"la Zorro de Nueve Colas",
				"tags":["Mage","Assassin"],
				"image":{"full":"Ahri.png"},
				"passive":{"name":"Absorción del alma","description":"orbes","image":{"full":"AhriP.png"}},
				"spells":[
					{"id":"AhriQ","name":"Orbe de engaño","description":"orbe","image":{"full":"AhriQ.png"}}
				]
			}
		}
	}`)
	cat, err := champions.Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	return cat
}

func testServer(t *testing.T) *Server {
	t.Helper()
	s, err := New(testCatalog(t), testChampCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	return testServer(t).Handler()
}

func TestHealthOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"ok":true`) || !strings.Contains(body, `"items":3`) || !strings.Contains(body, `"champions":2`) {
		t.Fatalf("health body: %s", body)
	}
}

func TestHealthNotReady(t *testing.T) {
	s := testServer(t)
	s.SetReady(false)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", rec.Code)
	}
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

func TestChampionsPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Aatrox", "Ahri", "golol — Campeones", "Campeones", "Parche 16.16.1", "Luchador", "2 campeones"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in page", want)
		}
	}
	if !strings.Contains(body, `href="/champions" class="is-on"`) {
		t.Fatal("champions nav should be active")
	}
}

func TestChampionsHTMXPartial(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions?role=fighter", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatal("partial should not be a full page")
	}
	if !strings.Contains(body, "Aatrox") {
		t.Fatal("fighter missing")
	}
	if strings.Contains(body, "Ahri") {
		t.Fatal("Ahri should be filtered out")
	}
	if !strings.Contains(body, "1 campeón") {
		t.Fatalf("count missing: %s", body)
	}
}

func TestChampionsRoleOR(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions?role=fighter&role=mage", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "Aatrox") || !strings.Contains(body, "Ahri") {
		t.Fatalf("OR leaked: %s", body)
	}
}

func TestChampionsPageChecksSelectedRole(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions?role=tank", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `value="tank" checked`) {
		t.Fatal("tank checkbox should be checked")
	}
	if strings.Contains(body, `value="mage" checked`) {
		t.Fatal("mage should not be checked")
	}
}

func TestChampionDetail(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions/Aatrox", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"Aatrox",
		"la Espada de los Oscuros",
		"splash/Aatrox_0.jpg",
		"Aspecto de la muerte",
		"La Espada de los Oscuros",
		"Cadenas infernales",
		"Deslizamiento sombrío",
		"El Aniquilador de mundos",
		`class="healing"`,
		`class="physicaldamage"`,
		"Luchador",
		"Tanque",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in detail", want)
		}
	}
	if strings.Contains(body, "<script") || strings.Contains(body, "alert(1)") {
		t.Fatal("script leaked into sanitized HTML")
	}
}

func TestChampionDetailNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions/Nope", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestItemsPageHasNav(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `href="/items" class="is-on"`) {
		t.Fatal("items nav should be active")
	}
	if !strings.Contains(body, `href="/champions"`) {
		t.Fatal("champions nav missing on items page")
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
