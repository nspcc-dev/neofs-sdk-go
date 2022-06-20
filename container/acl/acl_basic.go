package acl

import (
	"fmt"
	"strconv"
	"strings"
)

// Basic represents basic part of the NeoFS container's ACL. It includes
// common (pretty simple) access rules for operations inside the container.
// See NeoFS Specification for details.
//
// One can find some similarities with the traditional Unix permission, such as
//  division into scopes: user, group, others
//  op-permissions: read, write, etc.
//  sticky bit
// However, these similarities should only be used for better understanding,
// in general these mechanisms are different.
//
// Instances can be created using built-in var declaration, but look carefully
// at the default values, and how individual permissions are regulated.
// Some frequently used values are presented in exported variables.
//
// Basic instances are comparable: values can be compared directly using
// == operator.
type Basic struct {
	bits uint32
}

// FromBits decodes Basic from the numerical representation.
//
// See also Bits.
func (x *Basic) FromBits(bits uint32) {
	x.bits = bits
}

// Bits returns numerical encoding of Basic.
//
// See also FromBits.
func (x Basic) Bits() uint32 {
	return x.bits
}

// common bit sections.
const (
	opAmount  = 7
	bitsPerOp = 4

	bitPosFinal  = opAmount * bitsPerOp
	bitPosSticky = bitPosFinal + 1
)

// per-op bit order.
const (
	opBitPosBearer uint8 = iota
	opBitPosOthers
	opBitPosContainer
	opBitPosOwner
)

// DisableExtension makes Basic FINAL. FINAL indicates the ACL non-extendability
// in the related container.
//
// See also Extendable.
func (x *Basic) DisableExtension() {
	setBit(&x.bits, bitPosFinal)
}

// Extendable checks if Basic is NOT made FINAL using DisableExtension.
//
// Zero Basic is NOT FINAL or extendable.
func (x Basic) Extendable() bool {
	return !isBitSet(x.bits, bitPosFinal)
}

// MakeSticky makes Basic STICKY. STICKY indicates that only the owner of any
// particular object is allowed to operate on it.
//
// See also Sticky.
func (x *Basic) MakeSticky() {
	setBit(&x.bits, bitPosSticky)
}

// Sticky checks if Basic is made STICKY using MakeSticky.
//
// Zero Basic is NOT STICKY.
func (x Basic) Sticky() bool {
	return isBitSet(x.bits, bitPosSticky)
}

// checks if op is used by the storage nodes within replication mechanism.
func isReplicationOp(op Op) bool {
	//nolint:exhaustive
	switch op {
	case
		OpObjectGet,
		OpObjectHead,
		OpObjectPut,
		OpObjectSearch,
		OpObjectHash:
		return true
	}

	return false
}

// AllowOp allows the parties with the given role to the given operation.
// Op MUST be one of the Op enumeration. Role MUST be one of:
//  RoleOwner
//  RoleContainer
//  RoleOthers
// and if role is RoleContainer, op MUST NOT be:
//  OpObjectGet
//  OpObjectHead
//  OpObjectPut
//  OpObjectSearch
//  OpObjectHash
//
// See also IsOpAllowed.
func (x *Basic) AllowOp(op Op, role Role) {
	var bitPos uint8

	switch role {
	default:
		panic(fmt.Sprintf("unable to set rules for unsupported role %v", role))
	case RoleInnerRing:
		panic("basic ACL MUST NOT be modified for Inner Ring")
	case RoleOwner:
		bitPos = opBitPosOwner
	case RoleContainer:
		if isReplicationOp(op) {
			panic("basic ACL for container replication ops MUST NOT be modified")
		}

		bitPos = opBitPosContainer
	case RoleOthers:
		bitPos = opBitPosOthers
	}

	setOpBit(&x.bits, op, bitPos)
}

