package pool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	sdkClient "github.com/nspcc-dev/neofs-sdk-go/client"
	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/relations"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

var (
	relationsGet = relations.Get
)

type sdkClientWrapper struct {
	*sdkClient.Client

	nodeSession nodeSessionContainer
	addr        string
}

// nodeSessionContainer represents storage for a session token. It contains only basics session info: id, pub key, expiration.
// This token is used for the final session tokens creation for specific verb. This token is 1:1 for each node.
// Should be stored until token not expired.
type nodeSessionContainer interface {
	SetNodeSession(*session.Object)
	GetNodeSession() *session.Object
}

// client represents virtual connection to the single NeoFS network endpoint from which Pool is formed.
// This interface is expected to have exactly one production implementation - clientWrapper.
// Others are expected to be for test purposes only.
type internalClient interface {
	// see clientWrapper.balanceGet.
	balanceGet(context.Context, PrmBalanceGet) (accounting.Decimal, error)
	// see clientWrapper.containerPut.
	containerPut(context.Context, container.Container, user.Signer, PrmContainerPut) (cid.ID, error)
	// see clientWrapper.containerGet.
	containerGet(context.Context, cid.ID) (container.Container, error)
	// see clientWrapper.containerList.
	containerList(context.Context, user.ID) ([]cid.ID, error)
	// see clientWrapper.containerDelete.
	containerDelete(context.Context, cid.ID, neofscrypto.Signer, PrmContainerDelete) error
	// see clientWrapper.containerEACL.
	containerEACL(context.Context, cid.ID) (eacl.Table, error)
	// see clientWrapper.containerSetEACL.
	containerSetEACL(context.Context, eacl.Table, user.Signer, PrmContainerSetEACL) error
	// see clientWrapper.endpointInfo.
	endpointInfo(context.Context, prmEndpointInfo) (netmap.NodeInfo, error)
	// see clientWrapper.networkInfo.
	networkInfo(context.Context, prmNetworkInfo) (netmap.NetworkInfo, error)
	// see clientWrapper.objectPut.
	objectPut(context.Context, user.Signer, PrmObjectPut) (oid.ID, error)
	// see clientWrapper.objectDelete.
	objectDelete(context.Context, cid.ID, oid.ID, user.Signer, PrmObjectDelete) error
	// see clientWrapper.objectGet.
	objectGet(context.Context, cid.ID, oid.ID, neofscrypto.Signer, PrmObjectGet) (ResGetObject, error)
	// see clientWrapper.objectHead.
	objectHead(context.Context, cid.ID, oid.ID, user.Signer, PrmObjectHead) (object.Object, error)
	// see clientWrapper.objectRange.
	objectRange(context.Context, cid.ID, oid.ID, uint64, uint64, neofscrypto.Signer, PrmObjectRange) (ResObjectRange, error)
	// see clientWrapper.objectSearch.
	objectSearch(context.Context, cid.ID, user.Signer, PrmObjectSearch) (ResObjectSearch, error)
	// see clientWrapper.sessionCreate.
	sessionCreate(context.Context, user.Signer, prmCreateSession) (resCreateSession, error)

	clientStatus
	statisticUpdater
	nodeSessionContainer

	// see clientWrapper.dial.
	dial(ctx context.Context) error
	// see clientWrapper.restartIfUnhealthy.
	restartIfUnhealthy(ctx context.Context) (bool, bool)

	getClient() (*sdkClient.Client, error)
}

type statisticUpdater interface {
	updateErrorRate(err error)
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

	mu                sync.RWMutex // protect counters
	currentErrorCount uint32
	overallErrorCount uint64
}

func newClientStatusMonitor(addr string, errorThreshold uint32) clientStatusMonitor {
	return clientStatusMonitor{
		addr:           addr,
		healthy:        atomic.NewBool(true),
		errorThreshold: errorThreshold,
	}
}

