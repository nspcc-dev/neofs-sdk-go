package pool

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildPoolClientFailed(t *testing.T) {
	mockClientBuilder1 := func(_ string) (client, error) {
		return nil, errors.New("oops")
	}
	mockClientBuilder2 := func(addr string) (client, error) {
		mockCli := newMockClient(addr, test.RandomSignerRFC6979(t))
		mockCli.errOnDial()
		return mockCli, nil
	}

	for name, b := range map[string]clientBuilder{
		"build": mockClientBuilder1,
		"dial":  mockClientBuilder2,
	} {
		t.Run(name, func(t *testing.T) {
			opts := InitParameters{
				signer:     test.RandomSignerRFC6979(t),
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

func TestBuildPoolCreateSessionFailed(t *testing.T) {
	clientMockBuilder := func(addr string) (client, error) {
		mockCli := newMockClient(addr, test.RandomSignerRFC6979(t))
		mockCli.errOnCreateSession()
		return mockCli, nil
	}

	opts := InitParameters{
		signer:     test.RandomSignerRFC6979(t),
		nodeParams: []NodeParam{{1, "peer0", 1}},
	}
	opts.setClientBuilder(clientMockBuilder)

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.Error(t, err)
}

func TestBuildPoolOneNodeFailed(t *testing.T) {
	nodes := []NodeParam{
		{1, "peer0", 1},
		{2, "peer1", 1},
	}

	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)

		if addr == nodes[0].address {
			mockCli := newMockClient(addr, key)
			mockCli.errOnEndpointInfo()
			return mockCli, nil
		}

		return newMockClient(addr, key), nil
	}

	log, err := zap.NewProduction()
	require.NoError(t, err)
	opts := InitParameters{
		signer:                  test.RandomSignerRFC6979(t),
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

	expectedAuthKey := clientKeys[1].Public()
	condition := func() bool {
		cp, err := clientPool.connection()
		if err != nil {
			return false
		}
		st, _ := clientPool.cache.Get(formCacheKey(cp.address(), clientPool.signer))
		return st.AssertAuthKey(expectedAuthKey)
	}
	require.Never(t, condition, 900*time.Millisecond, 100*time.Millisecond)
	require.Eventually(t, condition, 3*time.Second, 300*time.Millisecond)
}

func TestBuildPoolZeroNodes(t *testing.T) {
	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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
		signer: test.RandomSigner(t),
	}
	_, err := NewPool(opts)
	require.Error(t, err)
}

func TestOneNode(t *testing.T) {
	key1 := test.RandomSignerRFC6979(t)
	mockClientBuilder := func(addr string) (client, error) {
		return newMockClient(addr, key1), nil
	}

	opts := InitParameters{
		signer:     test.RandomSignerRFC6979(t),
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
	st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.True(t, st.AssertAuthKey(key1.Public()))
}

func TestTwoNodes(t *testing.T) {
	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)
		return newMockClient(addr, key), nil
	}

	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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
	st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.True(t, assertAuthKeyForAny(st, clientKeys))
}

func assertAuthKeyForAny(st session.Object, clientKeys []neofscrypto.Signer) bool {
	for _, key := range clientKeys {
		expectedAuthKey := key.Public()
		if st.AssertAuthKey(expectedAuthKey) {
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

	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)

		if addr == nodes[0].address {
			return newMockClient(addr, key), nil
		}

		mockCli := newMockClient(addr, key)
		mockCli.errOnEndpointInfo()
		mockCli.errOnNetworkInfo()
		return mockCli, nil
	}

	opts := InitParameters{
		signer:                  test.RandomSignerRFC6979(t),
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

	for i := 0; i < 5; i++ {
		cp, err := pool.connection()
		require.NoError(t, err)
		st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
		require.True(t, assertAuthKeyForAny(st, clientKeys))
	}
}

func TestTwoFailed(t *testing.T) {
	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)
		mockCli := newMockClient(addr, key)
		mockCli.errOnEndpointInfo()
		return mockCli, nil
	}

	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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
	key := test.RandomSignerRFC6979(t)

	mockClientBuilder := func(addr string) (client, error) {
		mockCli := newMockClient(addr, key)
		mockCli.statusOnGetObject(apistatus.SessionTokenNotFound{})
		return mockCli, nil
	}

	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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

	// cache must contain session token
	cp, err := pool.connection()
	require.NoError(t, err)
	st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.True(t, st.AssertAuthKey(key.Public()))

	var prm PrmObjectGet
	prm.UseSession(session.Object{})

	_, err = pool.GetObject(ctx, cid.ID{}, oid.ID{}, prm)
	require.Error(t, err)

	// cache must not contain session token
	cp, err = pool.connection()
	require.NoError(t, err)
	_, ok := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.False(t, ok)

	var prm2 PrmObjectPut
	prm2.SetHeader(object.Object{})

	_, err = pool.PutObject(ctx, prm2)
	require.NoError(t, err)

	// cache must contain session token
	cp, err = pool.connection()
	require.NoError(t, err)
	st, _ = pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.True(t, st.AssertAuthKey(key.Public()))
}

func TestPriority(t *testing.T) {
	nodes := []NodeParam{
		{1, "peer0", 1},
		{2, "peer1", 100},
	}

	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)

		if addr == nodes[0].address {
			mockCli := newMockClient(addr, key)
			mockCli.errOnEndpointInfo()
			return mockCli, nil
		}

		return newMockClient(addr, key), nil
	}

	opts := InitParameters{
		signer:                  test.RandomSignerRFC6979(t),
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

	expectedAuthKey1 := clientKeys[0].Public()
	firstNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
		return st.AssertAuthKey(expectedAuthKey1)
	}

	expectedAuthKey2 := clientKeys[1].Public()
	secondNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
		return st.AssertAuthKey(expectedAuthKey2)
	}
	require.Never(t, secondNode, time.Second, 200*time.Millisecond)

	require.Eventually(t, secondNode, time.Second, 200*time.Millisecond)
	require.Never(t, firstNode, time.Second, 200*time.Millisecond)
}

