package items

import (
	"testing"

	"golol/internal/ddragon"
)

func TestApplyShopRolesEssenceReaver(t *testing.T) {
	cat := parseRiftItem(t, "3508", "Segador de esencia", "Damage", "CriticalStrike", "ManaRegen")
	it, _ := cat.Get("3508")
	if !it.MatchesRole(RoleSupport) || !it.MatchesRole(RoleMarksman) {
		t.Fatal("precondition: heuristic should put Essence Reaver in support+marksman")
	}

	if err := cat.ApplyShopRoles([]byte(`{
		"3508":{"shop":{"tags":["MARKSMAN","MANA_AND_REG","ONHIT_EFFECTS"]}}
	}`)); err != nil {
		t.Fatal(err)
	}
	it, _ = cat.Get("3508")
	if it.MatchesRole(RoleSupport) {
		t.Fatal("Essence Reaver should not be support after shop.tags")
	}
	if !it.MatchesRole(RoleMarksman) {
		t.Fatal("Essence Reaver should stay marksman")
	}
	if it.MatchesRole(RoleMage) {
		t.Fatal("MANA_AND_REG is a stat filter, not a class")
	}
}

func TestApplyShopRolesIgnoresStatTagsAndKeepsOrder(t *testing.T) {
	cat := catalogWith(itemWith("1028", "Ruby Crystal", "Health"))
	if err := cat.ApplyShopRoles([]byte(`{
		"1028":{"shop":{"tags":["SUPPORT","TANK","FIGHTER","MARKSMAN","ASSASSIN","MAGE"]}}
	}`)); err != nil {
		t.Fatal(err)
	}
	it, _ := cat.Get("1028")
	want := []Role{RoleFighter, RoleMarksman, RoleAssassin, RoleMage, RoleTank, RoleSupport}
	if len(it.ShopRoles) != len(want) {
		t.Fatalf("roles: %+v", it.ShopRoles)
	}
	for i, r := range want {
		if it.ShopRoles[i] != r {
			t.Fatalf("order %d: %s != %s", i, it.ShopRoles[i], r)
		}
		if !it.MatchesRole(r) {
			t.Fatalf("missing %s", r)
		}
	}
}

func TestApplyShopRolesMissingOrEmptyFallsBackToHeuristic(t *testing.T) {
	cat := catalogWith(
		itemWith("1004", "Faerie", "ManaRegen"),
		itemWith("1001", "Boots", "Boots"),
	)
	if err := cat.ApplyShopRoles([]byte(`{
		"1001":{"shop":{"tags":["BOOTS"]}},
		"9999":{"shop":{"tags":["SUPPORT"]}}
	}`)); err != nil {
		t.Fatal(err)
	}
	faerie, _ := cat.Get("1004")
	if !faerie.MatchesRole(RoleSupport) {
		t.Fatal("missing meraki id should keep ManaRegen support heuristic")
	}
	if len(faerie.ShopRoles) != 0 {
		t.Fatalf("shop roles: %+v", faerie.ShopRoles)
	}
	boots, _ := cat.Get("1001")
	if len(boots.ShopRoles) != 0 {
		t.Fatal("stat-only shop.tags should not replace the heuristic")
	}
	if boots.MatchesRole(RoleSupport) || boots.MatchesRole(RoleMage) {
		t.Fatal("boots should not pick up a class from BOOTS")
	}
}

func TestApplyShopRolesBadJSON(t *testing.T) {
	cat := catalogWith(itemWith("1", "A", "Damage"))
	if err := cat.ApplyShopRoles([]byte(`{`)); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestRoleOverridesBeatShopRoles(t *testing.T) {
	prev := roleOverrides
	t.Cleanup(func() { roleOverrides = prev })
	roleOverrides = map[string][]Role{"3508": {RoleMarksman}}

	it := itemWith("3508", "ER", "Damage", "CriticalStrike", "ManaRegen")
	it.ShopRoles = []Role{RoleSupport, RoleMarksman}
	if it.MatchesRole(RoleSupport) {
		t.Fatal("override should drop support even if shop.tags has it")
	}
	if !it.MatchesRole(RoleMarksman) {
		t.Fatal("override marksman")
	}
}

func TestFilterUsesShopRoles(t *testing.T) {
	er := itemWith("3508", "ER", "Damage", "CriticalStrike", "ManaRegen")
	faerie := itemWith("1004", "Faerie", "ManaRegen")
	cat := catalogWith(er, faerie)
	if err := cat.ApplyShopRoles([]byte(`{
		"3508":{"shop":{"tags":["MARKSMAN"]}}
	}`)); err != nil {
		t.Fatal(err)
	}
	got := Filter(cat.Items, Query{Role: RoleSupport})
	if len(got) != 1 || got[0].ID != "1004" {
		t.Fatalf("support: %+v", got)
	}
	got = Filter(cat.Items, Query{Role: RoleMarksman})
	if len(got) != 1 || got[0].ID != "3508" {
		t.Fatalf("marksman: %+v", got)
	}
}

func parseRiftItem(t *testing.T, id, name string, tags ...string) *Catalog {
	t.Helper()
	tagJSON := ""
	for i, tag := range tags {
		if i > 0 {
			tagJSON += ","
		}
		tagJSON += `"` + tag + `"`
	}
	raw := []byte(`{
		"data":{
			"` + id + `":{
				"name":"` + name + `",
				"gold":{"purchasable":true,"total":3050},
				"tags":[` + tagJSON + `],
				"maps":{"11":true},
				"image":{"full":"` + id + `.png"}
			}
		}
	}`)
	cat, err := Parse("16.17.1", "es_ES", ddragon.DefaultBaseURL, raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Items) != 1 {
		t.Fatalf("parse: %+v", cat.Items)
	}
	return cat
}

func catalogWith(list ...Item) *Catalog {
	cat := &Catalog{ByID: make(map[string]Item, len(list))}
	for _, it := range list {
		cat.Items = append(cat.Items, it)
		cat.ByID[it.ID] = it
	}
	return cat
}