// clientWrapper is used by default, alternative implementations are intended for testing purposes only.
type clientWrapper struct {
	clientMutex sync.RWMutex
	client      *sdkClient.Client
	prm         wrapperPrm

	clientStatusMonitor
	statisticCallback stat.OperationCallback

	nodeSessionMutex sync.RWMutex
	nodeSession      *session.Object

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

// getNewClient returns a new [sdkClient.Client] instance using internal parameters.
func (x *wrapperPrm) getNewClient(statisticCallback stat.OperationCallback) (*sdkClient.Client, error) {
	var prmInit sdkClient.PrmInit
	prmInit.SetResponseInfoCallback(x.responseInfoCallback)
	prmInit.SetStatisticCallback(statisticCallback)

	return sdkClient.New(prmInit)
}

// newWrapper creates a clientWrapper that implements the client interface.
func newWrapper(prm wrapperPrm) (*clientWrapper, error) {
	res := &clientWrapper{
		clientStatusMonitor: newClientStatusMonitor(prm.address, prm.errorThreshold),
		statisticCallback:   prm.statisticCallback,
	}

	oldCallBack := prm.responseInfoCallback
	prm.setResponseInfoCallback(func(info sdkClient.ResponseMetaInfo) error {
		newEpoch := info.Epoch()
		if newEpoch > res.epoch.Load() {
			res.epoch.Store(newEpoch)
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
	if _, err := c.endpointInfo(ctx, prmEndpointInfo{}); err == nil {
		return true, false
	} else if !errors.Is(err, errPoolClientUnhealthy) {
		wasHealthy = true
	}

	cl, err := c.prm.getNewClient(c.statisticMiddleware)
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

	if _, err := cl.EndpointInfo(ctx, sdkClient.PrmEndpointInfo{}); err != nil {
		c.setUnhealthy()
		return false, wasHealthy
	}

	c.setHealthy()
	return true, !wasHealthy
}

func (c *clientWrapper) getClient() (*sdkClient.Client, error) {
	c.clientMutex.RLock()
	defer c.clientMutex.RUnlock()
	if c.isHealthy() {
		return c.client, nil
	}
	return nil, errPoolClientUnhealthy
}

// balanceGet invokes sdkClient.BalanceGet parse response status to error and return result as is.
func (c *clientWrapper) balanceGet(ctx context.Context, prm PrmBalanceGet) (accounting.Decimal, error) {
	cl, err := c.getClient()
	if err != nil {
		return accounting.Decimal{}, err
	}

	var cliPrm sdkClient.PrmBalanceGet
	cliPrm.SetAccount(prm.account)

	res, err := cl.BalanceGet(ctx, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return accounting.Decimal{}, fmt.Errorf("balance get on client: %w", err)
	}

	return res, nil
}

// containerPut invokes sdkClient.ContainerPut parse response status to error and return result as is.
// It also waits for the container to appear on the network.
func (c *clientWrapper) containerPut(ctx context.Context, cont container.Container, signer user.Signer, prm PrmContainerPut) (cid.ID, error) {
	cl, err := c.getClient()
	if err != nil {
		return cid.ID{}, err
	}

	idCnr, err := cl.ContainerPut(ctx, cont, signer, prm.prmClient)
	c.updateErrorRate(err)
	if err != nil {
		return cid.ID{}, fmt.Errorf("container put on client: %w", err)
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	err = waitForContainerPresence(ctx, c, idCnr, &prm.waitParams)
	c.updateErrorRate(err)
	if err != nil {
		return cid.ID{}, fmt.Errorf("wait container presence on client: %w", err)
	}

	return idCnr, nil
}

// containerGet invokes sdkClient.ContainerGet parse response status to error and return result as is.
func (c *clientWrapper) containerGet(ctx context.Context, cnrID cid.ID) (container.Container, error) {
	cl, err := c.getClient()
	if err != nil {
		return container.Container{}, err
	}

	res, err := cl.ContainerGet(ctx, cnrID, sdkClient.PrmContainerGet{})
	c.updateErrorRate(err)
	if err != nil {
		return container.Container{}, fmt.Errorf("container get on client: %w", err)
	}

	return res, nil
}

// containerList invokes sdkClient.ContainerList parse response status to error and return result as is.
func (c *clientWrapper) containerList(ctx context.Context, ownerID user.ID) ([]cid.ID, error) {
	cl, err := c.getClient()
	if err != nil {
		return nil, err
	}

	res, err := cl.ContainerList(ctx, ownerID, sdkClient.PrmContainerList{})
	c.updateErrorRate(err)
	if err != nil {
		return nil, fmt.Errorf("container list on client: %w", err)
	}
	return res, nil
}

// containerDelete invokes sdkClient.ContainerDelete parse response status to error.
// It also waits for the container to be removed from the network.
func (c *clientWrapper) containerDelete(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm PrmContainerDelete) error {
	cl, err := c.getClient()
	if err != nil {
		return err
	}

	var cliPrm sdkClient.PrmContainerDelete
	if prm.stokenSet {
		cliPrm.WithinSession(prm.stoken)
	}

	err = cl.ContainerDelete(ctx, id, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return fmt.Errorf("container delete on client: %w", err)
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	return waitForContainerRemoved(ctx, c, id, &prm.waitParams)
}

// containerEACL invokes sdkClient.ContainerEACL parse response status to error and return result as is.
func (c *clientWrapper) containerEACL(ctx context.Context, id cid.ID) (eacl.Table, error) {
	cl, err := c.getClient()
	if err != nil {
		return eacl.Table{}, err
	}

	res, err := cl.ContainerEACL(ctx, id, sdkClient.PrmContainerEACL{})
	c.updateErrorRate(err)
	if err != nil {
		return eacl.Table{}, fmt.Errorf("get eacl on client: %w", err)
	}

	return res, nil
}

// containerSetEACL invokes sdkClient.ContainerSetEACL parse response status to error.
// It also waits for the EACL to appear on the network.
func (c *clientWrapper) containerSetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm PrmContainerSetEACL) error {
	cl, err := c.getClient()
	if err != nil {
		return err
	}

	var cliPrm sdkClient.PrmContainerSetEACL
	if prm.sessionSet {
		cliPrm.WithinSession(prm.session)
	}

	err = cl.ContainerSetEACL(ctx, table, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return fmt.Errorf("set eacl on client: %w", err)
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	var cIDp cid.ID
	if cID, set := table.CID(); set {
		cIDp = cID
	}

	err = waitForEACLPresence(ctx, c, cIDp, &table, &prm.waitParams)
	c.updateErrorRate(err)
	if err != nil {
		return fmt.Errorf("wait eacl presence on client: %w", err)
	}

	return nil
}

// endpointInfo invokes sdkClient.EndpointInfo parse response status to error and return result as is.
func (c *clientWrapper) endpointInfo(ctx context.Context, _ prmEndpointInfo) (netmap.NodeInfo, error) {
	cl, err := c.getClient()
	if err != nil {
		return netmap.NodeInfo{}, err
	}

	res, err := cl.EndpointInfo(ctx, sdkClient.PrmEndpointInfo{})
	c.updateErrorRate(err)
	if err != nil {
		return netmap.NodeInfo{}, fmt.Errorf("endpoint info on client: %w", err)
	}

	return res.NodeInfo(), nil
}

// networkInfo invokes sdkClient.NetworkInfo parse response status to error and return result as is.
func (c *clientWrapper) networkInfo(ctx context.Context, _ prmNetworkInfo) (netmap.NetworkInfo, error) {
	cl, err := c.getClient()
	if err != nil {
		return netmap.NetworkInfo{}, err
	}

	res, err := cl.NetworkInfo(ctx, sdkClient.PrmNetworkInfo{})
	c.updateErrorRate(err)
	if err != nil {
		return netmap.NetworkInfo{}, fmt.Errorf("network info on client: %w", err)
	}

	return res, nil
}

// objectPut writes object to NeoFS.
func (c *clientWrapper) objectPut(ctx context.Context, signer user.Signer, prm PrmObjectPut) (oid.ID, error) {
	cl, err := c.getClient()
	if err != nil {
		return oid.ID{}, err
	}

	var cliPrm sdkClient.PrmObjectPutInit
	cliPrm.SetCopiesNumber(prm.copiesNumber)
	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}
	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	wObj, err := cl.ObjectPutInit(ctx, prm.hdr, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return oid.ID{}, fmt.Errorf("init writing on API client: %w", err)
	}

	sz := prm.hdr.PayloadSize()

	if data := prm.hdr.Payload(); len(data) > 0 {
		if prm.payload != nil {
			prm.payload = io.MultiReader(bytes.NewReader(data), prm.payload)
		} else {
			prm.payload = bytes.NewReader(data)
			sz = uint64(len(data))
		}
	}

	if err = writePayload(wObj, prm.payload, sz); err != nil {
		c.updateErrorRate(err)

		return oid.ID{}, fmt.Errorf("writePayload: %w", err)
	}

	err = wObj.Close()
	c.updateErrorRate(err)
	if err != nil { // here err already carries both status and client errors
		return oid.ID{}, fmt.Errorf("client failure: %w", err)
	}

	return wObj.GetResult().StoredObjectID(), nil
}

func writePayload(wObj io.Writer, payload io.Reader, sz uint64) error {
	if payload == nil || wObj == nil {
		return nil
	}

	const defaultBufferSizePut = 3 << 20 // configure?
	if sz == 0 || sz > defaultBufferSizePut {
		sz = defaultBufferSizePut
	}

	buf := make([]byte, sz)

	var (
		n   int
		err error
	)

	for {
		n, err = payload.Read(buf)

		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("read payload: %w", err)
			}
		}

		if n == 0 {
			break
		}

		if _, err = wObj.Write(buf[:n]); err != nil {
			return fmt.Errorf("write payload: %w", err)
		}
	}

	return nil
}

// objectDelete invokes sdkClient.ObjectDelete parse response status to error.
func (c *clientWrapper) objectDelete(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectDelete) error {
	cl, err := c.getClient()
	if err != nil {
		return err
	}

	var cliPrm sdkClient.PrmObjectDelete
	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	_, err = cl.ObjectDelete(ctx, containerID, objectID, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return fmt.Errorf("delete object on client: %w", err)
	}
	return nil
}

// objectGet returns header and reader for object.
func (c *clientWrapper) objectGet(ctx context.Context, containerID cid.ID, objectID oid.ID, signer neofscrypto.Signer, prm PrmObjectGet) (ResGetObject, error) {
	cl, err := c.getClient()
	if err != nil {
		return ResGetObject{}, err
	}

	var cliPrm sdkClient.PrmObjectGet
	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	var res ResGetObject

	hdr, rObj, err := cl.ObjectGetInit(ctx, containerID, objectID, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return ResGetObject{}, fmt.Errorf("init object reading on client: %w", err)
	}

	res.Header = hdr
	res.Payload = &objectReadCloser{
		reader: rObj,
	}

	return res, nil
}

// objectHead invokes sdkClient.ObjectHead parse response status to error and return result as is.
func (c *clientWrapper) objectHead(ctx context.Context, containerID cid.ID, objectID oid.ID, signer user.Signer, prm PrmObjectHead) (object.Object, error) {
	cl, err := c.getClient()
	if err != nil {
		return object.Object{}, err
	}

	var cliPrm sdkClient.PrmObjectHead
	if prm.raw {
		cliPrm.MarkRaw()
	}

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	var obj object.Object

	res, err := cl.ObjectHead(ctx, containerID, objectID, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return obj, fmt.Errorf("read object header via client: %w", err)
	}
	if !res.ReadHeader(&obj) {
		return obj, errors.New("missing object header in response")
	}

	return obj, nil
}

// objectRange returns object range reader.
func (c *clientWrapper) objectRange(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, signer neofscrypto.Signer, prm PrmObjectRange) (ResObjectRange, error) {
	cl, err := c.getClient()
	if err != nil {
		return ResObjectRange{}, err
	}

	var cliPrm sdkClient.PrmObjectRange

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	res, err := cl.ObjectRangeInit(ctx, containerID, objectID, offset, length, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return ResObjectRange{}, fmt.Errorf("init payload range reading on client: %w", err)
	}

	return ResObjectRange{
		payload: res,
	}, nil
}

// objectSearch invokes sdkClient.ObjectSearchInit parse response status to error and return result as is.
func (c *clientWrapper) objectSearch(ctx context.Context, containerID cid.ID, signer user.Signer, prm PrmObjectSearch) (ResObjectSearch, error) {
	cl, err := c.getClient()
	if err != nil {
		return ResObjectSearch{}, err
	}

	var cliPrm sdkClient.PrmObjectSearch
	cliPrm.SetFilters(prm.filters)

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	res, err := cl.ObjectSearchInit(ctx, containerID, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return ResObjectSearch{}, fmt.Errorf("init object searching on client: %w", err)
	}

	return ResObjectSearch{r: res}, nil
}

// sessionCreate invokes sdkClient.SessionCreate parse response status to error and return result as is.
func (c *clientWrapper) sessionCreate(ctx context.Context, signer user.Signer, prm prmCreateSession) (resCreateSession, error) {
	cl, err := c.getClient()
	if err != nil {
		return resCreateSession{}, err
	}

	var cliPrm sdkClient.PrmSessionCreate
	cliPrm.SetExp(prm.exp)

	res, err := cl.SessionCreate(ctx, signer, cliPrm)
	c.updateErrorRate(err)
	if err != nil {
		return resCreateSession{}, fmt.Errorf("session creation on client: %w", err)
	}

	return resCreateSession{
		id:         res.ID(),
		sessionKey: res.PublicKey(),
	}, nil
}

func (c *clientWrapper) SetNodeSession(token *session.Object) {
	c.nodeSessionMutex.Lock()
	c.nodeSession = token
	c.nodeSessionMutex.Unlock()
}

func (c *clientWrapper) GetNodeSession() *session.Object {
	c.nodeSessionMutex.RLock()
	defer c.nodeSessionMutex.RUnlock()

	if c.nodeSession == nil {
		return nil
	}

	if c.nodeSession.ExpiredAt(c.epoch.Load()) {
		return nil
	}

	token := *c.nodeSession
	return &token
}

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

// WaitParams contains parameters used in polling is a something applied on NeoFS network.
type WaitParams struct {
	timeout      time.Duration
	pollInterval time.Duration
}

// SetTimeout specifies the time to wait for the operation to complete.
func (x *WaitParams) SetTimeout(timeout time.Duration) {
	x.timeout = timeout
}

// SetPollInterval specifies the interval, once it will check the completion of the operation.
func (x *WaitParams) SetPollInterval(tick time.Duration) {
	x.pollInterval = tick
}

func (x *WaitParams) setDefaults() {
	x.timeout = 120 * time.Second
	x.pollInterval = 5 * time.Second
}

// checkForPositive panics if any of the wait params isn't positive.
func (x *WaitParams) checkForPositive() {
	if x.timeout <= 0 || x.pollInterval <= 0 {
		panic("all wait params must be positive")
	}
}

type prmContext struct {
	defaultSession bool
	verb           session.ObjectVerb
	cnr            cid.ID

	objSet bool
	objs   []oid.ID
}

func (x *prmContext) useDefaultSession() {
	x.defaultSession = true
}

func (x *prmContext) useContainer(cnr cid.ID) {
	x.cnr = cnr
}

func (x *prmContext) useObjects(ids []oid.ID) {
	x.objs = ids
	x.objSet = true
}

func (x *prmContext) useVerb(verb session.ObjectVerb) {
	x.verb = verb
}

type prmCommon struct {
	signer user.Signer
	btoken *bearer.Token
	stoken *session.Object
}

// UseSigner specifies private signer to sign the requests.
// If signer is not provided, then Pool default signer is used.
func (x *prmCommon) UseSigner(signer user.Signer) {
	x.signer = signer
}

// UseBearer attaches bearer token to be used for the operation.
func (x *prmCommon) UseBearer(token bearer.Token) {
	x.btoken = &token
}

// UseSession specifies session within which operation should be performed.
func (x *prmCommon) UseSession(token session.Object) {
	x.stoken = &token
}

// PrmObjectPut groups parameters of PutObject operation.
type PrmObjectPut struct {
	prmCommon

	hdr object.Object

	payload io.Reader

	copiesNumber uint32
}

// SetHeader specifies header of the object.
func (x *PrmObjectPut) SetHeader(hdr object.Object) {
	x.hdr = hdr
}

// SetPayload specifies payload of the object.
func (x *PrmObjectPut) SetPayload(payload io.Reader) {
	x.payload = payload
}

// SetCopiesNumber sets number of object copies that is enough to consider put successful.
// Zero means using default behavior.
func (x *PrmObjectPut) SetCopiesNumber(copiesNumber uint32) {
	x.copiesNumber = copiesNumber
}

// PrmObjectDelete groups parameters of DeleteObject operation.
type PrmObjectDelete struct {
	prmCommon
}

// PrmObjectGet groups parameters of GetObject operation.
type PrmObjectGet struct {
	prmCommon
}

// PrmObjectHead groups parameters of HeadObject operation.
type PrmObjectHead struct {
	prmCommon

	raw bool
}

// MarkRaw marks an intent to read physically stored object.
func (x *PrmObjectHead) MarkRaw() {
	x.raw = true
}

// PrmObjectRange groups parameters of RangeObject operation.
type PrmObjectRange struct {
	prmCommon
}

// PrmObjectSearch groups parameters of SearchObjects operation.
type PrmObjectSearch struct {
	prmCommon

	filters object.SearchFilters
}

// SetFilters specifies filters by which to select objects.
func (x *PrmObjectSearch) SetFilters(filters object.SearchFilters) {
	x.filters = filters
}

// PrmContainerPut groups parameters of PutContainer operation.
type PrmContainerPut struct {
	prmClient sdkClient.PrmContainerPut

	waitParams    WaitParams
	waitParamsSet bool
}

// WithinSession specifies session to be used as a parameter of the base
// client's operation.
//
// See github.com/nspcc-dev/neofs-sdk-go/client.PrmContainerPut.WithinSession.
func (x *PrmContainerPut) WithinSession(s session.Container) {
	x.prmClient.WithinSession(s)
}

// SetWaitParams specifies timeout params to complete operation.
// If not provided the default one will be used.
// Panics if any of the wait params isn't positive.
func (x *PrmContainerPut) SetWaitParams(waitParams WaitParams) {
	waitParams.checkForPositive()
	x.waitParams = waitParams
	x.waitParamsSet = true
}

// PrmContainerDelete groups parameters of DeleteContainer operation.
type PrmContainerDelete struct {
	stoken    session.Container
	stokenSet bool

	waitParams    WaitParams
	waitParamsSet bool
}

// SetSessionToken specifies session within which operation should be performed.
func (x *PrmContainerDelete) SetSessionToken(token session.Container) {
	x.stoken = token
	x.stokenSet = true
}

// SetWaitParams specifies timeout params to complete operation.
// If not provided the default one will be used.
// Panics if any of the wait params isn't positive.
func (x *PrmContainerDelete) SetWaitParams(waitParams WaitParams) {
	waitParams.checkForPositive()
	x.waitParams = waitParams
	x.waitParamsSet = true
}

// PrmContainerSetEACL groups parameters of SetEACL operation.
type PrmContainerSetEACL struct {
	sessionSet bool
	session    session.Container

	waitParams    WaitParams
	waitParamsSet bool
}

// WithinSession specifies session to be used as a parameter of the base
// client's operation.
//
// See github.com/nspcc-dev/neofs-sdk-go/client.PrmContainerSetEACL.WithinSession.
func (x *PrmContainerSetEACL) WithinSession(s session.Container) {
	x.session = s
	x.sessionSet = true
}

// SetWaitParams specifies timeout params to complete operation.
// If not provided the default one will be used.
// Panics if any of the wait params isn't positive.
func (x *PrmContainerSetEACL) SetWaitParams(waitParams WaitParams) {
	waitParams.checkForPositive()
	x.waitParams = waitParams
	x.waitParamsSet = true
}

// PrmBalanceGet groups parameters of Balance operation.
type PrmBalanceGet struct {
	account user.ID
}

// SetAccount specifies identifier of the NeoFS account for which the balance is requested.
func (x *PrmBalanceGet) SetAccount(id user.ID) {
	x.account = id
}

// prmEndpointInfo groups parameters of sessionCreate operation.
type prmCreateSession struct {
	exp uint64
}

// setExp sets number of the last NeoFS epoch in the lifetime of the session after which it will be expired.
func (x *prmCreateSession) setExp(exp uint64) {
	x.exp = exp
}

// prmEndpointInfo groups parameters of endpointInfo operation.
type prmEndpointInfo struct{}

// prmNetworkInfo groups parameters of networkInfo operation.
type prmNetworkInfo struct{}

// resCreateSession groups resulting values of sessionCreate operation.
type resCreateSession struct {
	id []byte

	sessionKey []byte
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

	pool := &Pool{cache: cache}

	// we need our middleware integration in clientBuilder
	fillDefaultInitParams(&options, cache, pool.statisticMiddleware)

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

			var st session.Object
			err := initSessionForDuration(ctx, &st, clients[j], p.rebalanceParams.sessionExpirationDuration, p.signer)
			if err != nil {
				clients[j].setUnhealthy()
				if p.logger != nil {
					p.logger.Warn("failed to create neofs session token for client",
						zap.String("address", addr), zap.Error(err))
				}
				continue
			}

			_ = p.cache.Put(formCacheKey(addr, p.signer), st)
			atLeastOneHealthy = true
		}
		source := rand.NewSource(time.Now().UnixNano())
		sampl := newSampler(params.weights, source)

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

func fillDefaultInitParams(params *InitParameters, cache *sessionCache, statisticCallback stat.OperationCallback) {
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
			return newWrapper(prm)
		})
	}
}

