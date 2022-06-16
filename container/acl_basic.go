package container

import (
	"fmt"
	"strconv"
	"strings"
)

// BasicACL represents basic part of the NeoFS container's ACL. It includes
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
// Some frequently used values are presented in BasicACL* values.
//
// BasicACL instances are comparable: values can be compared directly using
// == operator.
type BasicACL struct {
	bits uint32
}

func (x *BasicACL) fromUint32(num uint32) {
	x.bits = num
}

func (x BasicACL) toUint32() uint32 {
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

// DisableExtension makes BasicACL FINAL. FINAL indicates the ACL non-extendability
// in the related container.
//
// See also Extendable.
func (x *BasicACL) DisableExtension() {
	setBit(&x.bits, bitPosFinal)
}

// Extendable checks if BasicACL is NOT made FINAL using DisableExtension.
//
// Zero BasicACL is NOT FINAL or extendable.
func (x BasicACL) Extendable() bool {
	return !isBitSet(x.bits, bitPosFinal)
}

// MakeSticky makes BasicACL STICKY. STICKY indicates that only the owner of any
// particular object is allowed to operate on it.
//
// See also Sticky.
func (x *BasicACL) MakeSticky() {
	setBit(&x.bits, bitPosSticky)
}

// Sticky checks if BasicACL is made STICKY using MakeSticky.
//
// Zero BasicACL is NOT STICKY.
func (x BasicACL) Sticky() bool {
	return isBitSet(x.bits, bitPosSticky)
}

// checks if op is used by the storage nodes within replication mechanism.
func isReplicationOp(op ACLOp) bool {
	//nolint:exhaustive
	switch op {
	case
		ACLOpObjectGet,
		ACLOpObjectHead,
		ACLOpObjectPut,
		ACLOpObjectSearch,
		ACLOpObjectHash:
		return true
	}

	return false
}

// AllowOp allows the parties with the given role to the given operation.
// Op MUST be one of the ACLOp enumeration. Role MUST be one of:
//  ACLRoleOwner
//  ACLRoleContainer
//  ACLRoleOthers
// and if role is ACLRoleContainer, op MUST NOT be:
//  ACLOpObjectGet
//  ACLOpObjectHead
//  ACLOpObjectPut
//  ACLOpObjectSearch
//  ACLOpObjectHash
//
// See also IsOpAllowed.
func (x *BasicACL) AllowOp(op ACLOp, role ACLRole) {
	var bitPos uint8

	switch role {
	default:
		panic(fmt.Sprintf("unable to set rules for unsupported role %v", role))
	case ACLRoleInnerRing:
		panic("basic ACL MUST NOT be modified for Inner Ring")
	case ACLRoleOwner:
		bitPos = opBitPosOwner
	case ACLRoleContainer:
		if isReplicationOp(op) {
			panic("basic ACL for container replication ops MUST NOT be modified")
		}

		bitPos = opBitPosContainer
	case ACLRoleOthers:
		bitPos = opBitPosOthers
	}

	setOpBit(&x.bits, op, bitPos)
}

// IsOpAllowed checks if parties with the given role are allowed to the given op
// according to the BasicACL rules. Op MUST be one of the ACLOp enumeration.
// Role MUST be one of the ACLRole enumeration.
//
// Members with ACLRoleContainer role have exclusive default access to the
// operations of the data replication mechanism:
//  ACLOpObjectGet
//  ACLOpObjectHead
//  ACLOpObjectPut
//  ACLOpObjectSearch
//  ACLOpObjectHash
//
// ACLRoleInnerRing members are allowed to data audit ops only:
//  ACLOpObjectGet
//  ACLOpObjectHead
//  ACLOpObjectHash
//  ACLOpObjectSearch
//
// Zero BasicACL prevents any role from accessing any operation in the absence
// of default rights.
//
// See also AllowOp.
func (x BasicACL) IsOpAllowed(op ACLOp, role ACLRole) bool {
	var bitPos uint8

	switch role {
	default:
		panic(fmt.Sprintf("role is unsupported %v", role))
	case ACLRoleInnerRing:
		switch op {
		case
			ACLOpObjectGet,
			ACLOpObjectHead,
			ACLOpObjectHash,
			ACLOpObjectSearch:
			return true
		default:
			return false
		}
	case ACLRoleOwner:
		bitPos = opBitPosOwner
	case ACLRoleContainer:
		if isReplicationOp(op) {
			return true
		}

		bitPos = opBitPosContainer
	case ACLRoleOthers:
		bitPos = opBitPosOthers
	}

	return isOpBitSet(x.bits, op, bitPos)
}

// AllowBearerRules allows bearer to provide extended ACL rules for the given
// operation. Bearer rules doesn't depend on container ACL
// // extensibility.
//
// See also AllowedBearerRules.
func (x *BasicACL) AllowBearerRules(op ACLOp) {
	setOpBit(&x.bits, op, opBitPosBearer)
}

// AllowedBearerRules checks if bearer rules are allowed using AllowBearerRules.
// Op MUST be one of the ACLOp enumeration.
//
// Zero BasicACL disallows bearer rules for any op.
func (x BasicACL) AllowedBearerRules(op ACLOp) bool {
	return isOpBitSet(x.bits, op, opBitPosBearer)
}

// EncodeToString encodes BasicACL into hexadecimal string.
//
// See also DecodeString.
func (x BasicACL) EncodeToString() string {
	return strconv.FormatUint(uint64(x.bits), 16)
}

// Names of the frequently used BasicACL values.
const (
	BasicACLNamePrivate              = "private"
	BasicACLNamePrivateExtended      = "eacl-private"
	BasicACLNamePublicRO             = "public-read"
	BasicACLNamePublicROExtended     = "eacl-public-read"
	BasicACLNamePublicRW             = "public-read-write"
	BasicACLNamePublicRWExtended     = "eacl-public-read-write"
	BasicACLNamePublicAppend         = "public-append"
	BasicACLNamePublicAppendExtended = "eacl-public-append"
)

// Frequently used BasicACL values (each value MUST NOT be modified, make a
// copy instead).
var (
	BasicACLPrivate              BasicACL // private
	BasicACLPrivateExtended      BasicACL // eacl-private
	BasicACLPublicRO             BasicACL // public-read
	BasicACLPublicROExtended     BasicACL // eacl-public-read
	BasicACLPublicRW             BasicACL // public-read-write
	BasicACLPublicRWExtended     BasicACL // eacl-public-read-write
	BasicACLPublicAppend         BasicACL // public-append
	BasicACLPublicAppendExtended BasicACL // eacl-public-append
)

// DecodeString decodes string calculated using EncodeToString. Also supports
// human-readable names (BasicACLName* constants).
func (x *BasicACL) DecodeString(s string) error {
	switch s {
	case BasicACLNamePrivate:
		*x = BasicACLPrivate
	case BasicACLNamePrivateExtended:
		*x = BasicACLPrivateExtended
	case BasicACLNamePublicRO:
		*x = BasicACLPublicRO
	case BasicACLNamePublicROExtended:
		*x = BasicACLPublicROExtended
	case BasicACLNamePublicRW:
		*x = BasicACLPublicRW
	case BasicACLNamePublicRWExtended:
		*x = BasicACLPublicRWExtended
	case BasicACLNamePublicAppend:
		*x = BasicACLPublicAppend
	case BasicACLNamePublicAppendExtended:
		*x = BasicACLPublicAppendExtended
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
