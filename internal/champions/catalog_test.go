package champions

import (
	"testing"

	"golol/internal/ddragon"
)

func sampleRoster(t *testing.T) *Catalog {
	t.Helper()
	raw := []byte(`{
		"type":"champion","version":"16.16.1",
		"data":{
			"Ahri":{
				"id":"Ahri","name":"Ahri","title":"la Zorro de Nueve Colas",
				"tags":["Mage","Assassin"],
				"image":{"full":"Ahri.png"},
				"passive":{"name":"Absorción del alma","description":"<healing>cura</healing>","image":{"full":"Ahri_SoulEater.png"}},
				"spells":[
					{"id":"AhriQ","name":"Orbe de engaño","description":"<magicDamage>daño</magicDamage>","cooldownBurn":"7","image":{"full":"AhriQ.png"}},
					{"id":"AhriW","name":"Fuego raposo","description":"zorros","cooldownBurn":"9/8/7/6/5","image":{"full":"AhriW.png"}},
					{"id":"AhriE","name":"Encanto","description":"encanta","cooldownBurn":"12","image":{"full":"AhriE.png"}},
					{"id":"AhriR","name":"Impulso espiritual","description":"dash","cooldownBurn":"130/105/80","image":{"full":"AhriR.png"}}
				]
			},
			"Aatrox":{
				"id":"Aatrox","name":"Aatrox","title":"la Espada de los Oscuros",
				"tags":["Fighter","Tank"],
				"image":{"full":"Aatrox.png"},
				"passive":{"name":"Aspecto de la muerte","description":"cura","image":{"full":"Aatrox_Passive.png"}},
				"spells":[
					{"id":"AatroxQ","name":"La Espada de los Oscuros","description":"golpe","cooldownBurn":"14/12/10/8/6","image":{"full":"AatroxQ.png"}}
				]
			},
			"blank":{
				"id":"Nope","name":"   ","tags":["Mage"]
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	return cat
}

func TestParseRoster(t *testing.T) {
	cat := sampleRoster(t)
	if len(cat.Champions) != 2 {
		t.Fatalf("expected 2 champs, got %d", len(cat.Champions))
	}
	if cat.Champions[0].ID != "Aatrox" || cat.Champions[1].ID != "Ahri" {
		t.Fatalf("sort: %+v %+v", cat.Champions[0], cat.Champions[1])
	}

	aatrox, ok := cat.Get("Aatrox")
	if !ok {
		t.Fatal("missing Aatrox")
	}
	if aatrox.Title != "la Espada de los Oscuros" {
		t.Fatalf("title: %s", aatrox.Title)
	}
	if aatrox.IconURL != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/champion/Aatrox.png" {
		t.Fatalf("icon: %s", aatrox.IconURL)
	}
	if aatrox.SplashURL != "https://ddragon.leagueoflegends.com/cdn/img/champion/splash/Aatrox_0.jpg" {
		t.Fatalf("splash: %s", aatrox.SplashURL)
	}
	if len(aatrox.Roles) != 2 || aatrox.Roles[0] != RoleFighter || aatrox.Roles[1] != RoleTank {
		t.Fatalf("roles: %+v", aatrox.Roles)
	}
	if len(aatrox.RoleLabels) != 2 || aatrox.RoleLabels[0] != "Luchador" || aatrox.RoleLabels[1] != "Tanque" {
		t.Fatalf("labels: %+v", aatrox.RoleLabels)
	}
	if aatrox.Passive.Key != "P" || aatrox.Passive.Name != "Aspecto de la muerte" {
		t.Fatalf("passive: %+v", aatrox.Passive)
	}
	if aatrox.Passive.IconURL != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/passive/Aatrox_Passive.png" {
		t.Fatalf("passive icon: %s", aatrox.Passive.IconURL)
	}
	if len(aatrox.Spells) != 1 || aatrox.Spells[0].Key != "Q" || aatrox.Spells[0].Cooldown != "14/12/10/8/6" {
		t.Fatalf("spells: %+v", aatrox.Spells)
	}
	if aatrox.Spells[0].Cost != "" || aatrox.Spells[0].Range != "" {
		t.Fatalf("empty meta: %+v", aatrox.Spells[0])
	}

	ahri, ok := cat.Get("Ahri")
	if !ok {
		t.Fatal("missing Ahri")
	}
	if len(ahri.Spells) != 4 || ahri.Spells[3].Key != "R" {
		t.Fatalf("ahri kit: %+v", ahri.Spells)
	}
}

func TestParseSpellCostAndRange(t *testing.T) {
	raw := []byte(`{
		"type":"champion","version":"16.16.1",
		"data":{
			"Ahri":{
				"id":"Ahri","name":"Ahri","partype":"Maná","tags":["Mage"],
				"image":{"full":"Ahri.png"},
				"passive":{"name":"P","description":"p","image":{"full":"P.png"}},
				"spells":[
					{"id":"AhriQ","name":"Orbe","description":"q","cooldownBurn":"7",
					 "costBurn":"55/65/75/85/95","resource":"{{ cost }} {{ abilityresourcename }}",
					 "rangeBurn":"970","image":{"full":"AhriQ.png"}},
					{"id":"AhriW","name":"Fuego","description":"w","cooldownBurn":"9",
					 "costBurn":"0","resource":"Sin coste","rangeBurn":"25000","image":{"full":"AhriW.png"}}
				]
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	ahri, _ := cat.Get("Ahri")
	if ahri.Spells[0].Cost != "55/65/75/85/95 Maná" {
		t.Fatalf("cost: %q", ahri.Spells[0].Cost)
	}
	if ahri.Spells[0].Range != "970" {
		t.Fatalf("range: %q", ahri.Spells[0].Range)
	}
	if ahri.Spells[1].Cost != "Sin coste" {
		t.Fatalf("no cost: %q", ahri.Spells[1].Cost)
	}
	if ahri.Spells[1].Range != "" {
		t.Fatalf("dummy range leaked: %q", ahri.Spells[1].Range)
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, []byte(`{`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetMissing(t *testing.T) {
	if _, ok := sampleRoster(t).Get("Zed"); ok {
		t.Fatal("Zed should be missing")
	}
	var nilCat *Catalog
	if _, ok := nilCat.Get("Aatrox"); ok {
		t.Fatal("nil catalog")
	}
}
