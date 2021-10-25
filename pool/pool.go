package pool

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"go.uber.org/zap"
)

// Client is a wrapper for client.Client to generate mock.
type Client interface {
	client.Client
}

// BuilderOptions contains options used to build connection pool.
type BuilderOptions struct {
	Key                     *ecdsa.PrivateKey
	Logger                  *zap.Logger
	NodeConnectionTimeout   time.Duration
	NodeRequestTimeout      time.Duration
	ClientRebalanceInterval time.Duration
	SessionExpirationEpoch  uint64
	weights                 []float64
	addresses               []string
	clientBuilder           func(opts ...client.Option) (client.Client, error)
}

// Builder is an interim structure used to collect node addresses/weights and
// build connection pool subsequently.
type Builder struct {
	addresses []string
	weights   []float64
}

// ContainerPollingParams contains parameters used in polling is a container created or not.
type ContainerPollingParams struct {
	CreationTimeout time.Duration
	PollInterval    time.Duration
}

// DefaultPollingParams creates ContainerPollingParams with default values.
func DefaultPollingParams() *ContainerPollingParams {
	return &ContainerPollingParams{
		CreationTimeout: 120 * time.Second,
		PollInterval:    5 * time.Second,
	}
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

	options.weights = adjustWeights(pb.weights)
	options.addresses = pb.addresses

	if options.clientBuilder == nil {
		options.clientBuilder = client.New
	}

	return newPool(ctx, options)
}

// Pool is an interface providing connection artifacts on request.
type Pool interface {
	client.Object
	client.Container
	Connection() (client.Client, *session.Token, error)
	OwnerID() *owner.ID
	WaitForContainerPresence(context.Context, *cid.ID, *ContainerPollingParams) error
	Close()

	PutObjectParam(ctx context.Context, params *client.PutObjectParams, callParam *CallParam) (*object.ID, error)
	DeleteObjectParam(ctx context.Context, params *client.DeleteObjectParams, callParam *CallParam) error
	GetObjectParam(ctx context.Context, params *client.GetObjectParams, callParam *CallParam) (*object.Object, error)
	GetObjectHeaderParam(ctx context.Context, params *client.ObjectHeaderParams, callParam *CallParam) (*object.Object, error)
	ObjectPayloadRangeDataParam(ctx context.Context, params *client.RangeDataParams, callParam *CallParam) ([]byte, error)
	ObjectPayloadRangeSHA256Param(ctx context.Context, params *client.RangeChecksumParams, callParam *CallParam) ([][32]byte, error)
	ObjectPayloadRangeTZParam(ctx context.Context, params *client.RangeChecksumParams, callParam *CallParam) ([][64]byte, error)
	SearchObjectParam(ctx context.Context, params *client.SearchObjectParams, callParam *CallParam) ([]*object.ID, error)
	PutContainerParam(ctx context.Context, cnr *container.Container, callParam *CallParam) (*cid.ID, error)
	GetContainerParam(ctx context.Context, cid *cid.ID, callParam *CallParam) (*container.Container, error)
	ListContainersParam(ctx context.Context, ownerID *owner.ID, callParam *CallParam) ([]*cid.ID, error)
	DeleteContainerParam(ctx context.Context, cid *cid.ID, callParam *CallParam) error
	GetEACLParam(ctx context.Context, cid *cid.ID, callParam *CallParam) (*client.EACLWithSignature, error)
	SetEACLParam(ctx context.Context, table *eacl.Table, callParam *CallParam) error
	AnnounceContainerUsedSpaceParam(ctx context.Context, announce []container.UsedSpaceAnnouncement, callParam *CallParam) error
}

type clientPack struct {
	client  client.Client
	healthy bool
	address string
}

type CallParam struct {
	Key *ecdsa.PrivateKey

	Options []client.CallOption
}

var _ Pool = (*pool)(nil)

type pool struct {
	lock        sync.RWMutex
	sampler     *Sampler
	key         *ecdsa.PrivateKey
	owner       *owner.ID
	clientPacks []*clientPack
	cancel      context.CancelFunc
	closedCh    chan struct{}
	cache       *SessionCache
}

