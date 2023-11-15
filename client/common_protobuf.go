package client

import (
	"encoding/binary"

	"google.golang.org/protobuf/encoding/protowire"
)

// NeoFS API fields numbers.
const (
	fieldNumRequestBody   = 1
	fieldNumRequestMeta   = 2
	fieldNumRequestVerify = 3

	fieldNumVersionMajor = 1
	fieldNumVersionMinor = 2

	fieldNumRequestMetaVersion = 1
	fieldNumRequestMetaTTL     = 3

	fieldNumVerifyHdrBodySig   = 1
	fieldNumVerifyHdrMetaSig   = 2
	fieldNumVerifyHdrOriginSig = 3

	fieldNumSigPubKey = 1
	fieldNumSigVal    = 2
	fieldNumSigScheme = 3
)

// writes tag for the field of the given type having specified number in some
// message into the given buffer. The buffer must have sufficient size according
// to Protocol Buffers V3 format. Returns number of bytes used.
func writeProtoTag(b []byte, num protowire.Number, typ protowire.Type) int {
	return binary.PutUvarint(b, protowire.EncodeTag(num, typ))
}
