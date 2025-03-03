package pool

import (
	"context"
	"errors"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

type mockClient struct {
	signer neofscrypto.Signer
	clientStatusMonitor

	errorOnDial          bool
	errorOnCreateSession bool
	errorOnEndpointInfo  bool
	errorOnNetworkInfo   bool
	errOnGetObject       error
	errOnPutObject       error
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
	var ni netmap.NetworkInfo

	if m.errorOnNetworkInfo {
		err := errors.New("endpoint info")
		m.updateErrorRate(err)
		return ni, err
	}

	ni.SetRawNetworkParameter(string(testutil.RandByteSlice(16)), testutil.RandByteSlice(16))
	ni.SetCurrentEpoch(uint64(time.Now().Unix()))
	ni.SetMaxObjectSize(1024)

	return ni, nil
}

func (m *mockClient) NetMapSnapshot(_ context.Context, _ client.PrmNetMapSnapshot) (netmap.NetMap, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectPutInit(_ context.Context, _ object.Object, _ user.Signer, _ client.PrmObjectPutInit) (client.ObjectWriter, error) {
	return nil, m.errOnPutObject
}

func (m *mockClient) ObjectGetInit(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectGet) (object.Object, *client.PayloadReader, error) {
	var hdr object.Object
	var pl client.PayloadReader
	m.updateErrorRate(m.errOnGetObject)

	return hdr, &pl, m.errOnGetObject
}

func (m *mockClient) ObjectHead(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectHead) (*object.Object, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectRangeInit(_ context.Context, _ cid.ID, _ oid.ID, _, _ uint64, _ user.Signer, _ client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectDelete(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectDelete) (oid.ID, error) {
	return oid.ID{}, nil
}

func (m *mockClient) ObjectHash(_ context.Context, _ cid.ID, _ oid.ID, _ user.Signer, _ client.PrmObjectHash) ([][]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) ObjectSearchInit(_ context.Context, _ cid.ID, _ user.Signer, _ client.PrmObjectSearch) (*client.ObjectListReader, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) SearchObjects(context.Context, cid.ID, object.SearchFilters, []string, string, neofscrypto.Signer, client.SearchObjectsOptions) ([]client.SearchResultItem, string, error) {
	// TODO implement me
	panic("implement me")
}

func (m *mockClient) SessionCreate(_ context.Context, signer user.Signer, _ client.PrmSessionCreate) (*client.ResSessionCreate, error) {
	if m.errorOnCreateSession {
		err := errors.New("create session")
		m.updateErrorRate(err)
		return nil, err
	}

	b := make([]byte, signer.Public().MaxEncodedSize())
	signer.Public().Encode(b)

	res := client.NewResSessionCreate(testutil.RandByteSlice(16), b)
	return &res, nil
}

func (m *mockClient) EndpointInfo(_ context.Context, _ client.PrmEndpointInfo) (*client.ResEndpointInfo, error) {
	if m.errorOnEndpointInfo {
		err := errors.New("endpoint info")
		m.updateErrorRate(err)
		return &client.ResEndpointInfo{}, err
	}

	var ni netmap.NodeInfo
	ni.SetNetworkEndpoints(m.addr)
	res := client.NewResEndpointInfo(version.Current(), ni)

	return &res, nil
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

func (m *mockClient) statusOnPutObject(err error) {
	m.errOnPutObject = err
}

func (m *mockClient) dial(context.Context) error {
	if m.errorOnDial {
		return errors.New("dial error")
	}
	return nil
}

func (m *mockClient) restartIfUnhealthy(ctx context.Context) (healthy bool, changed bool) {
	_, err := m.EndpointInfo(ctx, client.PrmEndpointInfo{})
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

func (m *mockClient) SetNodeSession(*session.Object, neofscrypto.PublicKey) {
}

func (m *mockClient) GetNodeSession(neofscrypto.PublicKey) *session.Object {
	return nil
}

func (m *mockClient) ResetSessions() {
}

func (m *mockClient) Close() error { panic("unimplemented") }