func isNodeValid(node NodeParam) error {
	_, _, err := client.ParseURI(node.address)
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

	healthyChanged := atomic.NewBool(false)
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
				cli.SetNodeSession(nil)
			}

			if changed {
				healthyChanged.Store(true)
			}
		}(j, cli)
	}
	wg.Wait()

	if healthyChanged.Load() {
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
	for k := 0; k < attempts; k++ {
		i := p.sampler.Next()
		if cp := p.clients[i]; cp.isHealthy() {
			return cp, nil
		}
	}

	return nil, errors.New("no healthy client")
}

func formCacheKey(address string, signer neofscrypto.Signer) string {
	return address + string(neofscrypto.PublicKeyBytes(signer.Public()))
}

// cacheKeyForSession generates cache key for a signed session token.
// It is used with pool methods compatible with [sdkClient.Client].
func cacheKeyForSession(address string, signer neofscrypto.Signer, verb session.ObjectVerb, cnr cid.ID) string {
	return fmt.Sprintf("%s%s%d%s", address, neofscrypto.PublicKeyBytes(signer.Public()), verb, cnr)
}

func (p *Pool) checkSessionTokenErr(err error, address string, cl internalClient) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, apistatus.ErrSessionTokenNotFound) || errors.Is(err, apistatus.ErrSessionTokenExpired) {
		p.cache.DeleteByPrefix(address)
		cl.SetNodeSession(nil)
		return true
	}

	return false
}

