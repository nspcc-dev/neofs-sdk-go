package eacl

import (
	"strconv"
)

// Action enumerates actions that may be applied within NeoFS access management.
// What and how specific Action affects depends on the specific context.
type Action int32

const (
	ActionUnspecified Action = iota // undefined (zero)
	ActionAllow                     // allow the op
	ActionDeny                      // deny the op
)

// ActionUnknown is an Action value used to mark action as undefined.
// Deprecated: use ActionUnspecified instead.
const ActionUnknown = ActionUnspecified

// Operation enumerates operations on NeoFS resources under access control.
type Operation int32

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
type Role int32

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
type Match int32

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
type FilterHeaderType int32

const (
	HeaderTypeUnspecified FilterHeaderType = iota // undefined (zero)
	HeaderFromRequest                             // protocol request X-Header
	HeaderFromObject                              // object attribute
	HeaderFromService                             // custom application-level attribute
)

// HeaderTypeUnknown is a FilterHeaderType value used to mark header type as undefined.
// Deprecated: use HeaderTypeUnspecified instead.
const HeaderTypeUnknown = HeaderTypeUnspecified

const (
	actionStringZero  = "ACTION_UNSPECIFIED"
	actionStringAllow = "ALLOW"
	actionStringDeny  = "DENY"
)

// EncodeToString returns string representation of Action.
//
// String mapping:
//   - ActionAllow: ALLOW;
//   - ActionDeny: DENY;
//   - ActionUnspecified, default: ACTION_UNSPECIFIED.
//
// Deprecated: use [Action.String] instead.
func (a Action) EncodeToString() string { return a.String() }

// String implements [fmt.Stringer] with the following string mapping:
//   - 0: ACTION_UNSPECIFIED
//   - [ActionAllow]: ALLOW
//   - [ActionDeny]: DENY
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [Action.DecodeString].
func (a Action) String() string {
	switch a {
	default:
		return strconv.FormatInt(int64(a), 10)
	case 0:
		return actionStringZero
	case ActionAllow:
		return actionStringAllow
	case ActionDeny:
		return actionStringDeny
	}
}

// DecodeString parses Action from a string representation. It is a reverse
// action to [Action.String].
//
// Returns true if s was parsed successfully.
func (a *Action) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*a = Action(n)
	case actionStringZero:
		*a = 0
	case actionStringAllow:
		*a = ActionAllow
	case actionStringDeny:
		*a = ActionDeny
	}
	return true
}

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
// Deprecated: use [Operation.String] instead.
func (o Operation) EncodeToString() string { return o.String() }

// String implements [fmt.Stringer] with the following string mapping:
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
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [Operation.DecodeString].
func (o Operation) String() string {
	switch o {
	default:
		return strconv.FormatInt(int64(o), 10)
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

// DecodeString parses Operation from a string representation. It is a reverse
// action to [Operation.String].
//
// Returns true if s was parsed successfully.
func (o *Operation) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*o = Operation(n)
	case opStringZero:
		*o = 0
	case opStringGet:
		*o = OperationGet
	case opStringHead:
		*o = OperationHead
	case opStringPut:
		*o = OperationPut
	case opStringDelete:
		*o = OperationDelete
	case opStringSearch:
		*o = OperationSearch
	case opStringRange:
		*o = OperationRange
	case opStringRangeHash:
		*o = OperationRangeHash
	}
	return true
}

const (
	roleStringZero   = "ROLE_UNSPECIFIED"
	roleStringUser   = "USER"
	roleStringSystem = "SYSTEM"
	roleStringOthers = "OTHERS"
)

// EncodeToString returns string representation of Role.
//
// String mapping:
//   - RoleUser: USER;
//   - RoleSystem: SYSTEM;
//   - RoleOthers: OTHERS;
//   - RoleUnspecified, default: ROLE_UNKNOWN.
//
// Deprecated: use [Role.String] instead.
func (r Role) EncodeToString() string { return r.String() }

// String implements [fmt.Stringer] with the following string mapping:
//   - 0: ROLE_UNSPECIFIED
//   - [RoleUser]: USER
//   - [RoleSystem]: SYSTEM
//   - [RoleOthers]: OTHERS
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [Role.DecodeString].
func (r Role) String() string {
	switch r {
	default:
		return strconv.FormatInt(int64(r), 10)
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

// DecodeString parses Role from a string representation. It is a reverse action
// to [Role.String].
//
// Returns true if s was parsed successfully.
func (r *Role) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*r = Role(n)
	case roleStringZero:
		*r = 0
	case roleStringUser:
		*r = RoleUser
	case roleStringSystem:
		*r = RoleSystem
	case roleStringOthers:
		*r = RoleOthers
	}
	return true
}

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
// Deprecated: use [Match.String] instead.
func (m Match) EncodeToString() string { return m.String() }

// String implements [fmt.Stringer] with the following string mapping:
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
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [Match.DecodeString].
func (m Match) String() string {
	switch m {
	default:
		return strconv.FormatInt(int64(m), 10)
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

// DecodeString parses Match from a string representation. It is a reverse action
// to [Match.String].
//
// Returns true if s was parsed successfully.
func (m *Match) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*m = Match(n)
	case matcherStringZero:
		*m = 0
	case matcherStringEqual:
		*m = MatchStringEqual
	case matcherStringNotEqual:
		*m = MatchStringNotEqual
	case matcherStringNotPresent:
		*m = MatchNotPresent
	case matcherStringNumGT:
		*m = MatchNumGT
	case matcherStringNumGE:
		*m = MatchNumGE
	case matcherStringNumLT:
		*m = MatchNumLT
	case matcherStringNumLE:
		*m = MatchNumLE
	}
	return true
}

const (
	headerTypeStringZero    = "HEADER_UNSPECIFIED"
	headerTypeStringRequest = "REQUEST"
	headerTypeStringObject  = "OBJECT"
	headerTypeStringService = "SERVICE"
)

// EncodeToString returns string representation of FilterHeaderType.
//
// String mapping:
//   - HeaderFromRequest: REQUEST;
//   - HeaderFromObject: OBJECT;
//   - HeaderTypeUnspecified, default: HEADER_UNSPECIFIED.
//
// Deprecated: use [HeaderTypeToString] instead.
func (h FilterHeaderType) EncodeToString() string { return h.String() }

// String implements [fmt.Stringer] with the following string mapping:
//   - 0: HEADER_UNSPECIFIED
//   - [HeaderFromRequest]: REQUEST
//   - [HeaderFromObject]: OBJECT
//   - [HeaderFromService]: SERVICE
//
// All other values are base-10 integers.
//
// The mapping is consistent and resilient to lib updates. At the same time,
// please note that this is not a NeoFS protocol format.
//
// String is reverse to [FilterHeaderType.DecodeString].
func (h FilterHeaderType) String() string {
	switch h {
	default:
		return strconv.FormatInt(int64(h), 10)
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

// DecodeString parses FilterHeaderType from a string representation. It is a
// reverse action to [FilterHeaderType.String].
//
// Returns true if s was parsed successfully.
func (h *FilterHeaderType) DecodeString(s string) bool {
	switch s {
	default:
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return false
		}
		*h = FilterHeaderType(n)
	case headerTypeStringZero:
		*h = 0
	case headerTypeStringRequest:
		*h = HeaderFromRequest
	case headerTypeStringObject:
		*h = HeaderFromObject
	case headerTypeStringService:
		*h = HeaderFromService
	}
	return true
}
