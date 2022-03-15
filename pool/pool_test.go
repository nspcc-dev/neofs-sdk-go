package pool

//go:generate mockgen -destination mock_test.go -package pool . Client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	sdkClient "github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildPoolClientFailed(t *testing.T) {
	clientBuilder := func(_ string) (client, error) {
		return nil, fmt.Errorf("error")
	}

	opts := InitParameters{
		key:           newPrivateKey(t),
		nodeParams:    []NodeParam{{1, "peer0", 1}},
		clientBuilder: clientBuilder,
	}

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.Error(t, err)
}

func TestBuildPoolCreateSessionFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	ni := &netmap.NodeInfo{}
	ni.SetAddresses("addr1", "addr2")

	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error session")).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResEndpointInfo{}, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	opts := InitParameters{
		key:           newPrivateKey(t),
		nodeParams:    []NodeParam{{1, "peer0", 1}},
		clientBuilder: clientBuilder,
	}

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.Error(t, err)
}

func newPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	p, err := keys.NewPrivateKey()
	require.NoError(t, err)
	return &p.PrivateKey
}

func TestBuildPoolOneNodeFailed(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	ni := &netmap.NodeInfo{}
	ni.SetAddresses("addr1", "addr2")

	var expectedToken *session.Token
	clientCount := -1
	clientBuilder := func(_ string) (client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockInvokes := 0
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			mockInvokes++
			if mockInvokes == 1 {
				expectedToken = newToken(t)
				return nil, fmt.Errorf("error session")
			}
			return expectedToken, nil
		}).AnyTimes()

		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient2.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()

		if clientCount == 0 {
			return mockClient, nil
		}
		return mockClient2, nil
	}

	log, err := zap.NewProduction()
	require.NoError(t, err)
	opts := InitParameters{
		key:                     newPrivateKey(t),
		clientBuilder:           clientBuilder,
		clientRebalanceInterval: 1000 * time.Millisecond,
		logger:                  log,
		nodeParams: []NodeParam{
			{9, "peer0", 1},
			{1, "peer1", 1},
		},
	}

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
		st := clientPool.cache.Get(formCacheKey(cp.address, clientPool.key))
		return st == expectedToken
	}
	require.Never(t, condition, 900*time.Millisecond, 100*time.Millisecond)
	require.Eventually(t, condition, 3*time.Second, 300*time.Millisecond)
}

func TestBuildPoolZeroNodes(t *testing.T) {
	opts := InitParameters{
		key: newPrivateKey(t),
	}
	_, err := NewPool(opts)
	require.Error(t, err)
}

func TestOneNode(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)

	tok := session.NewToken()
	uid, err := uuid.New().MarshalBinary()
	require.NoError(t, err)
	tok.SetID(uid)

	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(tok, nil)
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResEndpointInfo{}, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	opts := InitParameters{
		key:           newPrivateKey(t),
		nodeParams:    []NodeParam{{1, "peer0", 1}},
		clientBuilder: clientBuilder,
	}

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	cp, err := pool.connection()
	require.NoError(t, err)
	st := pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Equal(t, tok, st)
}

func TestTwoNodes(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return tok, err
		})
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResEndpointInfo{}, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
			{1, "peer1", 1},
		},
		clientBuilder: clientBuilder,
	}

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(context.Background())
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	cp, err := pool.connection()
	require.NoError(t, err)
	st := pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Contains(t, tokens, st)
}

func TestOneOfTwoFailed(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	var tokens []*session.Token
	clientCount := -1
	clientBuilder := func(_ string) (client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient2.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ ...interface{}) (*sdkClient.ResEndpointInfo, error) {
			return nil, fmt.Errorf("error")
		}).AnyTimes()
		mockClient2.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ ...interface{}) (*sdkClient.ResNetworkInfo, error) {
			return nil, fmt.Errorf("error")
		}).AnyTimes()

		if clientCount == 0 {
			return mockClient, nil
		}
		return mockClient2, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
			{9, "peer1", 1},
		},
		clientRebalanceInterval: 200 * time.Millisecond,
		clientBuilder:           clientBuilder,
	}

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
		st := pool.cache.Get(formCacheKey(cp.address, pool.key))
		require.Equal(t, tokens[0], st)
	}
}

func TestTwoFailed(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)

	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		return mockClient, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
			{1, "peer1", 1},
		},
		clientRebalanceInterval: 200 * time.Millisecond,
		clientBuilder:           clientBuilder,
	}

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
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return tok, err
		}).MaxTimes(3)

		mockClient.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("session token does not exist"))
		mockClient.EXPECT().PutObject(gomock.Any(), gomock.Any()).Return(nil, nil)

		return mockClient, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
		},
		clientRebalanceInterval: 30 * time.Second,
		clientBuilder:           clientBuilder,
	}

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
	st := pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Contains(t, tokens, st)

	var prm PrmObjectGet
	prm.SetAddress(address.Address{})

	_, err = pool.GetObject(ctx, prm)
	require.Error(t, err)

	// cache must not contain session token
	cp, err = pool.connection()
	require.NoError(t, err)
	st = pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Nil(t, st)

	var prm2 PrmObjectPut
	prm2.SetHeader(object.Object{})

	_, err = pool.PutObject(ctx, prm2)
	require.NoError(t, err)

	// cache must contain session token
	cp, err = pool.connection()
	require.NoError(t, err)
	st = pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Contains(t, tokens, st)
}

