package champions

// Filter returns champions that have every selected role, keeping catalog order.
func Filter(catalog []Champion, roles []Role) []Champion {
	out := make([]Champion, 0, len(catalog))
	for _, c := range catalog {
		if c.Matches(roles) {
			out = append(out, c)
		}
	}
	return out
}
