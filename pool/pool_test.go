package pool

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func anyValidPeerAddress(ind uint) string { return fmt.Sprintf("peer%d:8080", ind) }

func TestBuildPoolClientFailed(t *testing.T) {
	mockClientBuilder1 := func(_ string) (internalClient, error) {
		return nil, errors.New("oops")
	}
	mockClientBuilder2 := func(addr string) (internalClient, error) {
		mockCli := newMockClient(addr, neofscryptotest.Signer())
		mockCli.errOnDial()
		return mockCli, nil
	}

	for name, b := range map[string]clientBuilder{
		"build": mockClientBuilder1,
		"dial":  mockClientBuilder2,
	} {
		t.Run(name, func(t *testing.T) {
			opts := InitParameters{
				signer:     usertest.User().RFC6979,
				nodeParams: []NodeParam{{1, anyValidPeerAddress(0), 1}},
			}
			opts.setClientBuilder(b)

			pool, err := NewPool(opts)
			require.NoError(t, err)
			err = pool.Dial(context.Background())
			require.Error(t, err)
		})
	}
}

func TestBuildPoolOneNodeFailed(t *testing.T) {
	nodes := []NodeParam{
		{1, anyValidPeerAddress(0), 1},
		{2, anyValidPeerAddress(1), 1},
	}

	mockClientBuilder := func(addr string) (internalClient, error) {
		signer := neofscryptotest.Signer()
		if addr == nodes[0].address {
			mockCli := newMockClient(addr, signer)
			mockCli.errOnEndpointInfo()
			return mockCli, nil
		}

		return newMockClient(addr, signer), nil
	}

	log, err := zap.NewProduction()
	require.NoError(t, err)
	opts := InitParameters{
		signer:                  usertest.User().RFC6979,
		clientRebalanceInterval: 1000 * time.Millisecond,
		logger:                  log,
		nodeParams:              nodes,
	}
	opts.setClientBuilder(mockClientBuilder)

	clientPool, err := NewPool(opts)
	require.NoError(t, err)
	err = clientPool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientPool.Close })

	condition := func() bool {
		cp, err := clientPool.connection()
		if err != nil {
			return false
		}

		return cp.address() == nodes[1].address
	}
	require.Never(t, condition, 900*time.Millisecond, 100*time.Millisecond)
	require.Eventually(t, condition, 3*time.Second, 300*time.Millisecond)
}

func TestBuildPoolZeroNodes(t *testing.T) {
	opts := InitParameters{
		signer: usertest.User(),
	}
	_, err := NewPool(opts)
	require.Error(t, err)
}

func TestBuildPoolNoSigner(t *testing.T) {
	_, err := NewPool(InitParameters{})
	require.Error(t, err)
}

func TestBuildPoolWrongSigner(t *testing.T) {
	opts := InitParameters{
		signer: usertest.User(),
	}
	_, err := NewPool(opts)
	require.Error(t, err)
}

func TestOneNode(t *testing.T) {
	signer1 := neofscryptotest.Signer()
	mockClientBuilder := func(addr string) (internalClient, error) {
		return newMockClient(addr, signer1), nil
	}

	opts := InitParameters{
		signer:     usertest.User().RFC6979,
		nodeParams: []NodeParam{{1, anyValidPeerAddress(0), 1}},
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	cp, err := pool.connection()
	require.NoError(t, err)
	require.Equal(t, opts.nodeParams[0].address, cp.address())
}

func TestTwoNodes(t *testing.T) {
	mockClientBuilder := func(addr string) (internalClient, error) {
		return newMockClient(addr, neofscryptotest.Signer()), nil
	}

	opts := InitParameters{
		signer: usertest.User().RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
			{1, anyValidPeerAddress(1), 1},
		},
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	cp, err := pool.connection()
	require.NoError(t, err)
	require.True(t, assertAuthKeyForAny(cp.address(), opts.nodeParams))
}

func assertAuthKeyForAny(addr string, nodes []NodeParam) bool {
	for _, node := range nodes {
		if addr == node.address {
			return true
		}
	}
	return false
}

func TestOneOfTwoFailed(t *testing.T) {
	nodes := []NodeParam{
		{1, anyValidPeerAddress(0), 1},
		{9, anyValidPeerAddress(1), 1},
	}

	mockClientBuilder := func(addr string) (internalClient, error) {
		signer := neofscryptotest.Signer()
		if addr == nodes[0].address {
			return newMockClient(addr, signer), nil
		}

		mockCli := newMockClient(addr, signer)
		mockCli.errOnEndpointInfo()
		mockCli.errOnNetworkInfo()
		return mockCli, nil
	}

	opts := InitParameters{
		signer:                  usertest.User().RFC6979,
		nodeParams:              nodes,
		clientRebalanceInterval: 200 * time.Millisecond,
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)

	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	time.Sleep(2 * time.Second)

	for range 5 {
		cp, err := pool.connection()
		require.NoError(t, err)
		require.True(t, assertAuthKeyForAny(cp.address(), nodes))
	}
}

func TestTwoFailed(t *testing.T) {
	mockClientBuilder := func(addr string) (internalClient, error) {
		mockCli := newMockClient(addr, neofscryptotest.Signer())
		mockCli.errOnEndpointInfo()
		return mockCli, nil
	}

	opts := InitParameters{
		signer: usertest.User().RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
			{1, anyValidPeerAddress(1), 1},
		},
		clientRebalanceInterval: 200 * time.Millisecond,
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)

	t.Cleanup(func() { _ = pool.Close })

	time.Sleep(2 * time.Second)

	_, err = pool.connection()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no healthy")
}

