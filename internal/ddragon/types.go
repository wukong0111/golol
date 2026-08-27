package ddragon

// File is the top-level shape of Data Dragon item.json.
type File struct {
	Type    string          `json:"type"`
	Version string          `json:"version"`
	Data    map[string]Item `json:"data"`
}

// Item is a single Data Dragon item entry.
type Item struct {
	Name             string             `json:"name"`
	Description      string             `json:"description"`
	Colloq           string             `json:"colloq"`
	Plaintext        string             `json:"plaintext"`
	From             []string           `json:"from"`
	Into             []string           `json:"into"`
	Image            Image              `json:"image"`
	Gold             Gold               `json:"gold"`
	Tags             []string           `json:"tags"`
	Maps             map[string]bool    `json:"maps"`
	Stats            map[string]float64 `json:"stats"`
	Depth            int                `json:"depth"`
	InStore          *bool              `json:"inStore"`
	HideFromAll      bool               `json:"hideFromAll"`
	RequiredChampion string             `json:"requiredChampion"`
}

// Image locates an icon on the Data Dragon CDN.
type Image struct {
	Full string `json:"full"`
}

// ChampionFile is the top-level shape of Data Dragon championFull.json.
type ChampionFile struct {
	Type    string              `json:"type"`
	Format  string              `json:"format"`
	Version string              `json:"version"`
	Data    map[string]Champion `json:"data"`
}

// Champion is a full Data Dragon champion entry, including the kit.
type Champion struct {
	ID      string   `json:"id"`
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Title   string   `json:"title"`
	Image   Image    `json:"image"`
	Tags    []string `json:"tags"`
	Blurb   string   `json:"blurb"`
	Spells  []Spell  `json:"spells"`
	Passive Passive  `json:"passive"`
}

// Spell is one of Q/W/E/R.
type Spell struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CooldownBurn string `json:"cooldownBurn"`
	Image        Image  `json:"image"`
}

// Passive is the champion's innate ability.
type Passive struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       Image  `json:"image"`
}

// Gold is the shop pricing block.
type Gold struct {
	Base        int  `json:"base"`
	Total       int  `json:"total"`
	Sell        int  `json:"sell"`
	Purchasable bool `json:"purchasable"`
}

// AvailableInStore reports whether the item is listed in the shop.
// Data Dragon omits inStore on most entries; the documented default is true.
func (it Item) AvailableInStore() bool {
	if it.InStore == nil {
		return true
	}
	return *it.InStore
}
