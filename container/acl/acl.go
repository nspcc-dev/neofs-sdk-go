package acl

import "strconv"

// Op enumerates operations under access control inside container.
// Non-positive values are reserved and depend on context (e.g. unsupported op).
//
// Note that type conversion from- and to numerical types is not recommended,
// use corresponding constants and/or methods instead.
type Op uint32

const (
	opZero Op = iota // extreme value for testing

	OpObjectGet    // Object.Get rpc
	OpObjectHead   // Object.Head rpc
	OpObjectPut    // Object.Put rpc
	OpObjectDelete // Object.Delete rpc
	OpObjectSearch // Object.Search rpc
	OpObjectRange  // Object.GetRange rpc
	OpObjectHash   // Object.GetRangeHash rpc

	opLast // extreme value for testing
)

// String implements [fmt.Stringer].
func (x Op) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case OpObjectGet:
		return "OBJECT_GET"
	case OpObjectHead:
		return "OBJECT_HEAD"
	case OpObjectPut:
		return "OBJECT_PUT"
	case OpObjectDelete:
		return "OBJECT_DELETE"
	case OpObjectSearch:
		return "OBJECT_SEARCH"
	case OpObjectRange:
		return "OBJECT_RANGE"
	case OpObjectHash:
		return "OBJECT_HASH"
	}
}

// Role enumerates roles covered by container ACL. Each role represents
// some party which can be authenticated during container op execution.
// Non-positive values are reserved and depend on context (e.g. unsupported role).
//
// Note that type conversion from- and to numerical types is not recommended,
// use corresponding constants and/or methods instead.
type Role uint32

const (
	roleZero Role = iota // extreme value for testing

	RoleOwner     // container owner
	RoleContainer // nodes of the related container
	RoleInnerRing // Inner Ring nodes
	RoleOthers    // all others

	roleLast // extreme value for testing
)

// String implements [fmt.Stringer].
func (x Role) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case RoleOwner:
		return "OWNER"
	case RoleContainer:
		return "CONTAINER"
	case RoleInnerRing:
		return "INNER_RING"
	case RoleOthers:
		return "OTHERS"
	}
}
