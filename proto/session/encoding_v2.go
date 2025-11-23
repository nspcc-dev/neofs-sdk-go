package session

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldDelegationIssuer
	fieldDelegationSubjects
	fieldDelegationLifetime
	fieldDelegationVerbs
	fieldDelegationSignature
)

// MarshaledSize returns size of the DelegationInfo in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DelegationInfo) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldDelegationIssuer, x.Issuer) +
			proto.SizeRepeatedMessages(fieldDelegationSubjects, x.Subjects) +
			proto.SizeEmbedded(fieldDelegationLifetime, x.Lifetime) +
			proto.SizeRepeatedVarint(fieldDelegationVerbs, x.Verbs) +
			proto.SizeEmbedded(fieldDelegationSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the DelegationInfo in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [DelegationInfo.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *DelegationInfo) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldDelegationIssuer, x.Issuer)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldDelegationSubjects, x.Subjects)
		off += proto.MarshalToEmbedded(b[off:], fieldDelegationLifetime, x.Lifetime)
		off += proto.MarshalToRepeatedVarint(b[off:], fieldDelegationVerbs, x.Verbs)
		proto.MarshalToEmbedded(b[off:], fieldDelegationSignature, x.Signature)
	}
}

const (
	_ = iota
	fieldAccountOwnerID
	fieldAccountNnsName
)

// MarshaledSize returns size of the Account in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Target) MarshaledSize() int {
	var sz int
	if x != nil {
		switch id := x.Identifier.(type) {
		default:
			panic(fmt.Sprintf("unexpected identifier %T", x.Identifier))
		case nil:
		case *Target_OwnerId:
			sz = proto.SizeEmbedded(fieldAccountOwnerID, id.OwnerId)
		case *Target_NnsName:
			sz = proto.SizeBytes(fieldAccountNnsName, id.NnsName)
		}
	}
	return sz
}

// MarshalStable writes the Account in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Account.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Target) MarshalStable(b []byte) {
	if x != nil {
		switch id := x.Identifier.(type) {
		default:
			panic(fmt.Sprintf("unexpected identifier %T", x.Identifier))
		case nil:
		case *Target_OwnerId:
			proto.MarshalToEmbedded(b, fieldAccountOwnerID, id.OwnerId)
		case *Target_NnsName:
			proto.MarshalToBytes(b, fieldAccountNnsName, id.NnsName)
		}
	}
}

const (
	_ = iota
	fieldSessionContextV2Container
	fieldSessionContextV2Objects
	fieldSessionContextV2Verbs
)

// MarshaledSize returns size of the SessionContextV2 in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SessionContextV2) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSessionContextV2Container, x.Container) +
			proto.SizeRepeatedMessages(fieldSessionContextV2Objects, x.Objects) +
			proto.SizeRepeatedVarint(fieldSessionContextV2Verbs, x.Verbs)
	}
	return 0
}

// MarshalStable writes the SessionContextV2 in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionContextV2.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionContextV2) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSessionContextV2Container, x.Container)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldSessionContextV2Objects, x.Objects)
		proto.MarshalToRepeatedVarint(b[off:], fieldSessionContextV2Verbs, x.Verbs)
	}
}

const (
	_ = iota
	fieldTokenV2Version
	fieldTokenV2ID
	fieldTokenV2Issuer
	fieldTokenV2Subjects
	fieldTokenV2Lifetime
	fieldTokenV2Contexts
)

// MarshaledSize returns size of the SessionTokenV2_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenV2Version, x.Version) +
			proto.SizeBytes(fieldTokenV2ID, x.Id) +
			proto.SizeEmbedded(fieldTokenV2Issuer, x.Issuer) +
			proto.SizeRepeatedMessages(fieldTokenV2Subjects, x.Subjects) +
			proto.SizeEmbedded(fieldTokenV2Lifetime, x.Lifetime) +
			proto.SizeRepeatedMessages(fieldTokenV2Contexts, x.Contexts)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SessionTokenV2_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTokenV2Version, x.Version)
		off += proto.MarshalToBytes(b[off:], fieldTokenV2ID, x.Id)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenV2Issuer, x.Issuer)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldTokenV2Subjects, x.Subjects)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenV2Lifetime, x.Lifetime)
		proto.MarshalToRepeatedMessages(b[off:], fieldTokenV2Contexts, x.Contexts)
	}
}

const (
	_ = iota
	fieldTokenV2Body
	fieldTokenV2Signature
	fieldTokenV2DelegationChain
)

// MarshaledSize returns size of the SessionTokenV2 in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldTokenV2Body, x.Body) +
			proto.SizeEmbedded(fieldTokenV2Signature, x.Signature) +
			proto.SizeRepeatedMessages(fieldTokenV2DelegationChain, x.DelegationChain)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionTokenV2) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldTokenV2Body, x.Body)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenV2Signature, x.Signature)
		proto.MarshalToRepeatedMessages(b[off:], fieldTokenV2DelegationChain, x.DelegationChain)
	}
}
