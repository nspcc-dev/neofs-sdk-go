package pool

import (
	"bytes"
	"context"
	"crypto/ecdsa"
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
	sessionv2 "github.com/nspcc-dev/neofs-api-go/v2/session"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/token"
	"go.uber.org/zap"
)

// Client is a wrapper for client.Client to generate mock.
type Client interface {
	BalanceGet(context.Context, client.PrmBalanceGet) (*client.ResBalanceGet, error)
	ContainerPut(context.Context, client.PrmContainerPut) (*client.ResContainerPut, error)
	ContainerGet(context.Context, client.PrmContainerGet) (*client.ResContainerGet, error)
	ContainerList(context.Context, client.PrmContainerList) (*client.ResContainerList, error)
	ContainerDelete(context.Context, client.PrmContainerDelete) (*client.ResContainerDelete, error)
	ContainerEACL(context.Context, client.PrmContainerEACL) (*client.ResContainerEACL, error)
	ContainerSetEACL(context.Context, client.PrmContainerSetEACL) (*client.ResContainerSetEACL, error)
	EndpointInfo(context.Context, client.PrmEndpointInfo) (*client.ResEndpointInfo, error)
	NetworkInfo(context.Context, client.PrmNetworkInfo) (*client.ResNetworkInfo, error)
	ObjectPutInit(context.Context, client.PrmObjectPutInit) (*client.ObjectWriter, error)
	ObjectDelete(context.Context, client.PrmObjectDelete) (*client.ResObjectDelete, error)
	ObjectGetInit(context.Context, client.PrmObjectGet) (*client.ObjectReader, error)
	ObjectHead(context.Context, client.PrmObjectHead) (*client.ResObjectHead, error)
	ObjectRangeInit(context.Context, client.PrmObjectRange) (*client.ObjectRangeReader, error)
	ObjectSearchInit(context.Context, client.PrmObjectSearch) (*client.ObjectListReader, error)
	SessionCreate(context.Context, client.PrmSessionCreate) (*client.ResSessionCreate, error)
}

// InitParameters contains options used to create connection Pool.
type InitParameters struct {
	Key                       *ecdsa.PrivateKey
	Logger                    *zap.Logger
	NodeConnectionTimeout     time.Duration
	NodeRequestTimeout        time.Duration
	ClientRebalanceInterval   time.Duration
	SessionTokenThreshold     time.Duration
	SessionExpirationDuration uint64
	NodeParams                []NodeParam

	clientBuilder func(endpoint string) (Client, error)
}

type rebalanceParameters struct {
	nodesParams               []*nodesParam
	nodeRequestTimeout        time.Duration
	clientRebalanceInterval   time.Duration
	sessionExpirationDuration uint64
}

type nodesParam struct {
	priority  int
	addresses []string
	weights   []float64
}

