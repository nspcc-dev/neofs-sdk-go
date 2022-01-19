package headers

import (
	"strconv"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl/validator"
	objectSDK "github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

type sysObjHdr struct {
	k, v string
}

// Key returns header key.
func (s *sysObjHdr) Key() string {
	return s.k
}

// Value returns header value.
func (s *sysObjHdr) Value() string {
	return s.v
}

func idValue(id *objectSDK.ID) string {
	return id.String()
}

func cidValue(id *cid.ID) string {
	return id.String()
}

func ownerIDValue(id *owner.ID) string {
	return id.String()
}

func u64Value(v uint64) string {
	return strconv.FormatUint(v, 10)
}

func headersFromObject(obj *objectSDK.Object, addr *objectSDK.Address) []validator.Header {
	// TODO: optimize allocs
	res := make([]validator.Header, 0)

	for ; obj != nil; obj = obj.Parent() {
		res = append(res,
			cidHeader(addr.ContainerID()),
			// owner ID
			&sysObjHdr{
				k: acl.FilterObjectOwnerID,
				v: ownerIDValue(obj.OwnerID()),
			},
			// creation epoch
			&sysObjHdr{
				k: acl.FilterObjectCreationEpoch,
				v: u64Value(obj.CreationEpoch()),
			},
			// payload size
			&sysObjHdr{
				k: acl.FilterObjectPayloadLength,
				v: u64Value(obj.PayloadSize()),
			},
			oidHeader(addr.ObjectID()),
			// object version
			&sysObjHdr{
				k: acl.FilterObjectVersion,
				v: obj.Version().String(),
			},
			// payload hash
			&sysObjHdr{
				k: acl.FilterObjectPayloadHash,
				v: obj.PayloadChecksum().String(),
			},
			// object type
			&sysObjHdr{
				k: acl.FilterObjectType,
				v: obj.Type().String(),
			},
			// payload homomorphic hash
			&sysObjHdr{
				k: acl.FilterObjectHomomorphicHash,
				v: obj.PayloadHomomorphicHash().String(),
			},
		)

		attrs := obj.Attributes()
		hs := make([]validator.Header, 0, len(attrs))

		for i := range attrs {
			hs = append(hs, attrs[i])
		}

		res = append(res, hs...)
	}

	return res
}
