package types

import (
	"encoding/json"
	"strings"
)

type IsolationLevel int

const (
	IsolationLevelDefault IsolationLevel = iota
	IsolationLevelReadUncommitted
	IsolationLevelReadCommitted
	IsolationLevelRepeatableRead
	IsolationLevelSerializable
)

func (il IsolationLevel) String() string {
	switch il {
	case IsolationLevelReadUncommitted:
		return "READ UNCOMMITTED"
	case IsolationLevelReadCommitted:
		return "READ COMMITTED"
	case IsolationLevelRepeatableRead:
		return "REPEATABLE READ"
	case IsolationLevelSerializable:
		return "SERIALIZABLE"
	default:
		return "DEFAULT"
	}
}

func (il IsolationLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(il.String())
}

func (il *IsolationLevel) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		var i int
		if err := json.Unmarshal(data, &i); err != nil {
			return err
		}
		*il = IsolationLevel(i)
		return nil
	}
	str := strings.ToUpper(strings.TrimSpace(s))
	str = strings.ReplaceAll(str, "_", " ")
	switch str {
	case "READ UNCOMMITTED", "READUNCOMMITTED":
		*il = IsolationLevelReadUncommitted
	case "READ COMMITTED", "READCOMMITTED":
		*il = IsolationLevelReadCommitted
	case "REPEATABLE READ", "REPEATABLEREAD":
		*il = IsolationLevelRepeatableRead
	case "SERIALIZABLE":
		*il = IsolationLevelSerializable
	default:
		*il = IsolationLevelDefault
	}
	return nil
}