func TestSessionCache(t *testing.T) {
	usr := usertest.User()
	var mockCli *mockClient

	mockClientBuilder := func(addr string) (internalClient, error) {
		mockCli = newMockClient(addr, usr)
		return mockCli, nil
	}

	opts := InitParameters{
		signer: usr.RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
		clientRebalanceInterval: 30 * time.Second,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	cp, err := pool.connection()
	require.NoError(t, err)

	containerID := cidtest.ID()
	cacheKey := cacheKeyForSession(cp.address(), pool.signer, session.VerbObjectPut, containerID)

	hdr := objecttest.Object()
	hdr.SetContainerID(containerID)

	t.Run("no session token after pool creation", func(t *testing.T) {
		st, ok := pool.cache.Get(cacheKey)
		require.False(t, ok)
		require.False(t, st.AssertAuthKey(usr.Public()))
	})

	t.Run("session token was created after request", func(t *testing.T) {
		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.NoError(t, err)

		st, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
		require.True(t, st.AssertAuthKey(usr.Public()))
	})

	t.Run("session is not removed", func(t *testing.T) {
		// error on the next request to the node
		mockCli.statusOnPutObject(errors.New("some error"))

		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.Error(t, err)

		_, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
	})

	t.Run("session is removed, because of the special error", func(t *testing.T) {
		// error on the next request to the node
		mockCli.statusOnPutObject(apistatus.SessionTokenNotFound{})

		// make request,
		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.Error(t, err)

		// cache must not contain session token
		cp, err = pool.connection()
		require.NoError(t, err)
		_, ok := pool.cache.Get(cacheKey)
		require.False(t, ok)
	})

	t.Run("session created again", func(t *testing.T) {
		mockCli.statusOnPutObject(nil)

		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.NoError(t, err)

		_, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
	})
}

func TestPriority(t *testing.T) {
	nodes := []NodeParam{
		{1, anyValidPeerAddress(0), 1},
		{2, anyValidPeerAddress(1), 100},
	}

	mockClientBuilder := func(addr string) (internalClient, error) {
		signer := neofscryptotest.Signer()

		if addr == nodes[0].address {
			mockCli := newMockClient(addr, signer)
			mockCli.errOnEndpointInfo()
			return mockCli, nil
		}

		return newMockClient(addr, signer), nil
	}

	opts := InitParameters{
		signer:                  usertest.User().RFC6979,
		nodeParams:              nodes,
		clientRebalanceInterval: 1500 * time.Millisecond,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	firstNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		return cp.address() == nodes[0].address
	}

	secondNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		return cp.address() == nodes[1].address
	}
	require.Never(t, secondNode, time.Second, 200*time.Millisecond)

	require.Eventually(t, secondNode, time.Second, 200*time.Millisecond)
	require.Never(t, firstNode, time.Second, 200*time.Millisecond)
}

