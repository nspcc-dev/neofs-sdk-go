package refs

import (
	"crypto/sha256"

	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
	"google.golang.org/protobuf/encoding/protowire"
)

// Fixed-length values.
const (
	ContainerIDLength   = 1 + 1 + sha256.Size
	ObjectDLength       = 1 + 1 + sha256.Size
	ObjectAddressLength = 1 + 1 + ContainerIDLength + 1 + 1 + ObjectDLength
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
		sz = protoencoding.SizeBytes(FieldOwnerIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the OwnerID in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [OwnerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *OwnerID) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToBytes(b, FieldOwnerIDValue, x.Value)
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
		sz = protoencoding.SizeBytes(FieldContainerIDValue, x.Value)
	}
	return sz
}

// WriteContainerIDField writes container ID field with given number and fields
// into buf. Returns number of bytes written.
func WriteContainerIDField(buf []byte, num protowire.Number, value [sha256.Size]byte) int {
	off := protoencoding.WriteTagAndLength(buf, num, ContainerIDLength)
	return off + WriteContainerID(buf[off:], value)
}

// WriteContainerID writes container ID message with given fields into buf.
// Returns number of bytes written.
func WriteContainerID(buf []byte, value [sha256.Size]byte) int {
	off := protoencoding.WriteTagAndLength(buf, FieldContainerIDValue, sha256.Size)
	return off + copy(buf[off:], value[:])
}

// MarshalStable writes the ContainerID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ContainerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ContainerID) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToBytes(b, FieldContainerIDValue, x.Value)
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
		sz = protoencoding.SizeBytes(FieldObjectIDValue, x.Value)
	}
	return sz
}

// WriteObjectIDField writes object ID field with given number and fields into
// buf. Returns number of bytes written.
func WriteObjectIDField(buf []byte, num protowire.Number, value [sha256.Size]byte) int {
	off := protoencoding.WriteTagAndLength(buf, num, ObjectDLength)
	return off + WriteObjectID(buf[off:], value)
}

// WriteObjectID writes object ID message with given fields into buf. Returns
// number of bytes written.
func WriteObjectID(buf []byte, value [sha256.Size]byte) int {
	off := protoencoding.WriteTagAndLength(buf, FieldObjectIDValue, sha256.Size)
	return off + copy(buf[off:], value[:])
}

// MarshalStable writes the ObjectID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ObjectID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ObjectID) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToBytes(b, FieldObjectIDValue, x.Value)
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
		sz = protoencoding.SizeEmbedded(FieldAddressContainerID, x.ContainerId) +
			protoencoding.SizeEmbedded(FieldAddressObjectID, x.ObjectId)
	}
	return sz
}

// WriteObjectAddressField writes object address field with given number and
// fields into buf. Returns number of bytes written.
func WriteObjectAddressField(buf []byte, num protowire.Number, cnr [sha256.Size]byte, obj [sha256.Size]byte) int {
	off := protoencoding.WriteTagAndLength(buf, num, ObjectAddressLength)
	off += WriteContainerIDField(buf[off:], FieldAddressContainerID, cnr)
	off += WriteObjectIDField(buf[off:], FieldAddressObjectID, obj)
	return off
}

// MarshalStable writes the Address in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Address.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Address) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldAddressContainerID, x.ContainerId)
		protoencoding.MarshalToEmbedded(b[off:], FieldAddressObjectID, x.ObjectId)
	}
}

// Field numbers of [Version] message.
const (
	_ = iota
	FieldVersionMajor
	FieldVersionMinor
)

// CalculateVersionLength calculates length of API version message with given
// fields.
func CalculateVersionLength(major uint32, minor uint32) int {
	ln := protoencoding.SizeVarint(FieldVersionMajor, major)
	ln += protoencoding.SizeVarint(FieldVersionMinor, minor)
	return ln
}

// MarshaledSize returns size of the Version in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Version) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateVersionLength(x.Major, x.Minor)
}

