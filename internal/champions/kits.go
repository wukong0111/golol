package champions

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type merakiFile map[string]merakiChampion

type merakiChampion struct {
	Key       string    `json:"key"`
	Abilities merakiKit `json:"abilities"`
}

type merakiKit struct {
	P []merakiAbility `json:"P"`
	Q []merakiAbility `json:"Q"`
	W []merakiAbility `json:"W"`
	E []merakiAbility `json:"E"`
	R []merakiAbility `json:"R"`
}

type merakiAbility struct {
	Name     string           `json:"name"`
	Effects  []merakiEffect   `json:"effects"`
	Cooldown *merakiStatBlock `json:"cooldown"`
}

type merakiEffect struct {
	Description string        `json:"description"`
	Leveling    []merakiLevel `json:"leveling"`
}

type merakiLevel struct {
	Attribute string           `json:"attribute"`
	Modifiers []merakiModifier `json:"modifiers"`
}

type merakiModifier struct {
	Values []float64 `json:"values"`
	Units  []string  `json:"units"`
}

type merakiStatBlock struct {
	Modifiers []merakiModifier `json:"modifiers"`
}

// ApplyKits fills per-rank ability numbers from a Meraki champions.json dump.
// Missing champions or a bad payload leave the Data Dragon kit as-is.
func (c *Catalog) ApplyKits(raw []byte) error {
	if c == nil || len(raw) == 0 {
		return nil
	}
	var file merakiFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return fmt.Errorf("decode meraki champions: %w", err)
	}
	byID := make(map[string]merakiChampion, len(file))
	for name, ch := range file {
		id := strings.TrimSpace(ch.Key)
		if id == "" {
			id = name
		}
		byID[id] = ch
	}
	for i := range c.Champions {
		kit, ok := byID[c.Champions[i].ID]
		if !ok {
			continue
		}
		applyKit(&c.Champions[i], kit)
		c.ByID[c.Champions[i].ID] = c.Champions[i]
	}
	return nil
}

func applyKit(ch *Champion, kit merakiChampion) {
	attachSlot(&ch.Passive, kit.Abilities.P)
	slots := [][]merakiAbility{
		kit.Abilities.Q,
		kit.Abilities.W,
		kit.Abilities.E,
		kit.Abilities.R,
	}
	for i := range ch.Spells {
		if i >= len(slots) {
			break
		}
		attachSlot(&ch.Spells[i], slots[i])
	}
}

func attachSlot(ab *Ability, src []merakiAbility) {
	if ab.Cooldown == "" {
		if cd := kitCooldown(src); cd != "" {
			ab.Cooldown = cd
		}
	}
	forms := kitForms(src)
	switch len(forms) {
	case 0:
		return
	case 1:
		ab.Scaling = forms[0].Scaling
	default:
		ab.Forms = forms
	}
}

func kitCooldown(src []merakiAbility) string {
	for _, ab := range src {
		if ab.Cooldown == nil {
			continue
		}
		text := formatModifiers(ab.Cooldown.Modifiers)
		if text != "" {
			return text
		}
	}
	return ""
}

func kitForms(src []merakiAbility) []AbilityForm {
	var out []AbilityForm
	for _, ab := range src {
		var rows []ScaleRow
		for _, eff := range ab.Effects {
			if len(eff.Leveling) > 0 {
				for _, lv := range eff.Leveling {
					text := formatModifiers(lv.Modifiers)
					if strings.TrimSpace(lv.Attribute) == "" && text == "" {
						continue
					}
					rows = append(rows, ScaleRow{
						Label: translateAttr(lv.Attribute),
						Text:  text,
					})
				}
				continue
			}
			rows = append(rows, extractDescriptionScales(eff.Description)...)
		}
		rows = dedupeScaleRows(rows)
		if len(rows) == 0 {
			continue
		}
		out = append(out, AbilityForm{
			Name:    strings.TrimSpace(ab.Name),
			Scaling: rows,
		})
	}
	return out
}

func extractDescriptionScales(desc string) []ScaleRow {
	rows := extractLevelScales(desc)
	rows = append(rows, extractFormulaScales(desc)...)
	rows = append(rows, extractDurationScales(desc)...)
	return rows
}

