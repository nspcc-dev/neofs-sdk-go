package eacl

import (
	"strconv"

	v2acl "github.com/nspcc-dev/neofs-api-go/v2/acl"
)

// Action enumerates actions that may be applied within NeoFS access management.
// What and how specific Action affects depends on the specific context.
type Action uint32

const (
	ActionUnspecified Action = iota // undefined (zero)
	ActionAllow                     // allow the op
	ActionDeny                      // deny the op
)

// ActionUnknown is an Action value used to mark action as undefined.
// Deprecated: use ActionUnspecified instead.
const ActionUnknown = ActionUnspecified

// Operation enumerates operations on NeoFS resources under access control.
type Operation uint32

const (
	OperationUnspecified Operation = iota // undefined (zero)
	OperationGet                          // ObjectService.Get RPC
	OperationHead                         // ObjectService.Head RPC
	OperationPut                          // ObjectService.Put RPC
	OperationDelete                       // ObjectService.Delete RPC
	OperationSearch                       // ObjectService.Search RPC
	OperationRange                        // ObjectService.GetRange RPC
	OperationRangeHash                    // ObjectService.GetRangeHash RPC
)

// OperationUnknown is an Operation value used to mark operation as undefined.
// Deprecated: use OperationUnspecified instead.
const OperationUnknown = OperationUnspecified

// Role enumerates groups of subjects requesting access to NeoFS resources.
type Role uint32

const (
	RoleUnspecified Role = iota // undefined (zero)
	RoleUser                    // owner of the container requesting its objects
	RoleSystem                  // Deprecated: NeoFS storage and Inner Ring nodes
	RoleOthers                  // any other party
)

// RoleUnknown is a Role value used to mark role as undefined.
// Deprecated: use RoleUnspecified instead.
const RoleUnknown = RoleUnspecified

// Match enumerates operators to check attribute value compliance. What and how
// specific Match affects depends on the specific context.
type Match uint32

const (
	MatchUnspecified    Match = iota // undefined (zero)
	MatchStringEqual                 // string equality
	MatchStringNotEqual              // string inequality
	MatchNotPresent                  // attribute absence
	MatchNumGT                       // numeric "greater than" operator
	MatchNumGE                       // numeric "greater or equal than" operator
	MatchNumLT                       // is a numeric "less than" operator
	MatchNumLE                       // is a numeric "less or equal than" operator
)

// MatchUnknown is a Match value used to mark matcher as undefined.
// Deprecated: use MatchUnspecified instead.
const MatchUnknown = MatchUnspecified

// FilterHeaderType enumerates the classes of resource attributes processed
// within NeoFS access management.
type FilterHeaderType uint32

const (
	HeaderTypeUnspecified FilterHeaderType = iota // undefined (zero)
	HeaderFromRequest                             // protocol request X-Header
	HeaderFromObject                              // object attribute
	HeaderFromService                             // custom application-level attribute
)

// HeaderTypeUnknown is a FilterHeaderType value used to mark header type as undefined.
// Deprecated: use HeaderTypeUnspecified instead.
const HeaderTypeUnknown = HeaderTypeUnspecified

// ToV2 converts Action to v2 Action enum value.
func (a Action) ToV2() v2acl.Action { return v2acl.Action(a) }

// ActionFromV2 converts v2 Action enum value to Action.
func ActionFromV2(action v2acl.Action) Action { return Action(action) }

const (
	actionStringZero  = "ACTION_UNSPECIFIED"
	actionStringAllow = "ALLOW"
	actionStringDeny  = "DENY"
)

// ActionToString maps Action values to strings:
//   - 0: ACTION_UNSPECIFIED
//   - [ActionAllow]: ALLOW
//   - [ActionDeny]: DENY
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use [Action.String] to
// get any human-readable text for printing.
func ActionToString(a Action) string {
	switch a {
	default:
		return strconv.FormatUint(uint64(a), 10)
	case 0:
		return actionStringZero
	case ActionAllow:
		return actionStringAllow
	case ActionDeny:
		return actionStringDeny
	}
}

// ActionFromString maps strings to Action values in reverse to
// [ActionToString]. Returns false if s is incorrect.
func ActionFromString(s string) (Action, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return Action(n), true
		}
		return 0, false
	case "ACTION_UNSPECIFIED":
		return 0, true
	case "ALLOW":
		return ActionAllow, true
	case "DENY":
		return ActionDeny, true
	}
}

