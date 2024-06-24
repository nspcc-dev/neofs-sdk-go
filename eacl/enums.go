package eacl

import (
	"strconv"
)

// Action enumerates actions that may be applied within NeoFS access management.
// What and how specific Action affects depends on the specific context.
type Action uint32

// All supported [Action] values.
const (
	_           Action = iota
	ActionAllow        // allows something
	ActionDeny         // denies something
)

// Role enumerates groups of subjects requesting access to NeoFS resources.
type Role uint32

// All supported [Role] values.
const (
	_          Role = iota
	RoleUser        // owner of the container requesting its objects
	RoleSystem      // Deprecated: NeoFS storage and Inner Ring nodes
	RoleOthers      // any other party
)

// Match enumerates operators to check attribute value compliance. What and how
// specific Match affects depends on the specific context.
type Match uint32

// All supported Match values.
const (
	_                   Match = iota
	MatchStringEqual          // string equality
	MatchStringNotEqual       // string inequality
	MatchNotPresent           // attribute absence
	MatchNumGT                // numeric "greater than" operator
	MatchNumGE                // numeric "greater or equal than" operator
	MatchNumLT                // is a numeric "less than" operator
	MatchNumLE                // is a numeric "less or equal than" operator
)

// AttributeType enumerates the classes of resource attributes processed within
// NeoFS access management.
type AttributeType uint32

const (
	_                      AttributeType = iota
	AttributeAPIRequest                  // API request X-Header
	AttributeObject                      // object attribute
	AttributeCustomService               // custom service attribute
)

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (a Action) String() string {
	switch a {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(a), 10)
	case ActionAllow:
		return "ALLOW"
	case ActionDeny:
		return "DENY"
	}
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (r Role) String() string {
	switch r {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(r), 10)
	case RoleUser:
		return "USER"
	case RoleSystem:
		return "SYSTEM"
	case RoleOthers:
		return "OTHERS"
	}
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (m Match) String() string {
	switch m {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(m), 10)
	case MatchStringEqual:
		return "STRING_EQUAL"
	case MatchStringNotEqual:
		return "STRING_NOT_EQUAL"
	case MatchNotPresent:
		return "NOT_PRESENT"
	case MatchNumGT:
		return "NUMERIC_GT"
	case MatchNumGE:
		return "NUMERIC_GE"
	case MatchNumLT:
		return "NUMERIC_LT"
	case MatchNumLE:
		return "NUMERIC_LE"
	}
}

// String implements [fmt.Stringer].
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions.
func (x AttributeType) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case AttributeAPIRequest:
		return "API_REQUEST"
	case AttributeObject:
		return "OBJECT"
	case AttributeCustomService:
		return "CUSTOM_SERVICE"
	}
}