// IsOpAllowed checks if parties with the given role are allowed to the given op
// according to the Basic rules. Op MUST be one of the Op enumeration.
// Role MUST be one of the Role enumeration.
//
// Members with RoleContainer role have exclusive default access to the
// operations of the data replication mechanism:
//  OpObjectGet
//  OpObjectHead
//  OpObjectPut
//  OpObjectSearch
//  OpObjectHash
//
// RoleInnerRing members are allowed to data audit ops only:
//  OpObjectGet
//  OpObjectHead
//  OpObjectHash
//  OpObjectSearch
//
// Zero Basic prevents any role from accessing any operation in the absence
// of default rights.
//
// See also AllowOp.
func (x Basic) IsOpAllowed(op Op, role Role) bool {
	var bitPos uint8

	switch role {
	default:
		panic(fmt.Sprintf("role is unsupported %v", role))
	case RoleInnerRing:
		switch op {
		case
			OpObjectGet,
			OpObjectHead,
			OpObjectHash,
			OpObjectSearch:
			return true
		default:
			return false
		}
	case RoleOwner:
		bitPos = opBitPosOwner
	case RoleContainer:
		if isReplicationOp(op) {
			return true
		}

		bitPos = opBitPosContainer
	case RoleOthers:
		bitPos = opBitPosOthers
	}

	return isOpBitSet(x.bits, op, bitPos)
}

// AllowBearerRules allows bearer to provide extended ACL rules for the given
// operation. Bearer rules doesn't depend on container ACL
// // extensibility.
//
// See also AllowedBearerRules.
func (x *Basic) AllowBearerRules(op Op) {
	setOpBit(&x.bits, op, opBitPosBearer)
}

// AllowedBearerRules checks if bearer rules are allowed using AllowBearerRules.
// Op MUST be one of the Op enumeration.
//
// Zero Basic disallows bearer rules for any op.
func (x Basic) AllowedBearerRules(op Op) bool {
	return isOpBitSet(x.bits, op, opBitPosBearer)
}

// EncodeToString encodes Basic into hexadecimal string.
//
// See also DecodeString.
func (x Basic) EncodeToString() string {
	return strconv.FormatUint(uint64(x.bits), 16)
}

// Names of the frequently used Basic values.
const (
	NamePrivate              = "private"
	NamePrivateExtended      = "eacl-private"
	NamePublicRO             = "public-read"
	NamePublicROExtended     = "eacl-public-read"
	NamePublicRW             = "public-read-write"
	NamePublicRWExtended     = "eacl-public-read-write"
	NamePublicAppend         = "public-append"
	NamePublicAppendExtended = "eacl-public-append"
)

// Frequently used Basic values (each value MUST NOT be modified, make a
// copy instead).
var (
	Private              Basic // private
	PrivateExtended      Basic // eacl-private
	PublicRO             Basic // public-read
	PublicROExtended     Basic // eacl-public-read
	PublicRW             Basic // public-read-write
	PublicRWExtended     Basic // eacl-public-read-write
	PublicAppend         Basic // public-append
	PublicAppendExtended Basic // eacl-public-append
)

// DecodeString decodes string calculated using EncodeToString. Also supports
// human-readable names (Name* constants).
func (x *Basic) DecodeString(s string) (e error) {
	switch s {
	case NamePrivate:
		*x = Private
	case NamePrivateExtended:
		*x = PrivateExtended
	case NamePublicRO:
		*x = PublicRO
	case NamePublicROExtended:
		*x = PublicROExtended
	case NamePublicRW:
		*x = PublicRW
	case NamePublicRWExtended:
		*x = PublicRWExtended
	case NamePublicAppend:
		*x = PublicAppend
	case NamePublicAppendExtended:
		*x = PublicAppendExtended
	default:
		s = strings.TrimPrefix(strings.ToLower(s), "0x")

		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return fmt.Errorf("parse hex: %w", err)
		}

		x.bits = uint32(v)
	}

	return nil
}
