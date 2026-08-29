package items

import (
	"testing"

	"golol/internal/ddragon"
)

func TestParseSummonersRiftOnly(t *testing.T) {
	raw := []byte(`{
		"type":"item","version":"16.16.1",
		"data":{
			"1029":{
				"name":"Armadura de tela",
				"description":"<stats><attention>15</attention> de armadura</stats>",
				"plaintext":"armadura",
				"image":{"full":"1029.png"},
				"gold":{"base":300,"purchasable":true,"total":300,"sell":210},
				"tags":["Armor"],
				"maps":{"11":true,"12":true},
				"stats":{"FlatArmorMod":15}
			},
			"arena":{
				"name":"Arena Only",
				"gold":{"purchasable":true,"total":1},
				"tags":["Armor"],
				"maps":{"11":false,"30":true},
				"image":{"full":"1.png"}
			},
			"hidden":{
				"name":"Hidden",
				"gold":{"purchasable":true,"total":1},
				"tags":["Armor"],
				"maps":{"11":true},
				"hideFromAll":true,
				"image":{"full":"2.png"}
			},
			"notbuy":{
				"name":"Not purchasable",
				"gold":{"purchasable":false,"total":1},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"3.png"}
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Items) != 1 || cat.Items[0].ID != "1029" {
		t.Fatalf("expected only cloth armor, got %+v", cat.Items)
	}
	if cat.Items[0].IconURL != "https://ddragon.leagueoflegends.com/cdn/16.16.1/img/item/1029.png" {
		t.Fatalf("icon url: %s", cat.Items[0].IconURL)
	}
	if !cat.Items[0].Has("Armor") {
		t.Fatal("tags")
	}
	if len(cat.Items[0].Bonuses) != 1 || cat.Items[0].Bonuses[0].Amount != 15 || cat.Items[0].Bonuses[0].Name != "de armadura" {
		t.Fatalf("bonuses: %+v", cat.Items[0].Bonuses)
	}
}

func TestParseInStoreDefaultTrue(t *testing.T) {
	raw := []byte(`{
		"data":{
			"1":{
				"name":"Boots",
				"gold":{"purchasable":true,"total":300},
				"tags":["Boots"],
				"maps":{"11":true},
				"image":{"full":"1001.png"}
			},
			"2":{
				"name":"Out of shop",
				"inStore":false,
				"gold":{"purchasable":true,"total":300},
				"tags":["Boots"],
				"maps":{"11":true},
				"image":{"full":"x.png"}
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Items) != 1 || cat.Items[0].Name != "Boots" {
		t.Fatalf("inStore default: %+v", cat.Items)
	}
}

func TestParseDropsModeCopies(t *testing.T) {
	raw := []byte(`{
		"data":{
			"1029":{
				"name":"Armadura de tela",
				"gold":{"purchasable":true,"total":300},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"1029.png"}
			},
			"6631":{
				"name":"Cortasendas",
				"gold":{"purchasable":true,"total":3300},
				"tags":["Damage"],
				"maps":{"11":true},
				"depth":3,
				"image":{"full":"6631.png"}
			},
			"663193":{
				"name":"Protector pétreo de gárgola",
				"gold":{"purchasable":true,"total":2500},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"3193.png"}
			},
			"323003":{
				"name":"Bastón del arcángel",
				"gold":{"purchasable":true,"total":2900},
				"tags":["SpellDamage"],
				"maps":{"11":true},
				"depth":3,
				"image":{"full":"3003.png"}
			},
			"2051":{
				"name":"Cuerno del guardián",
				"gold":{"purchasable":true,"total":950},
				"tags":["Health"],
				"maps":{"11":true,"12":true},
				"image":{"full":"2051.png"}
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Items) != 2 {
		t.Fatalf("expected cloth + stridebreaker, got %+v", cat.Items)
	}
	if _, ok := cat.Get("1029"); !ok {
		t.Fatal("cloth armor is a Rift item")
	}
	if _, ok := cat.Get("6631"); !ok {
		t.Fatal("4-digit legendary must stay")
	}
	for _, id := range []string{"663193", "323003", "2051"} {
		if _, ok := cat.Get(id); ok {
			t.Fatalf("mode copy %s leaked into the shop", id)
		}
	}
}

func TestComponents(t *testing.T) {
	raw := []byte(`{
		"data":{
			"1029":{
				"name":"Cloth",
				"gold":{"purchasable":true,"total":300},
				"tags":["Armor"],
				"maps":{"11":true},
				"image":{"full":"1029.png"}
			},
			"1031":{
				"name":"Chain Vest",
				"from":["1029"],
				"gold":{"purchasable":true,"total":800},
				"tags":["Armor"],
				"maps":{"11":true},
				"depth":2,
				"image":{"full":"1031.png"}
			}
		}
	}`)
	cat, err := Parse("16.16.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	vest, ok := cat.Get("1031")
	if !ok {
		t.Fatal("missing vest")
	}
	parts := cat.Components(vest)
	if len(parts) != 1 || parts[0].ID != "1029" {
		t.Fatalf("components: %+v", parts)
	}
}