// Passives almost never have Meraki leveling tables; the numbers live in
// "4% : 8% (based on level)" inside the English description.
var levelScaleRe = regexp.MustCompile(`(?i)(-?\d+(?:\.\d+)?%?(?:\s*[/:]\s*-?\d+(?:\.\d+)?%?)+)\s*\(\s*based on [^)]+\)(?:\s*\(\+\s*([^)]+)\))?`)

// Utility spells (Gwen W, Pyke W) put the only numbers in "22 (+ 7% AP)".
var formulaScaleRe = regexp.MustCompile(`(?i)(-?\d+(?:\.\d+)?%?)\s*\(\+\s*([^)]+?)\)`)

var lastingRe = regexp.MustCompile(`(?i)lasting for (\d+(?:\.\d+)?) seconds`)

func extractLevelScales(desc string) []ScaleRow {
	if strings.TrimSpace(desc) == "" {
		return nil
	}
	flat := strings.NewReplacer("[", " ", "]", " ").Replace(desc)
	matches := levelScaleRe.FindAllStringSubmatchIndex(flat, -1)
	if len(matches) == 0 {
		return nil
	}
	var rows []ScaleRow
	for _, loc := range matches {
		expr := strings.TrimSpace(flat[loc[2]:loc[3]])
		bonus := ""
		if loc[4] >= 0 {
			bonus = strings.TrimSpace(flat[loc[4]:loc[5]])
		}
		before := flat[:loc[0]]
		after := ""
		if loc[1] < len(flat) {
			after = flat[loc[1]:]
		}
		label := inferScaleLabel(before + " " + after)
		text := formatScaleExpr(expr)
		if extra := translateBonus(bonus); extra != "" {
			text += " (+ " + extra + ")"
		}
		if suffix := trailingScaleUnit(after); suffix != "" {
			text += " " + suffix
		}
		if text == "" {
			continue
		}
		rows = append(rows, ScaleRow{Label: label, Text: text})
	}
	return rows
}

func extractFormulaScales(desc string) []ScaleRow {
	if strings.TrimSpace(desc) == "" {
		return nil
	}
	flat := strings.NewReplacer("[", " ", "]", " ").Replace(desc)
	matches := formulaScaleRe.FindAllStringSubmatchIndex(flat, -1)
	if len(matches) == 0 {
		return nil
	}
	levelSpans := levelScaleRe.FindAllStringSubmatchIndex(flat, -1)
	var rows []ScaleRow
	for _, loc := range matches {
		if insideSpans(loc[0], loc[1], levelSpans) {
			continue
		}
		base := strings.TrimSpace(flat[loc[2]:loc[3]])
		bonus := translateBonus(strings.TrimSpace(flat[loc[4]:loc[5]]))
		if base == "" || bonus == "" {
			continue
		}
		before := flat[:loc[0]]
		after := ""
		if loc[1] < len(flat) {
			after = flat[loc[1]:]
		}
		rows = append(rows, ScaleRow{
			Label: inferScaleLabel(before + " " + after),
			Text:  base + " (+ " + bonus + ")",
		})
	}
	return rows
}

func extractDurationScales(desc string) []ScaleRow {
	m := lastingRe.FindStringSubmatch(desc)
	if len(m) < 2 {
		return nil
	}
	return []ScaleRow{{Label: "Duración", Text: m[1] + " s"}}
}

func insideSpans(start, end int, spans [][]int) bool {
	for _, loc := range spans {
		if start >= loc[0] && end <= loc[1] {
			return true
		}
	}
	return false
}

func formatScaleExpr(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, ":") && !strings.Contains(raw, "/") {
		parts := strings.Split(raw, ":")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) + " – " + strings.TrimSpace(parts[1])
		}
	}
	parts := strings.Split(raw, "/")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, " / ")
}

func translateBonus(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	out := s
	for _, p := range bonusPhrases {
		out = strings.ReplaceAll(out, p.en, p.es)
	}
	return strings.Join(strings.Fields(out), " ")
}

func trailingScaleUnit(after string) string {
	s := strings.TrimSpace(after)
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	for _, p := range scaleTrailPhrases {
		if strings.HasPrefix(lower, strings.ToLower(p.en)) {
			return p.es
		}
	}
	return ""
}

