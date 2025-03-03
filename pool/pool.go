package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	sdkClient "github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/internal/uriutil"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"go.uber.org/zap"
)

const (
	// max GRPC message size.
	defaultBufferSize = 4194304 // 4MB

	defaultNodeSessionCacheSize = 100
)

type sdkClientInterface interface {
	Dial(prm sdkClient.PrmDial) error

	BalanceGet(ctx context.Context, prm sdkClient.PrmBalanceGet) (accounting.Decimal, error)

	ContainerPut(ctx context.Context, cont container.Container, signer neofscrypto.Signer, prm sdkClient.PrmContainerPut) (cid.ID, error)
	ContainerGet(ctx context.Context, id cid.ID, prm sdkClient.PrmContainerGet) (container.Container, error)
	ContainerList(ctx context.Context, ownerID user.ID, prm sdkClient.PrmContainerList) ([]cid.ID, error)
	ContainerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm sdkClient.PrmContainerDelete) error
	ContainerEACL(ctx context.Context, id cid.ID, prm sdkClient.PrmContainerEACL) (eacl.Table, error)
	ContainerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm sdkClient.PrmContainerSetEACL) error

	NetworkInfo(ctx context.Context, prm sdkClient.PrmNetworkInfo) (netmap.NetworkInfo, error)
	NetMapSnapshot(ctx context.Context, prm sdkClient.PrmNetMapSnapshot) (netmap.NetMap, error)

	ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm sdkClient.PrmObjectPutInit) (sdkClient.ObjectWriter, error)
	ObjectGetInit(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm sdkClient.PrmObjectGet) (object.Object, *sdkClient.PayloadReader, error)
	ObjectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm sdkClient.PrmObjectHead) (*object.Object, error)
	ObjectRangeInit(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer user.Signer, prm sdkClient.PrmObjectRange) (*sdkClient.ObjectRangeReader, error)
	ObjectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm sdkClient.PrmObjectDelete) (oid.ID, error)
	ObjectHash(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm sdkClient.PrmObjectHash) ([][]byte, error)
	ObjectSearchInit(ctx context.Context, containerID cid.ID, signer user.Signer, prm sdkClient.PrmObjectSearch) (*sdkClient.ObjectListReader, error)
	SearchObjects(ctx context.Context, containerID cid.ID, filters object.SearchFilters, attrs []string, cursor string, signer neofscrypto.Signer, opts sdkClient.SearchObjectsOptions) ([]sdkClient.SearchResultItem, string, error)

	SessionCreate(ctx context.Context, signer user.Signer, prm sdkClient.PrmSessionCreate) (*sdkClient.ResSessionCreate, error)

	EndpointInfo(ctx context.Context, prm sdkClient.PrmEndpointInfo) (*sdkClient.ResEndpointInfo, error)
}

type sdkClientWrapper struct {
	sdkClientInterface

	nodeSession nodeSessionContainer
	addr        string
}

// nodeSessionContainer represents storage for a session tokens. It contains only basics session info: id, pub key, expiration.
// This tokens are used for the final session tokens creation for specific verb. This token is 1:1 for each public key in the node.
// Should be stored until token not expired.
type nodeSessionContainer interface {
	SetNodeSession(*session.Object, neofscrypto.PublicKey)
	GetNodeSession(neofscrypto.PublicKey) *session.Object
	ResetSessions()
}

// client represents virtual connection to the single NeoFS network endpoint from which Pool is formed.
// This interface is expected to have exactly one production implementation - clientWrapper.
// Others are expected to be for test purposes only.
type internalClient interface {
	clientStatus
	nodeSessionContainer
	io.Closer

	// see clientWrapper.dial.
	dial(ctx context.Context) error
	// see clientWrapper.restartIfUnhealthy.
	restartIfUnhealthy(ctx context.Context) (bool, bool)

	getClient() (sdkClientInterface, error)
	getRawClient() (*sdkClient.Client, error)
}

// clientStatus provide access to some metrics for connection.
type clientStatus interface {
	// isHealthy checks if the connection can handle requests.
	isHealthy() bool
	// setUnhealthy marks client as unhealthy.
	setUnhealthy()
	// address return address of endpoint.
	address() string
	// currentErrorRate returns current errors rate.
	// After specific threshold connection is considered as unhealthy.
	// Pool.startRebalance routine can make this connection healthy again.
	currentErrorRate() uint32
	// overallErrorRate returns the number of all happened errors.
	overallErrorRate() uint64
}