func TestPriority(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	var tokens []*session.Token
	clientCount := -1
	clientBuilder := func(_ string) (client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient2.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient2.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

		if clientCount == 0 {
			return mockClient, nil
		}
		return mockClient2, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
			{2, "peer1", 100},
		},
		clientRebalanceInterval: 1500 * time.Millisecond,
		clientBuilder:           clientBuilder,
	}

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
		st := pool.cache.Get(formCacheKey(cp.address, pool.key))
		return st == tokens[0]
	}
	secondNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		st := pool.cache.Get(formCacheKey(cp.address, pool.key))
		return st == tokens[1]
	}
	require.Never(t, secondNode, time.Second, 200*time.Millisecond)

	require.Eventually(t, secondNode, time.Second, 200*time.Millisecond)
	require.Never(t, firstNode, time.Second, 200*time.Millisecond)
}

func TestSessionCacheWithKey(t *testing.T) {
	t.Skip("NeoFS API client can't be mocked") // neofs-sdk-go#85

	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return tok, err
		}).MaxTimes(2)

		mockClient.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

		return mockClient, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
		},
		clientRebalanceInterval: 30 * time.Second,
		clientBuilder:           clientBuilder,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := NewPool(opts)
	require.NoError(t, err)
	err = pool.Dial(ctx)
	require.NoError(t, err)

	// cache must contain session token
	cp, err := pool.connection()
	require.NoError(t, err)
	st := pool.cache.Get(formCacheKey(cp.address, pool.key))
	require.Contains(t, tokens, st)

	var prm PrmObjectGet
	prm.SetAddress(address.Address{})
	prm.UseKey(newPrivateKey(t))

	_, err = pool.GetObject(ctx, prm)
	require.NoError(t, err)
	require.Len(t, tokens, 2)
}

func newToken(t *testing.T) *session.Token {
	tok := session.NewToken()
	uid, err := uuid.New().MarshalBinary()
	require.NoError(t, err)
	tok.SetID(uid)

	return tok
}

func TestSessionTokenOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(&sdkClient.ResSessionCreate{}, nil).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResEndpointInfo{}, nil).AnyTimes()
		mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	opts := InitParameters{
		key: newPrivateKey(t),
		nodeParams: []NodeParam{
			{1, "peer0", 1},
		},
		clientBuilder: clientBuilder,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p, err := NewPool(opts)
	require.NoError(t, err)
	err = p.Dial(ctx)
	require.NoError(t, err)
	t.Cleanup(p.Close)

	anonKey := newPrivateKey(t)
	anonOwner := owner.NewIDFromPublicKey(&anonKey.PublicKey)

	var prm prmCommon
	prm.UseKey(anonKey)
	prm.useDefaultSession()
	cp, err := p.conn(ctx, prm)
	require.NoError(t, err)

	tkn := p.cache.Get(formCacheKey(cp.address, anonKey))
	require.True(t, anonOwner.Equal(tkn.OwnerID()))
}

func TestWaitPresence(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockClient := NewMockClient(ctrl)
	mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().NetworkInfo(gomock.Any(), gomock.Any()).Return(&sdkClient.ResNetworkInfo{}, nil).AnyTimes()
	mockClient.EXPECT().GetContainer(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	cache, err := newCache()
	require.NoError(t, err)

	inner := &innerPool{
		sampler: newSampler([]float64{1}, rand.NewSource(0)),
		clientPacks: []*clientPack{{
			client:  mockClient,
			healthy: true,
		}},
	}

	p := &Pool{
		innerPools: []*innerPool{inner},
		key:        newPrivateKey(t),
		cache:      cache,
	}

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(500 * time.Millisecond)
			cancel()
		}()

		err := p.WaitForContainerPresence(ctx, nil, DefaultPollingParams())
		require.Error(t, err)
		require.Contains(t, err.Error(), "context canceled")
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx := context.Background()
		err := p.WaitForContainerPresence(ctx, nil, &ContainerPollingParams{
			timeout:      500 * time.Millisecond,
			pollInterval: 5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		err := p.WaitForContainerPresence(ctx, nil, &ContainerPollingParams{
			timeout:      10 * time.Second,
			pollInterval: 500 * time.Millisecond,
		})
		require.NoError(t, err)
	})
}
