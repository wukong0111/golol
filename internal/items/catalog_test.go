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
				"description":"<stats>+15</stats>",
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
