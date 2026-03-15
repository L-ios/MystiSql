package rbac

import (
	"fmt"
	"strings"
)

// Permission represents an access right in the simplified format:
// <instance>:<database>:<action>
type Permission struct {
	Instance string
	Database string
	Action   string
}

// ParsePermissionFromString parses a string like "inst:db:action" into a Permission.
func ParsePermissionFromString(s string) (Permission, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return Permission{}, fmt.Errorf("invalid permission format: %s", s)
	}
	return Permission{Instance: parts[0], Database: parts[1], Action: parts[2]}, nil
}

func (p Permission) String() string {
	return p.Instance + ":" + p.Database + ":" + p.Action
}
