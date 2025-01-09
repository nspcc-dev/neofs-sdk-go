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
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noOtherClientCalls struct{}

func (noOtherClientCalls) Dial(client.PrmDial) error { panic("must not be called") }

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

func (noOtherClientCalls) ObjectRangeInit(context.Context, cid.ID, oid.ID, uint64, uint64, user.Signer, client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectDelete(context.Context, cid.ID, oid.ID, user.Signer, client.PrmObjectDelete) (oid.ID, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectHash(context.Context, cid.ID, oid.ID, user.Signer, client.PrmObjectHash) ([][]byte, error) {
	panic("must not be called")
}

func (noOtherClientCalls) ObjectSearchInit(context.Context, cid.ID, user.Signer, client.PrmObjectSearch) (*client.ObjectListReader, error) {
	panic("must not be called")
}

func (noOtherClientCalls) SearchObjects(context.Context, cid.ID, uint32, neofscrypto.Signer, client.SearchObjectsOptions) ([]client.SearchResultItem, string, error) {
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

func (x mockedClientWrapper) isHealthy() bool          { return true }
func (x mockedClientWrapper) setUnhealthy()            { panic("must not be called") }
func (x mockedClientWrapper) address() string          { return x.addr }
func (x mockedClientWrapper) currentErrorRate() uint32 { panic("must not be called") }
func (x mockedClientWrapper) overallErrorRate() uint64 { panic("must not be called") }
func (x mockedClientWrapper) SetNodeSession(*session.Object, neofscrypto.PublicKey) {
	panic("must not be called")
}
func (x mockedClientWrapper) GetNodeSession(neofscrypto.PublicKey) *session.Object {
	panic("must not be called")
}
func (x mockedClientWrapper) ResetSessions()             { panic("must not be called") }
func (x mockedClientWrapper) dial(context.Context) error { return nil }
func (x mockedClientWrapper) restartIfUnhealthy(context.Context) (bool, bool) {
	panic("must not be called")
}
func (x mockedClientWrapper) getClient() (sdkClientInterface, error) { panic("must not be called") }
func (x mockedClientWrapper) getRawClient() (*client.Client, error)  { panic("must not be called") }

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
	getOpts.WithinSession(sessiontest.Object())
	getOpts.WithBearerToken(bearertest.Token())
	getOpts.MarkRaw()
	getOpts.MarkLocal()
	getOpts.WithXHeaders("k1", "v1", "k2", "v2")

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
	t.Cleanup(p.Close)

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
	headOpts.WithinSession(sessiontest.Object())
	headOpts.WithBearerToken(bearertest.Token())
	headOpts.MarkRaw()
	headOpts.MarkLocal()
	headOpts.WithXHeaders("k1", "v1", "k2", "v2")

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
	t.Cleanup(p.Close)

	hdr, err := p.ObjectHead(context.Background(), cnrID, objID, usr, headOpts)
	require.Equal(t, err, headClient.err)
	require.Equal(t, hdr, &headClient.hdr)
}

type objectRangeOnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr     cid.ID
	objID   oid.ID
	off, ln uint64
	sgnr    user.Signer
	opts    client.PrmObjectRange
	// ret
	pld *client.ObjectRangeReader
	err error
}

func (x objectRangeOnlyClient) ObjectRangeInit(ctx context.Context, cnr cid.ID, objID oid.ID, off, ln uint64, signer user.Signer, opts client.PrmObjectRange) (*client.ObjectRangeReader, error) {
	switch {
	case ctx == nil:
		return nil, errors.New("[test] nil context")
	case cnr != x.cnr:
		return nil, errors.New("[test] wrong container")
	case objID != x.objID:
		return nil, errors.New("[test] wrong object ID")
	case off != x.off:
		return nil, errors.New("[test] wrong range offset")
	case ln != x.ln:
		return nil, errors.New("[test] wrong range length")
	case !assert.ObjectsAreEqual(signer, x.sgnr):
		return nil, errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return nil, errors.New("[test] wrong options")
	}
	return x.pld, x.err
}

type objectRangeOnlyClientWrapper struct {
	mockedClientWrapper
	c objectRangeOnlyClient
}

func (x objectRangeOnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_ObjectRangeInit(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	objID := oidtest.ID()
	const off, ln = 13, 42
	usr := usertest.User()

	var rangeOpts client.PrmObjectRange
	rangeOpts.WithinSession(sessiontest.Object())
	rangeOpts.WithBearerToken(bearertest.Token())
	rangeOpts.MarkRaw()
	rangeOpts.MarkLocal()
	rangeOpts.WithXHeaders("k1", "v1", "k2", "v2")

	rangeClient := objectRangeOnlyClient{
		cnr:   cnrID,
		objID: objID,
		off:   off,
		ln:    ln,
		sgnr:  usr,
		opts:  rangeOpts,
		pld:   nil, // no way to construct
		err:   errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectRangeOnlyClientWrapper, len(endpoints))
	for i := range endpoints {
		nodes[i].address = endpoints[i]
		cws[i].addr = endpoints[i]
		cws[i].c = rangeClient
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
	t.Cleanup(p.Close)

	pld, err := p.ObjectRangeInit(context.Background(), cnrID, objID, off, ln, usr, rangeOpts)
	require.Equal(t, err, rangeClient.err)
	require.Equal(t, pld, rangeClient.pld)
}

type objectHashOnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr   cid.ID
	objID oid.ID
	sgnr  user.Signer
	opts  client.PrmObjectHash
	// ret
	hs  [][]byte
	err error
}

func (x objectHashOnlyClient) ObjectHash(ctx context.Context, cnr cid.ID, objID oid.ID, signer user.Signer, opts client.PrmObjectHash) ([][]byte, error) {
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
	return x.hs, x.err
}

type objectHashOnlyClientWrapper struct {
	mockedClientWrapper
	c objectHashOnlyClient
}

func (x objectHashOnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_ObjectHash(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	objID := oidtest.ID()
	usr := usertest.User()

	var hashOpts client.PrmObjectHash
	hashOpts.WithinSession(sessiontest.Object())
	hashOpts.WithBearerToken(bearertest.Token())
	hashOpts.MarkLocal()
	hashOpts.WithXHeaders("k1", "v1", "k2", "v2")
	hashOpts.TillichZemorAlgo()
	hashOpts.SetRangeList(1, 2, 3, 4)
	hashOpts.UseSalt([]byte("any_salt"))

	hashClient := objectHashOnlyClient{
		cnr:   cnrID,
		objID: objID,
		sgnr:  usr,
		opts:  hashOpts,
		hs:    [][]byte{[]byte("hash1"), []byte("hash2")},
		err:   errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectHashOnlyClientWrapper, len(endpoints))
	for i := range endpoints {
		nodes[i].address = endpoints[i]
		cws[i].addr = endpoints[i]
		cws[i].c = hashClient
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
	t.Cleanup(p.Close)

	hs, err := p.ObjectHash(context.Background(), cnrID, objID, usr, hashOpts)
	require.Equal(t, err, hashClient.err)
	require.Equal(t, hs, hashClient.hs)
}

type objectSearchOnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr  cid.ID
	sgnr user.Signer
	opts client.PrmObjectSearch
	// ret
	rdr *client.ObjectListReader
	err error
}

func (x objectSearchOnlyClient) ObjectSearchInit(ctx context.Context, cnr cid.ID, signer user.Signer, opts client.PrmObjectSearch) (*client.ObjectListReader, error) {
	switch {
	case ctx == nil:
		return nil, errors.New("[test] nil context")
	case cnr != x.cnr:
		return nil, errors.New("[test] wrong container")
	case !assert.ObjectsAreEqual(signer, x.sgnr):
		return nil, errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return nil, errors.New("[test] wrong options")
	}
	return x.rdr, x.err
}

type objectSearchOnlyClientWrapper struct {
	mockedClientWrapper
	c objectSearchOnlyClient
}

func (x objectSearchOnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_ObjectSearchInit(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	usr := usertest.User()

	var sfs object.SearchFilters
	sfs.AddFilter("k1", "v1", object.MatchStringEqual)
	sfs.AddFilter("k2", "v2", object.MatchStringNotEqual)

	var searchOpts client.PrmObjectSearch
	searchOpts.WithinSession(sessiontest.Object())
	searchOpts.WithBearerToken(bearertest.Token())
	searchOpts.MarkLocal()
	searchOpts.WithXHeaders("k1", "v1", "k2", "v2")
	searchOpts.SetFilters(sfs)

	searchClient := objectSearchOnlyClient{
		cnr:  cnrID,
		sgnr: usr,
		opts: searchOpts,
		rdr:  nil, // no way to construct
		err:  errors.New("any error"),
	}
	endpoints := []string{"localhost:8080", "localhost:8081"}
	nodes := make([]NodeParam, len(endpoints))
	cws := make([]objectSearchOnlyClientWrapper, len(endpoints))
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
	t.Cleanup(p.Close)

	rdr, err := p.ObjectSearchInit(context.Background(), cnrID, usr, searchOpts)
	require.Equal(t, err, searchClient.err)
	require.Equal(t, rdr, searchClient.rdr)
}

type objectSearchV2OnlyClient struct {
	noOtherClientCalls
	// expected input
	cnr    cid.ID
	count  uint32
	signer neofscrypto.Signer
	opts   client.SearchObjectsOptions
	// ret
	items  []client.SearchResultItem
	cursor string
	err    error
}

func (x objectSearchV2OnlyClient) SearchObjects(ctx context.Context, cnr cid.ID, count uint32, signer neofscrypto.Signer, opts client.SearchObjectsOptions) ([]client.SearchResultItem, string, error) {
	switch {
	case ctx == nil:
		return nil, "", errors.New("[test] nil context")
	case cnr != x.cnr:
		return nil, "", errors.New("[test] wrong container")
	case count != x.count:
		return nil, "", errors.New("[test] wrong count")
	case !assert.ObjectsAreEqual(signer, x.signer):
		return nil, "", errors.New("[test] wrong signer")
	case !assert.ObjectsAreEqual(opts, x.opts):
		return nil, "", errors.New("[test] wrong options")
	}
	return x.items, x.cursor, x.err
}

type objectSearchV2OnlyClientWrapper struct {
	mockedClientWrapper
	c objectSearchV2OnlyClient
}

func (x objectSearchV2OnlyClientWrapper) getClient() (sdkClientInterface, error) { return x.c, nil }

func TestPool_SearchObjects(t *testing.T) {
	ctx := context.Background()
	cnrID := cidtest.ID()
	const count = 1000
	signer := neofscryptotest.Signer()

	var fs object.SearchFilters
	fs.AddFilter("k1", "v1", object.MatchStringEqual)
	fs.AddFilter("k2", "v2", object.MatchStringNotEqual)

	var opts client.SearchObjectsOptions
	opts.WithXHeaders("k1", "v1", "k2", "v2")
	opts.DisableForwarding()
	opts.WithSessionToken(sessiontest.Object())
	opts.WithBearerToken(bearertest.Token())
	opts.SetFilters(fs)
	opts.AttachAttributes([]string{"a1", "a2", "a3"})

	searchClient := objectSearchV2OnlyClient{
		cnr:    cnrID,
		count:  count,
		signer: signer,
		opts:   opts,
		items: []client.SearchResultItem{
			{ID: oidtest.ID(), Attributes: []string{"val_1_1", "val_1_2"}},
			{ID: oidtest.ID(), Attributes: []string{"val_2_1", "val_2_2"}},
			{ID: oidtest.ID(), Attributes: []string{"val_3_1", "val_3_2"}},
		},
		cursor: "any_cursor",
		err:    errors.New("any error"),
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
	t.Cleanup(p.Close)

	items, cursor, err := p.SearchObjects(ctx, cnrID, count, signer, opts)
	require.Equal(t, items, searchClient.items)
	require.Equal(t, cursor, searchClient.cursor)
	require.Equal(t, err, searchClient.err)
}
