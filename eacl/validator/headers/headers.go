package headers

import (
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	objectV2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	eaclSDK "github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/eacl/validator"
	objectSDK "github.com/nspcc-dev/neofs-sdk-go/object"
	sessionSDK "github.com/nspcc-dev/neofs-sdk-go/session"
)

// Option is a HeaderSource configuration change function.
type Option func(*cfg)

type cfg struct {
	storage ObjectStorage

	msg xHeaderSource

	addr *objectSDK.Address
}

// ObjectStorage is an interface that maps object address
// to object.
type ObjectStorage interface {
	Head(*objectSDK.Address) (*objectSDK.Object, error)
}

// Request is v2 compatible request interface.
type Request interface {
	GetMetaHeader() *session.RequestMetaHeader
}

// Response is v2 compatible response interface.
type Response interface {
	GetMetaHeader() *session.ResponseMetaHeader
}

type headerSource struct {
	*cfg
}

func defaultCfg() *cfg {
	return new(cfg)
}

// NewMessageHeaderSource creates, initializes and returns
// the implementation of the validator.TypedHeaderSource
// interface.
func NewMessageHeaderSource(opts ...Option) validator.TypedHeaderSource {
	cfg := defaultCfg()

	for i := range opts {
		opts[i](cfg)
	}

	return &headerSource{
		cfg: cfg,
	}
}

func (h *headerSource) HeadersOfType(typ eaclSDK.FilterHeaderType) ([]validator.Header, bool) {
	switch typ {
	default:
		return nil, true
	case eaclSDK.HeaderFromRequest:
		return requestHeaders(h.msg), true
	case eaclSDK.HeaderFromObject:
		return h.objectHeaders()
	}
}

func requestHeaders(msg xHeaderSource) []validator.Header {
	xHdrs := msg.GetXHeaders()

	res := make([]validator.Header, 0, len(xHdrs))

	for i := range xHdrs {
		res = append(res, sessionSDK.NewXHeaderFromV2(xHdrs[i]))
	}

	return res
}

func (h *headerSource) objectHeaders() ([]validator.Header, bool) {
	switch m := h.msg.(type) {
	default:
		panic(fmt.Sprintf("unexpected message type %T", h.msg))
	case *requestXHeaderSource:
		switch req := m.req.(type) {
		case *objectV2.GetRequest:
			return h.localObjectHeaders(h.addr)
		case *objectV2.HeadRequest:
			return h.localObjectHeaders(h.addr)
		case
			*objectV2.GetRangeRequest,
			*objectV2.GetRangeHashRequest,
			*objectV2.DeleteRequest:
			return addressHeaders(h.addr), true
		case *objectV2.PutRequest:
			if v, ok := req.GetBody().GetObjectPart().(*objectV2.PutObjectPartInit); ok {
				oV2 := new(objectV2.Object)
				oV2.SetObjectID(v.GetObjectID())
				oV2.SetHeader(v.GetHeader())

				if h.addr == nil {
					h.addr = objectSDK.NewAddress()
					h.addr.SetContainerID(cid.NewFromV2(v.GetHeader().GetContainerID()))
					h.addr.SetObjectID(objectSDK.NewIDFromV2(v.GetObjectID()))
				}

				hs := headersFromObject(objectSDK.NewFromV2(oV2), h.addr)

				return hs, true
			}
		case *objectV2.SearchRequest:
			return []validator.Header{cidHeader(
				cid.NewFromV2(
					req.GetBody().GetContainerID()),
			)}, true
		}
	case *responseXHeaderSource:
		switch resp := m.resp.(type) {
		default:
			hs, _ := h.localObjectHeaders(h.addr)
			return hs, true
		case *objectV2.GetResponse:
			if v, ok := resp.GetBody().GetObjectPart().(*objectV2.GetObjectPartInit); ok {
				oV2 := new(objectV2.Object)
				oV2.SetObjectID(v.GetObjectID())
				oV2.SetHeader(v.GetHeader())

				return headersFromObject(objectSDK.NewFromV2(oV2), h.addr), true
			}
		case *objectV2.HeadResponse:
			oV2 := new(objectV2.Object)

			var hdr *objectV2.Header

			switch v := resp.GetBody().GetHeaderPart().(type) {
			case *objectV2.ShortHeader:
				hdr = new(objectV2.Header)

				hdr.SetContainerID(h.addr.ContainerID().ToV2())
				hdr.SetVersion(v.GetVersion())
				hdr.SetCreationEpoch(v.GetCreationEpoch())
				hdr.SetOwnerID(v.GetOwnerID())
				hdr.SetObjectType(v.GetObjectType())
				hdr.SetPayloadLength(v.GetPayloadLength())
			case *objectV2.HeaderWithSignature:
				hdr = v.GetHeader()
			}

			oV2.SetHeader(hdr)

			return headersFromObject(objectSDK.NewFromV2(oV2), h.addr), true
		}
	}

	return nil, true
}

func (h *headerSource) localObjectHeaders(addr *objectSDK.Address) ([]validator.Header, bool) {
	obj, err := h.storage.Head(addr)
	if err == nil {
		return headersFromObject(obj, addr), true
	}

	return addressHeaders(addr), false
}

func cidHeader(cid *cid.ID) validator.Header {
	return &sysObjHdr{
		k: acl.FilterObjectContainerID,
		v: cidValue(cid),
	}
}

func oidHeader(oid *objectSDK.ID) validator.Header {
	return &sysObjHdr{
		k: acl.FilterObjectID,
		v: idValue(oid),
	}
}

func addressHeaders(addr *objectSDK.Address) []validator.Header {
	res := make([]validator.Header, 1, 2)
	res[0] = cidHeader(addr.ContainerID())

	if oid := addr.ObjectID(); oid != nil {
		res = append(res, oidHeader(oid))
	}

	return res
}
