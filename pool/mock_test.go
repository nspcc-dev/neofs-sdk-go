package pool

import (
	"context"
	"crypto/ecdsa"
	"errors"

	"github.com/google/uuid"
	sessionv2 "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"go.uber.org/atomic"
)

type mockClient struct {
	key        ecdsa.PrivateKey
	addr       string
	healthy    *atomic.Bool
	errorCount *atomic.Uint32

	errorOnCreateSession bool
	errorOnEndpointInfo  bool
	errorOnNetworkInfo   bool
	errorOnGetObject     error
}

func newMockClient(addr string, key ecdsa.PrivateKey) *mockClient {
	return &mockClient{
		key:        key,
		addr:       addr,
		healthy:    atomic.NewBool(true),
		errorCount: atomic.NewUint32(0),
	}
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

func (m *mockClient) errOnGetObject(err error) {
	m.errorOnGetObject = err
}

func newToken(key ecdsa.PrivateKey) *session.Object {
	var tok session.Object
	tok.SetID(uuid.New())
	pk := neofsecdsa.PublicKey(key.PublicKey)
	tok.SetAuthKey(&pk)

	return &tok
}

func (m *mockClient) balanceGet(context.Context, PrmBalanceGet) (*accounting.Decimal, error) {
	return nil, nil
}

func (m *mockClient) containerPut(context.Context, PrmContainerPut) (*cid.ID, error) {
	return nil, nil
}

func (m *mockClient) containerGet(context.Context, PrmContainerGet) (*container.Container, error) {
	return nil, nil
}

func (m *mockClient) containerList(context.Context, PrmContainerList) ([]cid.ID, error) {
	return nil, nil
}

func (m *mockClient) containerDelete(context.Context, PrmContainerDelete) error {
	return nil
}

func (m *mockClient) containerEACL(context.Context, PrmContainerEACL) (*eacl.Table, error) {
	return nil, nil
}

func (m *mockClient) containerSetEACL(context.Context, PrmContainerSetEACL) error {
	return nil
}

func (m *mockClient) endpointInfo(context.Context, prmEndpointInfo) (*netmap.NodeInfo, error) {
	if m.errorOnEndpointInfo {
		return nil, errors.New("error")
	}

	var ni netmap.NodeInfo
	ni.SetNetworkEndpoints(m.addr)
	return &ni, nil
}

func (m *mockClient) networkInfo(context.Context, prmNetworkInfo) (*netmap.NetworkInfo, error) {
	if m.errorOnNetworkInfo {
		return nil, errors.New("error")
	}

	var ni netmap.NetworkInfo
	return &ni, nil
}

func (m *mockClient) objectPut(context.Context, PrmObjectPut) (*oid.ID, error) {
	return nil, nil
}

func (m *mockClient) objectDelete(context.Context, PrmObjectDelete) error {
	return nil
}

func (m *mockClient) objectGet(context.Context, PrmObjectGet) (*ResGetObject, error) {
	return &ResGetObject{}, m.errorOnGetObject
}

func (m *mockClient) objectHead(context.Context, PrmObjectHead) (*object.Object, error) {
	return nil, nil
}

func (m *mockClient) objectRange(context.Context, PrmObjectRange) (*ResObjectRange, error) {
	return nil, nil
}

func (m *mockClient) objectSearch(context.Context, PrmObjectSearch) (*ResObjectSearch, error) {
	return nil, nil
}

func (m *mockClient) sessionCreate(context.Context, prmCreateSession) (*resCreateSession, error) {
	if m.errorOnCreateSession {
		return nil, errors.New("error")
	}

	tok := newToken(m.key)

	var v2tok sessionv2.Token
	tok.WriteToV2(&v2tok)

	return &resCreateSession{
		id:         v2tok.GetBody().GetID(),
		sessionKey: v2tok.GetBody().GetSessionKey(),
	}, nil
}

func (m *mockClient) isHealthy() bool {
	return m.healthy.Load()
}

func (m *mockClient) setHealthy(b bool) bool {
	return m.healthy.Swap(b) != b
}

func (m *mockClient) address() string {
	return m.addr
}

func (m *mockClient) errorRate() uint32 {
	return m.errorCount.Load()
}

func (m *mockClient) resetErrorCounter() {
	m.errorCount.Store(0)
}
