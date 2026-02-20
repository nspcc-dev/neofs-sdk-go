package refs

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Field numbers of [OwnerID] message.
const (
	_ = iota
	FieldOwnerIDValue
)

// MarshaledSize returns size of the OwnerID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *OwnerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldOwnerIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the OwnerID in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [OwnerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *OwnerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, FieldOwnerIDValue, x.Value)
	}
}

// Field numbers of [ContainerID] message.
const (
	_ = iota
	FieldContainerIDValue
)

// MarshaledSize returns size of the ContainerID in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *ContainerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldContainerIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the ContainerID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ContainerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ContainerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, FieldContainerIDValue, x.Value)
	}
}

// Field numbers of [ObjectID] message.
const (
	_ = iota
	FieldObjectIDValue
)

// MarshaledSize returns size of the ObjectID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *ObjectID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldObjectIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the ObjectID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ObjectID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, FieldObjectIDValue, x.Value)
	}
}

// Field numbers of [Address] message.
const (
	_ = iota
	FieldAddressContainerID
	FieldAddressObjectID
)

// MarshaledSize returns size of the Address in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Address) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldAddressContainerID, x.ContainerId) +
			proto.SizeEmbedded(FieldAddressObjectID, x.ObjectId)
	}
	return sz
}

// MarshalStable writes the Address in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Address.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Address) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldAddressContainerID, x.ContainerId)
		proto.MarshalToEmbedded(b[off:], FieldAddressObjectID, x.ObjectId)
	}
}

// Field numbers of [Version] message.
const (
	_ = iota
	FieldVersionMajor
	FieldVersionMinor
)

// MarshaledSize returns size of the Version in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Version) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldVersionMajor, x.Major) +
			proto.SizeVarint(FieldVersionMinor, x.Minor)
	}
	return sz
}

// MarshalStable writes the Version in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Version.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Version) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldVersionMajor, x.Major)
		proto.MarshalToVarint(b[off:], FieldVersionMinor, x.Minor)
	}
}

// Field numbers of [Signature] message.
const (
	_ = iota
	FieldSignatureKey
	FieldSignatureValue
	FieldSignatureScheme
)

// MarshaledSize returns size of the Signature in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Signature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldSignatureKey, x.Key) +
			proto.SizeBytes(FieldSignatureValue, x.Sign) +
			proto.SizeVarint(FieldSignatureScheme, int32(x.Scheme))
	}
	return sz
}

// MarshalStable writes the Signature in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Signature.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Signature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldSignatureKey, x.Key)
		off += proto.MarshalToBytes(b[off:], FieldSignatureValue, x.Sign)
		proto.MarshalToVarint(b[off:], FieldSignatureScheme, int32(x.Scheme))
	}
}

// Field numbers of [SignatureRFC6979] message.
const (
	_ = iota
	FieldSignatureRFC6979Key
	FieldSignatureRFC6979Value
)

// MarshaledSize returns size of the SignatureRFC6979 in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SignatureRFC6979) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldSignatureRFC6979Key, x.Key) +
			proto.SizeBytes(FieldSignatureRFC6979Value, x.Sign)
	}
	return sz
}

// MarshalStable writes the SignatureRFC6979 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SignatureRFC6979.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SignatureRFC6979) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldSignatureRFC6979Key, x.Key)
		proto.MarshalToBytes(b[off:], FieldSignatureRFC6979Value, x.Sign)
	}
}

// Field numbers of [SignatureRFC6979] message.
const (
	_ = iota
	FieldChecksumType
	FieldChecksumValue
)

// MarshaledSize returns size of the Checksum in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Checksum) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldChecksumType, int32(x.Type)) +
			proto.SizeBytes(FieldChecksumValue, x.Sum)
	}
	return sz
}

// MarshalStable writes the Checksum in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Checksum.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Checksum) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldChecksumType, int32(x.Type))
		proto.MarshalToBytes(b[off:], FieldChecksumValue, x.Sum)
	}
}

// Field numbers of [SubnetID] message.
const (
	_ = iota
	FieldSubnetIDValue
)

// MarshaledSize returns size of the SubnetID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SubnetID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeFixed32(FieldSubnetIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the SubnetID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToFixed32(b, FieldSubnetIDValue, x.Value)
	}
}