func initSessionForDuration(ctx context.Context, dst *session.Object, c internalClient, dur uint64, signer user.Signer) error {
	ni, err := c.networkInfo(ctx, prmNetworkInfo{})
	if err != nil {
		return err
	}

	epoch := ni.CurrentEpoch()

	var exp uint64
	if math.MaxUint64-epoch < dur {
		exp = math.MaxUint64
	} else {
		exp = epoch + dur
	}
	var prm prmCreateSession
	prm.setExp(exp)

	res, err := c.sessionCreate(ctx, signer, prm)
	if err != nil {
		return err
	}

	var id uuid.UUID

	err = id.UnmarshalBinary(res.id)
	if err != nil {
		return fmt.Errorf("invalid session token ID: %w", err)
	}

	var key neofsecdsa.PublicKey

	err = key.Decode(res.sessionKey)
	if err != nil {
		return fmt.Errorf("invalid public session key: %w", err)
	}

	dst.SetID(id)
	dst.SetAuthKey(&key)
	dst.SetExp(exp)

	return nil
}

type callContext struct {
	// base context for RPC
	context.Context

	client internalClient

	// client endpoint
	endpoint string

	// request signer
	signer user.Signer

	// flag to open default session if session token is missing
	sessionDefault bool
	sessionTarget  func(session.Object)
	sessionVerb    session.ObjectVerb
	sessionCnr     cid.ID
	sessionObjSet  bool
	sessionObjs    []oid.ID
}

