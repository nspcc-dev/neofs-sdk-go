package pool

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	apiclient "github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
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
	GetBalance(context.Context, *owner.ID, ...client.CallOption) (*client.BalanceOfRes, error)
	PutContainer(context.Context, *container.Container, ...client.CallOption) (*client.ContainerPutRes, error)
	GetContainer(context.Context, *cid.ID, ...client.CallOption) (*client.ContainerGetRes, error)
	ListContainers(context.Context, *owner.ID, ...client.CallOption) (*client.ContainerListRes, error)
	DeleteContainer(context.Context, *cid.ID, ...client.CallOption) (*client.ContainerDeleteRes, error)
	EACL(context.Context, *cid.ID, ...client.CallOption) (*client.EACLRes, error)
	SetEACL(context.Context, *eacl.Table, ...client.CallOption) (*client.SetEACLRes, error)
	AnnounceContainerUsedSpace(context.Context, []container.UsedSpaceAnnouncement, ...client.CallOption) (*client.AnnounceSpaceRes, error)
	EndpointInfo(context.Context, ...client.CallOption) (*client.EndpointInfoRes, error)
	NetworkInfo(context.Context, ...client.CallOption) (*client.NetworkInfoRes, error)
	PutObject(context.Context, *client.PutObjectParams, ...client.CallOption) (*client.ObjectPutRes, error)
	DeleteObject(context.Context, *client.DeleteObjectParams, ...client.CallOption) (*client.ObjectDeleteRes, error)
	GetObject(context.Context, *client.GetObjectParams, ...client.CallOption) (*client.ObjectGetRes, error)
	HeadObject(context.Context, *client.ObjectHeaderParams, ...client.CallOption) (*client.ObjectHeadRes, error)
	ObjectPayloadRangeData(context.Context, *client.RangeDataParams, ...client.CallOption) (*client.ObjectRangeRes, error)
	HashObjectPayloadRanges(context.Context, *client.RangeChecksumParams, ...client.CallOption) (*client.ObjectRangeHashRes, error)
	SearchObjects(context.Context, *client.SearchObjectParams, ...client.CallOption) (*client.ObjectSearchRes, error)
	AnnounceLocalTrust(context.Context, client.AnnounceLocalTrustPrm, ...client.CallOption) (*client.AnnounceLocalTrustRes, error)
	AnnounceIntermediateTrust(context.Context, client.AnnounceIntermediateTrustPrm, ...client.CallOption) (*client.AnnounceIntermediateTrustRes, error)
	CreateSession(context.Context, uint64, ...client.CallOption) (*client.CreateSessionRes, error)

	Raw() *apiclient.Client

	Conn() io.Closer
}

// BuilderOptions contains options used to build connection pool.
type BuilderOptions struct {
	Key                     *ecdsa.PrivateKey
	Logger                  *zap.Logger
	NodeConnectionTimeout   time.Duration
	NodeRequestTimeout      time.Duration
	ClientRebalanceInterval time.Duration
	SessionExpirationEpoch  uint64
	nodesParams             []*NodesParam
	clientBuilder           func(opts ...client.Option) (Client, error)
}

type NodesParam struct {
	priority  int
	addresses []string
	weights   []float64
}

type NodeParam struct {
	priority int
	address  string
	weight   float64
}

