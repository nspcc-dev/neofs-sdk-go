package eacl

import (
	"strconv"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
)

// Action enumerates actions that may be applied within NeoFS access management.
// What and how specific Action affects depends on the specific context.
//
// Non-positive values are reserved and depend on context (e.g. unsupported
// action).
//
// Note that type conversion from- and to numerical types is not recommended:
// use enum names only.
type Action v2acl.Action

// All supported Action values.
const (
	_           Action = iota
	ActionAllow        // allows something
	ActionDeny         // denies something
	lastAction
)

// String implements [fmt.Stringer].
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

// Role enumerates groups of subjects requesting access to NeoFS resources.
//
// Non-positive values are reserved and depend on context (e.g. unsupported
// role).
//
// Note that type conversion from- and to numerical types is not recommended:
// use enum names only.
type Role uint32

// All supported Match values.
const (
	_                  Role = iota
	RoleContainerOwner      // owner of the container requesting its objects
	RoleSystem              // Deprecated: NeoFS storage and Inner Ring nodes
	RoleOthers              // any other party
	lastRole
)

// String implements [fmt.Stringer].
func (x Role) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case RoleContainerOwner:
		return "CONTAINER_OWNER"
	case RoleSystem:
		return "SYSTEM"
	case RoleOthers:
		return "OTHERS"
	}
}

// Matcher enumerates operators to check value compliance. What and how specific
// Match affects depends on the specific context.
//
// Non-positive values are reserved and depend on context (e.g. unsupported
// matcher).
//
// Note that type conversion from- and to numerical types is not recommended:
// use enum names only.
type Matcher v2acl.MatchType

// All supported Match values.
const (
	_                   Matcher = iota
	MatchStringEqual            // string equality
	MatchStringNotEqual         // string inequality
	lastMatcher
)

// String implements [fmt.Stringer].
func (m Matcher) String() string {
	switch m {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(m), 10)
	case MatchStringEqual:
		return "STRING_EQUAL"
	case MatchStringNotEqual:
		return "STRING_NOT_EQUAL"
	}
}

// HeaderType enumerates the types of headers processed within NeoFS access
// management.
//
// Non-positive values are reserved and depend on context (e.g. unsupported
// type).
//
// Note that type conversion from- and to numerical types is not recommended:
// use enum names only.
type HeaderType v2acl.HeaderType

const (
	_                 HeaderType = iota
	HeaderFromRequest            // request X-Header
	HeaderFromObject             // object header
	HeaderFromService            // custom service header
	lastHeaderType
)

// String implements [fmt.Stringer].
func (h HeaderType) String() string {
	switch h {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(h), 10)
	case HeaderFromRequest:
		return "REQUEST"
	case HeaderFromObject:
		return "OBJECT"
	case HeaderFromService:
		return "SERVICE"
	}
}