func (p *Pool) initCallContext(ctx *callContext, cfg prmCommon, prmCtx prmContext) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	ctx.signer = cfg.signer
	if ctx.signer == nil {
		// use pool signer if caller didn't specify its own
		ctx.signer = p.signer
	}

	ctx.endpoint = cp.address()
	ctx.client = cp

	if ctx.sessionTarget != nil && cfg.stoken != nil {
		ctx.sessionTarget(*cfg.stoken)
	}

	// note that we don't override session provided by the caller
	ctx.sessionDefault = cfg.stoken == nil && prmCtx.defaultSession
	if ctx.sessionDefault {
		ctx.sessionVerb = prmCtx.verb
		ctx.sessionCnr = prmCtx.cnr
		ctx.sessionObjSet = prmCtx.objSet
		ctx.sessionObjs = prmCtx.objs
	}

	return err
}

// opens new session or uses cached one.
// Must be called only on initialized callContext with set sessionTarget.
func (p *Pool) openDefaultSession(ctx *callContext) error {
	cacheKey := formCacheKey(ctx.endpoint, ctx.signer)

	tok, ok := p.cache.Get(cacheKey)
	if !ok {
		// init new session
		err := initSessionForDuration(ctx, &tok, ctx.client, p.stokenDuration, ctx.signer)
		if err != nil {
			return fmt.Errorf("session API client: %w", err)
		}

		// cache the opened session
		p.cache.Put(cacheKey, tok)
	}

	tok.ForVerb(ctx.sessionVerb)
	tok.BindContainer(ctx.sessionCnr)

	if ctx.sessionObjSet {
		tok.LimitByObjects(ctx.sessionObjs...)
	}

	// sign the token
	if err := tok.Sign(ctx.signer); err != nil {
		return fmt.Errorf("sign token of the opened session: %w", err)
	}

	ctx.sessionTarget(tok)

	return nil
}

