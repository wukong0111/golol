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
		ch := Champion{
			ID:         id,
			Name:       name,
			Title:      strings.TrimSpace(src.Title),
			Roles:      roles,
			RoleLabels: labels,
			IconURL:    ddragon.ChampionIconURL(cdnBase, version, image),
			SplashURL:  ddragon.SplashURL(cdnBase, id),
			Passive:    parsePassive(cdnBase, version, src.Passive),
			Spells:     parseSpells(cdnBase, version, src.Spells),
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

func parseSpells(cdnBase, version string, src []ddragon.Spell) []Ability {
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
			Cooldown:    strings.TrimSpace(sp.CooldownBurn),
		})
	}
	return out
}

// Get returns a champion by Data Dragon id.
func (c *Catalog) Get(id string) (Champion, bool) {
	if c == nil {
		return Champion{}, false
	}
	ch, ok := c.ByID[id]
	return ch, ok
}