// WriteVersionField writes API version field with given number and fields into
// buf. Returns number of bytes written.
func WriteVersionField(buf []byte, num protowire.Number, major uint32, minor uint32) int {
	ln := CalculateVersionLength(major, minor)
	if ln == 0 {
		return 0
	}
	off := protoencoding.WriteTagAndLength(buf, num, ln)
	return off + WriteVersion(buf[off:], major, minor)
}

// WriteVersion writes API version message with given fields into buf. Returns
// number of bytes written.
func WriteVersion(buf []byte, major uint32, minor uint32) int {
	off := protoencoding.MarshalToVarint(buf, FieldVersionMajor, major)
	off += protoencoding.MarshalToVarint(buf[off:], FieldVersionMinor, minor)
	return off
}

// MarshalStable writes the Version in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Version.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Version) MarshalStable(b []byte) {
	if x != nil {
		WriteVersion(b, x.Major, x.Minor)
	}
}

// Field numbers of [Signature] message.
const (
	_ = iota
	FieldSignatureKey
	FieldSignatureValue
	FieldSignatureScheme
)

// CalculateSignatureLength calculates length of signature message with given
// fields.
func CalculateSignatureLength[SCHEME ~int32](pubKey []byte, value []byte, scheme SCHEME) int {
	ln := protoencoding.SizeBytes(FieldSignatureKey, pubKey)
	ln += protoencoding.SizeBytes(FieldSignatureValue, value)
	ln += protoencoding.SizeVarint(FieldSignatureScheme, scheme)
	return ln
}

// MarshaledSize returns size of the Signature in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Signature) MarshaledSize() int {
	if x == nil {
		return 0
	}
	return CalculateSignatureLength(x.Key, x.Sign, x.Scheme)
}

// WriteSignatureField writes signature field with given number and fields into
// buf. Returns number of bytes written.
func WriteSignatureField[SCHEME ~int32](buf []byte, num protowire.Number, pubKey []byte, value []byte, scheme SCHEME) int {
	ln := CalculateSignatureLength(pubKey, value, scheme)
	if ln == 0 {
		return 0
	}
	off := protoencoding.WriteTagAndLength(buf, num, ln)
	return off + WriteSignature(buf[off:], pubKey, value, scheme)
}

// WriteSignature writes signature message with given fields into buf. Returns
// number of bytes written.
func WriteSignature[SCHEME ~int32](buf []byte, pubKey []byte, value []byte, scheme SCHEME) int {
	off := protoencoding.MarshalToBytes(buf, FieldSignatureKey, pubKey)
	off += protoencoding.MarshalToBytes(buf[off:], FieldSignatureValue, value)
	off += protoencoding.MarshalToVarint(buf[off:], FieldSignatureScheme, scheme)
	return off
}

// MarshalStable writes the Signature in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Signature.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Signature) MarshalStable(b []byte) {
	if x != nil {
		WriteSignature(b, x.Key, x.Sign, x.Scheme)
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
		sz = protoencoding.SizeBytes(FieldSignatureRFC6979Key, x.Key) +
			protoencoding.SizeBytes(FieldSignatureRFC6979Value, x.Sign)
	}
	return sz
}

// MarshalStable writes the SignatureRFC6979 in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SignatureRFC6979.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SignatureRFC6979) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldSignatureRFC6979Key, x.Key)
		protoencoding.MarshalToBytes(b[off:], FieldSignatureRFC6979Value, x.Sign)
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
		sz = protoencoding.SizeVarint(FieldChecksumType, x.Type) +
			protoencoding.SizeBytes(FieldChecksumValue, x.Sum)
	}
	return sz
}

// MarshalStable writes the Checksum in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Checksum.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Checksum) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToVarint(b, FieldChecksumType, x.Type)
		protoencoding.MarshalToBytes(b[off:], FieldChecksumValue, x.Sum)
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
		sz = protoencoding.SizeFixed32(FieldSubnetIDValue, x.Value)
	}
	return sz
}

// MarshalStable writes the SubnetID in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [SubnetID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *SubnetID) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToFixed32(b, FieldSubnetIDValue, x.Value)
	}
}
