package pool

//go:generate mockgen -destination mock_test.go -source pool.go -mock_names client=MockClient -package pool . client

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error session")).AnyTimes()
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(&netmap.NodeInfo{}, nil).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
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
	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	var expectedToken *session.Token
	clientCount := -1
	clientBuilder := func(_ string) (client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := newToken(t)
			return &resCreateSession{
				sessionKey: tok.SessionKey(),
				id:         tok.ID(),
			}, nil
		}).AnyTimes()

		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			expectedToken = newToken(t)
			return &resCreateSession{
				sessionKey: expectedToken.SessionKey(),
				id:         expectedToken.ID(),
			}, nil
		}).AnyTimes()
		mockClient2.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient2.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

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
		return areEqualTokens(st, expectedToken)
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
	ctrl := gomock.NewController(t)

	tok := session.NewToken()
	uid, err := uuid.New().MarshalBinary()
	require.NoError(t, err)
	tok.SetID(uid)

	tokRes := &resCreateSession{
		id:         tok.ID(),
		sessionKey: tok.SessionKey(),
	}

	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).Return(tokRes, nil)
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(&netmap.NodeInfo{}, nil).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
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
	require.True(t, areEqualTokens(tok, st))
}

func areEqualTokens(t1, t2 *session.Token) bool {
	if t1 == nil || t2 == nil {
		return false
	}
	return bytes.Equal(t1.ID(), t2.ID()) &&
		bytes.Equal(t1.SessionKey(), t2.SessionKey())
}

func TestTwoNodes(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, err
		})
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(&netmap.NodeInfo{}, nil).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
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
	require.True(t, containsTokens(tokens, st))
}

func containsTokens(list []*session.Token, item *session.Token) bool {
	for _, tok := range list {
		if areEqualTokens(tok, item) {
			return true
		}
	}
	return false
}

func TestOneOfTwoFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	var tokens []*session.Token
	clientCount := -1
	clientBuilder := func(_ string) (client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, nil
		}).AnyTimes()
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, nil
		}).AnyTimes()
		mockClient2.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ ...interface{}) (*netmap.NodeInfo, error) {
			return nil, fmt.Errorf("error")
		}).AnyTimes()
		mockClient2.EXPECT().networkInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ ...interface{}) (*netmap.NetworkInfo, error) {
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
		require.True(t, areEqualTokens(tokens[0], st))
	}
}

func TestTwoFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).Return(&resCreateSession{}, nil).AnyTimes()
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
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
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*resCreateSession, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, err
		}).MaxTimes(3)
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

		mockClient.EXPECT().objectGet(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("session token does not exist"))
		mockClient.EXPECT().objectPut(gomock.Any(), gomock.Any()).Return(nil, nil)

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
	require.True(t, containsTokens(tokens, st))

	var prm PrmObjectGet
	prm.SetAddress(address.Address{})
	prm.UseSession(*session.NewToken())

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
	require.True(t, containsTokens(tokens, st))
}

func TestPriority(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	tokens := make([]*session.Token, 2)
	clientBuilder := func(endpoint string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := newToken(t)
			tokens[0] = tok
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, nil
		}).AnyTimes()
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := newToken(t)
			tokens[1] = tok
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, nil
		}).AnyTimes()
		mockClient2.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(&netmap.NodeInfo{}, nil).AnyTimes()
		mockClient2.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()

		if endpoint == "peer0" {
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
		return areEqualTokens(st, tokens[0])
	}
	secondNode := func() bool {
		cp, err := pool.connection()
		require.NoError(t, err)
		st := pool.cache.Get(formCacheKey(cp.address, pool.key))
		return areEqualTokens(st, tokens[1])
	}
	require.Never(t, secondNode, time.Second, 200*time.Millisecond)

	require.Eventually(t, secondNode, time.Second, 200*time.Millisecond)
	require.Never(t, firstNode, time.Second, 200*time.Millisecond)
}

func TestSessionCacheWithKey(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(_ string) (client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}) (*resCreateSession, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return &resCreateSession{
				id:         tok.ID(),
				sessionKey: tok.SessionKey(),
			}, err
		}).MaxTimes(2)

		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
		mockClient.EXPECT().objectGet(gomock.Any(), gomock.Any()).Return(nil, nil)

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
	require.True(t, containsTokens(tokens, st))

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
		mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).Return(&resCreateSession{}, nil).AnyTimes()
		mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(&netmap.NodeInfo{}, nil).AnyTimes()
		mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
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

	var anonOwner user.ID
	user.IDFromKey(&anonOwner, anonKey.PublicKey)

	var prm prmCommon
	prm.UseKey(anonKey)
	var prmCtx prmContext
	prmCtx.useDefaultSession()

	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = func(session.Token) {}
	err = p.initCallContext(&cc, prm, prmCtx)
	require.NoError(t, err)

	err = p.openDefaultSession(&cc)
	require.NoError(t, err)

	tkn := p.cache.Get(formCacheKey("peer0", anonKey))
	require.True(t, anonOwner.Equals(*tkn.OwnerID()))
}

func TestWaitPresence(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockClient := NewMockClient(ctrl)
	mockClient.EXPECT().sessionCreate(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().endpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().networkInfo(gomock.Any(), gomock.Any()).Return(&netmap.NetworkInfo{}, nil).AnyTimes()
	mockClient.EXPECT().containerGet(gomock.Any(), gomock.Any()).Return(&container.Container{}, nil).AnyTimes()

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(500 * time.Millisecond)
			cancel()
		}()

		err := waitForContainerPresence(ctx, mockClient, nil, &WaitParams{
			timeout:      120 * time.Second,
			pollInterval: 5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context canceled")
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		ctx := context.Background()
		err := waitForContainerPresence(ctx, mockClient, nil, &WaitParams{
			timeout:      500 * time.Millisecond,
			pollInterval: 5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		err := waitForContainerPresence(ctx, mockClient, nil, &WaitParams{
			timeout:      10 * time.Second,
			pollInterval: 500 * time.Millisecond,
		})
		require.NoError(t, err)
	})
}

func TestCopySessionTokenWithoutSignatureAndContext(t *testing.T) {
	from := sessiontest.SignedToken()
	to := copySessionTokenWithoutSignatureAndContext(*from)

	require.Equal(t, from.Nbf(), to.Nbf())
	require.Equal(t, from.Exp(), to.Exp())
	require.Equal(t, from.Iat(), to.Iat())
	require.Equal(t, from.ID(), to.ID())
	require.Equal(t, from.OwnerID().String(), to.OwnerID().String())
	require.Equal(t, from.SessionKey(), to.SessionKey())

	require.False(t, to.VerifySignature())

	t.Run("empty object context", func(t *testing.T) {
		octx := sessiontest.ObjectContext()
		from.SetContext(octx)
		to = copySessionTokenWithoutSignatureAndContext(*from)
		require.Nil(t, to.Context())
	})

	t.Run("empty container context", func(t *testing.T) {
		cctx := sessiontest.ContainerContext()
		from.SetContext(cctx)
		to = copySessionTokenWithoutSignatureAndContext(*from)
		require.Nil(t, to.Context())
	})
}
