package items

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golol/internal/ddragon"
)

// Catalog is the in-memory shop inventory for one Data Dragon patch.
type Catalog struct {
	Version string
	Locale  string
	Items   []Item
	ByID    map[string]Item
}

// Parse builds a Summoner's Rift purchasable catalog from item.json.
func Parse(version, locale, cdnBase string, raw []byte) (*Catalog, error) {
	var file ddragon.File
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("decode item.json: %w", err)
	}
	if version == "" {
		version = file.Version
	}

	cat := &Catalog{
		Version: version,
		Locale:  locale,
		ByID:    make(map[string]Item, len(file.Data)),
	}
	for id, src := range file.Data {
		if !isSummonersRiftPurchasable(src) {
			continue
		}
		name := strings.TrimSpace(src.Name)
		if name == "" {
			continue
		}
		image := src.Image.Full
		if image == "" {
			image = id + ".png"
		}
		it := Item{
			ID:          id,
			Name:        name,
			Description: src.Description,
			Plaintext:   src.Plaintext,
			Gold:        src.Gold.Total,
			Tags:        tagSet(src.Tags),
			From:        append([]string(nil), src.From...),
			Into:        append([]string(nil), src.Into...),
			Depth:       src.Depth,
			IconURL:     ddragon.IconURL(cdnBase, version, image),
		}
		cat.Items = append(cat.Items, it)
		cat.ByID[id] = it
	}
	sort.Slice(cat.Items, func(i, j int) bool {
		if cat.Items[i].Gold != cat.Items[j].Gold {
			return cat.Items[i].Gold < cat.Items[j].Gold
		}
		if cat.Items[i].Name != cat.Items[j].Name {
			return cat.Items[i].Name < cat.Items[j].Name
		}
		return cat.Items[i].ID < cat.Items[j].ID
	})
	return cat, nil
}

func isSummonersRiftPurchasable(src ddragon.Item) bool {
	if !src.Gold.Purchasable || src.HideFromAll || !src.AvailableInStore() {
		return false
	}
	if src.Maps == nil || !src.Maps[summonersRiftMap] {
		return false
	}
	return true
}

// Get returns an item by Data Dragon id.
func (c *Catalog) Get(id string) (Item, bool) {
	if c == nil {
		return Item{}, false
	}
	it, ok := c.ByID[id]
	return it, ok
}

// Components resolves build-from items that are themselves in the catalog.
func (c *Catalog) Components(it Item) []Item {
	if c == nil {
		return nil
	}
	out := make([]Item, 0, len(it.From))
	for _, id := range it.From {
		if part, ok := c.ByID[id]; ok {
			out = append(out, part)
		}
	}
	return out
}