func TestSessionCacheWithKey(t *testing.T) {
	mockClientBuilder := func(addr string) (internalClient, error) {
		return newMockClient(addr, neofscryptotest.Signer()), nil
	}

	opts := InitParameters{
		signer: usertest.User().RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
		clientRebalanceInterval: 30 * time.Second,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(ctx)
	require.NoError(t, err)

	cp, err := pool.connection()
	require.NoError(t, err)

	var prm client.PrmObjectDelete
	anonSigner := usertest.User()

	_, err = pool.ObjectDelete(ctx, cid.ID{}, oid.ID{}, anonSigner, prm)
	require.NoError(t, err)

	st, _ := pool.cache.Get(cacheKeyForSession(cp.address(), anonSigner, session.VerbObjectDelete, cid.ID{}))
	require.True(t, st.AssertAuthKey(anonSigner.Public()))
}

func TestSessionTokenOwner(t *testing.T) {
	mockClientBuilder := func(addr string) (internalClient, error) {
		return newMockClient(addr, neofscryptotest.Signer()), nil
	}

	opts := InitParameters{
		signer: usertest.User().RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := NewPool(opts)
	require.NoError(t, err)
	err = p.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = p.Close })

	cp, err := p.connection()
	require.NoError(t, err)

	anonSigner := usertest.User()

	containerID := cidtest.ID()
	hdr := objecttest.Object()
	hdr.SetContainerID(containerID)

	_, err = p.ObjectPutInit(ctx, hdr, anonSigner, client.PrmObjectPutInit{})
	require.NoError(t, err)

	cacheKey := cacheKeyForSession(cp.address(), anonSigner, session.VerbObjectPut, containerID)
	st, ok := p.cache.Get(cacheKey)
	require.True(t, ok)
	require.True(t, st.AssertAuthKey(anonSigner.Public()))

	require.True(t, st.VerifySignature())
	require.True(t, st.Issuer() == anonSigner.UserID())
}

func TestSessionTokenV2(t *testing.T) {
	usr := usertest.User()
	var mockCli *mockClient

	mockClientBuilder := func(addr string) (internalClient, error) {
		mockCli = newMockClient(addr, usr)
		return mockCli, nil
	}

	opts := InitParameters{
		signer: usr.RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
		clientRebalanceInterval: 30 * time.Second,
		useV2Sessions:           true,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	require.True(t, pool.useV2Sessions)

	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	cp, err := pool.connection()
	require.NoError(t, err)

	containerID := cidtest.ID()
	cacheKey := cacheKeyForSessionV2(cp.address(), pool.signer, containerID, "")

	hdr := objecttest.Object()
	hdr.SetContainerID(containerID)

	t.Run("no session token after pool creation", func(t *testing.T) {
		_, ok := pool.cache.GetV2(cacheKey)
		require.False(t, ok)
	})

	t.Run("created after request", func(t *testing.T) {
		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.NoError(t, err)

		stV2, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok, "v2 token should be in cache")
		require.Equal(t, usr.UserID(), stV2.Issuer())

		// Verify that v1 token is NOT in cache
		_, okV1 := pool.cache.Get(cacheKey)
		require.False(t, okV1, "v1 token should NOT be in cache when using v2")
	})

	t.Run("not removed on regular error", func(t *testing.T) {
		mockCli.statusOnPutObject(errors.New("some error"))

		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.Error(t, err)

		_, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok)
	})

	t.Run("removed on SessionTokenNotFound error", func(t *testing.T) {
		mockCli.statusOnPutObject(apistatus.SessionTokenNotFound{})

		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.Error(t, err)

		// cache must not contain session token
		cp, err = pool.connection()
		require.NoError(t, err)
		_, ok := pool.cache.GetV2(cacheKey)
		require.False(t, ok)
	})

	t.Run("created again", func(t *testing.T) {
		mockCli.statusOnPutObject(nil)

		_, err = pool.ObjectPutInit(ctx, hdr, usr, client.PrmObjectPutInit{})
		require.NoError(t, err)

		stV2, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok)

		require.Equal(t, usr.UserID(), stV2.Issuer())
		require.NoError(t, stV2.Validate())

		neoPubKey, err := keys.NewPublicKeyFromBytes(mockCli.nodeKey, elliptic.P256())
		require.NoError(t, err)

		ecdsaPubKey := (*ecdsa.PublicKey)(neoPubKey)

		userID := user.NewFromECDSAPublicKey(*ecdsaPubKey)
		ok, err = stV2.AssertAuthority(userID, nil)
		require.NoError(t, err)
		require.True(t, ok)

		t.Run("delete with the same token", func(t *testing.T) {
			var prm client.PrmObjectDelete

			_, err = pool.ObjectDelete(ctx, containerID, hdr.GetID(), usr, prm)
			require.NoError(t, err)

			stV2d, ok := pool.cache.GetV2(cacheKey)
			require.True(t, ok)
			require.Equal(t, stV2, stV2d)
		})
	})
}

