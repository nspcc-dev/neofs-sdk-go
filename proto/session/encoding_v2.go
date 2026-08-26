package session

import (
	"fmt"

	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
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
		sz = protoencoding.SizeVarint(FieldTokenLifetimeExp, x.Exp) +
			protoencoding.SizeVarint(FieldTokenLifetimeNbf, x.Nbf) +
			protoencoding.SizeVarint(FieldTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldTokenLifetimeExp, x.Exp)
		off += protoencoding.MarshalToVarint(b[off:], FieldTokenLifetimeNbf, x.Nbf)
		protoencoding.MarshalToVarint(b[off:], FieldTokenLifetimeIat, x.Iat)
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
			sz = protoencoding.SizeEmbedded(FieldTargetOwnerID, id.OwnerId)
		case *Target_NnsName:
			sz = protoencoding.SizeBytes(FieldTargetNNSName, id.NnsName)
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
			protoencoding.MarshalToEmbedded(b, FieldTargetOwnerID, id.OwnerId)
		case *Target_NnsName:
			protoencoding.MarshalToBytes(b, FieldTargetNNSName, id.NnsName)
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
		return protoencoding.SizeEmbedded(FieldSessionContextV2Container, x.Container) +
			protoencoding.SizeRepeatedVarint(FieldSessionContextV2Verbs, x.Verbs)
	}
	return 0
}

// MarshalStable writes the SessionContextV2 in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SessionContextV2.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SessionContextV2) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSessionContextV2Container, x.Container)
		protoencoding.MarshalToRepeatedVarint(b[off:], FieldSessionContextV2Verbs, x.Verbs)
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
		sz = protoencoding.SizeVarint(FieldSessionTokenV2BodyVersion, x.Version) +
			protoencoding.SizeBytes(FieldSessionTokenV2BodyAppdata, x.Appdata) +
			protoencoding.SizeEmbedded(FieldSessionTokenV2BodyIssuer, x.Issuer) +
			protoencoding.SizeRepeatedMessages(FieldSessionTokenV2BodySubjects, x.Subjects) +
			protoencoding.SizeEmbedded(FieldSessionTokenV2BodyLifetime, x.Lifetime) +
			protoencoding.SizeRepeatedMessages(FieldSessionTokenV2BodyContexts, x.Contexts) +
			protoencoding.SizeBool(FieldSessionTokenV2BodyFinal, x.Final)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SessionTokenV2_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldSessionTokenV2BodyVersion, x.Version)
		off += protoencoding.MarshalToBytes(b[off:], FieldSessionTokenV2BodyAppdata, x.Appdata)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenV2BodyIssuer, x.Issuer)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldSessionTokenV2BodySubjects, x.Subjects)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenV2BodyLifetime, x.Lifetime)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldSessionTokenV2BodyContexts, x.Contexts)
		protoencoding.MarshalToBool(b[off:], FieldSessionTokenV2BodyFinal, x.Final)
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
		sz = protoencoding.SizeEmbedded(FieldSessionTokenV2Body, x.Body) +
			protoencoding.SizeEmbedded(FieldSessionTokenV2Signature, x.Signature) +
			protoencoding.SizeEmbedded(FieldSessionTokenV2Origin, x.Origin)
	}
	return sz
}

// MarshalStable writes the SessionTokenV2 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SessionTokenV2.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SessionTokenV2) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSessionTokenV2Body, x.Body)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenV2Signature, x.Signature)
		protoencoding.MarshalToEmbedded(b[off:], FieldSessionTokenV2Origin, x.Origin)
	}
}