func inferScaleLabel(context string) string {
	lower := strings.ToLower(context)
	for _, p := range scaleLabelHints {
		if strings.Contains(lower, strings.ToLower(p.en)) {
			return p.es
		}
	}
	return "Según nivel"
}

func dedupeScaleRows(rows []ScaleRow) []ScaleRow {
	if len(rows) < 2 {
		return rows
	}
	seen := make(map[string]struct{}, len(rows))
	out := rows[:0]
	for _, row := range rows {
		key := row.Label + "\x00" + row.Text
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, row)
	}
	return out
}

func formatModifiers(mods []merakiModifier) string {
	var parts []string
	for _, mod := range mods {
		if len(mod.Values) == 0 {
			continue
		}
		unit := ""
		if len(mod.Units) > 0 {
			unit = translateUnit(mod.Units[0])
		}
		chunk := withUnit(formatValues(mod.Values), unit)
		if chunk == "" {
			continue
		}
		parts = append(parts, chunk)
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + " (+ " + strings.Join(parts[1:], " + ") + ")"
}

func formatValues(vals []float64) string {
	if len(vals) == 0 {
		return ""
	}
	if allEqual(vals) {
		return formatNum(vals[0])
	}
	if len(vals) >= 10 {
		return formatNum(vals[0]) + " – " + formatNum(vals[len(vals)-1])
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = formatNum(v)
	}
	return strings.Join(parts, " / ")
}

func allEqual(vals []float64) bool {
	if len(vals) <= 1 {
		return true
	}
	first := vals[0]
	for _, v := range vals[1:] {
		if math.Abs(v-first) > 0.0005 {
			return false
		}
	}
	return true
}

func formatNum(v float64) string {
	if math.Abs(v-math.Round(v)) < 0.0005 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func withUnit(values, unit string) string {
	if values == "" {
		return ""
	}
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return values
	}
	if strings.HasPrefix(unit, "%") {
		return values + unit
	}
	return values + " " + unit
}

func translateUnit(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if es, ok := unitES[u]; ok {
		return es
	}
	lower := strings.ToLower(u)
	for en, es := range unitES {
		if strings.ToLower(en) == lower {
			return es
		}
	}
	return u
}

func translateAttr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if es, ok := attrExact[s]; ok {
		return es
	}
	out := s
	for _, p := range attrPhrases {
		out = strings.ReplaceAll(out, p.en, p.es)
	}
	return capitalize(strings.Join(strings.Fields(out), " "))
}

