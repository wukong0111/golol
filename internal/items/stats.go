package items

import (
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Bonus is one numeric line from the item's <stats> block — the shop
// "mejoras": +75 attack damage, +400 health, 20 ability haste, …
// Rank orders known stats when several items are summed.
type Bonus struct {
	Amount  float64 `json:"amount"`
	Percent bool    `json:"percent"`
	Name    string  `json:"name"`
	Rank    int     `json:"rank"`
}

var (
	statsBlockRe = regexp.MustCompile(`(?is)<stats>(.*?)</stats>`)
	brSplitRe    = regexp.MustCompile(`(?i)<br\s*/?>`)
	htmlTagRe    = regexp.MustCompile(`<[^>]+>`)
	bonusLineRe  = regexp.MustCompile(`(?i)^\s*([+-]?\d+(?:\.\d+)?)(%?)\s+(.+?)\s*$`)
)

// Known stat names, Spanish and English, in tooltip order. Matching is
// exact against the text after a leading "de " / "of ".
var bonusNameGroups = [][]string{
	{"daño de ataque", "attack damage"},
	{"poder de habilidad", "ability power"},
	{"fuerza adaptable", "adaptive force"},
	{"vida", "health"},
	{"armadura", "armor"},
	{"resistencia mágica", "magic resist", "magic resistance"},
	{"velocidad de ataque", "attack speed"},
	{"probabilidad de impacto crítico", "critical strike chance"},
	{"daño de impacto crítico", "critical strike damage"},
	{"velocidad de habilidades", "ability haste"},
	{"reducción de enfriamiento", "cooldown reduction"},
	{"maná", "mana"},
	{"letalidad", "lethality"},
	{"penetración de armadura", "armor penetration"},
	{"penetración mágica", "magic penetration"},
	{"velocidad de movimiento", "movement speed", "move speed"},
	{"tenacidad", "tenacity"},
	{"robo de vida", "life steal", "lifesteal"},
	{"omnisucción", "omnivamp"},
	{"poder de curaciones y escudos", "heal and shield power"},
	{"regeneración de vida básica", "base health regen"},
	{"regeneración de maná básica", "base mana regen"},
	{"regeneración de vida cada 5 s", "health regen per 5 seconds"},
	{"regeneración de maná cada 5 s", "mana regen per 5 seconds"},
	{"oro cada 10 segundos", "gold per 10 seconds"},
}

// ParseBonuses reads the <stats> block of a Data Dragon description.
func ParseBonuses(description string) []Bonus {
	m := statsBlockRe.FindStringSubmatch(description)
	if m == nil {
		return nil
	}
	var out []Bonus
	for _, part := range brSplitRe.Split(m[1], -1) {
		text := strings.Join(strings.Fields(html.UnescapeString(htmlTagRe.ReplaceAllString(part, ""))), " ")
		if text == "" {
			continue
		}
		sm := bonusLineRe.FindStringSubmatch(text)
		if sm == nil {
			continue
		}
		amount, err := strconv.ParseFloat(sm[1], 64)
		if err != nil || amount == 0 {
			continue
		}
		name := sm[3]
		out = append(out, Bonus{
			Amount:  amount,
			Percent: sm[2] == "%",
			Name:    name,
			Rank:    bonusRank(name),
		})
	}
	return SumBonuses(out)
}

// SumBonuses adds matching lines (same name and percent/flat) and sorts
// the result in tooltip order. Zero totals are dropped.
func SumBonuses(lists ...[]Bonus) []Bonus {
	type key struct {
		percent bool
		name    string
	}
	acc := make(map[key]*Bonus)
	var order []key
	for _, list := range lists {
		for _, b := range list {
			k := key{b.Percent, b.Name}
			if cur, ok := acc[k]; ok {
				cur.Amount += b.Amount
				continue
			}
			cp := b
			cp.Rank = bonusRank(cp.Name)
			acc[k] = &cp
			order = append(order, k)
		}
	}
	out := make([]Bonus, 0, len(order))
	for _, k := range order {
		b := *acc[k]
		if b.Amount == 0 {
			continue
		}
		out = append(out, b)
	}
	sortBonuses(out)
	return out
}

// BonusIndex maps item id → parsed stat bonuses, omitting empty entries.
func BonusIndex(catalog []Item) map[string][]Bonus {
	out := make(map[string][]Bonus, len(catalog))
	for _, it := range catalog {
		if len(it.Bonuses) == 0 {
			continue
		}
		out[it.ID] = it.Bonuses
	}
	return out
}

func bonusRank(name string) int {
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.TrimPrefix(n, "de ")
	n = strings.TrimPrefix(n, "of ")
	for i, group := range bonusNameGroups {
		for _, known := range group {
			if n == known {
				return i
			}
		}
	}
	return len(bonusNameGroups)
}

func sortBonuses(out []Bonus) {
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Rank != out[j].Rank {
			return out[i].Rank < out[j].Rank
		}
		if out[i].Percent != out[j].Percent {
			return !out[i].Percent
		}
		return out[i].Name < out[j].Name
	})
}