// errPoolClientUnhealthy is an error to indicate that client in pool is unhealthy.
var errPoolClientUnhealthy = errors.New("pool client unhealthy")

// clientStatusMonitor count error rate and other statistics for connection.
type clientStatusMonitor struct {
	addr           string
	healthy        *atomic.Bool
	errorThreshold uint32

	mu                *sync.RWMutex // protect counters
	currentErrorCount uint32
	overallErrorCount uint64
}

func newClientStatusMonitor(addr string, errorThreshold uint32) clientStatusMonitor {
	m := clientStatusMonitor{
		addr:           addr,
		healthy:        &atomic.Bool{},
		mu:             &sync.RWMutex{},
		errorThreshold: errorThreshold,
	}

	m.healthy.Store(true)
	return m
}

// clientWrapper is used by default, alternative implementations are intended for testing purposes only.
type clientWrapper struct {
	clientMutex sync.RWMutex
	client      *sdkClient.Client
	prm         wrapperPrm

	clientStatusMonitor
	statisticCallback stat.OperationCallback

	nodeSessionCache *sessionCache

	epoch atomic.Uint64
}

// wrapperPrm is params to create clientWrapper.
type wrapperPrm struct {
	address              string
	signer               neofscrypto.Signer
	dialTimeout          time.Duration
	streamTimeout        time.Duration
	errorThreshold       uint32
	responseInfoCallback func(sdkClient.ResponseMetaInfo) error
	statisticCallback    stat.OperationCallback
	buffers              *sync.Pool
	nodeSessionCacheSize int
}

// setAddress sets endpoint to connect in NeoFS network.
func (x *wrapperPrm) setAddress(address string) {
	x.address = address
}

// setSigner sets sdkClient.Client private signer to be used for the protocol communication by default.
func (x *wrapperPrm) setSigner(signer neofscrypto.Signer) {
	x.signer = signer
}

// setDialTimeout sets the timeout for connection to be established.
func (x *wrapperPrm) setDialTimeout(timeout time.Duration) {
	x.dialTimeout = timeout
}

// setStreamTimeout sets the timeout for individual operations in streaming RPC.
func (x *wrapperPrm) setStreamTimeout(timeout time.Duration) {
	x.streamTimeout = timeout
}

// setErrorThreshold sets threshold after reaching which connection is considered unhealthy
// until Pool.startRebalance routing updates its status.
func (x *wrapperPrm) setErrorThreshold(threshold uint32) {
	x.errorThreshold = threshold
}

// setResponseInfoCallback sets callback that will be invoked after every response.
func (x *wrapperPrm) setResponseInfoCallback(f func(sdkClient.ResponseMetaInfo) error) {
	x.responseInfoCallback = f
}

// setStatisticCallback set callback for external statistic.
func (x *wrapperPrm) setStatisticCallback(statisticCallback stat.OperationCallback) {
	x.statisticCallback = statisticCallback
}

// setBuffers set buffer to message sign routine inside sdkClient.Client.
func (x *wrapperPrm) setBuffers(buffers *sync.Pool) {
	x.buffers = buffers
}

// SetNodeSessionCacheSize sets cache size for the basic sessions for node.
func (x *wrapperPrm) setNodeSessionCacheSize(cacheSize int) {
	x.nodeSessionCacheSize = cacheSize
}

// getNewClient returns a new [sdkClient.Client] instance using internal parameters.
func (x *wrapperPrm) getNewClient(statisticCallback stat.OperationCallback) (*sdkClient.Client, error) {
	var prmInit sdkClient.PrmInit
	prmInit.SetResponseInfoCallback(x.responseInfoCallback)
	prmInit.SetStatisticCallback(statisticCallback)
	prmInit.SetSignMessageBuffers(x.buffers)

	return sdkClient.New(prmInit)
}

