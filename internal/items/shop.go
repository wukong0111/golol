package items

import (
	"encoding/json"
	"fmt"
	"strings"
)

type merakiItemsFile map[string]merakiItem

type merakiItem struct {
	Shop merakiShop `json:"shop"`
}

type merakiShop struct {
	Tags []string `json:"tags"`
}

// ApplyShopRoles copies in-game class tabs from a Meraki items.json dump.
// Only FIGHTER / MARKSMAN / ASSASSIN / MAGE / TANK / SUPPORT are kept.
// Missing ids, empty class lists, or a bad payload leave the Data Dragon
// tag heuristic in place.
func (c *Catalog) ApplyShopRoles(raw []byte) error {
	if c == nil || len(raw) == 0 {
		return nil
	}
	var file merakiItemsFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return fmt.Errorf("decode meraki items: %w", err)
	}
	for i := range c.Items {
		src, ok := file[c.Items[i].ID]
		if !ok {
			continue
		}
		roles := shopRolesFromTags(src.Shop.Tags)
		if len(roles) == 0 {
			continue
		}
		c.Items[i].ShopRoles = roles
		c.ByID[c.Items[i].ID] = c.Items[i]
	}
	return nil
}

func shopRolesFromTags(tags []string) []Role {
	present := make(map[Role]struct{}, len(tags))
	for _, tag := range tags {
		if r, ok := merakiClass(tag); ok {
			present[r] = struct{}{}
		}
	}
	if len(present) == 0 {
		return nil
	}
	var out []Role
	for _, opt := range AllRoles {
		if _, ok := present[opt.ID]; ok {
			out = append(out, opt.ID)
		}
	}
	return out
}

func merakiClass(tag string) (Role, bool) {
	r := Role(strings.ToLower(strings.TrimSpace(tag)))
	switch r {
	case RoleFighter, RoleMarksman, RoleAssassin, RoleMage, RoleTank, RoleSupport:
		return r, true
	default:
		return "", false
	}
}
