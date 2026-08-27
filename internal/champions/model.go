package champions

// Champion is the page-facing view of a Data Dragon champion.
type Champion struct {
	ID         string
	Name       string
	Title      string
	Roles      []Role
	RoleLabels []string
	IconURL    string
	SplashURL  string
	Passive    Ability
	Spells     []Ability
}

// Ability is the innate (P) or one of Q/W/E/R.
type Ability struct {
	Key         string
	Name        string
	Description string
	IconURL     string
	Cooldown    string
}

// HasRole reports whether the champion carries the given class tag.
func (c Champion) HasRole(role Role) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Matches reports whether the champion belongs in the current checkbox set.
// An empty set matches everyone (OR of nothing would otherwise hide the roster).
func (c Champion) Matches(roles []Role) bool {
	if len(roles) == 0 {
		return true
	}
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}