// newWrapper creates a clientWrapper that implements the client interface.
func newWrapper(prm wrapperPrm) (*clientWrapper, error) {
	cacheSize := prm.nodeSessionCacheSize
	if cacheSize == 0 {
		cacheSize = defaultNodeSessionCacheSize
	}

	cache, err := newCache(cacheSize)
	if err != nil {
		return nil, fmt.Errorf("cacheNodeSessions: %w", err)
	}

	res := &clientWrapper{
		clientStatusMonitor: newClientStatusMonitor(prm.address, prm.errorThreshold),
		statisticCallback:   prm.statisticCallback,
		nodeSessionCache:    cache,
	}

	oldCallBack := prm.responseInfoCallback
	prm.setResponseInfoCallback(func(info sdkClient.ResponseMetaInfo) error {
		newEpoch := info.Epoch()
		if newEpoch > res.epoch.Load() {
			res.epoch.Store(newEpoch)
			cache.updateEpoch(newEpoch)
		}

		if oldCallBack != nil {
			return oldCallBack(info)
		}

		return nil
	})

	res.prm = prm

	// integrate clientWrapper middleware to handle errors and wrapped client health.
	cl, err := prm.getNewClient(res.statisticMiddleware)
	if err != nil {
		return nil, err
	}

	res.client = cl

	return res, nil
}

func (c *clientWrapper) statisticMiddleware(nodeKey []byte, endpoint string, method stat.Method, duration time.Duration, err error) {
	c.updateErrorRate(err)

	if c.statisticCallback != nil {
		c.statisticCallback(nodeKey, endpoint, method, duration, err)
	}
}

// dial establishes a connection to the server from the NeoFS network.
// Returns an error describing failure reason. If failed, the client
// SHOULD NOT be used.
func (c *clientWrapper) dial(ctx context.Context) error {
	cl, err := c.getClient()
	if err != nil {
		return err
	}

	var prmDial sdkClient.PrmDial
	prmDial.SetServerURI(c.prm.address)
	prmDial.SetTimeout(c.prm.dialTimeout)
	prmDial.SetStreamTimeout(c.prm.streamTimeout)
	prmDial.SetContext(ctx)

	if err = cl.Dial(prmDial); err != nil {
		c.setUnhealthy()
		return err
	}

	return nil
}

// restartIfUnhealthy checks healthy status of client and recreate it if status is unhealthy.
// Return current healthy status and indicating if status was changed by this function call.
func (c *clientWrapper) restartIfUnhealthy(ctx context.Context) (healthy, changed bool) {
	var wasHealthy bool

	cl, err := c.getRawClient()
	if err != nil {
		cl, err = c.prm.getNewClient(c.statisticMiddleware)
		if err != nil {
			c.setUnhealthy()
			return false, wasHealthy
		}

		var prmDial sdkClient.PrmDial
		prmDial.SetServerURI(c.prm.address)
		prmDial.SetTimeout(c.prm.dialTimeout)
		prmDial.SetStreamTimeout(c.prm.streamTimeout)
		prmDial.SetContext(ctx)

		if err := cl.Dial(prmDial); err != nil {
			c.setUnhealthy()
			return false, wasHealthy
		}

		c.clientMutex.Lock()
		c.client = cl
		c.clientMutex.Unlock()
	} else {
		wasHealthy = true
	}

	_, err = cl.EndpointInfo(ctx, sdkClient.PrmEndpointInfo{})
	if err != nil {
		c.setUnhealthy()
		return false, wasHealthy
	}

	c.setHealthy()
	return true, !wasHealthy
}

func (c *clientWrapper) getClient() (sdkClientInterface, error) {
	return c.getRawClient()
}

func (c *clientWrapper) getRawClient() (*sdkClient.Client, error) {
	c.clientMutex.RLock()
	defer c.clientMutex.RUnlock()
	if c.isHealthy() {
		return c.client, nil
	}
	return nil, errPoolClientUnhealthy
}

func (c *clientWrapper) nodeSessionKey(pubKey neofscrypto.PublicKey) string {
	return string(neofscrypto.PublicKeyBytes(pubKey))
}

func (c *clientWrapper) SetNodeSession(token *session.Object, pubKey neofscrypto.PublicKey) {
	c.nodeSessionCache.Put(c.nodeSessionKey(pubKey), *token)
}

func (c *clientWrapper) GetNodeSession(pubKey neofscrypto.PublicKey) *session.Object {
	tok, ok := c.nodeSessionCache.Get(c.nodeSessionKey(pubKey))
	if ok {
		return &tok
	}

	return nil
}

func (c *clientWrapper) ResetSessions() {
	c.nodeSessionCache.Purge()
}

func (c *clientWrapper) Close() error { return c.client.Close() }

func (c *clientStatusMonitor) isHealthy() bool {
	return c.healthy.Load()
}

func (c *clientStatusMonitor) setHealthy() {
	c.healthy.Store(true)
}

func (c *clientStatusMonitor) setUnhealthy() {
	c.healthy.Store(false)
}