func TestSessionCacheWithKey(t *testing.T) {
	key := test.RandomSignerRFC6979(t)

	mockClientBuilder := func(addr string) (client, error) {
		return newMockClient(addr, key), nil
	}

	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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

	// cache must contain session token
	cp, err := pool.connection()
	require.NoError(t, err)
	st, _ := pool.cache.Get(formCacheKey(cp.address(), pool.signer))
	require.True(t, st.AssertAuthKey(key.Public()))

	var prm PrmObjectDelete
	anonKey := test.RandomSignerRFC6979(t)
	prm.UseSigner(anonKey)

	err = pool.DeleteObject(ctx, cid.ID{}, oid.ID{}, prm)
	require.NoError(t, err)
	st, _ = pool.cache.Get(formCacheKey(cp.address(), anonKey))
	require.True(t, st.AssertAuthKey(key.Public()))
}

func TestSessionTokenOwner(t *testing.T) {
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		return newMockClient(addr, key), nil
	}

	opts := InitParameters{
		signer: test.RandomSignerRFC6979(t),
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

	anonKey := test.RandomSignerRFC6979(t)
	var anonOwner user.ID
	require.NoError(t, user.IDFromSigner(&anonOwner, anonKey))

	var prm prmCommon
	prm.UseSigner(anonKey)
	var prmCtx prmContext
	prmCtx.useDefaultSession()

	var tkn session.Object
	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = func(tok session.Object) {
		tkn = tok
	}
	err = p.initCallContext(&cc, prm, prmCtx)
	require.NoError(t, err)

	err = p.openDefaultSession(&cc)
	require.NoError(t, err)
	require.True(t, tkn.VerifySignature())
	require.True(t, tkn.Issuer().Equals(anonOwner))
}

func TestWaitPresence(t *testing.T) {
	mockCli := newMockClient("", test.RandomSignerRFC6979(t))

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(500 * time.Millisecond)
			cancel()
		}()

		var idCnr cid.ID

		err := waitForContainerPresence(ctx, mockCli, idCnr, &WaitParams{
			timeout:      120 * time.Second,
			pollInterval: 5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context canceled")
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx := context.Background()
		var idCnr cid.ID
		err := waitForContainerPresence(ctx, mockCli, idCnr, &WaitParams{
			timeout:      500 * time.Millisecond,
			pollInterval: 5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		var idCnr cid.ID
		err := waitForContainerPresence(ctx, mockCli, idCnr, &WaitParams{
			timeout:      10 * time.Second,
			pollInterval: 500 * time.Millisecond,
		})
		require.NoError(t, err)
	})
}

func TestStatusMonitor(t *testing.T) {
	monitor := newClientStatusMonitor("", 10)
	monitor.errorThreshold = 3

	count := 10
	for i := 0; i < count; i++ {
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

	var clientKeys []neofscrypto.Signer
	mockClientBuilder := func(addr string) (client, error) {
		key := test.RandomSignerRFC6979(t)
		clientKeys = append(clientKeys, key)

		if addr == nodes[0].address {
			mockCli := newMockClient(addr, key)
			mockCli.setThreshold(uint32(errorThreshold))
			mockCli.statusOnGetObject(apistatus.ServerInternal{})
			return mockCli, nil
		}

		return newMockClient(addr, key), nil
	}

	opts := InitParameters{
		signer:                  test.RandomSignerRFC6979(t),
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

	for i := 0; i < errorThreshold; i++ {
		conn, err := pool.connection()
		require.NoError(t, err)
		require.Equal(t, nodes[0].address, conn.address())
		_, err = conn.objectGet(ctx, cid.ID{}, oid.ID{}, PrmObjectGet{})
		require.Error(t, err)
	}

	conn, err := pool.connection()
	require.NoError(t, err)
	require.Equal(t, nodes[1].address, conn.address())
	_, err = conn.objectGet(ctx, cid.ID{}, oid.ID{}, PrmObjectGet{})
	require.NoError(t, err)
}