func newPool(ctx context.Context, options *BuilderOptions) (Pool, error) {
	cache := NewCache()

	clientPacks := make([]*clientPack, len(options.weights))
	var atLeastOneHealthy bool
	for i, address := range options.addresses {
		c, err := options.clientBuilder(client.WithDefaultPrivateKey(options.Key),
			client.WithURIAddress(address, nil),
			client.WithDialTimeout(options.NodeConnectionTimeout))
		if err != nil {
			return nil, err
		}
		var healthy bool
		st, err := c.CreateSession(ctx, options.SessionExpirationEpoch)
		if err != nil && options.Logger != nil {
			options.Logger.Warn("failed to create neofs session token for client",
				zap.String("address", address),
				zap.Error(err))
		} else if err == nil {
			healthy, atLeastOneHealthy = true, true
			_ = cache.Put(formCacheKey(address, options.Key), st)
		}
		clientPacks[i] = &clientPack{client: c, healthy: healthy, address: address}
	}

	if !atLeastOneHealthy {
		return nil, fmt.Errorf("at least one node must be healthy")
	}

	source := rand.NewSource(time.Now().UnixNano())
	sampler := NewSampler(options.weights, source)
	wallet, err := owner.NEO3WalletFromPublicKey(&options.Key.PublicKey)
	if err != nil {
		return nil, err
	}
	ownerID := owner.NewIDFromNeo3Wallet(wallet)

	ctx, cancel := context.WithCancel(ctx)
	pool := &pool{
		sampler:     sampler,
		key:         options.Key,
		owner:       ownerID,
		clientPacks: clientPacks,
		cancel:      cancel,
		closedCh:    make(chan struct{}),
		cache:       cache,
	}
	go startRebalance(ctx, pool, options)
	return pool, nil
}

func startRebalance(ctx context.Context, p *pool, options *BuilderOptions) {
	ticker := time.NewTimer(options.ClientRebalanceInterval)
	buffer := make([]float64, len(options.weights))

	for {
		select {
		case <-ctx.Done():
			close(p.closedCh)
			return
		case <-ticker.C:
			updateNodesHealth(ctx, p, options, buffer)
			ticker.Reset(options.ClientRebalanceInterval)
		}
	}
}

func updateNodesHealth(ctx context.Context, p *pool, options *BuilderOptions, bufferWeights []float64) {
	if len(bufferWeights) != len(p.clientPacks) {
		bufferWeights = make([]float64, len(p.clientPacks))
	}
	healthyChanged := false
	wg := sync.WaitGroup{}
	for i, cPack := range p.clientPacks {
		wg.Add(1)

		go func(i int, client client.Client) {
			defer wg.Done()
			ok := true
			tctx, c := context.WithTimeout(ctx, options.NodeRequestTimeout)
			defer c()
			if _, err := client.EndpointInfo(tctx); err != nil {
				ok = false
				bufferWeights[i] = 0
			}
			p.lock.RLock()
			cp := *p.clientPacks[i]
			p.lock.RUnlock()

			if ok {
				bufferWeights[i] = options.weights[i]
				if !cp.healthy {
					if tkn, err := client.CreateSession(ctx, options.SessionExpirationEpoch); err != nil {
						ok = false
						bufferWeights[i] = 0
					} else {
						_ = p.cache.Put(formCacheKey(cp.address, p.key), tkn)
					}
				}
			} else {
				p.cache.DeleteByPrefix(cp.address)
			}

			p.lock.Lock()
			if p.clientPacks[i].healthy != ok {
				p.clientPacks[i].healthy = ok
				healthyChanged = true
			}
			p.lock.Unlock()
		}(i, cPack.client)
	}
	wg.Wait()

	if healthyChanged {
		probabilities := adjustWeights(bufferWeights)
		source := rand.NewSource(time.Now().UnixNano())
		p.lock.Lock()
		p.sampler = NewSampler(probabilities, source)
		p.lock.Unlock()
	}
}

func adjustWeights(weights []float64) []float64 {
	adjusted := make([]float64, len(weights))
	sum := 0.0
	for _, weight := range weights {
		sum += weight
	}
	if sum > 0 {
		for i, weight := range weights {
			adjusted[i] = weight / sum
		}
	}

	return adjusted
}

func (p *pool) Connection() (client.Client, *session.Token, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, nil, err
	}

	token := p.cache.Get(formCacheKey(cp.address, p.key))
	return cp.client, token, nil
}

