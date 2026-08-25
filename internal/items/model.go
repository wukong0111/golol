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

// Tier classifies the item by Data Dragon depth.
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

// TierOf maps depth to a shop quality bucket.
func TierOf(depth int) Tier {
	switch {
	case depth >= 3:
		return TierLegendary
	case depth == 2:
		return TierEpic
	default:
		return TierBasic
	}
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
