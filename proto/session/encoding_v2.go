package session

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldTokenLifetimeExp
	fieldTokenLifetimeNbf
	fieldTokenLifetimeIat
)

// MarshaledSize returns size of the TokenLifetime in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *TokenLifetime) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenLifetimeExp, x.Exp) +
			proto.SizeVarint(fieldTokenLifetimeNbf, x.Nbf) +
			proto.SizeVarint(fieldTokenLifetimeIat, x.Iat)
	}
	return sz
}

// MarshalStable writes the TokenLifetime in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [TokenLifetime.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *TokenLifetime) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldTokenLifetimeExp, x.Exp)
		off += proto.MarshalToVarint(b[off:], fieldTokenLifetimeNbf, x.Nbf)
		proto.MarshalToVarint(b[off:], fieldTokenLifetimeIat, x.Iat)
	}
}

const (
	_ = iota
	fieldAccountOwnerID
	fieldAccountNnsName
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
			sz = proto.SizeEmbedded(fieldAccountOwnerID, id.OwnerId)
		case *Target_NnsName:
			sz = proto.SizeBytes(fieldAccountNnsName, id.NnsName)
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
			proto.MarshalToEmbedded(b, fieldAccountOwnerID, id.OwnerId)
		case *Target_NnsName:
			proto.MarshalToBytes(b, fieldAccountNnsName, id.NnsName)
		}
	}
}

const (
	_ = iota
	fieldSessionContextV2Container
	fieldSessionContextV2Verbs
)

// MarshaledSize returns size of the SessionContextV2 in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SessionContextV2) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldSessionContextV2Container, x.Container) +
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
		proto.MarshalToRepeatedVarint(b[off:], fieldSessionContextV2Verbs, x.Verbs)
	}
}

const (
	_ = iota
	fieldTokenV2Version
	fieldTokenV2Appdata
	fieldTokenV2Issuer
	fieldTokenV2Subjects
	fieldTokenV2Lifetime
	fieldTokenV2Contexts
	fieldTokenV2Final
)

// MarshaledSize returns size of the SessionTokenV2_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldTokenV2Version, x.Version) +
			proto.SizeBytes(fieldTokenV2Appdata, x.Appdata) +
			proto.SizeEmbedded(fieldTokenV2Issuer, x.Issuer) +
			proto.SizeRepeatedMessages(fieldTokenV2Subjects, x.Subjects) +
			proto.SizeEmbedded(fieldTokenV2Lifetime, x.Lifetime) +
			proto.SizeRepeatedMessages(fieldTokenV2Contexts, x.Contexts) +
			proto.SizeBool(fieldTokenV2Final, x.Final)
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
		off += proto.MarshalToBytes(b[off:], fieldTokenV2Appdata, x.Appdata)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenV2Issuer, x.Issuer)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldTokenV2Subjects, x.Subjects)
		off += proto.MarshalToEmbedded(b[off:], fieldTokenV2Lifetime, x.Lifetime)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldTokenV2Contexts, x.Contexts)
		proto.MarshalToBool(b[off:], fieldTokenV2Final, x.Final)
	}
}

const (
	_ = iota
	fieldTokenV2Body
	fieldTokenV2Signature
	fieldTokenV2Origin
)

// MarshaledSize returns size of the SessionTokenV2 in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SessionTokenV2) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldTokenV2Body, x.Body) +
			proto.SizeEmbedded(fieldTokenV2Signature, x.Signature) +
			proto.SizeEmbedded(fieldTokenV2Origin, x.Origin)
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
		proto.MarshalToEmbedded(b[off:], fieldTokenV2Origin, x.Origin)
	}
}
