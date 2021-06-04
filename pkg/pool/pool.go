package pool

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nspcc-dev/neofs-api-go/pkg/client"
	"github.com/nspcc-dev/neofs-api-go/pkg/owner"
	"github.com/nspcc-dev/neofs-api-go/pkg/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// BuilderOptions contains options used to build connection pool.
type BuilderOptions struct {
	Key                     *ecdsa.PrivateKey
	NodeConnectionTimeout   time.Duration
	NodeRequestTimeout      time.Duration
	ClientRebalanceInterval time.Duration
	KeepaliveTime           time.Duration
	KeepaliveTimeout        time.Duration
	KeepalivePermitWoStream bool
	SessionExpirationEpoch  uint64
	weights                 []float64
	connections             []*grpc.ClientConn
}

// Builder is an interim structure used to collect node addresses/weights and
// build connection pool subsequently.
type Builder struct {
	addresses []string
	weights   []float64
}

// AddNode adds address/weight pair to node PoolBuilder list.
func (pb *Builder) AddNode(address string, weight float64) *Builder {
	pb.addresses = append(pb.addresses, address)
	pb.weights = append(pb.weights, weight)
	return pb
}

// Build creates new pool based on current PoolBuilder state and options.
func (pb *Builder) Build(ctx context.Context, options *BuilderOptions) (Pool, error) {
	if len(pb.addresses) == 0 {
		return nil, errors.New("no NeoFS peers configured")
	}
	totalWeight := 0.0
	for _, w := range pb.weights {
		totalWeight += w
	}
	for i, w := range pb.weights {
		pb.weights[i] = w / totalWeight
	}
	var cons = make([]*grpc.ClientConn, len(pb.addresses))
	for i, address := range pb.addresses {
		con, err := func() (*grpc.ClientConn, error) {
			toctx, c := context.WithTimeout(ctx, options.NodeConnectionTimeout)
			defer c()
			return grpc.DialContext(toctx, address,
				grpc.WithInsecure(),
				grpc.WithBlock(),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                options.KeepaliveTime,
					Timeout:             options.KeepaliveTimeout,
					PermitWithoutStream: options.KeepalivePermitWoStream,
				}),
			)
		}()
		if err != nil {
			return nil, err
		}
		cons[i] = con
	}
	options.weights = pb.weights
	options.connections = cons
	return newPool(ctx, options)
}

// Pool is an interface providing connection artifacts on request.
type Pool interface {
	Connection() (client.Client, *session.Token, error)
	OwnerID() *owner.ID
}

type clientPack struct {
	client       client.Client
	sessionToken *session.Token
	healthy      bool
}

type pool struct {
	lock        sync.RWMutex
	sampler     *Sampler
	owner       *owner.ID
	clientPacks []*clientPack
}

func newPool(ctx context.Context, options *BuilderOptions) (Pool, error) {
	clientPacks := make([]*clientPack, len(options.weights))
	for i, con := range options.connections {
		c, err := client.New(client.WithDefaultPrivateKey(options.Key), client.WithGRPCConnection(con))
		if err != nil {
			return nil, err
		}
		st, err := c.CreateSession(ctx, options.SessionExpirationEpoch)
		if err != nil {
			address := "unknown"
			if epi, err := c.EndpointInfo(ctx); err == nil {
				address = epi.NodeInfo().Address()
			}
			return nil, fmt.Errorf("failed to create neofs session token for client %s: %w", address, err)
		}
		clientPacks[i] = &clientPack{client: c, sessionToken: st, healthy: true}
	}
	source := rand.NewSource(time.Now().UnixNano())
	sampler := NewSampler(options.weights, source)
	wallet, err := owner.NEO3WalletFromPublicKey(&options.Key.PublicKey)
	if err != nil {
		return nil, err
	}
	ownerID := owner.NewIDFromNeo3Wallet(wallet)

	pool := &pool{sampler: sampler, owner: ownerID, clientPacks: clientPacks}
	go func() {
		ticker := time.NewTimer(options.ClientRebalanceInterval)
		for range ticker.C {
			ok := true
			for i, clientPack := range pool.clientPacks {
				func() {
					tctx, c := context.WithTimeout(ctx, options.NodeRequestTimeout)
					defer c()
					if _, err := clientPack.client.EndpointInfo(tctx); err != nil {
						ok = false
					}
					pool.lock.Lock()
					pool.clientPacks[i].healthy = ok
					pool.lock.Unlock()
				}()
			}
			ticker.Reset(options.ClientRebalanceInterval)
		}
	}()
	return pool, nil
}

func (p *pool) Connection() (client.Client, *session.Token, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if len(p.clientPacks) == 1 {
		cp := p.clientPacks[0]
		if cp.healthy {
			return cp.client, cp.sessionToken, nil
		}
		return nil, nil, errors.New("no healthy client")
	}
	attempts := 3 * len(p.clientPacks)
	for k := 0; k < attempts; k++ {
		i := p.sampler.Next()
		if cp := p.clientPacks[i]; cp.healthy {
			return cp.client, cp.sessionToken, nil
		}
	}
	return nil, nil, errors.New("no healthy client")
}

func (p *pool) OwnerID() *owner.ID {
	return p.owner
}