func (p *pool) connection() (*clientPack, error) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if len(p.clientPacks) == 1 {
		cp := p.clientPacks[0]
		if cp.healthy {
			return cp, nil
		}
		return nil, errors.New("no healthy client")
	}
	attempts := 3 * len(p.clientPacks)
	for k := 0; k < attempts; k++ {
		i := p.sampler.Next()
		if cp := p.clientPacks[i]; cp.healthy {
			return cp, nil
		}
	}
	return nil, errors.New("no healthy client")
}

func (p *pool) OwnerID() *owner.ID {
	return p.owner
}

func (p *pool) conn(option []client.CallOption) (client.Client, []client.CallOption, error) {
	conn, token, err := p.Connection()
	if err != nil {
		return nil, nil, err
	}
	return conn, append([]client.CallOption{client.WithSession(token)}, option...), nil
}

func formCacheKey(address string, key *ecdsa.PrivateKey) string {
	k := keys.PrivateKey{PrivateKey: *key}
	return address + k.String()
}

func (p *pool) connParam(ctx context.Context, param *CallParam) (*clientPack, []client.CallOption, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, nil, err
	}

	key := p.key
	if param.Key != nil {
		key = param.Key
	}

	param.Options = append(param.Options, client.WithKey(key))
	cacheKey := formCacheKey(cp.address, key)
	token := p.cache.Get(cacheKey)
	if token == nil {
		token, err = cp.client.CreateSession(ctx, math.MaxUint32, param.Options...)
		if err != nil {
			return nil, nil, err
		}
		_ = p.cache.Put(cacheKey, token)
	}

	return cp, append([]client.CallOption{client.WithSession(token)}, param.Options...), nil
}

func (p *pool) PutObject(ctx context.Context, params *client.PutObjectParams, option ...client.CallOption) (*object.ID, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.PutObject(ctx, params, options...)
}

func (p *pool) DeleteObject(ctx context.Context, params *client.DeleteObjectParams, option ...client.CallOption) error {
	conn, options, err := p.conn(option)
	if err != nil {
		return err
	}
	return conn.DeleteObject(ctx, params, options...)
}

func (p *pool) GetObject(ctx context.Context, params *client.GetObjectParams, option ...client.CallOption) (*object.Object, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.GetObject(ctx, params, options...)
}

func (p *pool) GetObjectHeader(ctx context.Context, params *client.ObjectHeaderParams, option ...client.CallOption) (*object.Object, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.GetObjectHeader(ctx, params, options...)
}

func (p *pool) ObjectPayloadRangeData(ctx context.Context, params *client.RangeDataParams, option ...client.CallOption) ([]byte, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.ObjectPayloadRangeData(ctx, params, options...)
}

func (p *pool) ObjectPayloadRangeSHA256(ctx context.Context, params *client.RangeChecksumParams, option ...client.CallOption) ([][32]byte, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.ObjectPayloadRangeSHA256(ctx, params, options...)
}

func (p *pool) ObjectPayloadRangeTZ(ctx context.Context, params *client.RangeChecksumParams, option ...client.CallOption) ([][64]byte, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.ObjectPayloadRangeTZ(ctx, params, options...)
}

func (p *pool) SearchObject(ctx context.Context, params *client.SearchObjectParams, option ...client.CallOption) ([]*object.ID, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.SearchObject(ctx, params, options...)
}

func (p *pool) PutContainer(ctx context.Context, cnr *container.Container, option ...client.CallOption) (*cid.ID, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.PutContainer(ctx, cnr, options...)
}

func (p *pool) GetContainer(ctx context.Context, cid *cid.ID, option ...client.CallOption) (*container.Container, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.GetContainer(ctx, cid, options...)
}

func (p *pool) ListContainers(ctx context.Context, ownerID *owner.ID, option ...client.CallOption) ([]*cid.ID, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.ListContainers(ctx, ownerID, options...)
}

func (p *pool) DeleteContainer(ctx context.Context, cid *cid.ID, option ...client.CallOption) error {
	conn, options, err := p.conn(option)
	if err != nil {
		return err
	}
	return conn.DeleteContainer(ctx, cid, options...)
}

func (p *pool) GetEACL(ctx context.Context, cid *cid.ID, option ...client.CallOption) (*client.EACLWithSignature, error) {
	conn, options, err := p.conn(option)
	if err != nil {
		return nil, err
	}
	return conn.GetEACL(ctx, cid, options...)
}

func (p *pool) SetEACL(ctx context.Context, table *eacl.Table, option ...client.CallOption) error {
	conn, options, err := p.conn(option)
	if err != nil {
		return err
	}
	return conn.SetEACL(ctx, table, options...)
}