// NodeParam groups parameters of remote node.
type NodeParam struct {
	Priority int
	Address  string
	Weight   float64
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

type clientPack struct {
	client  Client
	healthy bool
	address string
}

type prmCommon struct {
	defaultSession bool
	verb           sessionv2.ObjectSessionVerb
	addr           *address.Address

	key    *ecdsa.PrivateKey
	btoken *token.BearerToken
	stoken *session.Token
}

func (x *prmCommon) useDefaultSession() {
	x.defaultSession = true
}

func (x *prmCommon) useAddress(addr *address.Address) {
	x.addr = addr
}

func (x *prmCommon) useVerb(verb sessionv2.ObjectSessionVerb) {
	x.verb = verb
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Pool default key is used.
func (x *prmCommon) UseKey(key *ecdsa.PrivateKey) {
	x.key = key
}

// UseBearer attaches bearer token to be used for the operation.
func (x *prmCommon) UseBearer(token *token.BearerToken) {
	x.btoken = token
}

// UseSession specifies session within which object should be read.
func (x *prmCommon) UseSession(token *session.Token) {
	x.stoken = token
}

// PrmObjectPut groups parameters of PutObject operation.
type PrmObjectPut struct {
	prmCommon

	hdr object.Object

	payload io.Reader
}

// SetHeader specifies header of the object.
func (x *PrmObjectPut) SetHeader(hdr object.Object) {
	x.hdr = hdr
}

// SetPayload specifies payload of the object.
func (x *PrmObjectPut) SetPayload(payload io.Reader) {
	x.payload = payload
}

// PrmObjectDelete groups parameters of DeleteObject operation.
type PrmObjectDelete struct {
	prmCommon

	addr address.Address
}

// SetAddress specifies NeoFS address of the object.
func (x *PrmObjectDelete) SetAddress(addr address.Address) {
	x.addr = addr
}

// PrmObjectGet groups parameters of GetObject operation.
type PrmObjectGet struct {
	prmCommon

	addr address.Address
}

// SetAddress specifies NeoFS address of the object.
func (x *PrmObjectGet) SetAddress(addr address.Address) {
	x.addr = addr
}

// PrmObjectHead groups parameters of HeadObject operation.
type PrmObjectHead struct {
	prmCommon

	addr address.Address
}

// SetAddress specifies NeoFS address of the object.
func (x *PrmObjectHead) SetAddress(addr address.Address) {
	x.addr = addr
}

// PrmObjectRange groups parameters of RangeObject operation.
type PrmObjectRange struct {
	prmCommon

	addr    address.Address
	off, ln uint64
}

// SetAddress specifies NeoFS address of the object.
func (x *PrmObjectRange) SetAddress(addr address.Address) {
	x.addr = addr
}

// SetOffset sets offset of the payload range to be read.
func (x *PrmObjectRange) SetOffset(offset uint64) {
	x.off = offset
}

// SetLength sets length of the payload range to be read.
func (x *PrmObjectRange) SetLength(length uint64) {
	x.ln = length
}

// PrmObjectSearch groups parameters of SearchObjects operation.
type PrmObjectSearch struct {
	prmCommon

	cnrID   cid.ID
	filters object.SearchFilters
}

// SetContainerID specifies the container in which to look for objects.
func (x *PrmObjectSearch) SetContainerID(cnrID cid.ID) {
	x.cnrID = cnrID
}

// SetFilters specifies filters by which to select objects.
func (x *PrmObjectSearch) SetFilters(filters object.SearchFilters) {
	x.filters = filters
}

// PrmContainerPut groups parameters of PutContainer operation.
type PrmContainerPut struct {
	prmCommon

	cnr *container.Container
}

// SetContainer specifies structured information about new NeoFS container.
func (x *PrmContainerPut) SetContainer(cnr *container.Container) {
	x.cnr = cnr
}

// PrmContainerGet groups parameters of GetContainer operation.
type PrmContainerGet struct {
	prmCommon

	cnrID *cid.ID
}

// SetContainerID specifies identifier of the container to be read.
func (x *PrmContainerGet) SetContainerID(cnrID *cid.ID) {
	x.cnrID = cnrID
}

// PrmContainerList groups parameters of ListContainers operation.
type PrmContainerList struct {
	prmCommon

	ownerID *owner.ID
}

// SetOwnerID specifies identifier of the NeoFS account to list the containers.
func (x *PrmContainerList) SetOwnerID(ownerID *owner.ID) {
	x.ownerID = ownerID
}

// PrmContainerDelete groups parameters of DeleteContainer operation.
type PrmContainerDelete struct {
	prmCommon

	cnrID *cid.ID
}

// SetContainerID specifies identifier of the NeoFS container to be removed.
func (x *PrmContainerDelete) SetContainerID(cnrID *cid.ID) {
	x.cnrID = cnrID
}

// PrmContainerEACL groups parameters of GetEACL operation.
type PrmContainerEACL struct {
	prmCommon

	cnrID *cid.ID
}

// SetContainerID specifies identifier of the NeoFS container to read the eACL table.
func (x *PrmContainerEACL) SetContainerID(cnrID *cid.ID) {
	x.cnrID = cnrID
}

// PrmContainerSetEACL groups parameters of SetEACL operation.
type PrmContainerSetEACL struct {
	prmCommon

	table *eacl.Table
}

// SetTable specifies eACL table structure to be set for the container.
func (x *PrmContainerSetEACL) SetTable(table *eacl.Table) {
	x.table = table
}

// PrmBalanceGet groups parameters of Balance operation.
type PrmBalanceGet struct {
	prmCommon

	ownerID *owner.ID
}

// SetOwnerID specifies identifier of the NeoFS account for which the balance is requested.
func (x *PrmBalanceGet) SetOwnerID(ownerID *owner.ID) {
	x.ownerID = ownerID
}

// Pool represents virtual connection to the NeoFS network to communicate
// with multiple NeoFS servers without thinking about switching between servers
// due to load balancing proportions or their unavailability.
// It is designed to provide a convenient abstraction from the multiple client.Client types.
//
// Pool can be created and initialized using NewPool function.
// Before executing the NeoFS operations using the Pool, connection to the
// servers MUST BE correctly established (see Dial method).
// Using the Pool before connecting have been established can lead to a panic.
// After the work, the Pool SHOULD BE closed (see Close method): it frees internal
// and system resources which were allocated for the period of work of the Pool.
// Calling Dial/Close methods during the communication process step strongly discouraged
// as it leads to undefined behavior.
//
// Each method which produces a NeoFS API call may return an error.
// Status of underlying server response is casted to built-in error instance.
// Certain statuses can be checked using `client` and standard `errors` packages.
// Note that package provides some helper functions to work with status returns
// (e.g. client.IsErrContainerNotFound, client.IsErrObjectNotFound).
//
// See pool package overview to get some examples.
type Pool struct {
	innerPools      []*innerPool
	key             *ecdsa.PrivateKey
	owner           *owner.ID
	cancel          context.CancelFunc
	closedCh        chan struct{}
	cache           *sessionCache
	stokenDuration  uint64
	stokenThreshold time.Duration
	rebalanceParams rebalanceParameters
	clientBuilder   func(endpoint string) (Client, error)
	logger          *zap.Logger
}

type innerPool struct {
	lock        sync.RWMutex
	sampler     *sampler
	clientPacks []*clientPack
}

const (
	defaultSessionTokenExpirationDuration = 100 // in blocks

	defaultSessionTokenThreshold = 5 * time.Second
	defaultRebalanceInterval     = 25 * time.Second
	defaultRequestTimeout        = 4 * time.Second
)

// NewPool create connection pool using parameters.
func NewPool(options InitParameters) (*Pool, error) {
	if options.Key == nil {
		return nil, fmt.Errorf("missed required parameter 'Key'")
	}

	nodesParams, err := adjustNodeParams(options.NodeParams)
	if err != nil {
		return nil, err
	}

	fillDefaultInitParams(&options)

	cache, err := newCache()
	if err != nil {
		return nil, fmt.Errorf("couldn't create cache: %w", err)
	}

	pool := &Pool{
		key:             options.Key,
		owner:           owner.NewIDFromPublicKey(&options.Key.PublicKey),
		cache:           cache,
		logger:          options.Logger,
		stokenDuration:  options.SessionExpirationDuration,
		stokenThreshold: options.SessionTokenThreshold,
		rebalanceParams: rebalanceParameters{
			nodesParams:               nodesParams,
			nodeRequestTimeout:        options.NodeRequestTimeout,
			clientRebalanceInterval:   options.ClientRebalanceInterval,
			sessionExpirationDuration: options.SessionExpirationDuration,
		},
		clientBuilder: options.clientBuilder,
	}

	return pool, nil
}

// Dial establishes a connection to the servers from the NeoFS network.
// Returns an error describing failure reason. If failed, the Pool
// SHOULD NOT be used.
func (p *Pool) Dial(ctx context.Context) error {
	inner := make([]*innerPool, len(p.rebalanceParams.nodesParams))
	var atLeastOneHealthy bool

	for i, params := range p.rebalanceParams.nodesParams {
		clientPacks := make([]*clientPack, len(params.weights))
		for j, addr := range params.addresses {
			c, err := p.clientBuilder(addr)
			if err != nil {
				return err
			}
			var healthy bool
			cliRes, err := createSessionTokenForDuration(ctx, c, p.rebalanceParams.sessionExpirationDuration)
			if err != nil && p.logger != nil {
				p.logger.Warn("failed to create neofs session token for client",
					zap.String("Address", addr),
					zap.Error(err))
			} else if err == nil {
				healthy, atLeastOneHealthy = true, true
				st := sessionTokenForOwner(p.owner, cliRes)
				_ = p.cache.Put(formCacheKey(addr, p.key), st)
			}
			clientPacks[j] = &clientPack{client: c, healthy: healthy, address: addr}
		}
		source := rand.NewSource(time.Now().UnixNano())
		sampl := newSampler(params.weights, source)

		inner[i] = &innerPool{
			sampler:     sampl,
			clientPacks: clientPacks,
		}
	}

	if !atLeastOneHealthy {
		return fmt.Errorf("at least one node must be healthy")
	}

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.closedCh = make(chan struct{})
	p.innerPools = inner

	go p.startRebalance(ctx)
	return nil
}

func fillDefaultInitParams(params *InitParameters) {
	if params.SessionExpirationDuration == 0 {
		params.SessionExpirationDuration = defaultSessionTokenExpirationDuration
	}

	if params.SessionTokenThreshold <= 0 {
		params.SessionTokenThreshold = defaultSessionTokenThreshold
	}

	if params.ClientRebalanceInterval <= 0 {
		params.ClientRebalanceInterval = defaultRebalanceInterval
	}

	if params.NodeRequestTimeout <= 0 {
		params.NodeRequestTimeout = defaultRequestTimeout
	}

	if params.clientBuilder == nil {
		params.clientBuilder = func(addr string) (Client, error) {
			var c client.Client

			var prmInit client.PrmInit
			prmInit.ResolveNeoFSFailures()
			prmInit.SetDefaultPrivateKey(*params.Key)

			c.Init(prmInit)

			var prmDial client.PrmDial
			prmDial.SetServerURI(addr)
			prmDial.SetTimeout(params.NodeConnectionTimeout)

			return &c, c.Dial(prmDial)
		}
	}
}

func adjustNodeParams(nodeParams []NodeParam) ([]*nodesParam, error) {
	if len(nodeParams) == 0 {
		return nil, errors.New("no NeoFS peers configured")
	}

	nodesParamsMap := make(map[int]*nodesParam)
	for _, param := range nodeParams {
		nodes, ok := nodesParamsMap[param.Priority]
		if !ok {
			nodes = &nodesParam{priority: param.Priority}
		}
		nodes.addresses = append(nodes.addresses, param.Address)
		nodes.weights = append(nodes.weights, param.Weight)
		nodesParamsMap[param.Priority] = nodes
	}

	nodesParams := make([]*nodesParam, 0, len(nodesParamsMap))
	for _, nodes := range nodesParamsMap {
		nodes.weights = adjustWeights(nodes.weights)
		nodesParams = append(nodesParams, nodes)
	}

	sort.Slice(nodesParams, func(i, j int) bool {
		return nodesParams[i].priority < nodesParams[j].priority
	})

	return nodesParams, nil
}

func (p *Pool) startRebalance(ctx context.Context) {
	ticker := time.NewTimer(p.rebalanceParams.clientRebalanceInterval)
	buffers := make([][]float64, len(p.rebalanceParams.nodesParams))
	for i, params := range p.rebalanceParams.nodesParams {
		buffers[i] = make([]float64, len(params.weights))
	}

	for {
		select {
		case <-ctx.Done():
			close(p.closedCh)
			return
		case <-ticker.C:
			p.updateNodesHealth(ctx, buffers)
			ticker.Reset(p.rebalanceParams.clientRebalanceInterval)
		}
	}
}

func (p *Pool) updateNodesHealth(ctx context.Context, buffers [][]float64) {
	wg := sync.WaitGroup{}
	for i, inner := range p.innerPools {
		wg.Add(1)

		bufferWeights := buffers[i]
		go func(i int, innerPool *innerPool) {
			defer wg.Done()
			p.updateInnerNodesHealth(ctx, i, bufferWeights)
		}(i, inner)
	}
	wg.Wait()
}

func (p *Pool) updateInnerNodesHealth(ctx context.Context, i int, bufferWeights []float64) {
	if i > len(p.innerPools)-1 {
		return
	}
	pool := p.innerPools[i]
	options := p.rebalanceParams

	healthyChanged := false
	wg := sync.WaitGroup{}

	var prmEndpoint client.PrmEndpointInfo

	for j, cPack := range pool.clientPacks {
		wg.Add(1)
		go func(j int, cli Client) {
			defer wg.Done()
			ok := true
			tctx, c := context.WithTimeout(ctx, options.nodeRequestTimeout)
			defer c()

			if _, err := cli.EndpointInfo(tctx, prmEndpoint); err != nil {
				ok = false
				bufferWeights[j] = 0
			}
			pool.lock.RLock()
			cp := *pool.clientPacks[j]
			pool.lock.RUnlock()

			if ok {
				bufferWeights[j] = options.nodesParams[i].weights[j]
				if !cp.healthy {
					cliRes, err := createSessionTokenForDuration(ctx, cli, options.sessionExpirationDuration)
					if err != nil {
						ok = false
						bufferWeights[j] = 0
					} else {
						tkn := p.newSessionToken(cliRes)

						_ = p.cache.Put(formCacheKey(cp.address, p.key), tkn)
					}
				}
			} else {
				p.cache.DeleteByPrefix(cp.address)
			}

			pool.lock.Lock()
			if pool.clientPacks[j].healthy != ok {
				pool.clientPacks[j].healthy = ok
				healthyChanged = true
			}
			pool.lock.Unlock()
		}(j, cPack.client)
	}
	wg.Wait()

	if healthyChanged {
		probabilities := adjustWeights(bufferWeights)
		source := rand.NewSource(time.Now().UnixNano())
		pool.lock.Lock()
		pool.sampler = newSampler(probabilities, source)
		pool.lock.Unlock()
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

func (p *Pool) Connection() (Client, *session.Token, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, nil, err
	}

	tok := p.cache.Get(formCacheKey(cp.address, p.key))
	return cp.client, tok, nil
}

func (p *Pool) connection() (*clientPack, error) {
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

func (p *Pool) OwnerID() *owner.ID {
	return p.owner
}

func formCacheKey(address string, key *ecdsa.PrivateKey) string {
	k := keys.PrivateKey{PrivateKey: *key}
	return address + k.String()
}

func (p *Pool) conn(ctx context.Context, cfg prmCommon) (*clientPack, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	key := p.key
	if cfg.key != nil {
		key = cfg.key
	}

	sessionToken := cfg.stoken
	if sessionToken == nil && cfg.defaultSession {
		cacheKey := formCacheKey(cp.address, key)
		sessionToken = p.cache.Get(cacheKey)
		if sessionToken == nil {
			cliRes, err := createSessionTokenForDuration(ctx, cp.client, p.stokenDuration)
			if err != nil {
				return nil, err
			}

			ownerID := owner.NewIDFromPublicKey(&key.PublicKey)
			sessionToken = sessionTokenForOwner(ownerID, cliRes)

			cfg.stoken = sessionToken

			_ = p.cache.Put(cacheKey, sessionToken)
		}
	}

	return cp, nil
}

func (p *Pool) checkSessionTokenErr(err error, address string) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), "session token does not exist") ||
		strings.Contains(err.Error(), "session token has been expired") {
		p.cache.DeleteByPrefix(address)
		return true
	}

	return false
}

func createSessionTokenForDuration(ctx context.Context, c Client, dur uint64) (*client.ResSessionCreate, error) {
	ni, err := c.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return nil, err
	}

	epoch := ni.Info().CurrentEpoch()

	var prm client.PrmSessionCreate
	if math.MaxUint64-epoch < dur {
		prm.SetExp(math.MaxUint64)
	} else {
		prm.SetExp(epoch + dur)
	}

	return c.SessionCreate(ctx, prm)
}