func TestSessionTokenV2DisableDelegation(t *testing.T) {
	originSigner := usertest.User()
	poolSigner := usertest.User()
	var mockCli *mockClient

	mockClientBuilder := func(addr string) (internalClient, error) {
		mockCli = newMockClient(addr, poolSigner)
		return mockCli, nil
	}

	opts := InitParameters{
		signer: poolSigner.RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
		clientRebalanceInterval: 30 * time.Second,
		useV2Sessions:           true,
	}
	opts.DisableSessionV2Delegation()
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	require.True(t, pool.disableDelegateSessionV2)

	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	containerID := cidtest.ID()
	testObject := objecttest.Object()
	testObject.SetContainerID(containerID)

	originToken := createOriginSessionV2Token(t, containerID, originSigner, poolSigner.UserID(), sessionv2.VerbObjectPut, sessionv2.VerbObjectDelete)
	require.NoError(t, originToken.Validate())

	var putPrm client.PrmObjectPutInit
	putPrm.WithinSessionV2(originToken)

	_, err = pool.ObjectPutInit(ctx, testObject, originSigner, putPrm)
	require.NoError(t, err)

	connection, err := pool.connection()
	require.NoError(t, err)

	tokenHash := hex.EncodeToString(sha256.New().Sum(originToken.SignedData()))
	cacheKey := cacheKeyForSessionV2(connection.address(), originSigner, containerID, tokenHash)
	cachedToken, isCached := pool.cache.GetV2(cacheKey)
	require.False(t, isCached)
	require.Zero(t, cachedToken)

	require.NoError(t, originToken.Validate())
	require.Equal(t, originSigner.UserID(), originToken.Issuer())

	var delPrm client.PrmObjectDelete
	delPrm.WithinSessionV2(originToken)

	_, err = pool.ObjectDelete(ctx, containerID, testObject.GetID(), originSigner, delPrm)
	require.NoError(t, err)

	cachedToken, isCached = pool.cache.GetV2(cacheKey)
	require.False(t, isCached)
	require.Zero(t, cachedToken)
}

func TestStatusMonitor(t *testing.T) {
	thresholdWindowSize := 1 * time.Second

	monitor := newClientStatusMonitor("", 10, thresholdWindowSize)

	count := 10
	for range count {
		monitor.incErrorRate()
	}

	require.Equal(t, uint64(count), monitor.overallErrorRate())
	require.Equal(t, uint32(count), monitor.currentErrorRate())

	time.Sleep(thresholdWindowSize * 2)
	monitor.incErrorRate()

	require.Equal(t, uint64(count+1), monitor.overallErrorRate())
	require.Equal(t, uint32(1), monitor.currentErrorRate())
}

func TestHandleError(t *testing.T) {
	monitor := newClientStatusMonitor("", 10, 10*time.Second)

	for i, tc := range []struct {
		err           error
		expectedError bool
		countError    bool
	}{
		{
			err:           nil,
			expectedError: false,
			countError:    false,
		},
		{
			err:           nil,
			expectedError: false,
			countError:    false,
		},
		{
			err:           errors.New("error"),
			expectedError: true,
			countError:    true,
		},
		{
			err:           errors.New("error"),
			expectedError: true,
			countError:    true,
		},
		{
			err:           apistatus.ObjectNotFound{},
			expectedError: true,
			countError:    false,
		},
		{
			err:           apistatus.ServerInternal{},
			expectedError: true,
			countError:    true,
		},
		{
			err:           apistatus.WrongMagicNumber{},
			expectedError: true,
			countError:    true,
		},
		{
			err:           apistatus.SignatureVerification{},
			expectedError: true,
			countError:    true,
		},
		{
			err:           apistatus.SignatureVerification{},
			expectedError: true,
			countError:    true,
		},
		{
			err:           apistatus.NodeUnderMaintenance{},
			expectedError: true,
			countError:    true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			errCount := monitor.currentErrorRate()
			monitor.updateErrorRate(tc.err)
			if tc.expectedError {
				require.Error(t, tc.err)
			} else {
				require.NoError(t, tc.err)
			}
			if tc.countError {
				errCount++
			}
			require.Equal(t, errCount, monitor.currentErrorRate())
		})
	}
}