func (c *clientStatusMonitor) address() string {
	return c.addr
}

func (c *clientStatusMonitor) incErrorRate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentErrorCount++
	c.overallErrorCount++
	if c.currentErrorCount >= c.errorThreshold {
		c.setUnhealthy()
		c.currentErrorCount = 0
	}
}

func (c *clientStatusMonitor) currentErrorRate() uint32 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentErrorCount
}

func (c *clientStatusMonitor) overallErrorRate() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.overallErrorCount
}

func (c *clientStatusMonitor) updateErrorRate(err error) {
	if err == nil {
		return
	}

	// count only this API errors
	if errors.Is(err, apistatus.ErrServerInternal) ||
		errors.Is(err, apistatus.ErrWrongMagicNumber) ||
		errors.Is(err, apistatus.ErrSignatureVerification) ||
		errors.Is(err, apistatus.ErrNodeUnderMaintenance) {
		c.incErrorRate()
		return
	}

	// don't count another API errors
	if errors.Is(err, apistatus.Error) {
		return
	}

	// non-status logic error that could be returned
	// from the SDK client; should not be considered
	// as a connection error
	var siErr *object.SplitInfoError
	if !errors.As(err, &siErr) {
		c.incErrorRate()
	}
}

// clientBuilder is a type alias of client constructors.
type clientBuilder = func(endpoint string) (internalClient, error)

// InitParameters contains values used to initialize connection Pool.
type InitParameters struct {
	signer                    user.Signer
	logger                    *zap.Logger
	nodeDialTimeout           time.Duration
	nodeStreamTimeout         time.Duration
	healthcheckTimeout        time.Duration
	clientRebalanceInterval   time.Duration
	sessionExpirationDuration uint64
	errorThreshold            uint32
	nodeParams                []NodeParam
	nodeSessionCacheSize      int

	clientBuilder clientBuilder

	statisticCallback stat.OperationCallback
}

// SetSigner specifies default signer to be used for the protocol communication by default.
// MUST be of [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme, for example,
// [neofsecdsa.SignerRFC6979] can be used.
func (x *InitParameters) SetSigner(signer user.Signer) {
	x.signer = signer
}

// SetLogger specifies logger.
func (x *InitParameters) SetLogger(logger *zap.Logger) {
	x.logger = logger
}

// SetNodeDialTimeout specifies the timeout for connection to be established.
func (x *InitParameters) SetNodeDialTimeout(timeout time.Duration) {
	x.nodeDialTimeout = timeout
}

// SetNodeStreamTimeout specifies the timeout for individual operations in streaming RPC.
func (x *InitParameters) SetNodeStreamTimeout(timeout time.Duration) {
	x.nodeStreamTimeout = timeout
}

// SetHealthcheckTimeout specifies the timeout for request to node to decide if it is alive.
//
// See also Pool.Dial.
func (x *InitParameters) SetHealthcheckTimeout(timeout time.Duration) {
	x.healthcheckTimeout = timeout
}

// SetClientRebalanceInterval specifies the interval for updating nodes health status.
//
// See also Pool.Dial.
func (x *InitParameters) SetClientRebalanceInterval(interval time.Duration) {
	x.clientRebalanceInterval = interval
}

// SetSessionExpirationDuration specifies the session token lifetime in epochs.
func (x *InitParameters) SetSessionExpirationDuration(expirationDuration uint64) {
	x.sessionExpirationDuration = expirationDuration
}

// SetErrorThreshold specifies the number of errors on connection after which node is considered as unhealthy.
func (x *InitParameters) SetErrorThreshold(threshold uint32) {
	x.errorThreshold = threshold
}

// AddNode append information about the node to which you want to connect.
func (x *InitParameters) AddNode(nodeParam NodeParam) {
	x.nodeParams = append(x.nodeParams, nodeParam)
}

// setClientBuilder sets clientBuilder used for client construction.
// Wraps setClientBuilderContext without a context.
func (x *InitParameters) setClientBuilder(builder clientBuilder) {
	x.clientBuilder = builder
}

// isMissingClientBuilder checks if client constructor was not specified.
func (x *InitParameters) isMissingClientBuilder() bool {
	return x.clientBuilder == nil
}

// SetStatisticCallback makes the Pool to pass [stat.OperationCallback] for external statistic.
func (x *InitParameters) SetStatisticCallback(statisticCallback stat.OperationCallback) {
	x.statisticCallback = statisticCallback
}

