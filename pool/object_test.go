package pool

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noOtherClientCalls struct{}

func (noOtherClientCalls) DialEndpoint(context.Context, string) error { panic("must not be called") }

func (noOtherClientCalls) BalanceGet(context.Context, client.PrmBalanceGet) (accounting.Decimal, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerPut(context.Context, container.Container, neofscrypto.Signer, client.PrmContainerPut) (cid.ID, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerGet(context.Context, cid.ID, client.PrmContainerGet) (container.Container, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerList(context.Context, user.ID, client.PrmContainerList) ([]cid.ID, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerDelete(context.Context, cid.ID, neofscrypto.Signer, client.PrmContainerDelete) error {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerEACL(context.Context, cid.ID, client.PrmContainerEACL) (eacl.Table, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ContainerSetEACL(context.Context, eacl.Table, user.Signer, client.PrmContainerSetEACL) error {
	panic("must not be called")
}

func (noOtherClientCalls) SetContainerAttribute(context.Context, client.SetContainerAttributeParameters, neofscrypto.Signature, client.SetContainerAttributeOptions) error {
	panic("must not be called")
}

func (noOtherClientCalls) RemoveContainerAttribute(context.Context, client.RemoveContainerAttributeParameters, neofscrypto.Signature, client.RemoveContainerAttributeOptions) error {
	panic("must not be called")
}

func (noOtherClientCalls) NetworkInfo(context.Context, client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	panic("must not be called")
}

func (noOtherClientCalls) NetMapSnapshot(context.Context, client.PrmNetMapSnapshot) (netmap.NetMap, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectPutInit(context.Context, object.Object, user.Signer, client.PrmObjectPutInit) (client.ObjectWriter, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectGetInit(context.Context, cid.ID, oid.ID, user.Signer, client.PrmObjectGet) (object.Object, *client.PayloadReader, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectHead(context.Context, cid.ID, oid.ID, user.Signer, client.PrmObjectHead) (*object.Object, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectDelete(context.Context, cid.ID, oid.ID, user.Signer, client.PrmObjectDelete) (oid.ID, error) {
	panic("must not be called")
}

func (noOtherClientCalls) SearchObjects(context.Context, cid.ID, object.SearchFilters, []string, string, neofscrypto.Signer, client.SearchObjectsOptions) ([]client.SearchResultItem, string, error) {
	panic("must not be called")
}

func (noOtherClientCalls) SessionCreate(context.Context, user.Signer, client.PrmSessionCreate) (*client.ResSessionCreate, error) {
	panic("must not be called")
}

func (noOtherClientCalls) EndpointInfo(context.Context, client.PrmEndpointInfo) (*client.ResEndpointInfo, error) {
	panic("must not be called")
}

type mockedClientWrapper struct {
	addr string
}

func (x mockedClientWrapper) isHealthy() bool { return true }
func (x mockedClientWrapper) setUnhealthy()   { panic("must not be called") }
func (x mockedClientWrapper) address() string { return x.addr }
func (x mockedClientWrapper) SetNodeSession(*session.Object, neofscrypto.PublicKey) {
	panic("must not be called")
}
func (x mockedClientWrapper) GetNodeSession(neofscrypto.PublicKey) *session.Object {
	panic("must not be called")
}
func (x mockedClientWrapper) ResetSessions()             { panic("must not be called") }
func (x mockedClientWrapper) dial(context.Context) error { return nil }
func (x mockedClientWrapper) restartIfUnhealthy(context.Context) (bool, bool) {
	return true, false
}
func (x mockedClientWrapper) getClient() (sdkClientInterface, error) { panic("must not be called") }
func (x mockedClientWrapper) getRawClient() (*client.Client, error)  { panic("must not be called") }
func (x mockedClientWrapper) Close() error                           { panic("must not be called") }

type objectGetOnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr   cid.ID
	objID oid.ID
	sgnr  user.Signer
	opts  client.PrmObjectGet
	// ret
	hdr object.Object
	pld *client.PayloadReader
	err error
}

func (x objectGetOnlyClient) ObjectGetInit(ctx context.Context, cnr cid.ID, objID oid.ID, signer user.Signer, opts client.PrmObjectGet) (object.Object, *client.PayloadReader, error) {
	switch {
	case ctx == nil:
		return object.Object{}, nil, errors.New("[test] nil context")
	case cnr != x.cnr:
		return object.Object{}, nil, errors.New("[test] wrong container")
	case objID != x.objID:
		return object.Object{}, nil, errors.New("[test] wrong object ID")
	case !assert.ObjectsAreEqual(signer, x.sgnr):
		return object.Object{}, nil, errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return object.Object{}, nil, errors.New("[test] wrong options")
	}
	return x.hdr, x.pld, x.err
}

type objectGetOnlyClientWrapper struct {
	mockedClientWrapper
	c objectGetOnlyClient
}

func (x objectGetOnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_ObjectGetInit(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	objID := oidtest.ID()
	usr := usertest.User()

	var getOpts client.PrmObjectGet
	getOpts.WithBearerToken(bearertest.Token())
	getOpts.MarkRaw()
	getOpts.MarkLocal()
	getOpts.WithXHeaders("k1", "v1", "k2", "v2") //nolint:staticcheck

	getClient := objectGetOnlyClient{
		cnr:   cnrID,
		objID: objID,
		sgnr:  usr,
		opts:  getOpts,
		hdr:   objecttest.Object(),
		pld:   nil, // no way to construct
		err:   errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectGetOnlyClientWrapper, len(endpoints))
	for i := range endpoints {
		nodes[i].address = endpoints[i]
		cws[i].addr = endpoints[i]
		cws[i].c = getClient
	}

	var poolOpts InitParameters
	poolOpts.setClientBuilder(func(endpoint string) (internalClient, error) {
		ind := slices.Index(endpoints, endpoint)
		if ind < 0 {
			return nil, fmt.Errorf("unexpected endpoint %q", endpoint)
		}
		return &cws[ind], nil
	})
	p, err := New(nodes, usertest.User().RFC6979, poolOpts)
	require.NoError(t, err)
	require.NoError(t, p.Dial(ctx))
	t.Cleanup(func() { _ = p.Close })

	hdr, pld, err := p.ObjectGetInit(context.Background(), cnrID, objID, usr, getOpts)
	require.Equal(t, err, getClient.err)
	require.Equal(t, hdr, getClient.hdr)
	require.Equal(t, pld, getClient.pld)
}

type objectHeadOnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr   cid.ID
	objID oid.ID
	sgnr  user.Signer
	opts  client.PrmObjectHead
	// ret
	hdr object.Object
	err error
}

func (x objectHeadOnlyClient) ObjectHead(ctx context.Context, cnr cid.ID, objID oid.ID, signer user.Signer, opts client.PrmObjectHead) (*object.Object, error) {
	switch {
	case ctx == nil:
		return nil, errors.New("[test] nil context")
	case cnr != x.cnr:
		return nil, errors.New("[test] wrong container")
	case objID != x.objID:
		return nil, errors.New("[test] wrong object ID")
	case !assert.ObjectsAreEqual(signer, x.sgnr):
		return nil, errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return nil, errors.New("[test] wrong options")
	}
	return &x.hdr, x.err
}

type objectHeadOnlyClientWrapper struct {
	mockedClientWrapper
	c objectHeadOnlyClient
}

func (x objectHeadOnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_ObjectHead(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	objID := oidtest.ID()
	usr := usertest.User()

	var headOpts client.PrmObjectHead
	headOpts.WithBearerToken(bearertest.Token())
	headOpts.MarkRaw()
	headOpts.MarkLocal()
	headOpts.WithXHeaders("k1", "v1", "k2", "v2") //nolint:staticcheck

	headClient := objectHeadOnlyClient{
		cnr:   cnrID,
		objID: objID,
		sgnr:  usr,
		opts:  headOpts,
		hdr:   objecttest.Object(),
		err:   errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectHeadOnlyClientWrapper, len(endpoints))
	for i := range endpoints {
		nodes[i].address = endpoints[i]
		cws[i].addr = endpoints[i]
		cws[i].c = headClient
	}

	var poolOpts InitParameters
	poolOpts.setClientBuilder(func(endpoint string) (internalClient, error) {
		ind := slices.Index(endpoints, endpoint)
		if ind < 0 {
			return nil, fmt.Errorf("unexpected endpoint %q", endpoint)
		}
		return &cws[ind], nil
	})
	p, err := New(nodes, usertest.User().RFC6979, poolOpts)
	require.NoError(t, err)
	require.NoError(t, p.Dial(ctx))
	t.Cleanup(func() { _ = p.Close })

	hdr, err := p.ObjectHead(context.Background(), cnrID, objID, usr, headOpts)
	require.Equal(t, err, headClient.err)
	require.Equal(t, hdr, &headClient.hdr)
}

type objectSearchV2OnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr       cid.ID
	filters   object.SearchFilters
	attrs     []string
	reqCursor string
	signer    neofscrypto.Signer
	opts      client.SearchObjectsOptions
	// ret
	items      []client.SearchResultItem
	respCursor string
	err        error
}

func (x objectSearchV2OnlyClient) SearchObjects(ctx context.Context, cnr cid.ID, filters object.SearchFilters, attrs []string, cursor string,
	signer neofscrypto.Signer, opts client.SearchObjectsOptions) ([]client.SearchResultItem, string, error) {
	switch {
	case ctx == nil:
		return nil, "", errors.New("[test] nil context")
	case cnr != x.cnr:
		return nil, "", errors.New("[test] wrong container")
	case !assert.ObjectsAreEqual(filters, x.filters):
		return nil, "", errors.New("[test] wrong filters")
	case !assert.ObjectsAreEqual(attrs, x.attrs):
		return nil, "", errors.New("[test] wrong attributes")
	case cursor != x.reqCursor:
		return nil, "", errors.New("[test] wrong cursor")
	case !assert.ObjectsAreEqual(signer, x.signer):
		return nil, "", errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return nil, "", errors.New("[test] wrong options")
	}
	return x.items, x.respCursor, x.err
}

type objectSearchV2OnlyClientWrapper struct {
	mockedClientWrapper
	c objectSearchV2OnlyClient
}

func (x objectSearchV2OnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_SearchObjects(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	const reqCursor = "any_request_cursor"
	signer := neofscryptotest.Signer()
	attrs := []string{"a1", "a2", "a3"}
	var fs object.SearchFilters
	fs.AddFilter("k1", "v1", object.MatchStringEqual)
	fs.AddFilter("k2", "v2", object.MatchStringNotEqual)

	var opts client.SearchObjectsOptions
	opts.WithXHeaders("k1", "v1", "k2", "v2") //nolint:staticcheck
	opts.DisableForwarding()
	opts.WithBearerToken(bearertest.Token())
	opts.SetCount(1000)

	searchClient := objectSearchV2OnlyClient{
		cnr:       cnrID,
		filters:   fs,
		attrs:     attrs,
		reqCursor: "any_request_cursor",
		signer:    signer,
		opts:      opts,
		items: []client.SearchResultItem{
			{ID: oidtest.ID(), Attributes: []string{"val_1_1", "val_1_2"}},
			{ID: oidtest.ID(), Attributes: []string{"val_2_1", "val_2_2"}},
			{ID: oidtest.ID(), Attributes: []string{"val_3_1", "val_3_2"}},
		},
		respCursor: "any_response_cursor",
		err:        errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectSearchV2OnlyClientWrapper, len(endpoints))
	for i := range endpoints {
		nodes[i].address = endpoints[i]
		cws[i].addr = endpoints[i]
		cws[i].c = searchClient
	}

	var poolOpts InitParameters
	poolOpts.setClientBuilder(func(endpoint string) (internalClient, error) {
		ind := slices.Index(endpoints, endpoint)
		if ind < 0 {
			return nil, fmt.Errorf("unexpected endpoint %q", endpoint)
		}
		return &cws[ind], nil
	})
	p, err := New(nodes, usertest.User().RFC6979, poolOpts)
	require.NoError(t, err)
	require.NoError(t, p.Dial(ctx))
	t.Cleanup(func() { _ = p.Close })

	items, cursor, err := p.SearchObjects(ctx, cnrID, fs, attrs, reqCursor, signer, opts)
	require.Equal(t, items, searchClient.items)
	require.Equal(t, cursor, searchClient.respCursor)
	require.Equal(t, err, searchClient.err)
}