// EncodeToString returns string representation of Action.
//
// String mapping:
//   - ActionAllow: ALLOW;
//   - ActionDeny: DENY;
//   - ActionUnspecified, default: ACTION_UNSPECIFIED.
//
// Deprecated: use [ActionToString] instead.
func (a Action) EncodeToString() string { return ActionToString(a) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [ActionToString] and [ActionFromString] for consistent
// mapping.
func (a Action) String() string {
	return ActionToString(a)
}

// DecodeString parses Action from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [ActionFromString] instead.
func (a *Action) DecodeString(s string) bool {
	if v, ok := ActionFromString(s); ok {
		*a = v
		return true
	}
	return false
}

// ToV2 converts Operation to v2 Operation enum value.
func (o Operation) ToV2() v2acl.Operation { return v2acl.Operation(o) }

// OperationFromV2 converts v2 Operation enum value to Operation.
func OperationFromV2(operation v2acl.Operation) Operation { return Operation(operation) }

const (
	opStringZero      = "OPERATION_UNSPECIFIED"
	opStringGet       = "GET"
	opStringHead      = "HEAD"
	opStringPut       = "PUT"
	opStringDelete    = "DELETE"
	opStringSearch    = "SEARCH"
	opStringRange     = "GETRANGE"
	opStringRangeHash = "GETRANGEHASH"
)

// OperationToString maps Operation values to strings:
//   - 0: OPERATION_UNSPECIFIED
//   - [OperationGet]: GET
//   - [OperationHead]: HEAD
//   - [OperationPut]: PUT
//   - [OperationDelete]: DELETE
//   - [OperationSearch]: SEARCH
//   - [OperationRange]: GETRANGE
//   - [OperationRangeHash]: GETRANGEHASH
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use [Operation.String] to
// get any human-readable text for printing.
func OperationToString(op Operation) string {
	switch op {
	default:
		return strconv.FormatUint(uint64(op), 10)
	case 0:
		return opStringZero
	case OperationGet:
		return opStringGet
	case OperationHead:
		return opStringHead
	case OperationPut:
		return opStringPut
	case OperationDelete:
		return opStringDelete
	case OperationSearch:
		return opStringSearch
	case OperationRange:
		return opStringRange
	case OperationRangeHash:
		return opStringRangeHash
	}
}

// OperationFromString maps strings to Operation values in reverse to
// [OperationToString]. Returns false if s is incorrect.
func OperationFromString(s string) (Operation, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return Operation(n), true
		}
		return 0, false
	case "OPERATION_UNSPECIFIED":
		return 0, true
	case opStringGet:
		return OperationGet, true
	case opStringHead:
		return OperationHead, true
	case opStringPut:
		return OperationPut, true
	case opStringDelete:
		return OperationDelete, true
	case opStringSearch:
		return OperationSearch, true
	case opStringRange:
		return OperationRange, true
	case opStringRangeHash:
		return OperationRangeHash, true
	}
}

// EncodeToString returns string representation of Operation.
//
// String mapping:
//   - OperationGet: GET;
//   - OperationHead: HEAD;
//   - OperationPut: PUT;
//   - OperationDelete: DELETE;
//   - OperationSearch: SEARCH;
//   - OperationRange: GETRANGE;
//   - OperationRangeHash: GETRANGEHASH;
//   - OperationUnspecified, default: OPERATION_UNSPECIFIED.
//
// Deprecated: use [OperationToString] instead.
func (o Operation) EncodeToString() string { return OperationToString(o) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [OperationToString] and [OperationFromString] for
// consistent mapping.
func (o Operation) String() string {
	return OperationToString(o)
}

// DecodeString parses Operation from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [OperationFromString] instead.
func (o *Operation) DecodeString(s string) bool {
	if v, ok := OperationFromString(s); ok {
		*o = v
		return true
	}
	return false
}

// ToV2 converts Role to v2 Role enum value.
func (r Role) ToV2() v2acl.Role { return v2acl.Role(r) }

// RoleFromV2 converts v2 Role enum value to Role.
func RoleFromV2(role v2acl.Role) Role { return Role(role) }

const (
	roleStringZero   = "ROLE_UNSPECIFIED"
	roleStringUser   = "USER"
	roleStringSystem = "SYSTEM"
	roleStringOthers = "OTHERS"
)

// RoleToString maps Role values to strings:
//   - 0: ROLE_UNSPECIFIED
//   - [RoleUser]: USER
//   - [RoleSystem]: SYSTEM
//   - [RoleOthers]: OTHERS
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use [Role.String] to
// get any human-readable text for printing.
func RoleToString(r Role) string {
	switch r {
	default:
		return strconv.FormatUint(uint64(r), 10)
	case 0:
		return roleStringZero
	case RoleUser:
		return roleStringUser
	case RoleSystem:
		return roleStringSystem
	case RoleOthers:
		return roleStringOthers
	}
}

// RoleFromString maps strings to Role values in reverse to [RoleToString].
// Returns false if s is incorrect.
func RoleFromString(s string) (Role, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return Role(n), true
		}
		return 0, false
	case "ROLE_UNSPECIFIED":
		return 0, true
	case roleStringUser:
		return RoleUser, true
	case roleStringSystem:
		return RoleSystem, true
	case roleStringOthers:
		return RoleOthers, true
	}
}

