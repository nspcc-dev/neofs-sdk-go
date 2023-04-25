package client

import (
	"crypto/rand"
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/accounting"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

type testResponse interface {
	SetMetaHeader(*session.ResponseMetaHeader)
	GetMetaHeader() *session.ResponseMetaHeader
}

func testOwner(t *testing.T, owner *refs.OwnerID, req any) {
	originalValue := owner.GetValue()
	owner.SetValue([]byte{1, 2, 3})
	// verification must fail
	require.Error(t, verifyServiceMessage(req))
	owner.SetValue(originalValue)
	require.NoError(t, verifyServiceMessage(req))
}

func testRequestSign(t *testing.T, meta *session.RequestMetaHeader, req request) {
	require.Error(t, verifyServiceMessage(req))

	// sign request
	require.NoError(t, signServiceMessage(signer, req))

	// verification must pass
	require.NoError(t, verifyServiceMessage(req))

	meta.SetOrigin(req.GetMetaHeader())
	req.SetMetaHeader(meta)

	// sign request
	require.NoError(t, signServiceMessage(signer, req))

	// verification must pass
	require.NoError(t, verifyServiceMessage(req))
}

func testRequestMeta(t *testing.T, meta *session.RequestMetaHeader, req serviceRequest) {
	// corrupt meta header
	meta.SetTTL(meta.GetTTL() + 1)

	// verification must fail
	require.Error(t, verifyServiceMessage(req))

	// restore meta header
	meta.SetTTL(meta.GetTTL() - 1)

	// corrupt origin verification header
	req.GetVerificationHeader().SetOrigin(nil)

	// verification must fail
	require.Error(t, verifyServiceMessage(req))
}

func testResponseSign(t *testing.T, meta *session.ResponseMetaHeader, resp testResponse) {
	require.Error(t, verifyServiceMessage(resp))

	// sign request
	require.NoError(t, signServiceMessage(signer, resp))

	// verification must pass
	require.NoError(t, verifyServiceMessage(resp))

	meta.SetOrigin(resp.GetMetaHeader())
	resp.SetMetaHeader(meta)

	// sign request
	require.NoError(t, signServiceMessage(signer, resp))

	// verification must pass
	require.NoError(t, verifyServiceMessage(resp))
}

func testResponseMeta(t *testing.T, meta *session.ResponseMetaHeader, req serviceResponse) {
	// corrupt meta header
	meta.SetTTL(meta.GetTTL() + 1)

	// verification must fail
	require.Error(t, verifyServiceMessage(req))

	// restore meta header
	meta.SetTTL(meta.GetTTL() - 1)

	// corrupt origin verification header
	req.GetVerificationHeader().SetOrigin(nil)

	// verification must fail
	require.Error(t, verifyServiceMessage(req))
}

func TestEmptyMessage(t *testing.T) {
	require.NoError(t, verifyServiceMessage(nil))
	require.NoError(t, signServiceMessage(signer, nil))
}

func TestBalanceRequest(t *testing.T) {
	var id user.ID
	require.NoError(t, user.IDFromSigner(&id, signer))

	var ownerID refs.OwnerID
	id.WriteToV2(&ownerID)

	body := accounting.BalanceRequestBody{}
	body.SetOwnerID(&ownerID)

	meta := &session.RequestMetaHeader{}
	meta.SetTTL(1)

	req := &accounting.BalanceRequest{}
	req.SetBody(&body)
	req.SetMetaHeader(meta)

	// add level to meta header matryoshka
	meta = &session.RequestMetaHeader{}
	testRequestSign(t, meta, req)

	testOwner(t, &ownerID, req)
	testRequestMeta(t, meta, req)
}

func TestBalanceResponse(t *testing.T) {
	dec := new(accounting.Decimal)
	dec.SetValue(100)

	body := new(accounting.BalanceResponseBody)
	body.SetBalance(dec)

	meta := new(session.ResponseMetaHeader)
	meta.SetTTL(1)

	resp := new(accounting.BalanceResponse)
	resp.SetBody(body)
	resp.SetMetaHeader(meta)

	// add level to meta header matryoshka
	meta = new(session.ResponseMetaHeader)
	testResponseSign(t, meta, resp)

	// corrupt body
	dec.SetValue(dec.GetValue() + 1)

	// verification must fail
	require.Error(t, verifyServiceMessage(resp))

	// restore body
	dec.SetValue(dec.GetValue() - 1)

	testResponseMeta(t, meta, resp)
}

func TestCreateRequest(t *testing.T) {
	var id user.ID
	require.NoError(t, user.IDFromSigner(&id, signer))

	var ownerID refs.OwnerID
	id.WriteToV2(&ownerID)

	body := session.CreateRequestBody{}
	body.SetOwnerID(&ownerID)
	body.SetExpiration(100)

	meta := &session.RequestMetaHeader{}
	meta.SetTTL(1)

	req := &session.CreateRequest{}
	req.SetBody(&body)
	req.SetMetaHeader(meta)

	// add level to meta header matryoshka
	meta = &session.RequestMetaHeader{}
	testRequestSign(t, meta, req)

	testOwner(t, &ownerID, req)

	// corrupt body
	body.SetExpiration(body.GetExpiration() + 1)

	// verification must fail
	require.Error(t, verifyServiceMessage(req))

	// restore body
	body.SetExpiration(body.GetExpiration() - 1)

	testRequestMeta(t, meta, req)
}

func TestCreateResponse(t *testing.T) {
	id := make([]byte, 8)
	_, err := rand.Read(id)
	require.NoError(t, err)

	sessionKey := make([]byte, 8)
	_, err = rand.Read(sessionKey)
	require.NoError(t, err)

	body := session.CreateResponseBody{}
	body.SetID(id)
	body.SetSessionKey(sessionKey)

	meta := &session.ResponseMetaHeader{}
	meta.SetTTL(1)

	req := &session.CreateResponse{}
	req.SetBody(&body)
	req.SetMetaHeader(meta)

	// add level to meta header matryoshka
	meta = &session.ResponseMetaHeader{}
	testResponseSign(t, meta, req)

	// corrupt body
	body.SetID([]byte{1})
	// verification must fail
	require.Error(t, verifyServiceMessage(req))
	// restore body
	body.SetID(id)

	// corrupt body
	body.SetSessionKey([]byte{1})
	// verification must fail
	require.Error(t, verifyServiceMessage(req))
	// restore body
	body.SetSessionKey(id)

	testResponseMeta(t, meta, req)
}