// SetNodeSessionCacheSize sets cache size for the basic sessions for node.
func (x *InitParameters) SetNodeSessionCacheSize(cacheSize int) {
	x.nodeSessionCacheSize = cacheSize
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
	priority int
	address  string
	weight   float64
}

// NewNodeParam creates NodeParam using parameters.
// Address parameter MUST follow the client requirements, see [sdkClient.PrmDial.SetServerURI] for details.
func NewNodeParam(priority int, address string, weight float64) (prm NodeParam) {
	prm.SetPriority(priority)
	prm.SetAddress(address)
	prm.SetWeight(weight)

	return
}

// SetPriority specifies priority of the node.
// Negative value is allowed. In the result node groups
// with the same priority will be sorted by descent.
func (x *NodeParam) SetPriority(priority int) {
	x.priority = priority
}

// SetAddress specifies address of the node.
func (x *NodeParam) SetAddress(address string) {
	x.address = address
}

// SetWeight specifies weight of the node.
func (x *NodeParam) SetWeight(weight float64) {
	x.weight = weight
}

// Pool represents virtual connection to the NeoFS network to communicate
// with multiple NeoFS servers without thinking about switching between servers
// due to load balancing proportions or their unavailability.
// It is designed to provide a convenient abstraction from the multiple sdkClient.client types.
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
// Certain statuses can be checked using `sdkClient` and standard `errors` packages.
//
// See pool package overview to get some examples.
type Pool struct {
	innerPools      []*innerPool
	signer          user.Signer
	cancel          context.CancelFunc
	closedCh        chan struct{}
	cache           *sessionCache
	stokenDuration  uint64
	rebalanceParams rebalanceParameters
	clientBuilder   clientBuilder
	logger          *zap.Logger

	statisticCallback stat.OperationCallback

	buffers *sync.Pool
}

type innerPool struct {
	lock    sync.RWMutex
	sampler *sampler
	clients []internalClient
}

const (
	defaultSessionTokenExpirationDuration = 100 // in blocks
	defaultErrorThreshold                 = 100

	defaultRebalanceInterval  = 25 * time.Second
	defaultHealthcheckTimeout = 4 * time.Second
	defaultDialTimeout        = 5 * time.Second
	defaultStreamTimeout      = 10 * time.Second
)

// DefaultOptions returns default option preset for Pool creation. It may be used like start point for configuration or
// like main configuration.
func DefaultOptions() InitParameters {
	params := InitParameters{
		sessionExpirationDuration: defaultSessionTokenExpirationDuration,
		errorThreshold:            defaultErrorThreshold,
		clientRebalanceInterval:   defaultRebalanceInterval,
		healthcheckTimeout:        defaultHealthcheckTimeout,
		nodeDialTimeout:           defaultDialTimeout,
		nodeStreamTimeout:         defaultStreamTimeout,
	}

	return params
}

// New creates connection pool using simple set of endpoints and parameters.
//
// See also [pool.DefaultOptions] and [pool.NewFlatNodeParams] for details.
//
// Returned errors:
//   - [neofscrypto.ErrIncorrectSigner]
func New(endpoints []NodeParam, signer user.Signer, options InitParameters) (*Pool, error) {
	if len(endpoints) == 0 {
		return nil, errors.New("empty endpoints")
	}

	options.nodeParams = endpoints
	options.signer = signer

	return NewPool(options)
}

// NewFlatNodeParams converts endpoints to appropriate NodeParam.
// It is useful for situations where all endpoints are equivalent.
func NewFlatNodeParams(endpoints []string) []NodeParam {
	if len(endpoints) == 0 {
		return nil
	}

	params := make([]NodeParam, 0, len(endpoints))

	for _, addr := range endpoints {
		params = append(params, NodeParam{
			priority: 1,
			address:  addr,
			weight:   1,
		})
	}

	return params
}

