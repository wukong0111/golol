package httpx

import (
	"encoding/json"
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
				"description":"<stats><attention>15</attention> de armadura</stats>",
				"plaintext":"Un poco de armadura",
				"gold":{"purchasable":true,"total":300},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"1029.png"}
			},
			"1042":{
				"name":"Daga",
				"description":"<stats><attention>10%</attention> de velocidad de ataque</stats>",
				"plaintext":"Velocidad",
				"gold":{"purchasable":true,"total":250},
				"tags":["AttackSpeed"],
				"maps":{"11":true},
				"image":{"full":"1042.png"}
			},
			"3082":{
				"name":"Guardabrazos",
				"description":"<stats><attention>40</attention> de armadura<br><attention>15%</attention> de velocidad de ataque</stats>",
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
					{"id":"AatroxQ","name":"La Espada de los Oscuros","description":"<physicalDamage>golpe</physicalDamage>","cooldownBurn":"14/12/10/8/6","costBurn":"0","resource":"Sin coste","rangeBurn":"25000","image":{"full":"AatroxQ.png"}},
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
	if err := cat.ApplyKits([]byte(`{
		"Aatrox":{
			"key":"Aatrox",
			"abilities":{
				"P":[{
					"name":"Deathbringer Stance",
					"effects":[{"description":"deal bonus magic damage equal to 4% : 8% (based on level) of the target's maximum health"}]
				}],
				"Q":[{
					"name":"The Darkin Blade",
					"effects":[{
						"leveling":[{
							"attribute":"Physical Damage",
							"modifiers":[
								{"values":[10,25,40,55,70],"units":["","","","",""]},
								{"values":[60,67.5,75,82.5,90],"units":["% AD","% AD","% AD","% AD","% AD"]}
							]
						}]
					}]
				}]
			}
		}
	}`)); err != nil {
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
	if !strings.Contains(body, `href="/builds"`) {
		t.Fatal("builds nav missing on champions page")
	}
	if !strings.Contains(body, `id="champ-roster"`) {
		t.Fatal("roster disclosure missing")
	}
	if !strings.Contains(body, "Mostrar") || !strings.Contains(body, "Ocultar") || !strings.Contains(body, "Cambiar") {
		t.Fatal("roster toggle labels missing")
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

func TestChampionsRoleAND(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/champions?role=fighter&role=mage", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "Aatrox") || strings.Contains(body, "Ahri") {
		t.Fatalf("AND leaked: %s", body)
	}
	if !strings.Contains(body, "Ningún campeón combina esos roles") {
		t.Fatalf("empty copy missing: %s", body)
	}

	req = httptest.NewRequest(http.MethodGet, "/champions?role=fighter&role=tank", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	body = rec.Body.String()
	if !strings.Contains(body, "Aatrox") {
		t.Fatal("fighter+tank should keep Aatrox")
	}
	if strings.Contains(body, "Ahri") {
		t.Fatal("Ahri should be filtered out")
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
		"Sin coste",
		"Daño físico",
		"10 / 25 / 40 / 55 / 70",
		"90% DA",
		"Daño mágico",
		"4% – 8%",
		"vida máxima del objetivo",
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
	if !strings.Contains(body, `href="/builds"`) {
		t.Fatal("builds nav missing on items page")
	}
}

func TestBuildsPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/builds", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		"golol — Builds",
		"creador de builds",
		`href="/builds" class="is-on"`,
		`id="add-build"`,
		"Añadir build",
		"Buscar campeón",
		"Buscar objeto",
		`data-item-slots="7"`,
		`data-kind="champion"`,
		`data-kind="item"`,
		`data-id="Aatrox"`,
		`data-id="Ahri"`,
		"Armadura de tela",
		"Daga",
		`id="builds-list"`,
		`src="/static/js/builds.js"`,
		"Colecciones",
		"Parche 16.16.1",
		"Selecciona una build para editarla",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in builds page", want)
		}
	}
	if strings.Contains(body, `href="/items" class="is-on"`) || strings.Contains(body, `href="/champions" class="is-on"`) {
		t.Fatal("other nav items should not be active on builds")
	}

	raw := jsonFromScript(t, body, "item-bonuses")
	var idx map[string][]items.Bonus
	if err := json.Unmarshal([]byte(raw), &idx); err != nil {
		t.Fatalf("item-bonuses json: %v (%s)", err, raw)
	}
	if len(idx["1029"]) != 1 || idx["1029"][0].Amount != 15 || idx["1029"][0].Name != "de armadura" {
		t.Fatalf("cloth bonuses: %+v", idx["1029"])
	}
	if len(idx["1042"]) != 1 || !idx["1042"][0].Percent || idx["1042"][0].Amount != 10 {
		t.Fatalf("dagger bonuses: %+v", idx["1042"])
	}
	if len(idx["3082"]) != 2 {
		t.Fatalf("warden bonuses: %+v", idx["3082"])
	}
}

func jsonFromScript(t *testing.T, body, id string) string {
	t.Helper()
	marker := `id="` + id + `"`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("script %s missing", id)
	}
	gt := strings.Index(body[i:], ">")
	if gt < 0 {
		t.Fatalf("script %s unclosed open tag", id)
	}
	start := i + gt + 1
	end := strings.Index(body[start:], "</script>")
	if end < 0 {
		t.Fatalf("script %s missing close", id)
	}
	return strings.TrimSpace(body[start : start+end])
}

func TestBuildsJS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/static/js/builds.js", nil)
	rec := httptest.NewRecorder()
	testHandler(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	src := string(body)
	for _, want := range []string{`golol.builds`, `ITEM_SLOTS`, `localStorage`, `item-bonuses`, `build-totals`, `Mejoras`, `item-flyout`, `/items/`} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in builds.js", want)
		}
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
	css := string(body)
	if !strings.Contains(css, "--gold") {
		t.Fatal("css not served")
	}
	if !strings.Contains(css, ".item[hidden]") || !strings.Contains(css, ".champ[hidden]") {
		t.Fatal("picker search needs [hidden] to beat display:flex")
	}
	if !strings.Contains(css, ".item-flyout") || !strings.Contains(css, ".item-flyout[hidden]") {
		t.Fatal("build item flyout missing")
	}
}