// EncodeToString returns string representation of Role.
//
// String mapping:
//   - RoleUser: USER;
//   - RoleSystem: SYSTEM;
//   - RoleOthers: OTHERS;
//   - RoleUnspecified, default: ROLE_UNKNOWN.
//
// Deprecated: use [RoleToString] instead.
func (r Role) EncodeToString() string { return RoleToString(r) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [RoleToString] and [RoleFromString] for consistent mapping.
func (r Role) String() string {
	return RoleToString(r)
}

// DecodeString parses Role from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [RoleFromString] instead.
func (r *Role) DecodeString(s string) bool {
	if v, ok := RoleFromString(s); ok {
		*r = v
		return true
	}
	return false
}

// ToV2 converts Match to v2 MatchType enum value.
func (m Match) ToV2() v2acl.MatchType { return v2acl.MatchType(m) }

// MatchFromV2 converts v2 MatchType enum value to Match.
func MatchFromV2(match v2acl.MatchType) Match { return Match(match) }

const (
	matcherStringZero       = "MATCH_TYPE_UNSPECIFIED"
	matcherStringEqual      = "STRING_EQUAL"
	matcherStringNotEqual   = "STRING_NOT_EQUAL"
	matcherStringNotPresent = "NOT_PRESENT"
	matcherStringNumGT      = "NUM_GT"
	matcherStringNumGE      = "NUM_GE"
	matcherStringNumLT      = "NUM_LT"
	matcherStringNumLE      = "NUM_LE"
)

// MatcherToString maps Match values to strings:
//   - 0: MATCH_TYPE_UNSPECIFIED
//   - [MatchStringEqual]: STRING_EQUAL
//   - [MatchStringNotEqual]: STRING_NOT_EQUAL
//   - [MatchNotPresent]: NOT_PRESENT
//   - [MatchNumGT]: NUM_GT
//   - [MatchNumGE]: NUM_GE
//   - [MatchNumLT]: NUM_LT
//   - [MatchNumLE]: NUM_LE
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use [Match.String] to
// get any human-readable text for printing.
func MatcherToString(m Match) string {
	switch m {
	default:
		return strconv.FormatUint(uint64(m), 10)
	case 0:
		return matcherStringZero
	case MatchStringEqual:
		return matcherStringEqual
	case MatchStringNotEqual:
		return matcherStringNotEqual
	case MatchNotPresent:
		return matcherStringNotPresent
	case MatchNumGT:
		return matcherStringNumGT
	case MatchNumGE:
		return matcherStringNumGE
	case MatchNumLT:
		return matcherStringNumLT
	case MatchNumLE:
		return matcherStringNumLE
	}
}

// MatcherFromString maps strings to Match values in reverse to
// [MatcherToString]. Returns false if s is incorrect.
func MatcherFromString(s string) (Match, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return Match(n), true
		}
		return 0, false
	case "MATCH_TYPE_UNSPECIFIED":
		return 0, true
	case matcherStringEqual:
		return MatchStringEqual, true
	case matcherStringNotEqual:
		return MatchStringNotEqual, true
	case matcherStringNotPresent:
		return MatchNotPresent, true
	case matcherStringNumGT:
		return MatchNumGT, true
	case matcherStringNumGE:
		return MatchNumGE, true
	case matcherStringNumLT:
		return MatchNumLT, true
	case matcherStringNumLE:
		return MatchNumLE, true
	}
}

