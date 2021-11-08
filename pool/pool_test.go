package pool

//go:generate mockgen -destination mock_test.go -package pool . Client

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuildPoolClientFailed(t *testing.T) {
	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		return nil, fmt.Errorf("error")
	}

	pb := new(Builder)
	pb.AddNode("peer0", 1)

	opts := &BuilderOptions{
		Key:           &key.PrivateKey,
		clientBuilder: clientBuilder,
	}

	_, err = pb.Build(context.TODO(), opts)
	require.Error(t, err)
}

func TestBuildPoolCreateSessionFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	ni := &netmap.NodeInfo{}
	ni.SetAddresses("addr1", "addr2")

	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error session")).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&client.EndpointInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	pb := new(Builder)
	pb.AddNode("peer0", 1)

	opts := &BuilderOptions{
		Key:           &key.PrivateKey,
		clientBuilder: clientBuilder,
	}

	_, err = pb.Build(context.TODO(), opts)
	require.Error(t, err)
}

func TestBuildPoolOneNodeFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	ni := &netmap.NodeInfo{}
	ni.SetAddresses("addr1", "addr2")

	var expectedToken *session.Token
	clientCount := -1
	clientBuilder := func(opts ...client.Option) (client.Client, error) {
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

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient2.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

		if clientCount == 0 {
			return mockClient, nil
		}
		return mockClient2, nil
	}

	pb := new(Builder)
	pb.AddNode("peer0", 9)
	pb.AddNode("peer1", 1)

	log, err := zap.NewProduction()
	require.NoError(t, err)
	opts := &BuilderOptions{
		Key:                     &key.PrivateKey,
		clientBuilder:           clientBuilder,
		ClientRebalanceInterval: 1000 * time.Millisecond,
		Logger:                  log,
	}

	clientPool, err := pb.Build(context.TODO(), opts)
	require.NoError(t, err)
	t.Cleanup(clientPool.Close)

	condition := func() bool {
		_, st, err := clientPool.Connection()
		return err == nil && st == expectedToken
	}
	require.Never(t, condition, 900*time.Millisecond, 100*time.Millisecond)
	require.Eventually(t, condition, 3*time.Second, 300*time.Millisecond)
}

func TestBuildPoolZeroNodes(t *testing.T) {
	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	opts := &BuilderOptions{
		Key: &key.PrivateKey,
	}
	_, err = pb.Build(context.TODO(), opts)
	require.Error(t, err)
}

func TestOneNode(t *testing.T) {
	ctrl := gomock.NewController(t)

	tok := session.NewToken()
	uid, err := uuid.New().MarshalBinary()
	require.NoError(t, err)
	tok.SetID(uid)

	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(tok, nil)
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&client.EndpointInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)

	opts := &BuilderOptions{
		Key:           &key.PrivateKey,
		clientBuilder: clientBuilder,
	}

	pool, err := pb.Build(context.Background(), opts)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, st, err := pool.Connection()
	require.NoError(t, err)
	require.Equal(t, tok, st)
}

func TestTwoNodes(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return tok, err
		})
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(&client.EndpointInfo{}, nil).AnyTimes()
		return mockClient, nil
	}

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)
	pb.AddNode("peer1", 1)

	opts := &BuilderOptions{
		Key:           &key.PrivateKey,
		clientBuilder: clientBuilder,
	}

	pool, err := pb.Build(context.Background(), opts)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, st, err := pool.Connection()
	require.NoError(t, err)
	require.Contains(t, tokens, st)
}

func TestOneOfTwoFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctrl2 := gomock.NewController(t)

	var tokens []*session.Token
	clientCount := -1
	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		clientCount++
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

		mockClient2 := NewMockClient(ctrl2)
		mockClient2.EXPECT().CreateSession(gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := newToken(t)
			tokens = append(tokens, tok)
			return tok, nil
		}).AnyTimes()
		mockClient2.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).DoAndReturn(func(_ interface{}, _ ...interface{}) (*client.EndpointInfo, error) {
			return nil, fmt.Errorf("error")
		}).AnyTimes()

		if clientCount == 0 {
			return mockClient, nil
		}
		return mockClient2, nil
	}

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)
	pb.AddNode("peer1", 9)

	opts := &BuilderOptions{
		Key:                     &key.PrivateKey,
		clientBuilder:           clientBuilder,
		ClientRebalanceInterval: 200 * time.Millisecond,
	}

	pool, err := pb.Build(context.Background(), opts)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	time.Sleep(2 * time.Second)

	for i := 0; i < 5; i++ {
		_, st, err := pool.Connection()
		require.NoError(t, err)
		require.Equal(t, tokens[0], st)
	}
}

func TestTwoFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error")).AnyTimes()
		return mockClient, nil
	}

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)
	pb.AddNode("peer1", 1)

	opts := &BuilderOptions{
		Key:                     &key.PrivateKey,
		clientBuilder:           clientBuilder,
		ClientRebalanceInterval: 200 * time.Millisecond,
	}

	pool, err := pb.Build(context.Background(), opts)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	time.Sleep(2 * time.Second)

	_, _, err = pool.Connection()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no healthy")
}

func TestSessionCache(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
			tok := session.NewToken()
			uid, err := uuid.New().MarshalBinary()
			require.NoError(t, err)
			tok.SetID(uid)
			tokens = append(tokens, tok)
			return tok, err
		}).MaxTimes(3)

		mockClient.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("session token does not exist"))
		mockClient.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("session token does not exist"))
		mockClient.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

		return mockClient, nil
	}

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)

	opts := &BuilderOptions{
		Key:                     &key.PrivateKey,
		clientBuilder:           clientBuilder,
		ClientRebalanceInterval: 30 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pb.Build(ctx, opts)
	require.NoError(t, err)

	// cache must contain session token
	_, st, err := pool.Connection()
	require.NoError(t, err)
	require.Contains(t, tokens, st)

	_, err = pool.GetObjectParam(ctx, nil, &CallParam{isRetry: true})
	require.Error(t, err)

	// cache must not contain session token
	_, st, err = pool.Connection()
	require.NoError(t, err)
	require.Nil(t, st)

	_, err = pool.PutObjectParam(ctx, nil, &CallParam{})
	require.NoError(t, err)

	// cache must contain session token
	_, st, err = pool.Connection()
	require.NoError(t, err)
	require.Contains(t, tokens, st)
}

func TestSessionCacheWithKey(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tokens []*session.Token
	clientBuilder := func(opts ...client.Option) (client.Client, error) {
		mockClient := NewMockClient(ctrl)
		mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_, _ interface{}, _ ...interface{}) (*session.Token, error) {
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

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)
	key2, err := keys.NewPrivateKey()
	require.NoError(t, err)

	pb := new(Builder)
	pb.AddNode("peer0", 1)

	opts := &BuilderOptions{
		Key:           &key.PrivateKey,
		clientBuilder: clientBuilder,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pb.Build(ctx, opts)
	require.NoError(t, err)

	// cache must contain session token
	_, st, err := pool.Connection()
	require.NoError(t, err)
	require.Contains(t, tokens, st)

	_, err = pool.GetObjectParam(ctx, nil, &CallParam{Key: &key2.PrivateKey})
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

func TestWaitPresence(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockClient := NewMockClient(ctrl)
	mockClient.EXPECT().CreateSession(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().EndpointInfo(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().GetContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	key, err := keys.NewPrivateKey()
	require.NoError(t, err)

	cache, err := NewCache()
	require.NoError(t, err)

	p := &pool{
		sampler: NewSampler([]float64{1}, rand.NewSource(0)),
		clientPacks: []*clientPack{{
			client:  mockClient,
			healthy: true,
		}},
		key:   &key.PrivateKey,
		cache: cache,
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
			CreationTimeout: 500 * time.Millisecond,
			PollInterval:    5 * time.Second,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("ok", func(t *testing.T) {
		ctx := context.Background()
		err := p.WaitForContainerPresence(ctx, nil, &ContainerPollingParams{
			CreationTimeout: 10 * time.Second,
			PollInterval:    500 * time.Millisecond,
		})
		require.NoError(t, err)
	})
}