func (p *Pool) removeSessionTokenAfterThreshold(cfg prmCommon) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	key := p.key
	if cfg.key != nil {
		key = cfg.key
	}

	ts, ok := p.cache.GetAccessTime(formCacheKey(cp.address, key))
	if ok && time.Since(ts) > p.stokenThreshold {
		p.cache.DeleteByPrefix(cp.address)
	}

	return nil
}

type callContext struct {
	// base context for RPC
	context.Context

	client Client

	// client endpoint
	endpoint string

	// request signer
	key *ecdsa.PrivateKey

	// flag to open default session if session token is missing
	sessionDefault bool
	sessionTarget  func(session.Token)
	sessionContext *session.ObjectContext
}

func (p *Pool) initCallContext(ctx *callContext, cfg prmCommon) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	ctx.key = cfg.key
	if ctx.key == nil {
		// use pool key if caller didn't specify its own
		ctx.key = p.key
	}

	ctx.endpoint = cp.address
	ctx.client = cp.client

	if ctx.sessionTarget != nil && cfg.stoken != nil {
		ctx.sessionTarget(*cfg.stoken)
	}

	// note that we don't override session provided by the caller
	ctx.sessionDefault = cfg.stoken == nil && cfg.defaultSession
	if ctx.sessionDefault {
		ctx.sessionContext = session.NewObjectContext()
		ctx.sessionContext.ToV2().SetVerb(cfg.verb)
		ctx.sessionContext.ApplyTo(cfg.addr)
	}

	return err
}

