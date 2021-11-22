package pool

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
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
	Object
	Container
	Connection() (client.Client, *session.Token, error)
	OwnerID() *owner.ID
	WaitForContainerPresence(context.Context, *cid.ID, *ContainerPollingParams) error
	Close()
}

type Object interface {
	PutObject(ctx context.Context, params *client.PutObjectParams, opts ...CallOption) (*object.ID, error)
	DeleteObject(ctx context.Context, params *client.DeleteObjectParams, opts ...CallOption) error
	GetObject(ctx context.Context, params *client.GetObjectParams, opts ...CallOption) (*object.Object, error)
	GetObjectHeader(ctx context.Context, params *client.ObjectHeaderParams, opts ...CallOption) (*object.Object, error)
	ObjectPayloadRangeData(ctx context.Context, params *client.RangeDataParams, opts ...CallOption) ([]byte, error)
	ObjectPayloadRangeSHA256(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][32]byte, error)
	ObjectPayloadRangeTZ(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][64]byte, error)
	SearchObject(ctx context.Context, params *client.SearchObjectParams, opts ...CallOption) ([]*object.ID, error)
}

type Container interface {
	PutContainer(ctx context.Context, cnr *container.Container, opts ...CallOption) (*cid.ID, error)
	GetContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) (*container.Container, error)
	ListContainers(ctx context.Context, ownerID *owner.ID, opts ...CallOption) ([]*cid.ID, error)
	DeleteContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) error
	GetEACL(ctx context.Context, cid *cid.ID, opts ...CallOption) (*client.EACLWithSignature, error)
	SetEACL(ctx context.Context, table *eacl.Table, opts ...CallOption) error
	AnnounceContainerUsedSpace(ctx context.Context, announce []container.UsedSpaceAnnouncement, opts ...CallOption) error
}

type clientPack struct {
	client  client.Client
	healthy bool
	address string
}

type CallOption func(config *callConfig)

type callConfig struct {
	isRetry bool

	key    *ecdsa.PrivateKey
	btoken *token.BearerToken
	stoken *session.Token
}

func WithKey(key *ecdsa.PrivateKey) CallOption {
	return func(config *callConfig) {
		config.key = key
	}
}

func WithBearer(token *token.BearerToken) CallOption {
	return func(config *callConfig) {
		config.btoken = token
	}
}

func WithSession(token *session.Token) CallOption {
	return func(config *callConfig) {
		config.stoken = token
	}
}

func retry() CallOption {
	return func(config *callConfig) {
		config.isRetry = true
	}
}

func cfgFromOpts(opts ...CallOption) *callConfig {
	var cfg = new(callConfig)
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
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
	cache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("couldn't create cache: %w", err)
	}

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

func formCacheKey(address string, key *ecdsa.PrivateKey) string {
	buf := make([]byte, 32)
	key.D.FillBytes(buf)
	return address + hex.EncodeToString(buf)
}

func (p *pool) conn(ctx context.Context, cfg *callConfig) (*clientPack, []client.CallOption, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, nil, err
	}

	clientCallOptions := make([]client.CallOption, 0, 3)

	key := p.key
	if cfg.key != nil {
		key = cfg.key
	}
	clientCallOptions = append(clientCallOptions, client.WithKey(key))

	sessionToken := cfg.stoken
	if sessionToken == nil {
		cacheKey := formCacheKey(cp.address, key)
		sessionToken = p.cache.Get(cacheKey)
		if sessionToken == nil {
			sessionToken, err = cp.client.CreateSession(ctx, math.MaxUint32, clientCallOptions...)
			if err != nil {
				return nil, nil, err
			}
			_ = p.cache.Put(cacheKey, sessionToken)
		}
	}
	clientCallOptions = append(clientCallOptions, client.WithSession(sessionToken))

	if cfg.btoken != nil {
		clientCallOptions = append(clientCallOptions, client.WithBearer(cfg.btoken))
	}

	return cp, clientCallOptions, nil
}

func (p *pool) checkSessionTokenErr(err error, address string) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), "session token does not exist") {
		p.cache.DeleteByPrefix(address)
		return true
	}

	return false
}

func (p *pool) PutObject(ctx context.Context, params *client.PutObjectParams, opts ...CallOption) (*object.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutObject(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.PutObject(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) DeleteObject(ctx context.Context, params *client.DeleteObjectParams, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	err = cp.client.DeleteObject(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.DeleteObject(ctx, params, opts...)
	}
	return err
}

func (p *pool) GetObject(ctx context.Context, params *client.GetObjectParams, opts ...CallOption) (*object.Object, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetObject(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetObject(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) GetObjectHeader(ctx context.Context, params *client.ObjectHeaderParams, opts ...CallOption) (*object.Object, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetObjectHeader(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetObjectHeader(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) ObjectPayloadRangeData(ctx context.Context, params *client.RangeDataParams, opts ...CallOption) ([]byte, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeData(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeData(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) ObjectPayloadRangeSHA256(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][32]byte, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeSHA256(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeSHA256(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) ObjectPayloadRangeTZ(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][64]byte, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeTZ(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeTZ(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) SearchObject(ctx context.Context, params *client.SearchObjectParams, opts ...CallOption) ([]*object.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.SearchObject(ctx, params, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.SearchObject(ctx, params, opts...)
	}
	return res, err
}

func (p *pool) PutContainer(ctx context.Context, cnr *container.Container, opts ...CallOption) (*cid.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutContainer(ctx, cnr, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.PutContainer(ctx, cnr, opts...)
	}
	return res, err
}

func (p *pool) GetContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) (*container.Container, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetContainer(ctx, cid, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetContainer(ctx, cid, opts...)
	}
	return res, err
}

func (p *pool) ListContainers(ctx context.Context, ownerID *owner.ID, opts ...CallOption) ([]*cid.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ListContainers(ctx, ownerID, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ListContainers(ctx, ownerID, opts...)
	}
	return res, err
}

func (p *pool) DeleteContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	err = cp.client.DeleteContainer(ctx, cid, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.DeleteContainer(ctx, cid, opts...)
	}
	return err
}

func (p *pool) GetEACL(ctx context.Context, cid *cid.ID, opts ...CallOption) (*client.EACLWithSignature, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetEACL(ctx, cid, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetEACL(ctx, cid, opts...)
	}
	return res, err
}

func (p *pool) SetEACL(ctx context.Context, table *eacl.Table, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	err = cp.client.SetEACL(ctx, table, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.SetEACL(ctx, table, opts...)
	}
	return err
}

func (p *pool) AnnounceContainerUsedSpace(ctx context.Context, announce []container.UsedSpaceAnnouncement, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	err = cp.client.AnnounceContainerUsedSpace(ctx, announce, options...)
	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.AnnounceContainerUsedSpace(ctx, announce, opts...)
	}
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