// opens default session (if sessionDefault is set), and calls f. If f returns
// session-related error then cached token is removed.
func (p *Pool) call(ctx *callContext, f func() error) error {
	var err error

	if ctx.sessionDefault {
		err = p.openDefaultSession(ctx)
		if err != nil {
			return fmt.Errorf("open default session: %w", err)
		}
	}

	err = f()
	_ = p.checkSessionTokenErr(err, ctx.endpoint, ctx.client)

	return err
}

// fillAppropriateSigner use pool signer if caller didn't specify its own.
func (p *Pool) fillAppropriateSigner(prm *prmCommon) {
	if prm.signer == nil {
		prm.signer = p.signer
	}
}

// PutObject writes an object through a remote server using NeoFS API protocol.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ObjectPutInit instead.
func (p *Pool) PutObject(ctx context.Context, prm PrmObjectPut) (oid.ID, error) {
	cnr, _ := prm.hdr.ContainerID()

	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(session.VerbObjectPut)
	prmCtx.useContainer(cnr)

	p.fillAppropriateSigner(&prm.prmCommon)

	var ctxCall callContext

	ctxCall.Context = ctx

	if err := p.initCallContext(&ctxCall, prm.prmCommon, prmCtx); err != nil {
		return oid.ID{}, fmt.Errorf("init call context: %w", err)
	}

	if ctxCall.sessionDefault {
		ctxCall.sessionTarget = prm.UseSession
		if err := p.openDefaultSession(&ctxCall); err != nil {
			return oid.ID{}, fmt.Errorf("open default session: %w", err)
		}
	}

	id, err := ctxCall.client.objectPut(ctx, prm.signer, prm)
	if err != nil {
		// removes session token from cache in case of token error
		p.checkSessionTokenErr(err, ctxCall.endpoint, ctxCall.client)
		return id, fmt.Errorf("init writing on API client: %w", err)
	}

	return id, nil
}

