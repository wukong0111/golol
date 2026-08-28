package champions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golol/internal/ddragon"
)

var spellKeys = []string{"Q", "W", "E", "R"}

// Catalog is the in-memory champion roster for one Data Dragon patch.
type Catalog struct {
	Version   string
	Locale    string
	Champions []Champion
	ByID      map[string]Champion
}

// Parse builds a roster from championFull.json.
func Parse(version, locale, cdnBase string, raw []byte) (*Catalog, error) {
	var file ddragon.ChampionFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("decode championFull.json: %w", err)
	}
	if version == "" {
		version = file.Version
	}

	cat := &Catalog{
		Version: version,
		Locale:  locale,
		ByID:    make(map[string]Champion, len(file.Data)),
	}
	for key, src := range file.Data {
		id := strings.TrimSpace(src.ID)
		if id == "" {
			id = key
		}
		name := strings.TrimSpace(src.Name)
		if name == "" {
			continue
		}
		image := src.Image.Full
		if image == "" {
			image = id + ".png"
		}
		roles, labels := rolesFromTags(src.Tags)
		partype := strings.TrimSpace(src.Partype)
		ch := Champion{
			ID:         id,
			Name:       name,
			Title:      strings.TrimSpace(src.Title),
			Roles:      roles,
			RoleLabels: labels,
			IconURL:    ddragon.ChampionIconURL(cdnBase, version, image),
			SplashURL:  ddragon.SplashURL(cdnBase, id),
			Passive:    parsePassive(cdnBase, version, src.Passive),
			Spells:     parseSpells(cdnBase, version, src.Spells, partype),
		}
		cat.Champions = append(cat.Champions, ch)
		cat.ByID[id] = ch
	}
	sort.Slice(cat.Champions, func(i, j int) bool {
		if cat.Champions[i].Name != cat.Champions[j].Name {
			return cat.Champions[i].Name < cat.Champions[j].Name
		}
		return cat.Champions[i].ID < cat.Champions[j].ID
	})
	return cat, nil
}

func parsePassive(cdnBase, version string, src ddragon.Passive) Ability {
	image := src.Image.Full
	return Ability{
		Key:         "P",
		Name:        strings.TrimSpace(src.Name),
		Description: src.Description,
		IconURL:     ddragon.PassiveIconURL(cdnBase, version, image),
	}
}

func parseSpells(cdnBase, version string, src []ddragon.Spell, partype string) []Ability {
	out := make([]Ability, 0, len(src))
	for i, sp := range src {
		key := ""
		if i < len(spellKeys) {
			key = spellKeys[i]
		}
		image := sp.Image.Full
		out = append(out, Ability{
			Key:         key,
			Name:        strings.TrimSpace(sp.Name),
			Description: sp.Description,
			IconURL:     ddragon.SpellIconURL(cdnBase, version, image),
			Cooldown:    burnStat(sp.CooldownBurn),
			Cost:        spellCost(sp, partype),
			Range:       spellRange(sp.RangeBurn),
		})
	}
	return out
}

func burnStat(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return ""
	}
	return s
}

func spellCost(sp ddragon.Spell, partype string) string {
	if strings.EqualFold(strings.TrimSpace(sp.Resource), "Sin coste") ||
		strings.EqualFold(strings.TrimSpace(sp.CostType), "Sin coste") {
		return "Sin coste"
	}
	cost := strings.TrimSpace(sp.CostBurn)
	if isZeroBurn(cost) {
		return ""
	}
	resource := interpolateResource(sp.Resource, cost, partype)
	if resource != "" {
		return resource
	}
	costType := interpolateResource(sp.CostType, cost, partype)
	if costType != "" {
		return costType
	}
	if partype != "" && !strings.EqualFold(partype, "Nada") {
		return cost + " " + partype
	}
	return cost
}

func spellRange(burn string) string {
	burn = strings.TrimSpace(burn)
	if burn == "" || burn == "0" || burn == "25000" {
		return ""
	}
	return burn
}

func interpolateResource(raw, costBurn, partype string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "{{ cost }}", strings.TrimSpace(costBurn))
	s = strings.ReplaceAll(s, "{{cost}}", strings.TrimSpace(costBurn))
	s = strings.ReplaceAll(s, "{{ abilityresourcename }}", partype)
	s = strings.ReplaceAll(s, "{{abilityresourcename}}", partype)
	s = strings.ReplaceAll(s, "@AbilityResourceName@", partype)
	s = strings.Join(strings.Fields(s), " ")
	if isZeroBurn(s) {
		return ""
	}
	return s
}

func isZeroBurn(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return true
	}
	for _, part := range strings.Split(s, "/") {
		if strings.TrimSpace(part) != "0" {
			return false
		}
	}
	return true
}

// Get returns a champion by Data Dragon id.
func (c *Catalog) Get(id string) (Champion, bool) {
	if c == nil {
		return Champion{}, false
	}
	ch, ok := c.ByID[id]
	return ch, ok
}
