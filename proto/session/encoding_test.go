package session_test

import (
	"testing"

	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/session"
)

// returns random session.XHeader with all non-zero fields.
func randXHeader() *session.XHeader {
	return &session.XHeader{
		Key: prototest.RandString(), Value: prototest.RandString(),
	}
}

// returns non-empty list of session.XHeader up to 10 elements. Each element may
// be nil and pointer to zero.
func randXHeaders() []*session.XHeader { return prototest.RandRepeated(randXHeader) }

func _randRequestMetaHeader(withOrigin bool) *session.RequestMetaHeader {
	v := &session.RequestMetaHeader{
		Version:      prototest.RandVersion(),
		Epoch:        prototest.RandUint64(),
		Ttl:          prototest.RandUint32(),
		XHeaders:     randXHeaders(),
		SessionToken: prototest.RandSessionToken(),
		BearerToken:  prototest.RandBearerToken(),
		MagicNumber:  prototest.RandUint64(),
	}
	if withOrigin {
		v.Origin = _randRequestMetaHeader(false)
	}
	return v
}

func randRequestMetaHeader() *session.RequestMetaHeader {
	return _randRequestMetaHeader(true)
}

func _randResponseMetaHeader(withOrigin bool) *session.ResponseMetaHeader {
	v := &session.ResponseMetaHeader{
		Version:  prototest.RandVersion(),
		Epoch:    prototest.RandUint64(),
		Ttl:      prototest.RandUint32(),
		XHeaders: randXHeaders(),
		Status:   prototest.RandStatus(),
	}
	if withOrigin {
		v.Origin = _randResponseMetaHeader(false)
	}
	return v
}

func randResponseMetaHeader() *session.ResponseMetaHeader {
	return _randResponseMetaHeader(true)
}

func _randRequestVerificationHeader(withOrigin bool) *session.RequestVerificationHeader {
	v := &session.RequestVerificationHeader{
		BodySignature:   prototest.RandSignature(),
		MetaSignature:   prototest.RandSignature(),
		OriginSignature: prototest.RandSignature(),
	}
	if withOrigin {
		v.Origin = _randRequestVerificationHeader(false)
	}
	return v
}

func randRequestVerificationHeader() *session.RequestVerificationHeader {
	return _randRequestVerificationHeader(true)
}

func _randResponseVerificationHeader(withOrigin bool) *session.ResponseVerificationHeader {
	v := &session.ResponseVerificationHeader{
		BodySignature:   prototest.RandSignature(),
		MetaSignature:   prototest.RandSignature(),
		OriginSignature: prototest.RandSignature(),
	}
	if withOrigin {
		v.Origin = _randResponseVerificationHeader(false)
	}
	return v
}

func randResponseVerificationHeader() *session.ResponseVerificationHeader {
	return _randResponseVerificationHeader(true)
}

func TestObjectSessionContext_Target_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.ObjectSessionContext_Target{
		prototest.RandObjectSessionTarget(),
	})
}

func TestObjectSessionContext_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.ObjectSessionContext{
		prototest.RandObjectSessionContext(),
	})
}

func TestContainerSessionContext_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.ContainerSessionContext{
		prototest.RandContainerSessionContext(),
	})
}

func TestSessionToken_Body_TokenLifetime_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.SessionToken_Body_TokenLifetime{
		prototest.RandSessionTokenLifetime(),
	})
}

func TestSessionToken_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.SessionToken_Body{
		prototest.RandSessionTokenBody(),
	})
}

func TestSessionToken_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.SessionToken{
		prototest.RandSessionToken(),
	})
}

func TestCreateRequest_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.CreateRequest_Body{
		{
			OwnerId:    prototest.RandOwnerID(),
			Expiration: prototest.RandUint64(),
		},
	})
}

func TestCreateResponse_Body_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.CreateResponse_Body{
		{
			Id:         prototest.RandBytes(),
			SessionKey: prototest.RandBytes(),
		},
	})
}

func TestXHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.XHeader{
		{Key: prototest.RandString()},
		{Value: prototest.RandString()},
		randXHeader(),
	})
}

func TestRequestMetaHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.RequestMetaHeader{
		randRequestMetaHeader(),
	})
}

func TestResponseMetaHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.ResponseMetaHeader{
		randResponseMetaHeader(),
	})
}

func TestRequestVerificationHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.RequestVerificationHeader{
		randRequestVerificationHeader(),
	})
}

func TestResponseVerificationHeader_MarshalStable(t *testing.T) {
	prototest.TestMarshalStable(t, []*session.ResponseVerificationHeader{
		randResponseVerificationHeader(),
	})
}