// Builder is an interim structure used to collect node addresses/weights and
// build connection pool subsequently.
type Builder struct {
	nodeParams []NodeParam
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
func (pb *Builder) AddNode(address string, priority int, weight float64) *Builder {
	pb.nodeParams = append(pb.nodeParams, NodeParam{
		address:  address,
		priority: priority,
		weight:   weight,
	})
	return pb
}

// Build creates new pool based on current PoolBuilder state and options.
func (pb *Builder) Build(ctx context.Context, options *BuilderOptions) (Pool, error) {
	if len(pb.nodeParams) == 0 {
		return nil, errors.New("no NeoFS peers configured")
	}

	nodesParams := make(map[int]*NodesParam)
	for _, param := range pb.nodeParams {
		nodes, ok := nodesParams[param.priority]
		if !ok {
			nodes = &NodesParam{priority: param.priority}
		}
		nodes.addresses = append(nodes.addresses, param.address)
		nodes.weights = append(nodes.weights, param.weight)
		nodesParams[param.priority] = nodes
	}

	for _, nodes := range nodesParams {
		nodes.weights = adjustWeights(nodes.weights)
		options.nodesParams = append(options.nodesParams, nodes)
	}

	sort.Slice(options.nodesParams, func(i, j int) bool {
		return options.nodesParams[i].priority < options.nodesParams[j].priority
	})

	if options.clientBuilder == nil {
		options.clientBuilder = func(opts ...client.Option) (Client, error) {
			return client.New(opts...)
		}
	}

	return newPool(ctx, options)
}

// Pool is an interface providing connection artifacts on request.
type Pool interface {
	Object
	Container
	Accounting
	Connection() (Client, *session.Token, error)
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
	GetEACL(ctx context.Context, cid *cid.ID, opts ...CallOption) (*eacl.Table, error)
	SetEACL(ctx context.Context, table *eacl.Table, opts ...CallOption) error
	AnnounceContainerUsedSpace(ctx context.Context, announce []container.UsedSpaceAnnouncement, opts ...CallOption) error
}

type Accounting interface {
	Balance(ctx context.Context, owner *owner.ID, opts ...CallOption) (*accounting.Decimal, error)
}

type clientPack struct {
	client  Client
	healthy bool
	address string
}

type CallOption func(config *callConfig)

type callConfig struct {
	isRetry           bool
	useDefaultSession bool

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

func useDefaultSession() CallOption {
	return func(config *callConfig) {
		config.useDefaultSession = true
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
	innerPools []*innerPool
	key        *ecdsa.PrivateKey
	owner      *owner.ID
	cancel     context.CancelFunc
	closedCh   chan struct{}
	cache      *SessionCache
}

type innerPool struct {
	lock        sync.RWMutex
	sampler     *Sampler
	clientPacks []*clientPack
}

func newPool(ctx context.Context, options *BuilderOptions) (Pool, error) {
	wallet, err := owner.NEO3WalletFromPublicKey(&options.Key.PublicKey)
	if err != nil {
		return nil, err
	}

	cache, err := NewCache()
	if err != nil {
		return nil, fmt.Errorf("couldn't create cache: %w", err)
	}

	ownerID := owner.NewIDFromNeo3Wallet(wallet)

	inner := make([]*innerPool, len(options.nodesParams))
	var atLeastOneHealthy bool
	for i, params := range options.nodesParams {
		clientPacks := make([]*clientPack, len(params.weights))
		for j, address := range params.addresses {
			c, err := options.clientBuilder(client.WithDefaultPrivateKey(options.Key),
				client.WithURIAddress(address, nil),
				client.WithDialTimeout(options.NodeConnectionTimeout))
			if err != nil {
				return nil, err
			}
			var healthy bool
			cliRes, err := c.CreateSession(ctx, options.SessionExpirationEpoch)
			if err != nil && options.Logger != nil {
				options.Logger.Warn("failed to create neofs session token for client",
					zap.String("address", address),
					zap.Error(err))
			} else if err == nil {
				healthy, atLeastOneHealthy = true, true
				st := sessionTokenForOwner(ownerID, cliRes)
				_ = cache.Put(formCacheKey(address, options.Key), st)
			}
			clientPacks[j] = &clientPack{client: c, healthy: healthy, address: address}
		}
		source := rand.NewSource(time.Now().UnixNano())
		sampler := NewSampler(params.weights, source)

		inner[i] = &innerPool{
			sampler:     sampler,
			clientPacks: clientPacks,
		}
	}

	if !atLeastOneHealthy {
		return nil, fmt.Errorf("at least one node must be healthy")
	}

	ctx, cancel := context.WithCancel(ctx)
	pool := &pool{
		innerPools: inner,
		key:        options.Key,
		owner:      ownerID,
		cancel:     cancel,
		closedCh:   make(chan struct{}),
		cache:      cache,
	}
	go startRebalance(ctx, pool, options)
	return pool, nil
}

func startRebalance(ctx context.Context, p *pool, options *BuilderOptions) {
	ticker := time.NewTimer(options.ClientRebalanceInterval)
	buffers := make([][]float64, len(options.nodesParams))
	for i, params := range options.nodesParams {
		buffers[i] = make([]float64, len(params.weights))
	}

	for {
		select {
		case <-ctx.Done():
			close(p.closedCh)
			return
		case <-ticker.C:
			updateNodesHealth(ctx, p, options, buffers)
			ticker.Reset(options.ClientRebalanceInterval)
		}
	}
}

func updateNodesHealth(ctx context.Context, p *pool, options *BuilderOptions, buffers [][]float64) {
	wg := sync.WaitGroup{}
	for i, inner := range p.innerPools {
		wg.Add(1)

		bufferWeights := buffers[i]
		go func(i int, innerPool *innerPool) {
			defer wg.Done()
			updateInnerNodesHealth(ctx, p, i, options, bufferWeights)
		}(i, inner)
	}
	wg.Wait()
}

func updateInnerNodesHealth(ctx context.Context, pool *pool, i int, options *BuilderOptions, bufferWeights []float64) {
	if i > len(pool.innerPools)-1 {
		return
	}
	p := pool.innerPools[i]

	healthyChanged := false
	wg := sync.WaitGroup{}
	for j, cPack := range p.clientPacks {
		wg.Add(1)
		go func(j int, client Client) {
			defer wg.Done()
			ok := true
			tctx, c := context.WithTimeout(ctx, options.NodeRequestTimeout)
			defer c()
			if _, err := client.EndpointInfo(tctx); err != nil {
				ok = false
				bufferWeights[j] = 0
			}
			p.lock.RLock()
			cp := *p.clientPacks[j]
			p.lock.RUnlock()

			if ok {
				bufferWeights[j] = options.nodesParams[i].weights[j]
				if !cp.healthy {
					if cliRes, err := client.CreateSession(ctx, options.SessionExpirationEpoch); err != nil {
						ok = false
						bufferWeights[j] = 0
					} else {
						tkn := pool.newSessionToken(cliRes)

						_ = pool.cache.Put(formCacheKey(cp.address, pool.key), tkn)
					}
				}
			} else {
				pool.cache.DeleteByPrefix(cp.address)
			}

			p.lock.Lock()
			if p.clientPacks[j].healthy != ok {
				p.clientPacks[j].healthy = ok
				healthyChanged = true
			}
			p.lock.Unlock()
		}(j, cPack.client)
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

func (p *pool) Connection() (Client, *session.Token, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, nil, err
	}

	token := p.cache.Get(formCacheKey(cp.address, p.key))
	return cp.client, token, nil
}

func (p *pool) connection() (*clientPack, error) {
	for _, inner := range p.innerPools {
		cp, err := inner.connection()
		if err == nil {
			return cp, nil
		}
	}

	return nil, errors.New("no healthy client")
}

func (p *innerPool) connection() (*clientPack, error) {
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
	k := keys.PrivateKey{PrivateKey: *key}
	return address + k.String()
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
	if sessionToken == nil && cfg.useDefaultSession {
		cacheKey := formCacheKey(cp.address, key)
		sessionToken = p.cache.Get(cacheKey)
		if sessionToken == nil {
			cliRes, err := cp.client.CreateSession(ctx, math.MaxUint32, clientCallOptions...)
			if err != nil {
				return nil, nil, err
			}

			sessionToken = p.newSessionToken(cliRes)

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
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutObject(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.PutObject(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.ID(), nil
}

func (p *pool) DeleteObject(ctx context.Context, params *client.DeleteObjectParams, opts ...CallOption) error {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}

	res, err := cp.client.DeleteObject(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.DeleteObject(ctx, params, opts...)
	}

	// here err already carries both status and client errors

	return err
}

func (p *pool) GetObject(ctx context.Context, params *client.GetObjectParams, opts ...CallOption) (*object.Object, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetObject(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetObject(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Object(), nil
}

func (p *pool) GetObjectHeader(ctx context.Context, params *client.ObjectHeaderParams, opts ...CallOption) (*object.Object, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.HeadObject(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetObjectHeader(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Object(), nil
}

func (p *pool) ObjectPayloadRangeData(ctx context.Context, params *client.RangeDataParams, opts ...CallOption) ([]byte, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ObjectPayloadRangeData(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeData(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Data(), nil
}

func copyRangeChecksumParams(prm *client.RangeChecksumParams) *client.RangeChecksumParams {
	var prmCopy client.RangeChecksumParams

	prmCopy.WithAddress(prm.Address())
	prmCopy.WithSalt(prm.Salt())
	prmCopy.WithRangeList(prm.RangeList()...)

	return &prmCopy
}

func (p *pool) ObjectPayloadRangeSHA256(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][32]byte, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// FIXME: pretty bad approach but we should not mutate params through the pointer
	//  If non-SHA256 algo is set then we need to reset it.
	params = copyRangeChecksumParams(params)
	// SHA256 by default, no need to do smth

	res, err := cp.client.HashObjectPayloadRanges(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeSHA256(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	cliHashes := res.Hashes()

	hs := make([][sha256.Size]byte, len(cliHashes))

	for i := range cliHashes {
		if ln := len(cliHashes[i]); ln != sha256.Size {
			return nil, fmt.Errorf("invalid SHA256 checksum size %d", ln)
		}

		copy(hs[i][:], cliHashes[i])
	}

	return hs, nil
}

func (p *pool) ObjectPayloadRangeTZ(ctx context.Context, params *client.RangeChecksumParams, opts ...CallOption) ([][64]byte, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// FIXME: pretty bad approach but we should not mutate params through the pointer
	//  We need to set Tillich-Zemor algo.
	params = copyRangeChecksumParams(params)
	params.TZ()

	res, err := cp.client.HashObjectPayloadRanges(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ObjectPayloadRangeTZ(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	cliHashes := res.Hashes()

	hs := make([][client.TZSize]byte, len(cliHashes))

	for i := range cliHashes {
		if ln := len(cliHashes[i]); ln != client.TZSize {
			return nil, fmt.Errorf("invalid TZ checksum size %d", ln)
		}

		copy(hs[i][:], cliHashes[i])
	}

	return hs, nil
}

func (p *pool) SearchObject(ctx context.Context, params *client.SearchObjectParams, opts ...CallOption) ([]*object.ID, error) {
	cfg := cfgFromOpts(append(opts, useDefaultSession())...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.SearchObjects(ctx, params, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.SearchObject(ctx, params, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.IDList(), nil
}

func (p *pool) PutContainer(ctx context.Context, cnr *container.Container, opts ...CallOption) (*cid.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.PutContainer(ctx, cnr, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.PutContainer(ctx, cnr, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.ID(), nil
}

func (p *pool) GetContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) (*container.Container, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.GetContainer(ctx, cid, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetContainer(ctx, cid, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Container(), nil
}

func (p *pool) ListContainers(ctx context.Context, ownerID *owner.ID, opts ...CallOption) ([]*cid.ID, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.ListContainers(ctx, ownerID, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.ListContainers(ctx, ownerID, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.IDList(), nil
}

func (p *pool) DeleteContainer(ctx context.Context, cid *cid.ID, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	res, err := cp.client.DeleteContainer(ctx, cid, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.DeleteContainer(ctx, cid, opts...)
	}

	// here err already carries both status and client errors

	return err
}

func (p *pool) GetEACL(ctx context.Context, cid *cid.ID, opts ...CallOption) (*eacl.Table, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}
	res, err := cp.client.EACL(ctx, cid, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.GetEACL(ctx, cid, opts...)
	}

	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Table(), nil
}

func (p *pool) SetEACL(ctx context.Context, table *eacl.Table, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	res, err := cp.client.SetEACL(ctx, table, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.SetEACL(ctx, table, opts...)
	}

	// here err already carries both status and client errors

	return err
}

func (p *pool) AnnounceContainerUsedSpace(ctx context.Context, announce []container.UsedSpaceAnnouncement, opts ...CallOption) error {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return err
	}
	res, err := cp.client.AnnounceContainerUsedSpace(ctx, announce, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if p.checkSessionTokenErr(err, cp.address) && !cfg.isRetry {
		opts = append(opts, retry())
		return p.AnnounceContainerUsedSpace(ctx, announce, opts...)
	}

	// here err already carries both status and client errors

	return err
}

func (p *pool) Balance(ctx context.Context, o *owner.ID, opts ...CallOption) (*accounting.Decimal, error) {
	cfg := cfgFromOpts(opts...)
	cp, options, err := p.conn(ctx, cfg)
	if err != nil {
		return nil, err
	}

	res, err := cp.client.GetBalance(ctx, o, options...)
	if err == nil {
		// reflect status failures in err
		err = apistatus.ErrFromStatus(res.Status())
	}

	if err != nil {
		return nil, err
	}

	return res.Amount(), nil
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

// Close closes the pool and releases all the associated resources.
func (p *pool) Close() {
	p.cancel()
	<-p.closedCh
}

// creates new session token from CreateSession call result.
func (p *pool) newSessionToken(cliRes *client.CreateSessionRes) *session.Token {
	return sessionTokenForOwner(p.owner, cliRes)
}

// creates new session token with specified owner from CreateSession call result.
func sessionTokenForOwner(id *owner.ID, cliRes *client.CreateSessionRes) *session.Token {
	st := session.NewToken()
	st.SetOwnerID(id)
	st.SetID(cliRes.ID())
	st.SetSessionKey(cliRes.SessionKey())

	return st
}
