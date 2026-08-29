package items

import "strings"

const summonersRiftMap = "11"

// Item is the shop-facing view of a Data Dragon item.
type Item struct {
	ID          string
	Name        string
	Description string
	Plaintext   string
	Gold        int
	Tags        map[string]struct{}
	From        []string
	Into        []string
	Depth       int
	IconURL     string
	// ShopRoles are the in-game class tabs from Meraki shop.tags.
	// Empty means MatchesRole falls back to the Data Dragon tag heuristic.
	ShopRoles []Role
}

// Has reports whether the item carries a Data Dragon tag.
func (it Item) Has(tag string) bool {
	_, ok := it.Tags[tag]
	return ok
}

// HasAny reports whether the item carries at least one of the tags.
func (it Item) HasAny(tags ...string) bool {
	for _, tag := range tags {
		if it.Has(tag) {
			return true
		}
	}
	return false
}

// MatchesGroup is true if the item has any tag in the group.
func (it Item) MatchesGroup(tags []string) bool {
	return it.HasAny(tags...)
}

// Tier is the shop quality bucket (básico / épico / legendario).
type Tier int

const (
	TierBasic Tier = iota
	TierEpic
	TierLegendary
)

func (t Tier) Label() string {
	switch t {
	case TierEpic:
		return "Épicos"
	case TierLegendary:
		return "Legendarios"
	default:
		return "Básicos"
	}
}

// Tier is the All Items shop bucket.
// Data Dragon depth only counts recipe layers, so Rabadon's Deathcap
// (Needlessly Large Rod ×2) is depth 2 but still a finished legendary.
func (it Item) Tier() Tier {
	if len(it.From) == 0 {
		return TierBasic
	}
	if len(it.Into) == 0 || it.Depth >= 3 {
		return TierLegendary
	}
	return TierEpic
}

func tagSet(tags []string) map[string]struct{} {
	out := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		out[tag] = struct{}{}
	}
	return out
}
