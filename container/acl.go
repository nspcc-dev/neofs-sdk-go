package container

import "strconv"

// ACLOp enumerates operations under access control inside container.
// Non-positive values are reserved and depend on context (e.g. unsupported op).
//
// Note that type conversion from- and to numerical types is not recommended,
// use corresponding constants and/or methods instead.
type ACLOp uint32

const (
	aclOpZero ACLOp = iota // extreme value for testing

	ACLOpObjectGet    // Object.Get rpc
	ACLOpObjectHead   // Object.Head rpc
	ACLOpObjectPut    // Object.Put rpc
	ACLOpObjectDelete // Object.Delete rpc
	ACLOpObjectSearch // Object.Search rpc
	ACLOpObjectRange  // Object.GetRange rpc
	ACLOpObjectHash   // Object.GetRangeHash rpc

	aclOpLast // extreme value for testing
)

// String implements fmt.Stringer.
func (x ACLOp) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case ACLOpObjectGet:
		return "OBJECT_GET"
	case ACLOpObjectHead:
		return "OBJECT_HEAD"
	case ACLOpObjectPut:
		return "OBJECT_PUT"
	case ACLOpObjectDelete:
		return "OBJECT_DELETE"
	case ACLOpObjectSearch:
		return "OBJECT_SEARCH"
	case ACLOpObjectRange:
		return "OBJECT_RANGE"
	case ACLOpObjectHash:
		return "OBJECT_HASH"
	}
}

// ACLRole enumerates roles covered by container ACL. Each role represents
// some party which can be authenticated during container op execution.
// Non-positive values are reserved and depend on context (e.g. unsupported role).
//
// Note that type conversion from- and to numerical types is not recommended,
// use corresponding constants and/or methods instead.
type ACLRole uint32

const (
	aclRoleZero ACLRole = iota // extreme value for testing

	ACLRoleOwner     // container owner
	ACLRoleContainer // nodes of the related container
	ACLRoleInnerRing // Inner Ring nodes
	ACLRoleOthers    // all others

	aclRoleLast // extreme value for testing
)

// String implements fmt.Stringer.
func (x ACLRole) String() string {
	switch x {
	default:
		return "UNKNOWN#" + strconv.FormatUint(uint64(x), 10)
	case ACLRoleOwner:
		return "OWNER"
	case ACLRoleContainer:
		return "CONTAINER"
	case ACLRoleInnerRing:
		return "INNER_RING"
	case ACLRoleOthers:
		return "OTHERS"
	}
}
