package champions

import "testing"

func TestFilterEmptyRolesReturnsAll(t *testing.T) {
	cat := sampleRoster(t)
	got := Filter(cat.Champions, nil)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilterSingleRole(t *testing.T) {
	cat := sampleRoster(t)
	got := Filter(cat.Champions, []Role{RoleMage})
	if len(got) != 1 || got[0].ID != "Ahri" {
		t.Fatalf("mage: %+v", got)
	}
	got = Filter(cat.Champions, []Role{RoleFighter})
	if len(got) != 1 || got[0].ID != "Aatrox" {
		t.Fatalf("fighter: %+v", got)
	}
}

func TestFilterAND(t *testing.T) {
	cat := sampleRoster(t)
	got := Filter(cat.Champions, []Role{RoleFighter, RoleMage})
	if len(got) != 0 {
		t.Fatalf("AND of disjoint roles should be empty, got %+v", got)
	}
	got = Filter(cat.Champions, []Role{RoleFighter, RoleTank})
	if len(got) != 1 || got[0].ID != "Aatrox" {
		t.Fatalf("fighter+tank: %+v", got)
	}
	got = Filter(cat.Champions, []Role{RoleMage, RoleAssassin})
	if len(got) != 1 || got[0].ID != "Ahri" {
		t.Fatalf("mage+assassin: %+v", got)
	}
}

func TestFilterUnknownIgnored(t *testing.T) {
	got := ParseRoles([]string{"mage", "dragon", "mage", "tank"})
	if len(got) != 2 || got[0] != RoleMage || got[1] != RoleTank {
		t.Fatalf("parse: %+v", got)
	}
}
