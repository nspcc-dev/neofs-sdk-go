package refs

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldOwnerIDValue
)

func (x *OwnerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldOwnerIDValue, x.Value)
	}
	return sz
}

func (x *OwnerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalBytes(b, fieldOwnerIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldVersionMajor
	fieldVersionMinor
)

func (x *Version) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldVersionMajor, x.Major) +
			proto.SizeVarint(fieldVersionMinor, x.Minor)
	}
	return sz
}

func (x *Version) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldVersionMajor, x.Major)
		proto.MarshalVarint(b[off:], fieldVersionMinor, x.Minor)
	}
}

const (
	_ = iota
	fieldContainerIDValue
)

func (x *ContainerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldContainerIDValue, x.Value)
	}
	return sz
}

func (x *ContainerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalBytes(b, fieldContainerIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldObjectIDValue
)

func (x *ObjectID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldObjectIDValue, x.Value)
	}
	return sz
}

func (x *ObjectID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalBytes(b, fieldObjectIDValue, x.Value)
	}
}

const (
	_ = iota
	fieldAddressContainer
	fieldAddressObject
)

func (x *Address) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldAddressContainer, x.ContainerId) +
			proto.SizeNested(fieldAddressObject, x.ObjectId)
	}
	return sz
}

func (x *Address) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldAddressContainer, x.ContainerId)
		proto.MarshalNested(b[off:], fieldAddressObject, x.ObjectId)
	}
}

const (
	_ = iota
	fieldSubnetVal
)

func (x *SubnetID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeFixed32(fieldSubnetVal, x.Value)
	}
	return sz
}

func (x *SubnetID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalFixed32(b, fieldSubnetVal, x.Value)
	}
}

const (
	_ = iota
	fieldSignatureKey
	fieldSignatureVal
	fieldSignatureScheme
)

func (x *Signature) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSignatureKey, x.Key) +
			proto.SizeBytes(fieldSignatureVal, x.Sign) +
			proto.SizeVarint(fieldSignatureScheme, int32(x.Scheme))
	}
	return sz
}

func (x *Signature) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldSignatureKey, x.Key)
		off += proto.MarshalBytes(b[off:], fieldSignatureVal, x.Sign)
		proto.MarshalVarint(b[off:], fieldSignatureScheme, int32(x.Scheme))
	}
}

const (
	_ = iota
	fieldSigRFC6979Key
	fieldSigRFC6979Val
)

func (x *SignatureRFC6979) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldSigRFC6979Key, x.Key) +
			proto.SizeBytes(fieldSigRFC6979Val, x.Sign)
	}
	return sz
}

func (x *SignatureRFC6979) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldSigRFC6979Key, x.Key)
		proto.MarshalBytes(b[off:], fieldSigRFC6979Val, x.Sign)
	}
}

const (
	_ = iota
	fieldChecksumType
	fieldChecksumValue
)

func (x *Checksum) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldChecksumType, int32(x.Type)) +
			proto.SizeBytes(fieldChecksumValue, x.Sum)
	}
	return sz
}

func (x *Checksum) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldChecksumType, int32(x.Type))
		proto.MarshalBytes(b[off:], fieldChecksumValue, x.Sum)
	}
}