type callContextWithRetry struct {
	callContext

	noRetry bool
}

func (p *Pool) initCallContextWithRetry(ctx *callContextWithRetry, cfg prmCommon) error {
	err := p.initCallContext(&ctx.callContext, cfg)
	if err != nil {
		return err
	}

	// don't retry if session was specified by the caller
	ctx.noRetry = cfg.stoken != nil

	return nil
}

// opens new session or uses cached one.
// Must be called only on initialized callContext with set sessionTarget.
func (p *Pool) openDefaultSession(ctx *callContext) error {
	cacheKey := formCacheKey(ctx.endpoint, ctx.key)

	tok := p.cache.Get(cacheKey)
	if tok == nil {
		// open new session
		cliRes, err := createSessionTokenForDuration(ctx, ctx.client, p.stokenDuration)
		if err != nil {
			return fmt.Errorf("session API client: %w", err)
		}

		tok = sessionTokenForOwner(owner.NewIDFromPublicKey(&ctx.key.PublicKey), cliRes)
		// cache the opened session
		p.cache.Put(cacheKey, tok)
	}

	tokToSign := *tok
	tokToSign.SetContext(ctx.sessionContext)

	// sign the token
	if err := tokToSign.Sign(ctx.key); err != nil {
		return fmt.Errorf("sign token of the opened session: %w", err)
	}

	ctx.sessionTarget(tokToSign)

	return nil
}

