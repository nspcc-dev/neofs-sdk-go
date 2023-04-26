package pool

import (
	"context"
	"errors"

	"github.com/google/uuid"
	sessionv2 "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	sdkClient "github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

type mockClient struct {
	signer neofscrypto.Signer
	clientStatusMonitor

	errorOnDial          bool
	errorOnCreateSession bool
	errorOnEndpointInfo  bool
	errorOnNetworkInfo   bool
	stOnGetObject        apistatus.Status
}

func newMockClient(addr string, signer neofscrypto.Signer) *mockClient {
	return &mockClient{
		signer:              signer,
		clientStatusMonitor: newClientStatusMonitor(addr, 10),
	}
}

func (m *mockClient) setThreshold(threshold uint32) {
	m.errorThreshold = threshold
}

func (m *mockClient) errOnCreateSession() {
	m.errorOnCreateSession = true
}

func (m *mockClient) errOnEndpointInfo() {
	m.errorOnEndpointInfo = true
}

func (m *mockClient) errOnNetworkInfo() {
	m.errorOnEndpointInfo = true
}

func (m *mockClient) errOnDial() {
	m.errorOnDial = true
	m.errOnCreateSession()
	m.errOnEndpointInfo()
	m.errOnNetworkInfo()
}

func (m *mockClient) statusOnGetObject(st apistatus.Status) {
	m.stOnGetObject = st
}

func newToken(signer neofscrypto.Signer) *session.Object {
	var tok session.Object
	tok.SetID(uuid.New())

	public := signer.Public()

	tok.SetAuthKey(public)

	return &tok
}

func (m *mockClient) balanceGet(context.Context, PrmBalanceGet) (accounting.Decimal, error) {
	return accounting.Decimal{}, nil
}

func (m *mockClient) containerPut(context.Context, PrmContainerPut) (cid.ID, error) {
	return cid.ID{}, nil
}

func (m *mockClient) containerGet(context.Context, PrmContainerGet) (container.Container, error) {
	return container.Container{}, nil
}

func (m *mockClient) containerList(context.Context, PrmContainerList) ([]cid.ID, error) {
	return nil, nil
}

func (m *mockClient) containerDelete(context.Context, PrmContainerDelete) error {
	return nil
}

func (m *mockClient) containerEACL(context.Context, PrmContainerEACL) (eacl.Table, error) {
	return eacl.Table{}, nil
}

func (m *mockClient) containerSetEACL(context.Context, PrmContainerSetEACL) error {
	return nil
}

func (m *mockClient) endpointInfo(context.Context, prmEndpointInfo) (netmap.NodeInfo, error) {
	var ni netmap.NodeInfo

	if m.errorOnEndpointInfo {
		return ni, m.handleError(nil, errors.New("error"))
	}

	ni.SetNetworkEndpoints(m.addr)
	return ni, nil
}

func (m *mockClient) networkInfo(context.Context, prmNetworkInfo) (netmap.NetworkInfo, error) {
	var ni netmap.NetworkInfo

	if m.errorOnNetworkInfo {
		return ni, m.handleError(nil, errors.New("error"))
	}

	return ni, nil
}

func (m *mockClient) objectPut(context.Context, PrmObjectPut) (oid.ID, error) {
	return oid.ID{}, nil
}

func (m *mockClient) objectDelete(context.Context, PrmObjectDelete) error {
	return nil
}

func (m *mockClient) objectGet(context.Context, PrmObjectGet) (ResGetObject, error) {
	var res ResGetObject

	if m.stOnGetObject == nil {
		return res, nil
	}

	status := apistatus.ErrFromStatus(m.stOnGetObject)
	return res, m.handleError(status, nil)
}

func (m *mockClient) objectHead(context.Context, PrmObjectHead) (object.Object, error) {
	return object.Object{}, nil
}

func (m *mockClient) objectRange(context.Context, PrmObjectRange) (ResObjectRange, error) {
	return ResObjectRange{}, nil
}

func (m *mockClient) objectSearch(context.Context, PrmObjectSearch) (ResObjectSearch, error) {
	return ResObjectSearch{}, nil
}

func (m *mockClient) sessionCreate(context.Context, prmCreateSession) (resCreateSession, error) {
	if m.errorOnCreateSession {
		return resCreateSession{}, m.handleError(nil, errors.New("error"))
	}

	tok := newToken(m.signer)

	var v2tok sessionv2.Token
	tok.WriteToV2(&v2tok)

	return resCreateSession{
		id:         v2tok.GetBody().GetID(),
		sessionKey: v2tok.GetBody().GetSessionKey(),
	}, nil
}

func (m *mockClient) dial(context.Context) error {
	if m.errorOnDial {
		return errors.New("dial error")
	}
	return nil
}

func (m *mockClient) restartIfUnhealthy(ctx context.Context) (healthy bool, changed bool) {
	_, err := m.endpointInfo(ctx, prmEndpointInfo{})
	healthy = err == nil
	changed = healthy != m.isHealthy()
	if healthy {
		m.setHealthy()
	} else {
		m.setUnhealthy()
	}
	return
}

func (m *mockClient) getClient() (*sdkClient.Client, error) {
	return nil, errors.New("now supported to return sdkClient from mockClient")
}
