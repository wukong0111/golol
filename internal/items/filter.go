package items

import "sort"

// StatFilter is one combinable checkbox. Matching any tag in Tags is enough
// for that checkbox; multiple checkboxes are AND-ed together.
type StatFilter struct {
	ID    string
	Label string
	Tags  []string
}

// AllStatFilters is the left-bar checklist, in shop-ish order.
var AllStatFilters = []StatFilter{
	{ID: "Damage", Label: "Daño de ataque", Tags: []string{"Damage"}},
	{ID: "CriticalStrike", Label: "Prob. crítico", Tags: []string{"CriticalStrike"}},
	{ID: "AttackSpeed", Label: "Velocidad de ataque", Tags: []string{"AttackSpeed"}},
	{ID: "OnHit", Label: "On-hit", Tags: []string{"OnHit"}},
	{ID: "ArmorPenetration", Label: "Pen. armadura / letalidad", Tags: []string{"ArmorPenetration"}},
	{ID: "SpellDamage", Label: "Poder de habilidad", Tags: []string{"SpellDamage"}},
	{ID: "Mana", Label: "Maná / regen", Tags: []string{"Mana", "ManaRegen"}},
	{ID: "MagicPenetration", Label: "Penetración mágica", Tags: []string{"MagicPenetration"}},
	{ID: "Health", Label: "Vida / regen", Tags: []string{"Health", "HealthRegen"}},
	{ID: "Armor", Label: "Armadura", Tags: []string{"Armor"}},
	{ID: "SpellBlock", Label: "Resistencia mágica", Tags: []string{"SpellBlock", "MagicResist"}},
	{ID: "AbilityHaste", Label: "Aceleración", Tags: []string{"AbilityHaste", "CooldownReduction"}},
	{ID: "Movement", Label: "Movimiento", Tags: []string{"Boots", "NonbootsMovement"}},
	{ID: "LifeSteal", Label: "Robo de vida / vamp", Tags: []string{"LifeSteal", "SpellVamp"}},
}

// ParseStats keeps only known filter IDs, in checklist order, without duplicates.
func ParseStats(raw []string) []StatFilter {
	seen := make(map[string]struct{}, len(raw))
	var out []StatFilter
	for _, filter := range AllStatFilters {
		for _, id := range raw {
			if id != filter.ID {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, filter)
		}
	}
	return out
}

// Query is the shop filter state.
type Query struct {
	Role  Role
	Stats []StatFilter
}

// Matches applies role ∩ AND(stats).
func (it Item) Matches(q Query) bool {
	if !it.MatchesRole(q.Role) {
		return false
	}
	for _, stat := range q.Stats {
		if !it.MatchesGroup(stat.Tags) {
			return false
		}
	}
	return true
}

// Filter returns items matching q, keeping catalog order.
func Filter(catalog []Item, q Query) []Item {
	out := make([]Item, 0, len(catalog))
	for _, it := range catalog {
		if it.Matches(q) {
			out = append(out, it)
		}
	}
	return out
}

// Group is a tier bucket on the grid.
type Group struct {
	Tier  Tier
	Label string
	Items []Item
}

// GroupByTier splits items into básicos / épicos / legendarios.
func GroupByTier(catalog []Item) []Group {
	buckets := [3][]Item{}
	for _, it := range catalog {
		t := it.Tier()
		buckets[t] = append(buckets[t], it)
	}
	var groups []Group
	for _, t := range []Tier{TierBasic, TierEpic, TierLegendary} {
		if len(buckets[t]) == 0 {
			continue
		}
		items := buckets[t]
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Gold != items[j].Gold {
				return items[i].Gold < items[j].Gold
			}
			return items[i].Name < items[j].Name
		})
		groups = append(groups, Group{Tier: t, Label: t.Label(), Items: items})
	}
	return groups
}