// DeleteObject marks an object for deletion from the container using NeoFS API protocol.
// As a marker, a special unit called a tombstone is placed in the container.
// It confirms the user's intent to delete the object, and is itself a container object.
// Explicit deletion is done asynchronously, and is generally not guaranteed.
// Deprecated: use ObjectDelete instead.
func (p *Pool) DeleteObject(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectDelete) error {
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(session.VerbObjectDelete)
	prmCtx.useContainer(containerID)

	if prm.stoken == nil { // collect phy objects only if we are about to open default session
		var tokens relations.Tokens
		tokens.Bearer = prm.btoken

		relatives, linkerID, err := relationsGet(ctx, p, containerID, objectID, tokens, prm.signer)
		if err != nil {
			return fmt.Errorf("failed to collect relatives: %w", err)
		}

		if len(relatives) != 0 {
			prmCtx.useContainer(containerID)
			objList := append(relatives, objectID)
			if linkerID != nil {
				objList = append(objList, *linkerID)
			}

			prmCtx.useObjects(objList)
		}
	}

	p.fillAppropriateSigner(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return err
	}

	return p.call(&cc, func() error {
		if err = cc.client.objectDelete(ctx, containerID, objectID, prm.signer, prm); err != nil {
			return fmt.Errorf("remove object via client: %w", err)
		}

		return nil
	})
}

// RawClient returns single client instance to have possibility to work with exact one.
func (p *Pool) RawClient() (*sdkClient.Client, error) {
	conn, err := p.connection()
	if err != nil {
		return nil, err
	}

	return conn.getClient()
}

type objectReadCloser struct {
	reader *sdkClient.PayloadReader
}

// Read implements io.Reader of the object payload.
func (x *objectReadCloser) Read(p []byte) (int, error) {
	return x.reader.Read(p)
}

// Close implements io.Closer of the object payload.
func (x *objectReadCloser) Close() error {
	return x.reader.Close()
}

// ResGetObject is designed to provide object header nad read one object payload from NeoFS system.
type ResGetObject struct {
	Header object.Object

	Payload io.ReadCloser
}

// GetObject reads object header and initiates reading an object payload through a remote server using NeoFS API protocol.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ObjectGetInit instead.
func (p *Pool) GetObject(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectGet) (ResGetObject, error) {
	p.fillAppropriateSigner(&prm.prmCommon)

	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	var res ResGetObject

	err := p.initCallContext(&cc, prm.prmCommon, prmContext{})
	if err != nil {
		return res, err
	}

	return res, p.call(&cc, func() error {
		res, err = cc.client.objectGet(ctx, containerID, objectID, prm.signer, prm)
		return err
	})
}