func TestSwitchAfterErrorThreshold(t *testing.T) {
	nodes := []NodeParam{
		{1, anyValidPeerAddress(0), 1},
		{2, anyValidPeerAddress(1), 100},
	}

	errorThreshold := 5

	mockClientBuilder := func(addr string) (internalClient, error) {
		signer := neofscryptotest.Signer()
		if addr == nodes[0].address {
			mockCli := newMockClient(addr, signer)
			mockCli.setThreshold(uint32(errorThreshold))
			mockCli.statusOnGetObject(apistatus.ServerInternal{})
			return mockCli, nil
		}

		return newMockClient(addr, signer), nil
	}

	usr := usertest.User()

	opts := InitParameters{
		signer:                  usr.RFC6979,
		nodeParams:              nodes,
		clientRebalanceInterval: 30 * time.Second,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	for range errorThreshold {
		conn, err := pool.connection()
		require.NoError(t, err)
		require.Equal(t, nodes[0].address, conn.address())
		sdkClient, err := conn.getClient()
		require.NoError(t, err)
		_, _, err = sdkClient.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, usr, client.PrmObjectGet{})

		require.Error(t, err)
	}

	conn, err := pool.connection()
	require.NoError(t, err)
	require.Equal(t, nodes[1].address, conn.address())

	sdkClient, err := conn.getClient()
	require.NoError(t, err)
	_, _, err = sdkClient.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, usr, client.PrmObjectGet{})
	require.NoError(t, err)
}

type dialCloseOnlyClient struct {
	internalClient
	closeErr error
}

func (x dialCloseOnlyClient) dial(context.Context) error { return nil }
func (x dialCloseOnlyClient) Close() error               { return x.closeErr }

func TestPool_Close(t *testing.T) {
	require.Implements(t, (*io.Closer)(nil), new(Pool))

	const n = 10
	ns := make([]NodeParam, n)
	errs := make([]error, n)
	for i := range n {
		ns[i] = NodeParam{
			priority: i,
			address:  anyValidPeerAddress(uint(i)),
			weight:   float64(i),
		}
		errs[i] = fmt.Errorf("error#%d", i)
	}

	var opts InitParameters
	opts.setClientBuilder(func(endpoint string) (internalClient, error) {
		ind := slices.IndexFunc(ns, func(n NodeParam) bool { return n.address == endpoint })
		require.True(t, ind >= 0)
		return dialCloseOnlyClient{closeErr: errs[ind]}, nil
	})

	p, err := New(ns, usertest.User().RFC6979, opts)
	require.NoError(t, err)
	require.NoError(t, p.Dial(context.Background()))

	err = p.Close()
	require.Error(t, err)

	for i := range errs {
		require.ErrorIs(t, err, errs[i])
	}

	type multiError = interface{ Unwrap() []error }
	require.Implements(t, (*multiError)(nil), err)

	unwrapped := err.(multiError).Unwrap()
	require.Len(t, unwrapped, len(ns))
	for i := range unwrapped {
		require.Implements(t, (*multiError)(nil), unwrapped[i])
		pes := unwrapped[i].(multiError).Unwrap()
		require.Len(t, pes, 1)
		require.ErrorIs(t, errs[i], pes[0])
	}
}

func BenchmarkSlidingWindow(b *testing.B) {
	sw := newSlidingWindow(15*time.Second, 100)

	for b.Loop() {
		sw.Allow()
	}
}

func createOriginSessionV2Token(t *testing.T, containerID cid.ID, signer user.Signer, gatewayID user.ID, verbs ...sessionv2.Verb) sessionv2.Token {
	var tok sessionv2.Token
	tok.SetVersion(sessionv2.TokenCurrentVersion)
	tok.SetNonce(sessionv2.RandomNonce())
	tok.SetIat(time.Now())
	tok.SetNbf(time.Now())
	tok.SetExp(time.Now().Add(1 * time.Hour))

	err := tok.AddSubject(sessionv2.NewTargetUser(gatewayID))
	require.NoError(t, err)

	ctxV2, err := sessionv2.NewContext(containerID, verbs)
	require.NoError(t, err)
	err = tok.AddContext(ctxV2)
	require.NoError(t, err)

	err = tok.Sign(signer)
	require.NoError(t, err)

	return tok
}

