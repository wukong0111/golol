package items

// Role is a shop class filter. The empty value means "all items".
type Role string

const (
	RoleAll      Role = ""
	RoleFighter  Role = "fighter"
	RoleMarksman Role = "marksman"
	RoleAssassin Role = "assassin"
	RoleMage     Role = "mage"
	RoleTank     Role = "tank"
	RoleSupport  Role = "support"
)

// RoleOption is a radio in the role bar.
type RoleOption struct {
	ID    Role
	Label string
}

// AllRoles is the exclusive role bar, in shop order, including "Todos".
var AllRoles = []RoleOption{
	{ID: RoleAll, Label: "Todos"},
	{ID: RoleFighter, Label: "Luchador"},
	{ID: RoleMarksman, Label: "Tirador"},
	{ID: RoleAssassin, Label: "Asesino"},
	{ID: RoleMage, Label: "Mago"},
	{ID: RoleTank, Label: "Tanque"},
	{ID: RoleSupport, Label: "Soporte"},
}

// ParseRole maps a query value to a known role. Unknown values become "all".
func ParseRole(raw string) Role {
	switch Role(raw) {
	case RoleFighter, RoleMarksman, RoleAssassin, RoleMage, RoleTank, RoleSupport:
		return Role(raw)
	default:
		return RoleAll
	}
}

// roleOverrides replaces the tag heuristic for specific item IDs.
// Empty by default; fill in if a well-known item lands in the wrong tab.
var roleOverrides = map[string][]Role{}

// MatchesRole reports whether the item belongs in the given class tab.
// An item may match several roles. RoleAll matches everything.
func (it Item) MatchesRole(role Role) bool {
	if role == RoleAll {
		return true
	}
	if roles, ok := roleOverrides[it.ID]; ok {
		for _, r := range roles {
			if r == role {
				return true
			}
		}
		return false
	}
	return it.inferredRoles()[role]
}

func (it Item) inferredRoles() map[Role]bool {
	out := make(map[Role]bool, 6)
	if it.Has("Damage") && (it.Has("Health") || it.Has("HealthRegen")) {
		out[RoleFighter] = true
	}
	if it.Has("CriticalStrike") || (it.Has("AttackSpeed") && it.Has("Damage")) {
		out[RoleMarksman] = true
	}
	if it.Has("ArmorPenetration") {
		out[RoleAssassin] = true
	}
	if it.Has("SpellDamage") {
		out[RoleMage] = true
	}
	if it.Has("Armor") || it.Has("SpellBlock") || it.Has("MagicResist") {
		out[RoleTank] = true
	}
	if it.Has("GoldPer") || it.Has("Aura") || (it.Has("ManaRegen") && !it.Has("SpellDamage")) {
		out[RoleSupport] = true
	}
	return out
}
