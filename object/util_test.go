package object_test

import (
	"crypto/sha256"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
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
	}
	anyValidVersion   = version.New(88789927, 2018985309)
	anyValidContainer = cid.ID{245, 94, 164, 207, 217, 233, 175, 75, 123, 153, 174, 8, 20, 135, 96, 204, 179, 93, 183, 250, 180,
		255, 162, 182, 222, 220, 99, 125, 136, 117, 206, 34}
	anyValidUser = user.ID{53, 59, 15, 5, 52, 131, 255, 198, 8, 98, 41, 184, 229, 237, 140, 215, 52, 129, 211, 214, 90, 145, 237,
		137, 153}
	anySHA256Hash = [sha256.Size]byte{233, 204, 37, 189, 15, 146, 210, 138, 178, 74, 213, 141, 199, 249, 94, 20, 73, 133, 16,
		154, 241, 152, 3, 205, 101, 210, 153, 141, 139, 30, 216, 125}
	anyTillichZemorHash = [tz.Size]byte{160, 149, 6, 167, 41, 70, 29, 61, 190, 154, 30, 117, 180, 150, 91, 146, 24, 16, 195, 213, 216,
		106, 119, 203, 178, 159, 37, 1, 252, 208, 87, 23, 165, 19, 22, 96, 50, 28, 145, 235, 127, 107, 86, 216, 51, 226, 84, 242,
		94, 186, 90, 81, 184, 236, 118, 65, 58, 69, 110, 232, 22, 249, 131, 173}
)

func protoIDFromBytes(b []byte) *refs.ObjectID {
	var id refs.ObjectID
	id.SetValue(b)
	return &id
}
