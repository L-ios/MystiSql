package types

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