// NewPool creates connection pool using parameters.
//
// Returned errors:
//   - [neofscrypto.ErrIncorrectSigner]
func NewPool(options InitParameters) (*Pool, error) {
	if options.signer == nil {
		return nil, fmt.Errorf("missed required parameter 'Signer'")
	}
	if options.signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return nil, fmt.Errorf("%w: expected ECDSA_DETERMINISTIC_SHA256 scheme", neofscrypto.ErrIncorrectSigner)
	}

	for _, node := range options.nodeParams {
		if err := isNodeValid(node); err != nil {
			return nil, fmt.Errorf("node: %w", err)
		}
	}

	nodesParams, err := adjustNodeParams(options.nodeParams)
	if err != nil {
		return nil, err
	}

	cache, err := newCache(defaultSessionCacheSize)
	if err != nil {
		return nil, fmt.Errorf("couldn't create cache: %w", err)
	}

	buffers := &sync.Pool{}
	buffers.New = func() any {
		b := make([]byte, defaultBufferSize)
		return &b
	}

	pool := &Pool{cache: cache, buffers: buffers}

	// we need our middleware integration in clientBuilder
	fillDefaultInitParams(&options, cache, pool.statisticMiddleware, pool.buffers)

	pool.signer = options.signer
	pool.logger = options.logger
	pool.stokenDuration = options.sessionExpirationDuration
	pool.rebalanceParams = rebalanceParameters{
		nodesParams:               nodesParams,
		nodeRequestTimeout:        options.healthcheckTimeout,
		clientRebalanceInterval:   options.clientRebalanceInterval,
		sessionExpirationDuration: options.sessionExpirationDuration,
	}
	pool.clientBuilder = options.clientBuilder
	pool.statisticCallback = options.statisticCallback

	return pool, nil
}