func (p *pool) AnnounceContainerUsedSpace(ctx context.Context, announce []container.UsedSpaceAnnouncement, option ...client.CallOption) error {
	conn, options, err := p.conn(option)
	if err != nil {
		return err
	}
	return conn.AnnounceContainerUsedSpace(ctx, announce, options...)
}

func (p *pool) checkSessionTokenErr(err error, address string) {
	if err == nil {
		return
	}

	if strings.Contains(err.Error(), "session token does not exist") {
		p.cache.DeleteByPrefix(address)
	}
}

func (p *pool) PutObjectParam(ctx context.Context, params *client.PutObjectParams, callParam *CallParam) (*object.ID, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutObject(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) DeleteObjectParam(ctx context.Context, params *client.DeleteObjectParams, callParam *CallParam) error {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return err
	}
	err = cp.client.DeleteObject(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return err
}

func (p *pool) GetObjectParam(ctx context.Context, params *client.GetObjectParams, callParam *CallParam) (*object.Object, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetObject(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) GetObjectHeaderParam(ctx context.Context, params *client.ObjectHeaderParams, callParam *CallParam) (*object.Object, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetObjectHeader(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) ObjectPayloadRangeDataParam(ctx context.Context, params *client.RangeDataParams, callParam *CallParam) ([]byte, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeData(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) ObjectPayloadRangeSHA256Param(ctx context.Context, params *client.RangeChecksumParams, callParam *CallParam) ([][32]byte, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeSHA256(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) ObjectPayloadRangeTZParam(ctx context.Context, params *client.RangeChecksumParams, callParam *CallParam) ([][64]byte, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeTZ(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) SearchObjectParam(ctx context.Context, params *client.SearchObjectParams, callParam *CallParam) ([]*object.ID, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.SearchObject(ctx, params, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) PutContainerParam(ctx context.Context, cnr *container.Container, callParam *CallParam) (*cid.ID, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutContainer(ctx, cnr, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) GetContainerParam(ctx context.Context, cid *cid.ID, callParam *CallParam) (*container.Container, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetContainer(ctx, cid, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) ListContainersParam(ctx context.Context, ownerID *owner.ID, callParam *CallParam) ([]*cid.ID, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ListContainers(ctx, ownerID, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) DeleteContainerParam(ctx context.Context, cid *cid.ID, callParam *CallParam) error {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return err
	}
	err = cp.client.DeleteContainer(ctx, cid, options...)
	p.checkSessionTokenErr(err, cp.address)
	return err
}

func (p *pool) GetEACLParam(ctx context.Context, cid *cid.ID, callParam *CallParam) (*client.EACLWithSignature, error) {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetEACL(ctx, cid, options...)
	p.checkSessionTokenErr(err, cp.address)
	return res, err
}

func (p *pool) SetEACLParam(ctx context.Context, table *eacl.Table, callParam *CallParam) error {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return err
	}
	err = cp.client.SetEACL(ctx, table, options...)
	p.checkSessionTokenErr(err, cp.address)
	return err
}

func (p *pool) AnnounceContainerUsedSpaceParam(ctx context.Context, announce []container.UsedSpaceAnnouncement, callParam *CallParam) error {
	cp, options, err := p.connParam(ctx, callParam)
	if err != nil {
		return err
	}
	err = cp.client.AnnounceContainerUsedSpace(ctx, announce, options...)
	p.checkSessionTokenErr(err, cp.address)
	return err
}

func (p *pool) WaitForContainerPresence(ctx context.Context, cid *cid.ID, pollParams *ContainerPollingParams) error {
	conn, _, err := p.Connection()
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, pollParams.CreationTimeout)
	defer cancel()
	ticker := time.NewTimer(pollParams.PollInterval)
	defer ticker.Stop()
	wdone := wctx.Done()
	done := ctx.Done()
	for {
		select {
		case <-done:
			return ctx.Err()
		case <-wdone:
			return wctx.Err()
		case <-ticker.C:
			_, err = conn.GetContainer(ctx, cid)
			if err == nil {
				return nil
			}
			ticker.Reset(pollParams.PollInterval)
		}
	}
}

// Cloce closes the pool and releases all the associated resources.
func (p *pool) Close() {
	p.cancel()
	<-p.closedCh
}