// opens default session (if sessionDefault is set), and calls f. If f returns
// session-related error (*), and retrying is enabled, then f is called once more.
//
// (*) in this case cached token is removed.
func (p *Pool) callWithRetry(ctx *callContextWithRetry, f func() error) error {
	var err error

	if ctx.sessionDefault {
		err = p.openDefaultSession(&ctx.callContext)
		if err != nil {
			return fmt.Errorf("open default session: %w", err)
		}
	}

	err = f()

	if p.checkSessionTokenErr(err, ctx.endpoint) && !ctx.noRetry {
		// don't retry anymore
		ctx.noRetry = true
		return p.callWithRetry(ctx, f)
	}

	return err
}

// PutObject writes an object through a remote server using NeoFS API protocol.
func (p *Pool) PutObject(ctx context.Context, prm PrmObjectPut) (*oid.ID, error) {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbPut)
	prm.useAddress(newAddressFromCnrID(prm.hdr.ContainerID()))

	// Put object is different from other object service methods. Put request
	// can't be resent in case of session token failures (i.e. session token is
	// invalid due to lifetime expiration or server restart). The reason is that
	// object's payload can be provided as a stream that should be read only once.
	//
	// To solve this issue, pool regenerates session tokens upon each request.
	// In case of subsequent requests, pool avoids session token initialization
	// by checking when the session token was accessed for the last time. If it
	// hits a threshold, session token is removed from cache for a new one to be
	// issued.
	err := p.removeSessionTokenAfterThreshold(prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var ctxCall callContext

	ctxCall.Context = ctx

	err = p.initCallContext(&ctxCall, prm.prmCommon)
	if err != nil {
		return nil, fmt.Errorf("init call context")
	}

	var cliPrm client.PrmObjectPutInit

	wObj, err := ctxCall.client.ObjectPutInit(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("init writing on API client: %w", err)
	}

	if ctxCall.sessionDefault {
		ctxCall.sessionTarget = wObj.WithinSession
		err = p.openDefaultSession(&ctxCall)
		if err != nil {
			return nil, fmt.Errorf("open default session: %w", err)
		}
	}

	wObj.UseKey(*ctxCall.key)

	if prm.btoken != nil {
		wObj.WithBearerToken(*prm.btoken)
	}

	if wObj.WriteHeader(prm.hdr) {
		sz := prm.hdr.PayloadSize()

		if data := prm.hdr.Payload(); len(data) > 0 {
			if prm.payload != nil {
				prm.payload = io.MultiReader(bytes.NewReader(data), prm.payload)
			} else {
				prm.payload = bytes.NewReader(data)
				sz = uint64(len(data))
			}
		}

		if prm.payload != nil {
			const defaultBufferSizePut = 3 << 20 // configure?

			if sz == 0 || sz > defaultBufferSizePut {
				sz = defaultBufferSizePut
			}

			buf := make([]byte, sz)

			var n int

			for {
				n, err = prm.payload.Read(buf)
				if n > 0 {
					if !wObj.WritePayloadChunk(buf[:n]) {
						break
					}

					continue
				}

				if errors.Is(err, io.EOF) {
					break
				}

				return nil, fmt.Errorf("read payload: %w", err)
			}
		}
	}

	res, err := wObj.Close()
	if err != nil { // here err already carries both status and client errors
		// removes session token from cache in case of token error
		p.checkSessionTokenErr(err, ctxCall.endpoint)
		return nil, fmt.Errorf("client failure: %w", err)
	}

	var id oid.ID

	if !res.ReadStoredObjectID(&id) {
		return nil, errors.New("missing ID of the stored object")
	}

	return &id, nil
}

