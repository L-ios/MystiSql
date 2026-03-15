package oidc

// MapRolesFromClaims extracts roles from a claims payload using the provided
// roleClaim key. The input may be a string, a slice of strings, or a slice of
// interfaces. Returns a flat slice of role names.
func MapRolesFromClaims(claims map[string]interface{}, roleClaim string) []string {
	if roleClaim == "" {
		// Fallback to a common default claim name
		roleClaim = "roles"
	}
	raw, ok := claims[roleClaim]
	if !ok {
		// Try a generic "roles" key as a fallback
		if v, ok := claims["roles"]; ok {
			raw = v
		} else {
			return nil
		}
	}
	return toStringSlice(raw)
}

func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []string:
		return t
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, a := range t {
			if s, ok := a.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
