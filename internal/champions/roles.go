package champions

// Role is a champion class filter. An empty selection means "all champions".
type Role string

const (
	RoleFighter  Role = "fighter"
	RoleMarksman Role = "marksman"
	RoleAssassin Role = "assassin"
	RoleMage     Role = "mage"
	RoleTank     Role = "tank"
	RoleSupport  Role = "support"
)

// RoleOption is a checkbox in the role pane.
type RoleOption struct {
	ID    Role
	Label string
	Tag   string
}

// AllRoles is the checklist, in shop order. There is no "Todos": empty checks
// already mean every champion.
var AllRoles = []RoleOption{
	{ID: RoleFighter, Label: "Luchador", Tag: "Fighter"},
	{ID: RoleMarksman, Label: "Tirador", Tag: "Marksman"},
	{ID: RoleAssassin, Label: "Asesino", Tag: "Assassin"},
	{ID: RoleMage, Label: "Mago", Tag: "Mage"},
	{ID: RoleTank, Label: "Tanque", Tag: "Tank"},
	{ID: RoleSupport, Label: "Soporte", Tag: "Support"},
}

// ParseRoles keeps only known role IDs, in checklist order, without duplicates.
func ParseRoles(raw []string) []Role {
	seen := make(map[Role]struct{}, len(raw))
	var out []Role
	for _, opt := range AllRoles {
		for _, id := range raw {
			if Role(id) != opt.ID {
				continue
			}
			if _, ok := seen[opt.ID]; ok {
				continue
			}
			seen[opt.ID] = struct{}{}
			out = append(out, opt.ID)
		}
	}
	return out
}

func rolesFromTags(tags []string) ([]Role, []string) {
	seen := make(map[Role]struct{}, len(tags))
	var roles []Role
	var labels []string
	for _, opt := range AllRoles {
		for _, tag := range tags {
			if tag != opt.Tag {
				continue
			}
			if _, ok := seen[opt.ID]; ok {
				continue
			}
			seen[opt.ID] = struct{}{}
			roles = append(roles, opt.ID)
			labels = append(labels, opt.Label)
		}
	}
	return roles, labels
}