// DeleteObject marks an object for deletion from the container using NeoFS API protocol.
// As a marker, a special unit called a tombstone is placed in the container.
// It confirms the user's intent to delete the object, and is itself a container object.
// Explicit deletion is done asynchronously, and is generally not guaranteed.
func (p *Pool) DeleteObject(ctx context.Context, prm PrmObjectDelete) error {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbDelete)
	prm.useAddress(&prm.addr)

	var cliPrm client.PrmObjectDelete

	var cc callContextWithRetry

	cc.Context = ctx
	cc.sessionTarget = cliPrm.WithinSession

	err := p.initCallContextWithRetry(&cc, prm.prmCommon)
	if err != nil {
		return err
	}

	if cnr := prm.addr.ContainerID(); cnr != nil {
		cliPrm.FromContainer(*cnr)
	}

	if obj := prm.addr.ObjectID(); obj != nil {
		cliPrm.ByID(*obj)
	}

	cliPrm.UseKey(*cc.key)

	return p.callWithRetry(&cc, func() error {
		_, err := cc.client.ObjectDelete(ctx, cliPrm)
		if err != nil {
			return fmt.Errorf("remove object via client: %w", err)
		}

		return nil
	})
}

type objectReadCloser client.ObjectReader

// Read implements io.Reader of the object payload.
func (x *objectReadCloser) Read(p []byte) (int, error) {
	return (*client.ObjectReader)(x).Read(p)
}

// Close implements io.Closer of the object payload.
func (x *objectReadCloser) Close() error {
	_, err := (*client.ObjectReader)(x).Close()
	return err
}

// ResGetObject is designed to provide object header nad read one object payload from NeoFS system.
type ResGetObject struct {
	Header object.Object

	Payload io.ReadCloser
}

