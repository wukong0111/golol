package items

import "testing"

func itemWith(id, name string, tags ...string) Item {
	return Item{ID: id, Name: name, Tags: tagSet(tags), Gold: 1000}
}

func TestStatAND(t *testing.T) {
	armorAS := itemWith("1", "Randuin", "Armor", "AttackSpeed")
	armorOnly := itemWith("2", "Cloth", "Armor")
	asOnly := itemWith("3", "Dagger", "AttackSpeed")

	q := Query{Stats: ParseStats([]string{"Armor", "AttackSpeed"})}
	got := Filter([]Item{armorAS, armorOnly, asOnly}, q)
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("AND Armor+AS: got %+v", got)
	}
}

func TestSingleStat(t *testing.T) {
	armor := itemWith("1", "Cloth", "Armor")
	ap := itemWith("2", "Tome", "SpellDamage")
	q := Query{Stats: ParseStats([]string{"Armor"})}
	got := Filter([]Item{armor, ap}, q)
	if len(got) != 1 || got[0].ID != "1" {
		t.Fatalf("Armor only: got %+v", got)
	}
}

func TestNoStatsReturnsAll(t *testing.T) {
	a := itemWith("1", "A", "Armor")
	b := itemWith("2", "B", "Damage")
	got := Filter([]Item{a, b}, Query{})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestSpellBlockMatchesMagicResistFilter(t *testing.T) {
	negatron := itemWith("1057", "Negatron", "SpellBlock")
	q := Query{Stats: ParseStats([]string{"SpellBlock"})}
	if !negatron.Matches(q) {
		t.Fatal("SpellBlock tag should match Resistencia mágica")
	}
	magicResistTag := itemWith("x", "JakSho", "MagicResist")
	if !magicResistTag.Matches(q) {
		t.Fatal("MagicResist tag should match Resistencia mágica")
	}
}

func TestAbilityHasteOrCooldownReduction(t *testing.T) {
	q := Query{Stats: ParseStats([]string{"AbilityHaste"})}
	if !itemWith("1", "Mote", "AbilityHaste").Matches(q) {
		t.Fatal("AbilityHaste should match")
	}
	if !itemWith("2", "Kindle", "CooldownReduction").Matches(q) {
		t.Fatal("CooldownReduction should match Aceleración")
	}
}

func TestUnknownStatsIgnored(t *testing.T) {
	got := ParseStats([]string{"Nope", "Armor", "Armor", "Damage"})
	if len(got) != 2 || got[0].ID != "Damage" || got[1].ID != "Armor" {
		t.Fatalf("ParseStats order/dedup: %+v", got)
	}
}

func TestTankRole(t *testing.T) {
	thornmail := itemWith("3075", "Thornmail", "Health", "Armor")
	ie := itemWith("3031", "Infinity Edge", "CriticalStrike", "Damage")
	q := Query{Role: RoleTank}
	got := Filter([]Item{thornmail, ie}, q)
	if len(got) != 1 || got[0].ID != "3075" {
		t.Fatalf("tank filter: %+v", got)
	}
}

func TestMageRole(t *testing.T) {
	raba := itemWith("3089", "Rabadon", "SpellDamage")
	boots := itemWith("1001", "Boots", "Boots")
	q := Query{Role: RoleMage}
	got := Filter([]Item{raba, boots}, q)
	if len(got) != 1 || got[0].ID != "3089" {
		t.Fatalf("mage filter: %+v", got)
	}
}

func TestFighterNeedsDamageAndHealth(t *testing.T) {
	steraks := itemWith("3053", "Sterak", "Damage", "Health")
	longSword := itemWith("1036", "Long Sword", "Damage")
	belt := itemWith("1011", "Giant's Belt", "Health")
	q := Query{Role: RoleFighter}
	got := Filter([]Item{steraks, longSword, belt}, q)
	if len(got) != 1 || got[0].ID != "3053" {
		t.Fatalf("fighter filter: %+v", got)
	}
}

func TestMarksman(t *testing.T) {
	ie := itemWith("3031", "IE", "CriticalStrike", "Damage")
	pd := itemWith("3046", "PD", "AttackSpeed", "CriticalStrike")
	wit := itemWith("3091", "Wit's End", "AttackSpeed", "SpellBlock")
	q := Query{Role: RoleMarksman}
	got := Filter([]Item{ie, pd, wit}, q)
	if len(got) != 2 {
		t.Fatalf("marksman expected 2 (IE, PD), got %+v", got)
	}
}

func TestAssassin(t *testing.T) {
	youmuu := itemWith("3142", "Youmuu", "ArmorPenetration", "Damage")
	ie := itemWith("3031", "IE", "CriticalStrike", "Damage")
	q := Query{Role: RoleAssassin}
	got := Filter([]Item{youmuu, ie}, q)
	if len(got) != 1 || got[0].ID != "3142" {
		t.Fatalf("assassin filter: %+v", got)
	}
}

func TestSupport(t *testing.T) {
	relic := itemWith("3862", "Relic", "GoldPer", "Health")
	faerie := itemWith("1004", "Faerie", "ManaRegen")
	lostChapter := itemWith("3802", "Lost Chapter", "SpellDamage", "ManaRegen")
	q := Query{Role: RoleSupport}
	got := Filter([]Item{relic, faerie, lostChapter}, q)
	if len(got) != 2 {
		t.Fatalf("support expected relic+faerie, got %+v", got)
	}
}

func TestRoleAndStatsIntersect(t *testing.T) {
	zhonya := itemWith("3157", "Zhonya", "Armor", "SpellDamage", "Active")
	thorn := itemWith("3075", "Thornmail", "Health", "Armor")
	raba := itemWith("3089", "Rabadon", "SpellDamage")
	q := Query{Role: RoleTank, Stats: ParseStats([]string{"SpellDamage"})}
	got := Filter([]Item{zhonya, thorn, raba}, q)
	if len(got) != 1 || got[0].ID != "3157" {
		t.Fatalf("tank ∩ AP: %+v", got)
	}
}

func TestParseRoleUnknownIsAll(t *testing.T) {
	if ParseRole("jungler") != RoleAll {
		t.Fatal("unknown role should be all")
	}
	if ParseRole("tank") != RoleTank {
		t.Fatal("tank")
	}
}

func TestGroupByTier(t *testing.T) {
	items := []Item{
		{ID: "3", Name: "C", Depth: 3, Gold: 3000},
		{ID: "1", Name: "A", Depth: 1, Gold: 300},
		{ID: "2", Name: "B", Depth: 2, Gold: 800},
		{ID: "0", Name: "Z", Depth: 0, Gold: 400},
	}
	groups := GroupByTier(items)
	if len(groups) != 3 {
		t.Fatalf("groups: %d", len(groups))
	}
	if groups[0].Label != "Básicos" || len(groups[0].Items) != 2 {
		t.Fatalf("basic: %+v", groups[0])
	}
	if groups[0].Items[0].ID != "1" || groups[0].Items[1].ID != "0" {
		t.Fatalf("basic sort by gold: %+v", groups[0].Items)
	}
	if groups[1].Label != "Épicos" || groups[2].Label != "Legendarios" {
		t.Fatalf("labels: %s %s", groups[1].Label, groups[2].Label)
	}
}
