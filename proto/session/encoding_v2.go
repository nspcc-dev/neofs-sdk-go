package session

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [TokenLifetime] message.
const (
	_ = iota
	FieldTokenLifetimeExp
	FieldTokenLifetimeNbf
	FieldTokenLifetimeIat
)

// MarshaledSize returns size of the TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldTokenLifetimeExp, x.Exp) +
			proto.SizeVarint(FieldTokenLifetimeNbf, x.Nbf) +
			proto.SizeVarint(FieldTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldTokenLifetimeExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], FieldTokenLifetimeNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], FieldTokenLifetimeIat, x.Iat)
	}
}

// Field numbers of [Target] message.
const (
	_ = iota
	FieldTargetOwnerID
	FieldTargetNNSName
)

// MarshaledSize returns size of the Target in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Target) MarshaledSize() int {
	var sz int
	if x != nil {
		switch id := x.Identifier.(type) {
		default:
			panic(fmt.Sprintf("unexpected identifier %T", x.Identifier))
		case nil:
		case *Target_OwnerId:
			sz = proto.SizeEmbedded(FieldTargetOwnerID, id.OwnerId)
		case *Target_NnsName:
			sz = proto.SizeBytes(FieldTargetNNSName, id.NnsName)
		}
	}
	return sz
}

// MarshalStable writes the Target in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Account.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Target) MarshalStable(b []byte) {
	if x != nil {
		switch id := x.Identifier.(type) {
		default:
			panic(fmt.Sprintf("unexpected identifier %T", x.Identifier))
		case nil:
		case *Target_OwnerId:
			proto.MarshalToEmbedded(b, FieldTargetOwnerID, id.OwnerId)
		case *Target_NnsName:
			proto.MarshalToBytes(b, FieldTargetNNSName, id.NnsName)
		}
	}
}

// Field numbers of [SessionContextV2] message.
const (
	_ = iota
	FieldSessionContextV2Container
	FieldSessionContextV2Verbs
)

// MarshaledSize returns size of the SessionContextV2 in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SessionContextV2) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldSessionContextV2Container, x.Container) +
			proto.SizeRepeatedVarint(FieldSessionContextV2Verbs, x.Verbs)
	}
	return 0
}

// MarshalStable writes the SessionContextV2 in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionContextV2.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionContextV2) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSessionContextV2Container, x.Container)
		proto.MarshalToRepeatedVarint(b[off:], FieldSessionContextV2Verbs, x.Verbs)
	}
}

// Field numbers of [SessionTokenV2_Body] message.
const (
	_ = iota
	FieldSessionTokenV2BodyVersion
	FieldSessionTokenV2BodyAppdata
	FieldSessionTokenV2BodyIssuer
	FieldSessionTokenV2BodySubjects
	FieldSessionTokenV2BodyLifetime
	FieldSessionTokenV2BodyContexts
	FieldSessionTokenV2BodyFinal
)

// MarshaledSize returns size of the SessionTokenV2_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldSessionTokenV2BodyVersion, x.Version) +
			proto.SizeBytes(FieldSessionTokenV2BodyAppdata, x.Appdata) +
			proto.SizeEmbedded(FieldSessionTokenV2BodyIssuer, x.Issuer) +
			proto.SizeRepeatedMessages(FieldSessionTokenV2BodySubjects, x.Subjects) +
			proto.SizeEmbedded(FieldSessionTokenV2BodyLifetime, x.Lifetime) +
			proto.SizeRepeatedMessages(FieldSessionTokenV2BodyContexts, x.Contexts) +
			proto.SizeBool(FieldSessionTokenV2BodyFinal, x.Final)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SessionTokenV2_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldSessionTokenV2BodyVersion, x.Version)
		off += proto.MarshalToBytes(b[off:], FieldSessionTokenV2BodyAppdata, x.Appdata)
		off += proto.MarshalToEmbedded(b[off:], FieldSessionTokenV2BodyIssuer, x.Issuer)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldSessionTokenV2BodySubjects, x.Subjects)
		off += proto.MarshalToEmbedded(b[off:], FieldSessionTokenV2BodyLifetime, x.Lifetime)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldSessionTokenV2BodyContexts, x.Contexts)
		proto.MarshalToBool(b[off:], FieldSessionTokenV2BodyFinal, x.Final)
	}
}

// Field numbers of [SessionTokenV2] message.
const (
	_ = iota
	FieldSessionTokenV2Body
	FieldSessionTokenV2Signature
	FieldSessionTokenV2Origin
)

// MarshaledSize returns size of the SessionTokenV2 in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldSessionTokenV2Body, x.Body) +
			proto.SizeEmbedded(FieldSessionTokenV2Signature, x.Signature) +
			proto.SizeEmbedded(FieldSessionTokenV2Origin, x.Origin)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionTokenV2) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSessionTokenV2Body, x.Body)
		off += proto.MarshalToEmbedded(b[off:], FieldSessionTokenV2Signature, x.Signature)
		proto.MarshalToEmbedded(b[off:], FieldSessionTokenV2Origin, x.Origin)
	}
}
