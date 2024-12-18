package object_test

import (
	"crypto/sha256"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

const (
	anyValidExpirationEpoch = uint64(6053221788077248524)
	anyValidCreationEpoch   = uint64(13233261290750647837)
	anyValidPayloadSize     = uint64(5544264194415343420)
	anyValidType            = object.Type(2082391263)
)

var (
	anyValidSplitIDBytes = []byte{224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41}
	anyValidSplitID      = object.NewSplitIDFromV2(anyValidSplitIDBytes)
	anyValidIDs          = []oid.ID{
		{178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8,
			139, 247, 174, 53, 60},
		{229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243,
			57, 253, 127, 179, 235},
		{206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235,
			126, 173, 229, 161, 202, 197, 242},
		{173, 160, 45, 58, 200, 168, 116, 142, 235, 209, 231, 80, 235, 186, 6, 132, 99, 95, 14, 39, 237, 139, 87, 66, 244, 72, 96,
			69, 13, 83, 81, 172},
		{238, 167, 85, 68, 91, 254, 165, 81, 182, 145, 16, 91, 35, 224, 17, 46, 164, 138, 86, 50, 196, 148, 215, 210, 247, 29, 44,
			153, 203, 20, 137, 169},
		{226, 165, 123, 249, 146, 166, 187, 202, 244, 12, 156, 43, 207, 204, 40, 230, 145, 34, 212, 152, 148, 112, 44, 21, 195,
			207, 249, 112, 34, 81, 145, 194},
		{119, 231, 221, 167, 7, 141, 50, 77, 49, 23, 194, 169, 82, 56, 150, 162, 103, 20, 124, 174, 16, 64, 169, 172, 79, 238, 242,
			146, 87, 88, 5, 147},
		{139, 94, 241, 189, 238, 91, 251, 21, 52, 85, 200, 121, 189, 78, 16, 173, 74, 174, 20, 47, 223, 68, 172, 82, 113, 185,
			171, 241, 195, 191, 186, 87},
		{57, 171, 109, 41, 105, 25, 146, 224, 164, 89, 61, 178, 179, 158, 67, 40, 32, 154, 122, 174, 117, 221, 138, 168, 135, 149, 238,
			61, 68, 58, 34, 189},
		{110, 233, 102, 232, 136, 68, 233, 22, 158, 100, 49, 20, 181, 95, 219, 143, 53, 250, 237, 113, 64, 25, 48, 11, 54, 207,
			56, 98, 99, 136, 207, 21},
	}
	anyValidVersions = []version.Version{
		version.New(88789927, 2018985309),
		version.New(525747025, 171993162),
	}
	anyValidContainers = []cid.ID{
		{245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180, 255, 162, 182, 222,
			220, 99, 125, 136, 117, 206, 34},
		{217, 213, 19, 152, 91, 248, 2, 180, 17, 177, 248, 226, 163, 200, 56, 31, 123, 24, 182, 144, 148, 180, 248, 192, 155,
			253, 104, 220, 69, 102, 174, 5},
		{135, 89, 149, 219, 185, 209, 233, 137, 224, 211, 141, 70, 193, 205, 248, 254, 226, 30, 114, 177, 245, 171, 29, 90, 212,
			15, 51, 86, 142, 101, 155, 141},
	}
	anyValidUsers = []user.ID{
		{53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129, 211, 214, 90, 145, 237, 137, 153},
		{53, 214, 113, 220, 69, 70, 98, 242, 115, 99, 188, 86, 53, 223, 243, 238, 11, 245, 251, 169, 115, 202, 247, 184, 221},
		{53, 248, 195, 15, 196, 254, 124, 23, 169, 198, 208, 15, 219, 229, 62, 150, 151, 159, 221, 73, 224, 229, 106, 42, 222},
	}
	anySHA256Hash = [sha256.Size]byte{233, 204, 37, 189, 15, 146, 210, 138, 178, 74, 213, 141, 199, 249, 94, 20, 73, 133, 16,
		154, 241, 152, 3, 205, 101, 210, 153, 141, 139, 30, 216, 125}
	anyTillichZemorHash = [tz.Size]byte{160, 149, 6, 167, 41, 70, 29, 61, 190, 154, 30, 117, 180, 150, 91, 146, 24, 16, 195, 213, 216,
		106, 119, 203, 178, 159, 37, 1, 252, 208, 87, 23, 165, 19, 22, 96, 50, 28, 145, 235, 127, 107, 86, 216, 51, 226, 84, 242,
		94, 186, 90, 81, 184, 236, 118, 65, 58, 69, 110, 232, 22, 249, 131, 173}
	anyValidSignatures = []neofscrypto.Signature{
		neofscrypto.NewSignatureFromRawKey(1277002296, []byte("pub_1"), []byte("sig_1")),
		neofscrypto.NewSignatureFromRawKey(1242896683, []byte("pub_2"), []byte("sig_2")),
	}
	anyValidChecksums = []checksum.Checksum{
		checksum.New(1974315742, []byte("checksum_1")),
		checksum.New(1922538608, []byte("checksum_2")),
		checksum.New(126384577, []byte("checksum_3")),
		checksum.New(1001923429, []byte("checksum_4")),
	}
	anyValidRegularPayload = []byte("Hello, world!")
	// corresponds to anyValidRegularPayload.
	anyValidPayloadChecksum = [sha256.Size]byte{49, 95, 91, 219, 118, 208, 120, 196, 59, 138, 192, 6, 78, 74, 1, 100, 97, 43, 31, 206,
		119, 200, 105, 52, 91, 252, 148, 199, 88, 148, 237, 211}
	emptySHA256Hash = [sha256.Size]byte{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65,
		228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85}
)

func protoIDFromBytes(b []byte) *refs.ObjectID {
	return &refs.ObjectID{Value: b}
}

func protoUserIDFromBytes(b []byte) *refs.OwnerID {
	return &refs.OwnerID{Value: b}
}

func protoContainerIDFromBytes(b []byte) *refs.ContainerID {
	return &refs.ContainerID{Value: b}
}
