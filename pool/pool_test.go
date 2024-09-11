package pool

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

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
				nodeParams: []NodeParam{{1, "peer0", 1}},
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
		{1, "peer0", 1},
		{2, "peer1", 1},
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
	t.Cleanup(clientPool.Close)

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
		nodeParams: []NodeParam{{1, "peer0", 1}},
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(pool.Close)

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
			{1, "peer0", 1},
			{1, "peer1", 1},
		},
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(pool.Close)

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
		{1, "peer0", 1},
		{9, "peer1", 1},
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
	t.Cleanup(pool.Close)

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
			{1, "peer0", 1},
			{1, "peer1", 1},
		},
		clientRebalanceInterval: 200 * time.Millisecond,
	}
	opts.setClientBuilder(mockClientBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)

	t.Cleanup(pool.Close)

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
			{1, "peer0", 1},
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
	t.Cleanup(pool.Close)

	cp, err := pool.connection()
	require.NoError(t, err)

	var containerID cid.ID
	cacheKey := cacheKeyForSession(cp.address(), pool.signer, session.VerbObjectGet, containerID)

	t.Run("no session token after pool creation", func(t *testing.T) {
		st, ok := pool.cache.Get(cacheKey)
		require.False(t, ok)
		require.False(t, st.AssertAuthKey(usr.Public()))
	})

	t.Run("session token was created after request", func(t *testing.T) {
		_, _, err = pool.ObjectGetInit(ctx, containerID, oid.ID{}, usr, client.PrmObjectGet{})
		require.NoError(t, err)

		st, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
		require.True(t, st.AssertAuthKey(usr.Public()))
	})

	t.Run("session is not removed", func(t *testing.T) {
		// error on the next request to the node
		mockCli.statusOnGetObject(errors.New("some error"))

		_, _, err = pool.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, usr, client.PrmObjectGet{})
		require.Error(t, err)

		_, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
	})

	t.Run("session is removed, because of the special error", func(t *testing.T) {
		// error on the next request to the node
		mockCli.statusOnGetObject(apistatus.SessionTokenNotFound{})

		// make request,
		_, _, err = pool.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, usr, client.PrmObjectGet{})
		require.Error(t, err)

		// cache must not contain session token
		cp, err = pool.connection()
		require.NoError(t, err)
		_, ok := pool.cache.Get(cacheKey)
		require.False(t, ok)
	})

	t.Run("session created again", func(t *testing.T) {
		mockCli.statusOnGetObject(nil)

		_, _, err = pool.ObjectGetInit(ctx, cid.ID{}, oid.ID{}, usr, client.PrmObjectGet{})
		require.NoError(t, err)

		_, ok := pool.cache.Get(cacheKey)
		require.True(t, ok)
	})
}

func TestPriority(t *testing.T) {
	nodes := []NodeParam{
		{1, "peer0", 1},
		{2, "peer1", 100},
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
	t.Cleanup(pool.Close)

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
			{1, "peer0", 1},
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
			{1, "peer0", 1},
		},
	}
	opts.setClientBuilder(mockClientBuilder)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := NewPool(opts)
	require.NoError(t, err)
	err = p.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(p.Close)

	cp, err := p.connection()
	require.NoError(t, err)

	anonSigner := usertest.User()

	var containerID cid.ID

	_, _, err = p.ObjectGetInit(ctx, containerID, oid.ID{}, anonSigner, client.PrmObjectGet{})
	require.NoError(t, err)

	cacheKey := cacheKeyForSession(cp.address(), anonSigner, session.VerbObjectGet, containerID)
	st, ok := p.cache.Get(cacheKey)
	require.True(t, ok)
	require.True(t, st.AssertAuthKey(anonSigner.Public()))

	require.True(t, st.VerifySignature())
	require.True(t, st.Issuer() == anonSigner.UserID())
}

func TestStatusMonitor(t *testing.T) {
	monitor := newClientStatusMonitor("", 10)
	monitor.errorThreshold = 3

	count := 10
	for range count {
		monitor.incErrorRate()
	}

	require.Equal(t, uint64(count), monitor.overallErrorRate())
	require.Equal(t, uint32(1), monitor.currentErrorRate())
}

func TestHandleError(t *testing.T) {
	monitor := newClientStatusMonitor("", 10)

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
		{1, "peer0", 1},
		{2, "peer1", 100},
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
	t.Cleanup(pool.Close)

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