// EncodeToString returns string representation of Match.
//
// String mapping:
//   - MatchStringEqual: STRING_EQUAL;
//   - MatchStringNotEqual: STRING_NOT_EQUAL;
//   - MatchNotPresent: NOT_PRESENT;
//   - MatchNumGT: NUM_GT;
//   - MatchNumGE: NUM_GE;
//   - MatchNumLT: NUM_LT;
//   - MatchNumLE: NUM_LE;
//   - MatchUnspecified, default: MATCH_TYPE_UNSPECIFIED.
//
// Deprecated: use [MatcherToString] instead.
func (m Match) EncodeToString() string { return MatcherToString(m) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [MatcherToString] and [MatcherFromString] for consistent
// mapping.
func (m Match) String() string {
	return MatcherToString(m)
}

// DecodeString parses Match from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [MatcherFromString] instead.
func (m *Match) DecodeString(s string) bool {
	if v, ok := MatcherFromString(s); ok {
		*m = v
		return true
	}
	return false
}

// ToV2 converts FilterHeaderType to v2 HeaderType enum value.
func (h FilterHeaderType) ToV2() v2acl.HeaderType { return v2acl.HeaderType(h) }

// FilterHeaderTypeFromV2 converts v2 HeaderType enum value to FilterHeaderType.
func FilterHeaderTypeFromV2(header v2acl.HeaderType) FilterHeaderType {
	return FilterHeaderType(header)
}

const (
	headerTypeStringZero    = "HEADER_UNSPECIFIED"
	headerTypeStringRequest = "REQUEST"
	headerTypeStringObject  = "OBJECT"
	headerTypeStringService = "SERVICE"
)

// HeaderTypeToString maps FilterHeaderType values to strings:
//   - 0: HEADER_UNSPECIFIED
//   - [HeaderFromRequest]: REQUEST
//   - [HeaderFromObject]: OBJECT
//   - [HeaderFromService]: SERVICE
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format. Use
// [FilterHeaderType.String] to get any human-readable text for printing.
func HeaderTypeToString(h FilterHeaderType) string {
	switch h {
	default:
		return strconv.FormatUint(uint64(h), 10)
	case 0:
		return headerTypeStringZero
	case HeaderFromRequest:
		return headerTypeStringRequest
	case HeaderFromObject:
		return headerTypeStringObject
	case HeaderFromService:
		return headerTypeStringService
	}
}

// HeaderTypeFromString maps strings to FilterHeaderType values in reverse to
// [MatcherToString]. Returns false if s is incorrect.
func HeaderTypeFromString(s string) (FilterHeaderType, bool) {
	switch s {
	default:
		if n, err := strconv.ParseUint(s, 10, 32); err == nil {
			return FilterHeaderType(n), true
		}
		return 0, false
	case "HEADER_UNSPECIFIED":
		return 0, true
	case headerTypeStringRequest:
		return HeaderFromRequest, true
	case headerTypeStringObject:
		return HeaderFromObject, true
	case headerTypeStringService:
		return HeaderFromService, true
	}
}

// EncodeToString returns string representation of FilterHeaderType.
//
// String mapping:
//   - HeaderFromRequest: REQUEST;
//   - HeaderFromObject: OBJECT;
//   - HeaderTypeUnspecified, default: HEADER_UNSPECIFIED.
//
// Deprecated: use [HeaderTypeToString] instead.
func (h FilterHeaderType) EncodeToString() string { return HeaderTypeToString(h) }

// String implements fmt.Stringer.
//
// String is designed to be human-readable, and its format MAY differ between
// SDK versions. Use [HeaderTypeToString] and [HeaderTypeFromString] for
// consistent mapping.
func (h FilterHeaderType) String() string {
	return HeaderTypeToString(h)
}

// DecodeString parses FilterHeaderType from a string representation.
// It is a reverse action to EncodeToString().
//
// Returns true if s was parsed successfully.
// Deprecated: use [HeaderTypeFromString] instead.
func (h *FilterHeaderType) DecodeString(s string) bool {
	if v, ok := HeaderTypeFromString(s); ok {
		*h = v
		return true
	}
	return false
}