func capitalize(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

type phrase struct {
	en, es string
}

var attrExact = map[string]string{
	"Magic Damage":                             "Daño mágico",
	"Physical Damage":                          "Daño físico",
	"True Damage":                              "Daño verdadero",
	"Total Magic Damage":                       "Daño mágico total",
	"Total Physical Damage":                    "Daño físico total",
	"Bonus Magic Damage":                       "Daño mágico adicional",
	"Bonus Physical Damage":                    "Daño físico adicional",
	"Bonus Movement Speed":                     "Velocidad de movimiento adicional",
	"Bonus Attack Speed":                       "Velocidad de ataque adicional",
	"Bonus Attack Damage":                      "Daño de ataque adicional",
	"Shield Strength":                          "Escudo",
	"Heal":                                     "Curación",
	"Healing":                                  "Curación",
	"Slow":                                     "Ralentización",
	"Root Duration":                            "Duración del enraizamiento",
	"Stun Duration":                            "Duración del aturdimiento",
	"Disable Duration":                         "Duración de la inhabilitación",
	"Magic Damage Per Tick":                    "Daño mágico por pulso",
	"Magic Damage per Tick":                    "Daño mágico por pulso",
	"Minion Damage":                            "Daño a súbditos",
	"Maximum Magic Damage":                     "Daño mágico máximo",
	"Minimum Magic Damage":                     "Daño mágico mínimo",
	"Maximum Physical Damage":                  "Daño físico máximo",
	"Minimum Physical Damage":                  "Daño físico mínimo",
	"First Cast Damage":                        "Daño (1.er lanzamiento)",
	"First Sweetspot Damage":                   "Daño de punto dulce (1.er lanzamiento)",
	"Second Cast Damage":                       "Daño (2.º lanzamiento)",
	"Second Sweetspot Damage":                  "Daño de punto dulce (2.º lanzamiento)",
	"Third Cast Damage":                        "Daño (3.er lanzamiento)",
	"Third Sweetspot Damage":                   "Daño de punto dulce (3.er lanzamiento)",
	"Maximum Non-Minion Non-Sweetspot Damage":  "Daño máximo (no súbditos, sin punto dulce)",
	"Maximum Non-Minion Sweetspot Damage":      "Daño máximo de punto dulce (no súbditos)",
	"Damage Per Pass":                          "Daño por pasada",
	"Initial Flame Magic Damage":               "Daño mágico (llama inicial)",
	"Subsequent Flame Magic Damage":            "Daño mágico (llamas posteriores)",
	"Increased Initial Flame Minion Damage":    "Daño aumentado a súbditos (llama inicial)",
	"Increased Subsequent Flame Minion Damage": "Daño aumentado a súbditos (llamas posteriores)",
}

var attrPhrases = []phrase{
	{en: "Magic Damage", es: "daño mágico"},
	{en: "Physical Damage", es: "daño físico"},
	{en: "True Damage", es: "daño verdadero"},
	{en: "Mixed Damage", es: "daño mixto"},
	{en: "Movement Speed", es: "velocidad de movimiento"},
	{en: "Attack Speed", es: "velocidad de ataque"},
	{en: "Attack Damage", es: "daño de ataque"},
	{en: "Magic Resistance", es: "resistencia mágica"},
	{en: "Shield Strength", es: "escudo"},
	{en: "Root Duration", es: "duración del enraizamiento"},
	{en: "Stun Duration", es: "duración del aturdimiento"},
	{en: "Slow Duration", es: "duración de la ralentización"},
	{en: "Fear Duration", es: "duración del temor"},
	{en: "Silence Duration", es: "duración del silencio"},
	{en: "Disable Duration", es: "duración de la inhabilitación"},
	{en: "Per Tick", es: "por pulso"},
	{en: "per Tick", es: "por pulso"},
	{en: "Per Second", es: "por segundo"},
	{en: "per Second", es: "por segundo"},
	{en: "Per Hit", es: "por golpe"},
	{en: "per Hit", es: "por golpe"},
	{en: "Per Stack", es: "por acumulación"},
	{en: "On-Hit", es: "al impacto"},
	{en: "Sweetspot", es: "punto dulce"},
	{en: "First Cast", es: "1.er lanzamiento"},
	{en: "Second Cast", es: "2.º lanzamiento"},
	{en: "Third Cast", es: "3.er lanzamiento"},
	{en: "Non-Minion", es: "no súbditos"},
	{en: "Per Pass", es: "por pasada"},
	{en: "Initial Flame", es: "llama inicial"},
	{en: "Subsequent Flame", es: "llama posterior"},
	{en: "Initial ", es: "inicial "},
	{en: "Subsequent ", es: "posterior "},
	{en: "First ", es: "1.er "},
	{en: "Second ", es: "2.º "},
	{en: "Third ", es: "3.er "},
	{en: "Single-Target", es: "a un objetivo"},
	{en: "Non-Champion", es: "no campeones"},
	{en: "Healing", es: "curación"},
	{en: "Minion", es: "súbditos"},
	{en: "Monster", es: "monstruos"},
	{en: "Champion", es: "campeones"},
	{en: "Heal", es: "curación"},
	{en: "Shield", es: "escudo"},
	{en: "Slow", es: "ralentización"},
	{en: "Damage", es: "daño"},
	{en: "Bonus", es: "adicional"},
	{en: "Total", es: "total"},
	{en: "Maximum", es: "máximo"},
	{en: "Minimum", es: "mínimo"},
	{en: "Increased", es: "aumentado"},
	{en: "Reduced", es: "reducido"},
	{en: "Enhanced", es: "mejorado"},
	{en: "Empowered", es: "potenciado"},
	{en: "Duration", es: "duración"},
	{en: "Cooldown", es: "enfriamiento"},
}

var scaleLabelHints = []phrase{
	{en: "bonus armor and bonus magic resistance", es: "Resistencias"},
	{en: "bonus magic resistance", es: "Resistencia mágica adicional"},
	{en: "magic resistance", es: "Resistencia mágica"},
	{en: "bonus armor", es: "Armadura adicional"},
	{en: "bonus attack speed", es: "Velocidad de ataque adicional"},
	{en: "attack speed", es: "Velocidad de ataque"},
	{en: "bonus movement speed", es: "Velocidad de movimiento adicional"},
	{en: "movement speed", es: "Velocidad de movimiento"},
	{en: "bonus attack damage", es: "DA adicional"},
	{en: "magic damage", es: "Daño mágico"},
	{en: "physical damage", es: "Daño físico"},
	{en: "true damage", es: "Daño verdadero"},
	{en: "life steal", es: "Robo de vida"},
	{en: "heal herself", es: "Curación"},
	{en: "heal himself", es: "Curación"},
	{en: "heal for", es: "Curación"},
	{en: "healing", es: "Curación"},
	{en: "heals", es: "Curación"},
	{en: "shield", es: "Escudo"},
	{en: "stun", es: "Aturdimiento"},
	{en: "slow", es: "Ralentización"},
	{en: "lethality", es: "Letalidad"},
}

var scaleTrailPhrases = []phrase{
	{en: "of the target's maximum health", es: "de la vida máxima del objetivo"},
	{en: "of the target's current health", es: "de la vida actual del objetivo"},
	{en: "of the target's missing health", es: "de la vida faltante del objetivo"},
	{en: "of his maximum health", es: "de su vida máxima"},
	{en: "of her maximum health", es: "de su vida máxima"},
	{en: "of maximum health", es: "de vida máxima"},
}

var bonusPhrases = []phrase{
	{en: "per 100 bonus health", es: "por 100 de vida adicional"},
	{en: "per 1 Lethality", es: "por 1 de letalidad"},
	{en: "per 100 AP", es: "por 100 PH"},
	{en: "bonus magic resistance", es: "resistencia mágica adicional"},
	{en: "bonus armor", es: "armadura adicional"},
	{en: "bonus health", es: "vida adicional"},
	{en: "bonus AD", es: "DA adicional"},
	{en: "bonus AP", es: "PH adicional"},
	{en: "maximum health", es: "vida máxima"},
	{en: " AP", es: " PH"},
	{en: " AD", es: " DA"},
}

var unitES = map[string]string{
	"% AP":                              "% PH",
	"% AD":                              "% DA",
	"% bonus AD":                        "% DA adicional",
	"% bonus AP":                        "% PH adicional",
	"%":                                 "%",
	" seconds":                          "s",
	"seconds":                           "s",
	"% per 100 AP":                      "% por 100 PH",
	"% of target's maximum health":      "% de la vida máxima del objetivo",
	"%  of target's maximum health":     "% de la vida máxima del objetivo",
	"% of target's missing health":      "% de la vida faltante del objetivo",
	"% bonus health":                    "% de vida adicional",
	"% maximum health":                  "% de vida máxima",
	"% of his bonus health":             "% de su vida adicional",
	"% per 100 bonus AD":                "% por 100 DA adicional",
	"% bonus armor":                     "% de armadura adicional",
	"%  of the target's maximum health": "% de la vida máxima del objetivo",
	"% bonus magic resistance":          "% de resistencia mágica adicional",
	"% maximum mana":                    "% de maná máximo",
	"% of target's armor":               "% de la armadura del objetivo",
	"% armor":                           "% de armadura",
	"% of target's current health":      "% de la vida actual del objetivo",
	"% of maximum health":               "% de vida máxima",
	"based on critical strike chance":   "según probabilidad de crítico",
}

func init() {
	sortLongestFirst(attrPhrases)
	sortLongestFirst(scaleLabelHints)
	sortLongestFirst(scaleTrailPhrases)
	sortLongestFirst(bonusPhrases)
}

func sortLongestFirst(list []phrase) {
	sort.Slice(list, func(i, j int) bool {
		if len(list[i].en) != len(list[j].en) {
			return len(list[i].en) > len(list[j].en)
		}
		return list[i].en < list[j].en
	})
}