func TestSessionTokenV2Delegation(t *testing.T) {
	user1 := usertest.User()
	poolSigner := usertest.User()
	var mockCli *mockClient

	mockClientBuilder := func(addr string) (internalClient, error) {
		mockCli = newMockClient(addr, poolSigner)
		return mockCli, nil
	}

	opts := InitParameters{
		signer: poolSigner.RFC6979,
		nodeParams: []NodeParam{
			{1, anyValidPeerAddress(0), 1},
		},
		clientRebalanceInterval: 30 * time.Second,
		useV2Sessions:           true,
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)

	err = pool.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close })

	containerID := cidtest.ID()
	hdr := objecttest.Object()
	hdr.SetContainerID(containerID)

	t.Run("delegation: user creates token, pool creates new with origin", func(t *testing.T) {
		originToken := createOriginSessionV2Token(t, containerID, user1, poolSigner.UserID(), sessionv2.VerbObjectPut, sessionv2.VerbObjectDelete)
		require.NoError(t, originToken.Validate())

		var prm client.PrmObjectPutInit
		prm.WithinSessionV2(originToken)

		_, err = pool.ObjectPutInit(ctx, hdr, user1, prm)
		require.NoError(t, err)

		cp, err := pool.connection()
		require.NoError(t, err)

		tokenHash := hex.EncodeToString(sha256.New().Sum(originToken.SignedData()))
		cacheKey := cacheKeyForSessionV2(cp.address(), user1, containerID, tokenHash)
		tokV2, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok, "v2 token should be in cache")

		require.Equal(t, poolSigner.UserID(), tokV2.Issuer())

		require.NotNil(t, tokV2.Origin())
		require.Equal(t, originToken.Issuer(), tokV2.Origin().Issuer())

		var prmDel client.PrmObjectDelete
		prmDel.WithinSessionV2(originToken)

		_, err = pool.ObjectDelete(ctx, containerID, hdr.GetID(), user1, prmDel)
		require.NoError(t, err)
	})

	t.Run("mixed: delegated and non-delegated requests", func(t *testing.T) {
		// First: regular request without delegation
		var prmRegular client.PrmObjectPutInit
		_, err := pool.ObjectPutInit(ctx, hdr, user1, prmRegular)
		require.NoError(t, err)

		cp, err := pool.connection()
		require.NoError(t, err)
		cacheKey := cacheKeyForSessionV2(cp.address(), user1, containerID, "")

		tokV2Regular, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok)
		require.Nil(t, tokV2Regular.Origin())

		// Second: request with delegation
		originToken := createOriginSessionV2Token(t, containerID, user1, poolSigner.UserID(), sessionv2.VerbObjectPut)

		var prmDelegated client.PrmObjectPutInit
		prmDelegated.WithinSessionV2(originToken)

		_, err = pool.ObjectPutInit(ctx, hdr, user1, prmDelegated)
		require.NoError(t, err)

		cacheKey = hex.EncodeToString(sha256.New().Sum(originToken.SignedData()))
		cacheKey = cacheKeyForSessionV2(cp.address(), user1, containerID, cacheKey)

		tokV2Delegated, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok)
		require.NotNil(t, tokV2Delegated.Origin())
		require.Equal(t, originToken.Issuer(), tokV2Delegated.Origin().Issuer())
	})

	t.Run("delegation: removed on SessionTokenNotFound error", func(t *testing.T) {
		originToken := createOriginSessionV2Token(t, containerID, user1, poolSigner.UserID(), sessionv2.VerbObjectPut)

		var prm client.PrmObjectPutInit
		prm.WithinSessionV2(originToken)

		_, err := pool.ObjectPutInit(ctx, hdr, user1, prm)
		require.NoError(t, err)

		cp, err := pool.connection()
		require.NoError(t, err)
		tokenHash := hex.EncodeToString(sha256.New().Sum(originToken.SignedData()))
		cacheKey := cacheKeyForSessionV2(cp.address(), user1, containerID, tokenHash)

		_, ok := pool.cache.GetV2(cacheKey)
		require.True(t, ok)

		mockCli.statusOnPutObject(apistatus.SessionTokenNotFound{})

		var prm2 client.PrmObjectPutInit
		prm2.WithinSessionV2(originToken)

		_, err = pool.ObjectPutInit(ctx, hdr, user1, prm2)
		require.Error(t, err)

		_, ok = pool.cache.GetV2(cacheKey)
		require.False(t, ok, "delegated token should be removed from cache on SessionTokenNotFound")
	})
}
