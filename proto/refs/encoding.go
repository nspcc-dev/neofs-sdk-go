package refs

import (
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

const (
	_ = iota
	fieldOwnerIDValue
)

// MarshaledSize returns size of the OwnerID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *OwnerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldOwnerIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the OwnerID in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [OwnerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *OwnerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, fieldOwnerIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldContainerIDValue
)

// MarshaledSize returns size of the ContainerID in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *ContainerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldContainerIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the ContainerID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ContainerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ContainerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, fieldContainerIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldObjectIDValue
)

// MarshaledSize returns size of the ObjectID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *ObjectID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldObjectIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the ObjectID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ObjectID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, fieldObjectIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldAddressContainer
	fieldAddressObject
)

// MarshaledSize returns size of the Address in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Address) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldAddressContainer, x.ContainerId) +
			proto.SizeEmbedded(fieldAddressObject, x.ObjectId)
	}
	return sz
}

// MarshalStable writes the Address in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Address.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Address) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldAddressContainer, x.ContainerId)
		proto.MarshalToEmbedded(b[off:], fieldAddressObject, x.ObjectId)
	}
}

const (
	_ = iota
	fieldVersionMajor
	fieldVersionMinor
)

// MarshaledSize returns size of the Version in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Version) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldVersionMajor, x.Major) +
			proto.SizeVarint(fieldVersionMinor, x.Minor)
	}
	return sz
}

// MarshalStable writes the Version in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Version.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Version) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldVersionMajor, x.Major)
		proto.MarshalToVarint(b[off:], fieldVersionMinor, x.Minor)
	}
}

const (
	_ = iota
	fieldSignatureKey
	fieldSignatureVal
	fieldSignatureScheme
)

// MarshaledSize returns size of the Signature in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Signature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSignatureKey, x.Key) +
			proto.SizeBytes(fieldSignatureVal, x.Sign) +
			proto.SizeVarint(fieldSignatureScheme, int32(x.Scheme))
	}
	return sz
}

// MarshalStable writes the Signature in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Signature.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Signature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldSignatureKey, x.Key)
		off += proto.MarshalToBytes(b[off:], fieldSignatureVal, x.Sign)
		proto.MarshalToVarint(b[off:], fieldSignatureScheme, int32(x.Scheme))
	}
}

const (
	_ = iota
	fieldSigRFC6979Key
	fieldSigRFC6979Val
)

// MarshaledSize returns size of the SignatureRFC6979 in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *SignatureRFC6979) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSigRFC6979Key, x.Key) +
			proto.SizeBytes(fieldSigRFC6979Val, x.Sign)
	}
	return sz
}

// MarshalStable writes the SignatureRFC6979 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SignatureRFC6979.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SignatureRFC6979) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldSigRFC6979Key, x.Key)
		proto.MarshalToBytes(b[off:], fieldSigRFC6979Val, x.Sign)
	}
}

const (
	_ = iota
	fieldChecksumType
	fieldChecksumValue
)

// MarshaledSize returns size of the Checksum in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Checksum) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldChecksumType, int32(x.Type)) +
			proto.SizeBytes(fieldChecksumValue, x.Sum)
	}
	return sz
}

// MarshalStable writes the Checksum in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Checksum.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Checksum) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldChecksumType, int32(x.Type))
		proto.MarshalToBytes(b[off:], fieldChecksumValue, x.Sum)
	}
}

const (
	_ = iota
	fieldSubnetVal
)

// MarshaledSize returns size of the SubnetID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *SubnetID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeFixed32(fieldSubnetVal, x.Value)
	}
	return sz
}

// MarshalStable writes the SubnetID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToFixed32(b, fieldSubnetVal, x.Value)
	}
}