// GetObject reads object header and initiates reading an object payload through a remote server using NeoFS API protocol.
func (p *Pool) GetObject(ctx context.Context, prm PrmObjectGet) (*ResGetObject, error) {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbGet)
	prm.useAddress(&prm.addr)

	var cliPrm client.PrmObjectGet

	var cc callContextWithRetry

	cc.Context = ctx
	cc.sessionTarget = cliPrm.WithinSession

	err := p.initCallContextWithRetry(&cc, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	if cnr := prm.addr.ContainerID(); cnr != nil {
		cliPrm.FromContainer(*cnr)
	}

	if obj := prm.addr.ObjectID(); obj != nil {
		cliPrm.ByID(*obj)
	}

	var res ResGetObject

	err = p.callWithRetry(&cc, func() error {
		rObj, err := cc.client.ObjectGetInit(ctx, cliPrm)
		if err != nil {
			return fmt.Errorf("init object reading on client: %w", err)
		}

		rObj.UseKey(*cc.key)

		if !rObj.ReadHeader(&res.Header) {
			_, err = rObj.Close()
			return fmt.Errorf("read header: %w", err)
		}

		res.Payload = (*objectReadCloser)(rObj)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// HeadObject reads object header through a remote server using NeoFS API protocol.
func (p *Pool) HeadObject(ctx context.Context, prm PrmObjectHead) (*object.Object, error) {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbHead)
	prm.useAddress(&prm.addr)

	var cliPrm client.PrmObjectHead

	var cc callContextWithRetry

	cc.Context = ctx
	cc.sessionTarget = cliPrm.WithinSession

	err := p.initCallContextWithRetry(&cc, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	if cnr := prm.addr.ContainerID(); cnr != nil {
		cliPrm.FromContainer(*cnr)
	}

	if obj := prm.addr.ObjectID(); obj != nil {
		cliPrm.ByID(*obj)
	}

	cliPrm.UseKey(*cc.key)

	var obj object.Object

	err = p.callWithRetry(&cc, func() error {
		res, err := cc.client.ObjectHead(ctx, cliPrm)
		if err != nil {
			return fmt.Errorf("read object header via client: %w", err)
		}

		if !res.ReadHeader(&obj) {
			return errors.New("missing object header in response")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &obj, nil
}

// ResObjectRange is designed to read payload range of one object
// from NeoFS system.
//
// Must be initialized using Pool.ObjectRange, any other
// usage is unsafe.
type ResObjectRange struct {
	payload *client.ObjectRangeReader
}

// Read implements io.Reader of the object payload.
func (x *ResObjectRange) Read(p []byte) (int, error) {
	return x.payload.Read(p)
}

// Close ends reading the payload range and returns the result of the operation
// along with the final results. Must be called after using the ResObjectRange.
func (x *ResObjectRange) Close() error {
	_, err := x.payload.Close()
	return err
}

// ObjectRange initiates reading an object's payload range through a remote
// server using NeoFS API protocol.
func (p *Pool) ObjectRange(ctx context.Context, prm PrmObjectRange) (*ResObjectRange, error) {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbRange)
	prm.useAddress(&prm.addr)

	var cliPrm client.PrmObjectRange

	cliPrm.SetOffset(prm.off)
	cliPrm.SetLength(prm.ln)

	var cc callContextWithRetry

	cc.Context = ctx
	cc.sessionTarget = cliPrm.WithinSession

	err := p.initCallContextWithRetry(&cc, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	if cnr := prm.addr.ContainerID(); cnr != nil {
		cliPrm.FromContainer(*cnr)
	}

	if obj := prm.addr.ObjectID(); obj != nil {
		cliPrm.ByID(*obj)
	}

	var res ResObjectRange

	err = p.callWithRetry(&cc, func() error {
		var err error

		res.payload, err = cc.client.ObjectRangeInit(ctx, cliPrm)
		if err != nil {
			return fmt.Errorf("init payload range reading on client: %w", err)
		}

		res.payload.UseKey(*cc.key)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// ResObjectSearch is designed to read list of object identifiers from NeoFS system.
//
// Must be initialized using Pool.SearchObjects, any other usage is unsafe.
type ResObjectSearch struct {
	r *client.ObjectListReader
}

// Read reads another list of the object identifiers.
func (x *ResObjectSearch) Read(buf []oid.ID) (int, error) {
	n, ok := x.r.Read(buf)
	if !ok {
		_, err := x.r.Close()
		if err == nil {
			return n, io.EOF
		}

		return n, err
	}

	return n, nil
}

// Iterate iterates over the list of found object identifiers.
// f can return true to stop iteration earlier.
//
// Returns an error if object can't be read.
func (x *ResObjectSearch) Iterate(f func(oid.ID) bool) error {
	return x.r.Iterate(f)
}

// Close ends reading list of the matched objects and returns the result of the operation
// along with the final results. Must be called after using the ResObjectSearch.
func (x *ResObjectSearch) Close() {
	_, _ = x.r.Close()
}

// SearchObjects initiates object selection through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching of matched objects
// is done using the ResObjectSearch. Exactly one return value is non-nil.
// Resulting reader must be finally closed.
func (p *Pool) SearchObjects(ctx context.Context, prm PrmObjectSearch) (*ResObjectSearch, error) {
	prm.useDefaultSession()
	prm.useVerb(sessionv2.ObjectVerbSearch)
	prm.useAddress(newAddressFromCnrID(&prm.cnrID))

	var cliPrm client.PrmObjectSearch

	cliPrm.InContainer(prm.cnrID)
	cliPrm.SetFilters(prm.filters)

	var cc callContextWithRetry

	cc.Context = ctx
	cc.sessionTarget = cliPrm.WithinSession

	err := p.initCallContextWithRetry(&cc, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var res ResObjectSearch

	err = p.callWithRetry(&cc, func() error {
		var err error

		res.r, err = cc.client.ObjectSearchInit(ctx, cliPrm)
		if err != nil {
			return fmt.Errorf("init object searching on client: %w", err)
		}

		res.r.UseKey(*cc.key)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// PutContainer sends request to save container in NeoFS.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see GetContainer).
func (p *Pool) PutContainer(ctx context.Context, prm PrmContainerPut) (*cid.ID, error) {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var cliPrm client.PrmContainerPut

	if prm.cnr != nil {
		cliPrm.SetContainer(*prm.cnr)
	}

	res, err := cp.client.ContainerPut(ctx, cliPrm)
	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.ID(), nil
}

// GetContainer reads NeoFS container by ID.
func (p *Pool) GetContainer(ctx context.Context, prm PrmContainerGet) (*container.Container, error) {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var cliPrm client.PrmContainerGet

	if prm.cnrID != nil {
		cliPrm.SetContainer(*prm.cnrID)
	}

	res, err := cp.client.ContainerGet(ctx, cliPrm)
	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Container(), nil
}

// ListContainers requests identifiers of the account-owned containers.
func (p *Pool) ListContainers(ctx context.Context, prm PrmContainerList) ([]cid.ID, error) {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var cliPrm client.PrmContainerList

	if prm.ownerID != nil {
		cliPrm.SetAccount(*prm.ownerID)
	}

	res, err := cp.client.ContainerList(ctx, cliPrm)
	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Containers(), nil
}

// DeleteContainer sends request to remove the NeoFS container.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see GetContainer).
func (p *Pool) DeleteContainer(ctx context.Context, prm PrmContainerDelete) error {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return err
	}

	var cliPrm client.PrmContainerDelete

	if prm.cnrID != nil {
		cliPrm.SetContainer(*prm.cnrID)
	}

	if prm.stoken != nil {
		cliPrm.SetSessionToken(*prm.stoken)
	}

	_, err = cp.client.ContainerDelete(ctx, cliPrm)

	// here err already carries both status and client errors

	return err
}

// GetEACL reads eACL table of the NeoFS container.
func (p *Pool) GetEACL(ctx context.Context, prm PrmContainerEACL) (*eacl.Table, error) {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var cliPrm client.PrmContainerEACL

	if prm.cnrID != nil {
		cliPrm.SetContainer(*prm.cnrID)
	}

	res, err := cp.client.ContainerEACL(ctx, cliPrm)
	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Table(), nil
}

// SetEACL sends request to update eACL table of the NeoFS container.
//
// Operation is asynchronous and no guaranteed even in the absence of errors.
// The required time is also not predictable.
//
// Success can be verified by reading by identifier (see GetEACL).
func (p *Pool) SetEACL(ctx context.Context, prm PrmContainerSetEACL) error {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return err
	}

	var cliPrm client.PrmContainerSetEACL

	if prm.table != nil {
		cliPrm.SetTable(*prm.table)
	}

	_, err = cp.client.ContainerSetEACL(ctx, cliPrm)

	// here err already carries both status and client errors

	return err
}

// Balance requests current balance of the NeoFS account.
func (p *Pool) Balance(ctx context.Context, prm PrmBalanceGet) (*accounting.Decimal, error) {
	cp, err := p.conn(ctx, prm.prmCommon)
	if err != nil {
		return nil, err
	}

	var cliPrm client.PrmBalanceGet

	if prm.ownerID != nil {
		cliPrm.SetAccount(*prm.ownerID)
	}

	res, err := cp.client.BalanceGet(ctx, cliPrm)
	if err != nil { // here err already carries both status and client errors
		return nil, err
	}

	return res.Amount(), nil
}

// WaitForContainerPresence waits until the container is found on the NeoFS network.
func (p *Pool) WaitForContainerPresence(ctx context.Context, cid *cid.ID, pollParams *ContainerPollingParams) error {
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

	var cliPrm client.PrmContainerGet

	if cid != nil {
		cliPrm.SetContainer(*cid)
	}

	for {
		select {
		case <-done:
			return ctx.Err()
		case <-wdone:
			return wctx.Err()
		case <-ticker.C:
			_, err = conn.ContainerGet(ctx, cliPrm)
			if err == nil {
				return nil
			}
			ticker.Reset(pollParams.PollInterval)
		}
	}
}

// Close closes the Pool and releases all the associated resources.
func (p *Pool) Close() {
	p.cancel()
	<-p.closedCh
}

// creates new session token from SessionCreate call result.
func (p *Pool) newSessionToken(cliRes *client.ResSessionCreate) *session.Token {
	return sessionTokenForOwner(p.owner, cliRes)
}

// creates new session token with specified owner from SessionCreate call result.
func sessionTokenForOwner(id *owner.ID, cliRes *client.ResSessionCreate) *session.Token {
	st := session.NewToken()
	st.SetOwnerID(id)
	st.SetID(cliRes.ID())
	st.SetSessionKey(cliRes.PublicKey())

	return st
}

func newAddressFromCnrID(cnrID *cid.ID) *address.Address {
	addr := address.NewAddress()
	addr.SetContainerID(cnrID)
	return addr
}