// HeadObject reads object header through a remote server using NeoFS API protocol.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ObjectHead instead.
func (p *Pool) HeadObject(ctx context.Context, containerID cid.ID, objectID oid.ID, prm PrmObjectHead) (object.Object, error) {
	p.fillAppropriateSigner(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	var obj object.Object

	err := p.initCallContext(&cc, prm.prmCommon, prmContext{})
	if err != nil {
		return obj, err
	}

	return obj, p.call(&cc, func() error {
		obj, err = cc.client.objectHead(ctx, containerID, objectID, prm.signer, prm)
		return err
	})
}

// ResObjectRange is designed to read payload range of one object
// from NeoFS system.
//
// Must be initialized using Pool.ObjectRange, any other
// usage is unsafe.
type ResObjectRange struct {
	payload *sdkClient.ObjectRangeReader
}

// Read implements io.Reader of the object payload.
func (x *ResObjectRange) Read(p []byte) (int, error) {
	return x.payload.Read(p)
}

// Close ends reading the payload range and returns the result of the operation
// along with the final results. Must be called after using the ResObjectRange.
func (x *ResObjectRange) Close() error {
	return x.payload.Close()
}

// ObjectRange initiates reading an object's payload range through a remote
// server using NeoFS API protocol.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ObjectRangeInit instead.
func (p *Pool) ObjectRange(ctx context.Context, containerID cid.ID, objectID oid.ID, offset, length uint64, prm PrmObjectRange) (ResObjectRange, error) {
	p.fillAppropriateSigner(&prm.prmCommon)

	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	var res ResObjectRange

	err := p.initCallContext(&cc, prm.prmCommon, prmContext{})
	if err != nil {
		return res, err
	}

	return res, p.call(&cc, func() error {
		res, err = cc.client.objectRange(ctx, containerID, objectID, offset, length, prm.signer, prm)
		return err
	})
}

// ResObjectSearch is designed to read list of object identifiers from NeoFS system.
//
// Must be initialized using Pool.SearchObjects, any other usage is unsafe.
type ResObjectSearch struct {
	r *sdkClient.ObjectListReader
}

// Read reads another list of the object identifiers.
func (x *ResObjectSearch) Read(buf []oid.ID) (int, error) {
	n, ok := x.r.Read(buf)
	if !ok {
		err := x.r.Close()
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
	_ = x.r.Close()
}

// SearchObjects initiates object selection through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching of matched objects
// is done using the ResObjectSearch. Resulting reader must be finally closed.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ObjectSearchInit instead.
func (p *Pool) SearchObjects(ctx context.Context, containerID cid.ID, prm PrmObjectSearch) (ResObjectSearch, error) {
	p.fillAppropriateSigner(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	var res ResObjectSearch

	err := p.initCallContext(&cc, prm.prmCommon, prmContext{})
	if err != nil {
		return res, err
	}

	return res, p.call(&cc, func() error {
		res, err = cc.client.objectSearch(ctx, containerID, prm.signer, prm)
		return err
	})
}

// PutContainer sends request to save container in NeoFS and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//
//	polling interval: 5s
//	waiting timeout: 120s
//
// Success can be verified by reading by identifier (see [Pool.GetContainer]).
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ContainerPut instead.
func (p *Pool) PutContainer(ctx context.Context, cont container.Container, signer user.Signer, prm PrmContainerPut) (cid.ID, error) {
	cp, err := p.connection()
	if err != nil {
		return cid.ID{}, err
	}

	return cp.containerPut(ctx, cont, signer, prm)
}

// GetContainer reads NeoFS container by ID.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ContainerGet instead.
func (p *Pool) GetContainer(ctx context.Context, id cid.ID) (container.Container, error) {
	cp, err := p.connection()
	if err != nil {
		return container.Container{}, err
	}

	return cp.containerGet(ctx, id)
}

// ListContainers requests identifiers of the account-owned containers.
// Deprecated: use ContainerList instead.
func (p *Pool) ListContainers(ctx context.Context, ownerID user.ID) ([]cid.ID, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.containerList(ctx, ownerID)
}

// DeleteContainer sends request to remove the NeoFS container and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//
//	polling interval: 5s
//	waiting timeout: 120s
//
// Success can be verified by reading by identifier (see GetContainer).
// Deprecated: use ContainerDelete instead.
func (p *Pool) DeleteContainer(ctx context.Context, id cid.ID, signer neofscrypto.Signer, prm PrmContainerDelete) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	return cp.containerDelete(ctx, id, signer, prm)
}

// GetEACL reads eACL table of the NeoFS container.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use ContainerEACL instead.
func (p *Pool) GetEACL(ctx context.Context, id cid.ID) (eacl.Table, error) {
	cp, err := p.connection()
	if err != nil {
		return eacl.Table{}, err
	}

	return cp.containerEACL(ctx, id)
}

// SetEACL sends request to update eACL table of the NeoFS container and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//
//	polling interval: 5s
//	waiting timeout: 120s
//
// Success can be verified by reading by identifier (see GetEACL).
// Deprecated: use ContainerSetEACL instead.
func (p *Pool) SetEACL(ctx context.Context, table eacl.Table, signer user.Signer, prm PrmContainerSetEACL) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	return cp.containerSetEACL(ctx, table, signer, prm)
}

// Balance requests current balance of the NeoFS account.
//
// Main return value MUST NOT be processed on an erroneous return.
// Deprecated: use BalanceGet instead.
func (p *Pool) Balance(ctx context.Context, prm PrmBalanceGet) (accounting.Decimal, error) {
	cp, err := p.connection()
	if err != nil {
		return accounting.Decimal{}, err
	}

	return cp.balanceGet(ctx, prm)
}

// waitForContainerPresence waits until the container is found on the NeoFS network.
func waitForContainerPresence(ctx context.Context, cli internalClient, cnrID cid.ID, waitParams *WaitParams) error {
	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		_, err := cli.containerGet(ctx, cnrID)
		return err == nil
	})
}

// waitForEACLPresence waits until the container eacl is applied on the NeoFS network.
func waitForEACLPresence(ctx context.Context, cli internalClient, cnrID cid.ID, table *eacl.Table, waitParams *WaitParams) error {
	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		eaclTable, err := cli.containerEACL(ctx, cnrID)
		if err == nil {
			return eacl.EqualTables(*table, eaclTable)
		}
		return false
	})
}

// waitForContainerRemoved waits until the container is removed from the NeoFS network.
func waitForContainerRemoved(ctx context.Context, cli internalClient, cnrID cid.ID, waitParams *WaitParams) error {
	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		_, err := cli.containerGet(ctx, cnrID)
		return errors.Is(err, apistatus.ErrContainerNotFound)
	})
}

// waitFor await that given condition will be met in waitParams time.
func waitFor(ctx context.Context, params *WaitParams, condition func(context.Context) bool) error {
	wctx, cancel := context.WithTimeout(ctx, params.timeout)
	defer cancel()
	ticker := time.NewTimer(params.pollInterval)
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
			if condition(ctx) {
				return nil
			}
			ticker.Reset(params.pollInterval)
		}
	}
}

// Close closes the Pool and releases all the associated resources.
func (p *Pool) Close() {
	p.cancel()
	<-p.closedCh
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
		Client:      cl,
		nodeSession: conn,
		addr:        conn.address(),
	}, nil
}

func (p *Pool) statisticMiddleware(nodeKey []byte, endpoint string, method stat.Method, duration time.Duration, err error) {
	if p.statisticCallback != nil {
		p.statisticCallback(nodeKey, endpoint, method, duration, err)
	}
}