// Dial establishes a connection to the servers from the NeoFS network.
// It also starts a routine that checks the health of the nodes and
// updates the weights of the nodes for balancing.
// Returns an error describing failure reason.
//
// If failed, the Pool SHOULD NOT be used.
//
// See also InitParameters.SetClientRebalanceInterval.
func (p *Pool) Dial(ctx context.Context) error {
	var (
		atLeastOneHealthy bool
		err               error
	)
	inner := make([]*innerPool, len(p.rebalanceParams.nodesParams))

	for i, params := range p.rebalanceParams.nodesParams {
		clients := make([]internalClient, len(params.weights))
		for j, addr := range params.addresses {
			clients[j], err = p.clientBuilder(addr)
			if err != nil {
				if p.logger != nil {
					p.logger.Warn("failed to build client", zap.String("address", addr), zap.Error(err))
				}
				continue
			}
			if err := clients[j].dial(ctx); err != nil {
				if p.logger != nil {
					p.logger.Warn("failed to dial client", zap.String("address", addr), zap.Error(err))
				}
				continue
			}

			atLeastOneHealthy = true
		}
		sampl := newSampler(params.weights, safeRand{})

		inner[i] = &innerPool{
			sampler: sampl,
			clients: clients,
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

func fillDefaultInitParams(params *InitParameters, cache *sessionCache, statisticCallback stat.OperationCallback, buffers *sync.Pool) {
	if params.sessionExpirationDuration == 0 {
		params.sessionExpirationDuration = defaultSessionTokenExpirationDuration
	}

	if params.errorThreshold == 0 {
		params.errorThreshold = defaultErrorThreshold
	}

	if params.clientRebalanceInterval <= 0 {
		params.clientRebalanceInterval = defaultRebalanceInterval
	}

	if params.healthcheckTimeout <= 0 {
		params.healthcheckTimeout = defaultHealthcheckTimeout
	}

	if params.nodeDialTimeout <= 0 {
		params.nodeDialTimeout = defaultDialTimeout
	}

	if params.nodeStreamTimeout <= 0 {
		params.nodeStreamTimeout = defaultStreamTimeout
	}

	if params.isMissingClientBuilder() {
		params.setClientBuilder(func(addr string) (internalClient, error) {
			var prm wrapperPrm
			prm.setAddress(addr)
			prm.setSigner(params.signer)
			prm.setDialTimeout(params.nodeDialTimeout)
			prm.setStreamTimeout(params.nodeStreamTimeout)
			prm.setErrorThreshold(params.errorThreshold)
			prm.setResponseInfoCallback(func(info sdkClient.ResponseMetaInfo) error {
				cache.updateEpoch(info.Epoch())
				return nil
			})
			prm.setStatisticCallback(statisticCallback)
			prm.setBuffers(buffers)
			prm.setNodeSessionCacheSize(params.nodeSessionCacheSize)
			return newWrapper(prm)
		})
	}
}

func isNodeValid(node NodeParam) error {
	_, _, err := uriutil.Parse(node.address)
	return err
}

func adjustNodeParams(nodeParams []NodeParam) ([]*nodesParam, error) {
	if len(nodeParams) == 0 {
		return nil, errors.New("no NeoFS peers configured")
	}

	nodesParamsMap := make(map[int]*nodesParam)
	for _, param := range nodeParams {
		nodes, ok := nodesParamsMap[param.priority]
		if !ok {
			nodes = &nodesParam{priority: param.priority}
		}
		nodes.addresses = append(nodes.addresses, param.address)
		nodes.weights = append(nodes.weights, param.weight)
		nodesParamsMap[param.priority] = nodes
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

// startRebalance runs loop to monitor connection healthy status.
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
		go func(i int, _ *innerPool) {
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

	healthyChanged := &atomic.Bool{}
	wg := sync.WaitGroup{}

	for j, cli := range pool.clients {
		wg.Add(1)
		go func(j int, cli internalClient) {
			defer wg.Done()

			tctx, c := context.WithTimeout(ctx, options.nodeRequestTimeout)
			defer c()

			healthy, changed := cli.restartIfUnhealthy(tctx)
			if healthy {
				bufferWeights[j] = options.nodesParams[i].weights[j]
			} else {
				bufferWeights[j] = 0
				p.cache.DeleteByPrefix(cli.address())
				cli.ResetSessions()
			}

			if changed {
				healthyChanged.Store(true)
			}
		}(j, cli)
	}
	wg.Wait()

	if healthyChanged.Load() {
		probabilities := adjustWeights(bufferWeights)
		pool.lock.Lock()
		pool.sampler = newSampler(probabilities, safeRand{})
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

func (p *Pool) connection() (internalClient, error) {
	for _, inner := range p.innerPools {
		cp, err := inner.connection()
		if err == nil {
			return cp, nil
		}
	}

	return nil, errors.New("no healthy client")
}

func (p *innerPool) connection() (internalClient, error) {
	p.lock.RLock() // need lock because of using p.sampler
	defer p.lock.RUnlock()
	if len(p.clients) == 1 {
		cp := p.clients[0]
		if cp.isHealthy() {
			return cp, nil
		}
		return nil, errors.New("no healthy client")
	}
	attempts := 3 * len(p.clients)
	for range attempts {
		i := p.sampler.next()
		if cp := p.clients[i]; cp.isHealthy() {
			return cp, nil
		}
	}

	return nil, errors.New("no healthy client")
}

// cacheKeyForSession generates cache key for a signed session token.
// It is used with pool methods compatible with [sdkClient.Client].
func cacheKeyForSession(address string, signer neofscrypto.Signer, verb session.ObjectVerb, cnr cid.ID) string {
	return fmt.Sprintf("%s%s%d%s", address, neofscrypto.PublicKeyBytes(signer.Public()), verb, cnr)
}

func (p *Pool) checkSessionTokenErr(err error, address string, cl nodeSessionContainer) {
	if err == nil {
		return
	}

	if errors.Is(err, apistatus.ErrSessionTokenNotFound) || errors.Is(err, apistatus.ErrSessionTokenExpired) {
		p.cache.DeleteByPrefix(address)
		cl.ResetSessions()
	}
}

// RawClient returns single client instance to have possibility to work with exact one.
func (p *Pool) RawClient() (*sdkClient.Client, error) {
	conn, err := p.connection()
	if err != nil {
		return nil, err
	}

	return conn.getRawClient()
}

// Close closes the Pool and releases all the associated resources.
func (p *Pool) Close() error {
	p.cancel()
	<-p.closedCh

	es := make([]error, 0, len(p.innerPools))
	for i := range p.innerPools {
		pes := make([]error, 0, len(p.innerPools[i].clients))
		for _, c := range p.innerPools[i].clients {
			if c != nil {
				pes = append(pes, c.Close())
			}
		}
		es = append(es, errors.Join(pes...))
	}
	return errors.Join(es...)
}

func (p *Pool) sdkClient() (*sdkClientWrapper, error) {
	conn, err := p.connection()
	if err != nil {
		return nil, fmt.Errorf("connection: %w", err)
	}

	cl, err := conn.getClient()
	if err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}

	return &sdkClientWrapper{
		sdkClientInterface: cl,
		nodeSession:        conn,
		addr:               conn.address(),
	}, nil
}

func (p *Pool) statisticMiddleware(nodeKey []byte, endpoint string, method stat.Method, duration time.Duration, err error) {
	if p.statisticCallback != nil {
		p.statisticCallback(nodeKey, endpoint, method, duration, err)
	}
}
