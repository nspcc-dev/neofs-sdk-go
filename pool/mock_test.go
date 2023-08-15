package pool

import (
	"context"
	"errors"

	"github.com/google/uuid"
	sessionv2 "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type mockClient struct {
	signer neofscrypto.Signer
	clientStatusMonitor

	errorOnDial          bool
	errorOnCreateSession bool
	errorOnEndpointInfo  bool
	errorOnNetworkInfo   bool
	errOnGetObject       error
}

func (m *mockClient) Dial(_ client.PrmDial) error {
	if m.errorOnDial {
		return errors.New("dial error")
	}
	return nil
}

func (m *mockClient) BalanceGet(_ context.Context, _ client.PrmBalanceGet) (accounting.Decimal, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerPut(_ context.Context, _ container.Container, _ neofscrypto.Signer, _ client.PrmContainerPut) (cid.ID, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerGet(_ context.Context, _ cid.ID, _ client.PrmContainerGet) (container.Container, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerList(_ context.Context, _ user.ID, _ client.PrmContainerList) ([]cid.ID, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerDelete(_ context.Context, _ cid.ID, _ neofscrypto.Signer, _ client.PrmContainerDelete) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerEACL(_ context.Context, _ cid.ID, _ client.PrmContainerEACL) (eacl.Table, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ContainerSetEACL(_ context.Context, _ eacl.Table, _ user.Signer, _ client.PrmContainerSetEACL) error {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) NetworkInfo(_ context.Context, _ client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) NetMapSnapshot(_ context.Context, _ client.PrmNetMapSnapshot) (netmap.NetMap, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectPutInit(_ context.Context, _ object.Object, _ user.Signer, _ client.PrmObjectPutInit) (client.ObjectWriter, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectGetInit(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectGet) (object.Object, *client.PayloadReader, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectHead(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectHead) (*client.ResObjectHead, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectRangeInit(_ context.Context, _ cid.ID, _ oid.ID, _, _ uint64, _ user.Signer, _ client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectDelete(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectDelete) (oid.ID, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectHash(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectHash) ([][]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectSearchInit(_ context.Context, _ cid.ID, _ user.Signer, _ client.PrmObjectSearch) (*client.ObjectListReader, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) SessionCreate(_ context.Context, _ user.Signer, _ client.PrmSessionCreate) (*client.ResSessionCreate, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) EndpointInfo(_ context.Context, _ client.PrmEndpointInfo) (*client.ResEndpointInfo, error) {
	// TODO implement me
	panic("implement me")
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

func (m *mockClient) statusOnGetObject(err error) {
	m.errOnGetObject = err
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

func (m *mockClient) containerPut(context.Context, container.Container, user.Signer, PrmContainerPut) (cid.ID, error) {
	return cid.ID{}, nil
}

func (m *mockClient) containerGet(context.Context, cid.ID) (container.Container, error) {
	return container.Container{}, nil
}

func (m *mockClient) containerList(context.Context, user.ID) ([]cid.ID, error) {
	return nil, nil
}

func (m *mockClient) containerDelete(context.Context, cid.ID, neofscrypto.Signer, PrmContainerDelete) error {
	return nil
}

func (m *mockClient) containerEACL(context.Context, cid.ID) (eacl.Table, error) {
	return eacl.Table{}, nil
}

func (m *mockClient) containerSetEACL(context.Context, eacl.Table, user.Signer, PrmContainerSetEACL) error {
	return nil
}

func (m *mockClient) endpointInfo(context.Context, prmEndpointInfo) (netmap.NodeInfo, error) {
	var ni netmap.NodeInfo

	if m.errorOnEndpointInfo {
		err := errors.New("endpoint info")
		m.updateErrorRate(err)
		return ni, err
	}

	ni.SetNetworkEndpoints(m.addr)
	return ni, nil
}

func (m *mockClient) networkInfo(context.Context, prmNetworkInfo) (netmap.NetworkInfo, error) {
	var ni netmap.NetworkInfo

	if m.errorOnNetworkInfo {
		err := errors.New("network info")
		m.updateErrorRate(err)
		return ni, err
	}

	return ni, nil
}

func (m *mockClient) objectPut(context.Context, user.Signer, PrmObjectPut) (oid.ID, error) {
	return oid.ID{}, nil
}

func (m *mockClient) objectDelete(context.Context, cid.ID, oid.ID, user.Signer, PrmObjectDelete) error {
	return nil
}

func (m *mockClient) objectGet(context.Context, cid.ID, oid.ID, user.Signer, PrmObjectGet) (ResGetObject, error) {
	var res ResGetObject

	if m.errOnGetObject == nil {
		return res, nil
	}

	m.updateErrorRate(m.errOnGetObject)
	return res, m.errOnGetObject
}

func (m *mockClient) objectHead(context.Context, cid.ID, oid.ID, user.Signer, PrmObjectHead) (object.Object, error) {
	return object.Object{}, nil
}

func (m *mockClient) objectRange(context.Context, cid.ID, oid.ID, uint64, uint64, user.Signer, PrmObjectRange) (ResObjectRange, error) {
	return ResObjectRange{}, nil
}

func (m *mockClient) objectSearch(context.Context, cid.ID, user.Signer, PrmObjectSearch) (ResObjectSearch, error) {
	return ResObjectSearch{}, nil
}

func (m *mockClient) sessionCreate(context.Context, user.Signer, prmCreateSession) (resCreateSession, error) {
	if m.errorOnCreateSession {
		err := errors.New("create session")
		m.updateErrorRate(err)
		return resCreateSession{}, err
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

func (m *mockClient) getClient() (sdkClientInterface, error) {
	return m, nil
}

func (m *mockClient) getRawClient() (*client.Client, error) {
	return nil, errors.New("now supported to return sdkClient from mockClient")
}

func (m *mockClient) SetNodeSession(*session.Object) {
}

func (m *mockClient) GetNodeSession() *session.Object {
	return nil
}
