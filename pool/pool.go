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
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	sdkClient "github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	"github.com/nspcc-dev/neofs-sdk-go/object/address"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"go.uber.org/zap"
)

// client represents virtual connection to the single NeoFS network endpoint from which Pool is formed.
type client interface {
	balanceGet(context.Context, PrmBalanceGet) (*accounting.Decimal, error)
	containerPut(context.Context, PrmContainerPut) (*cid.ID, error)
	containerGet(context.Context, PrmContainerGet) (*container.Container, error)
	containerList(context.Context, PrmContainerList) ([]cid.ID, error)
	containerDelete(context.Context, PrmContainerDelete) error
	containerEACL(context.Context, PrmContainerEACL) (*eacl.Table, error)
	containerSetEACL(context.Context, PrmContainerSetEACL) error
	endpointInfo(context.Context, prmEndpointInfo) (*netmap.NodeInfo, error)
	networkInfo(context.Context, prmNetworkInfo) (*netmap.NetworkInfo, error)
	objectPut(context.Context, PrmObjectPut) (*oid.ID, error)
	objectDelete(context.Context, PrmObjectDelete) error
	objectGet(context.Context, PrmObjectGet) (*ResGetObject, error)
	objectHead(context.Context, PrmObjectHead) (*object.Object, error)
	objectRange(context.Context, PrmObjectRange) (*ResObjectRange, error)
	objectSearch(context.Context, PrmObjectSearch) (*ResObjectSearch, error)
	sessionCreate(context.Context, prmCreateSession) (*resCreateSession, error)
}

// clientWrapper is used by default, alternative implementations are intended for testing purposes only.
type clientWrapper struct {
	client sdkClient.Client
	key    ecdsa.PrivateKey
}

type wrapperPrm struct {
	address              string
	key                  ecdsa.PrivateKey
	timeout              time.Duration
	responseInfoCallback func(sdkClient.ResponseMetaInfo) error
}

func (x *wrapperPrm) setAddress(address string) {
	x.address = address
}

func (x *wrapperPrm) setKey(key ecdsa.PrivateKey) {
	x.key = key
}

func (x *wrapperPrm) setTimeout(timeout time.Duration) {
	x.timeout = timeout
}

func (x *wrapperPrm) setResponseInfoCallback(f func(sdkClient.ResponseMetaInfo) error) {
	x.responseInfoCallback = f
}

func newWrapper(prm wrapperPrm) (*clientWrapper, error) {
	var prmInit sdkClient.PrmInit
	prmInit.ResolveNeoFSFailures()
	prmInit.SetDefaultPrivateKey(prm.key)
	prmInit.SetResponseInfoCallback(prm.responseInfoCallback)

	res := &clientWrapper{key: prm.key}

	res.client.Init(prmInit)

	var prmDial sdkClient.PrmDial
	prmDial.SetServerURI(prm.address)
	prmDial.SetTimeout(prm.timeout)

	err := res.client.Dial(prmDial)
	if err != nil {
		return nil, fmt.Errorf("client dial: %w", err)
	}

	return res, nil
}

func (c *clientWrapper) balanceGet(ctx context.Context, prm PrmBalanceGet) (*accounting.Decimal, error) {
	var cliPrm sdkClient.PrmBalanceGet
	cliPrm.SetAccount(prm.ownerID)

	res, err := c.client.BalanceGet(ctx, cliPrm)
	if err != nil {
		return nil, err
	}
	return res.Amount(), nil
}

func (c *clientWrapper) containerPut(ctx context.Context, prm PrmContainerPut) (*cid.ID, error) {
	var cliPrm sdkClient.PrmContainerPut
	cliPrm.SetContainer(prm.cnr)

	res, err := c.client.ContainerPut(ctx, cliPrm)
	if err != nil {
		return nil, err
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	return res.ID(), waitForContainerPresence(ctx, c, res.ID(), &prm.waitParams)
}

func (c *clientWrapper) containerGet(ctx context.Context, prm PrmContainerGet) (*container.Container, error) {
	var cliPrm sdkClient.PrmContainerGet
	cliPrm.SetContainer(prm.cnrID)

	res, err := c.client.ContainerGet(ctx, cliPrm)
	if err != nil {
		return nil, err
	}
	return res.Container(), nil
}

func (c *clientWrapper) containerList(ctx context.Context, prm PrmContainerList) ([]cid.ID, error) {
	var cliPrm sdkClient.PrmContainerList
	cliPrm.SetAccount(prm.ownerID)

	res, err := c.client.ContainerList(ctx, cliPrm)
	if err != nil {
		return nil, err
	}
	return res.Containers(), nil
}

func (c *clientWrapper) containerDelete(ctx context.Context, prm PrmContainerDelete) error {
	var cliPrm sdkClient.PrmContainerDelete
	cliPrm.SetContainer(prm.cnrID)
	if prm.stokenSet {
		cliPrm.SetSessionToken(prm.stoken)
	}

	if _, err := c.client.ContainerDelete(ctx, cliPrm); err != nil {
		return err
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	return waitForContainerRemoved(ctx, c, &prm.cnrID, &prm.waitParams)
}

func (c *clientWrapper) containerEACL(ctx context.Context, prm PrmContainerEACL) (*eacl.Table, error) {
	var cliPrm sdkClient.PrmContainerEACL
	cliPrm.SetContainer(prm.cnrID)

	res, err := c.client.ContainerEACL(ctx, cliPrm)
	if err != nil {
		return nil, err
	}
	return res.Table(), nil
}

func (c *clientWrapper) containerSetEACL(ctx context.Context, prm PrmContainerSetEACL) error {
	var cliPrm sdkClient.PrmContainerSetEACL
	cliPrm.SetTable(prm.table)

	if _, err := c.client.ContainerSetEACL(ctx, cliPrm); err != nil {
		return err
	}

	if !prm.waitParamsSet {
		prm.waitParams.setDefaults()
	}

	var cIDp *cid.ID
	if cID, set := prm.table.CID(); set {
		cIDp = &cID
	}

	return waitForEACLPresence(ctx, c, cIDp, &prm.table, &prm.waitParams)
}

func (c *clientWrapper) endpointInfo(ctx context.Context, _ prmEndpointInfo) (*netmap.NodeInfo, error) {
	res, err := c.client.EndpointInfo(ctx, sdkClient.PrmEndpointInfo{})
	if err != nil {
		return nil, err
	}
	return res.NodeInfo(), nil
}

func (c *clientWrapper) networkInfo(ctx context.Context, _ prmNetworkInfo) (*netmap.NetworkInfo, error) {
	res, err := c.client.NetworkInfo(ctx, sdkClient.PrmNetworkInfo{})
	if err != nil {
		return nil, err
	}
	return res.Info(), nil
}

func (c *clientWrapper) objectPut(ctx context.Context, prm PrmObjectPut) (*oid.ID, error) {
	var cliPrm sdkClient.PrmObjectPutInit
	wObj, err := c.client.ObjectPutInit(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("init writing on API client: %w", err)
	}

	if prm.stoken != nil {
		wObj.WithinSession(*prm.stoken)
	}
	if prm.key != nil {
		wObj.UseKey(*prm.key)
	}

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
		return nil, fmt.Errorf("client failure: %w", err)
	}

	var id oid.ID

	if !res.ReadStoredObjectID(&id) {
		return nil, errors.New("missing ID of the stored object")
	}

	return &id, nil
}

func (c *clientWrapper) objectDelete(ctx context.Context, prm PrmObjectDelete) error {
	var cliPrm sdkClient.PrmObjectDelete

	if cnr, set := prm.addr.ContainerID(); set {
		cliPrm.FromContainer(cnr)
	}

	if obj, set := prm.addr.ObjectID(); set {
		cliPrm.ByID(obj)
	}

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	if prm.key != nil {
		cliPrm.UseKey(*prm.key)
	}
	_, err := c.client.ObjectDelete(ctx, cliPrm)
	return err
}

func (c *clientWrapper) objectGet(ctx context.Context, prm PrmObjectGet) (*ResGetObject, error) {
	var cliPrm sdkClient.PrmObjectGet

	if cnr, set := prm.addr.ContainerID(); set {
		cliPrm.FromContainer(cnr)
	}

	if obj, set := prm.addr.ObjectID(); set {
		cliPrm.ByID(obj)
	}

	if cnr, set := prm.addr.ContainerID(); set {
		cliPrm.FromContainer(cnr)
	}

	if obj, set := prm.addr.ObjectID(); set {
		cliPrm.ByID(obj)
	}

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	var res ResGetObject

	rObj, err := c.client.ObjectGetInit(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("init object reading on client: %w", err)
	}

	if prm.key != nil {
		rObj.UseKey(*prm.key)
	}

	if !rObj.ReadHeader(&res.Header) {
		_, err = rObj.Close()
		return nil, fmt.Errorf("read header: %w", err)
	}

	res.Payload = (*objectReadCloser)(rObj)

	return &res, nil
}

func (c *clientWrapper) objectHead(ctx context.Context, prm PrmObjectHead) (*object.Object, error) {
	var cliPrm sdkClient.PrmObjectHead

	if cnr, set := prm.addr.ContainerID(); set {
		cliPrm.FromContainer(cnr)
	}

	if obj, set := prm.addr.ObjectID(); set {
		cliPrm.ByID(obj)
	}

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	if prm.key != nil {
		cliPrm.UseKey(*prm.key)
	}

	var obj object.Object

	res, err := c.client.ObjectHead(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("read object header via client: %w", err)
	}
	if !res.ReadHeader(&obj) {
		return nil, errors.New("missing object header in response")
	}

	return &obj, nil
}

func (c *clientWrapper) objectRange(ctx context.Context, prm PrmObjectRange) (*ResObjectRange, error) {
	var cliPrm sdkClient.PrmObjectRange

	cliPrm.SetOffset(prm.off)
	cliPrm.SetLength(prm.ln)

	if cnr, set := prm.addr.ContainerID(); set {
		cliPrm.FromContainer(cnr)
	}

	if obj, set := prm.addr.ObjectID(); set {
		cliPrm.ByID(obj)
	}

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	res, err := c.client.ObjectRangeInit(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("init payload range reading on client: %w", err)
	}
	if prm.key != nil {
		res.UseKey(*prm.key)
	}

	return &ResObjectRange{payload: res}, nil
}

func (c *clientWrapper) objectSearch(ctx context.Context, prm PrmObjectSearch) (*ResObjectSearch, error) {
	var cliPrm sdkClient.PrmObjectSearch

	cliPrm.InContainer(prm.cnrID)
	cliPrm.SetFilters(prm.filters)

	if prm.stoken != nil {
		cliPrm.WithinSession(*prm.stoken)
	}

	if prm.btoken != nil {
		cliPrm.WithBearerToken(*prm.btoken)
	}

	res, err := c.client.ObjectSearchInit(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("init object searching on client: %w", err)
	}
	if prm.key != nil {
		res.UseKey(*prm.key)
	}

	return &ResObjectSearch{r: res}, nil
}

func (c *clientWrapper) sessionCreate(ctx context.Context, prm prmCreateSession) (*resCreateSession, error) {
	var cliPrm sdkClient.PrmSessionCreate
	cliPrm.SetExp(prm.exp)

	res, err := c.client.SessionCreate(ctx, cliPrm)
	if err != nil {
		return nil, fmt.Errorf("session creation on client: %w", err)
	}

	return &resCreateSession{
		id:         res.ID(),
		sessionKey: res.PublicKey(),
	}, nil
}

// InitParameters contains values used to initialize connection Pool.
type InitParameters struct {
	key                       *ecdsa.PrivateKey
	logger                    *zap.Logger
	nodeDialTimeout           time.Duration
	healthcheckTimeout        time.Duration
	clientRebalanceInterval   time.Duration
	sessionExpirationDuration uint64
	nodeParams                []NodeParam

	clientBuilder func(endpoint string) (client, error)
}

// SetKey specifies default key to be used for the protocol communication by default.
func (x *InitParameters) SetKey(key *ecdsa.PrivateKey) {
	x.key = key
}

// SetLogger specifies logger.
func (x *InitParameters) SetLogger(logger *zap.Logger) {
	x.logger = logger
}

// SetNodeDialTimeout specifies the timeout for connection to be established.
func (x *InitParameters) SetNodeDialTimeout(timeout time.Duration) {
	x.nodeDialTimeout = timeout
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

// AddNode append information about the node to which you want to connect.
func (x *InitParameters) AddNode(nodeParam NodeParam) {
	x.nodeParams = append(x.nodeParams, nodeParam)
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

type clientPack struct {
	client  client
	healthy bool
	address string
}

type prmContext struct {
	defaultSession bool
	verb           sessionv2.ObjectSessionVerb
	addr           *address.Address
}

func (x *prmContext) useDefaultSession() {
	x.defaultSession = true
}

func (x *prmContext) useAddress(addr *address.Address) {
	x.addr = addr
}

func (x *prmContext) useVerb(verb sessionv2.ObjectSessionVerb) {
	x.verb = verb
}

type prmCommon struct {
	key    *ecdsa.PrivateKey
	btoken *bearer.Token
	stoken *session.Token
}

// UseKey specifies private key to sign the requests.
// If key is not provided, then Pool default key is used.
func (x *prmCommon) UseKey(key *ecdsa.PrivateKey) {
	x.key = key
}

// UseBearer attaches bearer token to be used for the operation.
func (x *prmCommon) UseBearer(token bearer.Token) {
	x.btoken = &token
}

// UseSession specifies session within which operation should be performed.
func (x *prmCommon) UseSession(token session.Token) {
	x.stoken = &token
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
	cnr container.Container

	waitParams    WaitParams
	waitParamsSet bool
}

// SetContainer specifies structured information about new NeoFS container.
func (x *PrmContainerPut) SetContainer(cnr container.Container) {
	x.cnr = cnr
}

// SetWaitParams specifies timeout params to complete operation.
// If not provided the default one will be used.
// Panics if any of the wait params isn't positive.
func (x *PrmContainerPut) SetWaitParams(waitParams WaitParams) {
	waitParams.checkForPositive()
	x.waitParams = waitParams
	x.waitParamsSet = true
}

// PrmContainerGet groups parameters of GetContainer operation.
type PrmContainerGet struct {
	cnrID cid.ID
}

// SetContainerID specifies identifier of the container to be read.
func (x *PrmContainerGet) SetContainerID(cnrID cid.ID) {
	x.cnrID = cnrID
}

// PrmContainerList groups parameters of ListContainers operation.
type PrmContainerList struct {
	ownerID owner.ID
}

// SetOwnerID specifies identifier of the NeoFS account to list the containers.
func (x *PrmContainerList) SetOwnerID(ownerID owner.ID) {
	x.ownerID = ownerID
}

// PrmContainerDelete groups parameters of DeleteContainer operation.
type PrmContainerDelete struct {
	cnrID cid.ID

	stoken    session.Token
	stokenSet bool

	waitParams    WaitParams
	waitParamsSet bool
}

// SetContainerID specifies identifier of the NeoFS container to be removed.
func (x *PrmContainerDelete) SetContainerID(cnrID cid.ID) {
	x.cnrID = cnrID
}

// SetSessionToken specifies session within which operation should be performed.
func (x *PrmContainerDelete) SetSessionToken(token session.Token) {
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

// PrmContainerEACL groups parameters of GetEACL operation.
type PrmContainerEACL struct {
	cnrID cid.ID
}

// SetContainerID specifies identifier of the NeoFS container to read the eACL table.
func (x *PrmContainerEACL) SetContainerID(cnrID cid.ID) {
	x.cnrID = cnrID
}

// PrmContainerSetEACL groups parameters of SetEACL operation.
type PrmContainerSetEACL struct {
	table eacl.Table

	waitParams    WaitParams
	waitParamsSet bool
}

// SetTable specifies eACL table structure to be set for the container.
func (x *PrmContainerSetEACL) SetTable(table eacl.Table) {
	x.table = table
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
	ownerID owner.ID
}

// SetOwnerID specifies identifier of the NeoFS account for which the balance is requested.
func (x *PrmBalanceGet) SetOwnerID(ownerID owner.ID) {
	x.ownerID = ownerID
}

// prmEndpointInfo groups parameters of sessionCreate operation.
type prmCreateSession struct {
	exp uint64
}

// SetExp sets number of the last NeoFS epoch in the lifetime of the session after which it will be expired.
func (x *prmCreateSession) SetExp(exp uint64) {
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
// Note that package provides some helper functions to work with status returns
// (e.g. sdkClient.IsErrContainerNotFound, sdkClient.IsErrObjectNotFound).
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
	rebalanceParams rebalanceParameters
	clientBuilder   func(endpoint string) (client, error)
	logger          *zap.Logger
}

type innerPool struct {
	lock        sync.RWMutex
	sampler     *sampler
	clientPacks []*clientPack
}

const (
	defaultSessionTokenExpirationDuration = 100 // in blocks

	defaultRebalanceInterval = 25 * time.Second
	defaultRequestTimeout    = 4 * time.Second
)

// NewPool creates connection pool using parameters.
func NewPool(options InitParameters) (*Pool, error) {
	if options.key == nil {
		return nil, fmt.Errorf("missed required parameter 'Key'")
	}

	nodesParams, err := adjustNodeParams(options.nodeParams)
	if err != nil {
		return nil, err
	}

	cache, err := newCache()
	if err != nil {
		return nil, fmt.Errorf("couldn't create cache: %w", err)
	}

	fillDefaultInitParams(&options, cache)

	pool := &Pool{
		key:            options.key,
		owner:          owner.NewIDFromPublicKey(&options.key.PublicKey),
		cache:          cache,
		logger:         options.logger,
		stokenDuration: options.sessionExpirationDuration,
		rebalanceParams: rebalanceParameters{
			nodesParams:               nodesParams,
			nodeRequestTimeout:        options.healthcheckTimeout,
			clientRebalanceInterval:   options.clientRebalanceInterval,
			sessionExpirationDuration: options.sessionExpirationDuration,
		},
		clientBuilder: options.clientBuilder,
	}

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
			st, err := createSessionTokenForDuration(ctx, c, p.owner, p.rebalanceParams.sessionExpirationDuration)
			if err != nil && p.logger != nil {
				p.logger.Warn("failed to create neofs session token for client",
					zap.String("Address", addr),
					zap.Error(err))
			} else if err == nil {
				healthy, atLeastOneHealthy = true, true
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

func fillDefaultInitParams(params *InitParameters, cache *sessionCache) {
	if params.sessionExpirationDuration == 0 {
		params.sessionExpirationDuration = defaultSessionTokenExpirationDuration
	}

	if params.clientRebalanceInterval <= 0 {
		params.clientRebalanceInterval = defaultRebalanceInterval
	}

	if params.healthcheckTimeout <= 0 {
		params.healthcheckTimeout = defaultRequestTimeout
	}

	if params.clientBuilder == nil {
		params.clientBuilder = func(addr string) (client, error) {
			var prm wrapperPrm
			prm.setAddress(addr)
			prm.setKey(*params.key)
			prm.setTimeout(params.nodeDialTimeout)
			prm.setResponseInfoCallback(func(info sdkClient.ResponseMetaInfo) error {
				cache.updateEpoch(info.Epoch())
				return nil
			})
			return newWrapper(prm)
		}
	}
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

	var prmEndpoint prmEndpointInfo

	for j, cPack := range pool.clientPacks {
		wg.Add(1)
		go func(j int, cli client) {
			defer wg.Done()
			ok := true
			tctx, c := context.WithTimeout(ctx, options.nodeRequestTimeout)
			defer c()

			if _, err := cli.endpointInfo(tctx, prmEndpoint); err != nil {
				ok = false
				bufferWeights[j] = 0
			}
			pool.lock.RLock()
			cp := *pool.clientPacks[j]
			pool.lock.RUnlock()

			if ok {
				bufferWeights[j] = options.nodesParams[i].weights[j]
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

func createSessionTokenForDuration(ctx context.Context, c client, ownerID *owner.ID, dur uint64) (*session.Token, error) {
	ni, err := c.networkInfo(ctx, prmNetworkInfo{})
	if err != nil {
		return nil, err
	}

	epoch := ni.CurrentEpoch()

	var exp uint64
	if math.MaxUint64-epoch < dur {
		exp = math.MaxUint64
	} else {
		exp = epoch + dur
	}
	var prm prmCreateSession
	prm.SetExp(exp)

	res, err := c.sessionCreate(ctx, prm)
	if err != nil {
		return nil, err
	}

	return sessionTokenForOwner(ownerID, res, exp), nil
}

type callContext struct {
	// base context for RPC
	context.Context

	client client

	// client endpoint
	endpoint string

	// request signer
	key *ecdsa.PrivateKey

	// flag to open default session if session token is missing
	sessionDefault bool
	sessionTarget  func(session.Token)
	sessionContext *session.ObjectContext
}

func (p *Pool) initCallContext(ctx *callContext, cfg prmCommon, prmCtx prmContext) error {
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
	ctx.sessionDefault = cfg.stoken == nil && prmCtx.defaultSession
	if ctx.sessionDefault {
		ctx.sessionContext = session.NewObjectContext()
		ctx.sessionContext.ToV2().SetVerb(prmCtx.verb)
		ctx.sessionContext.ApplyTo(prmCtx.addr)
	}

	return err
}

// opens new session or uses cached one.
// Must be called only on initialized callContext with set sessionTarget.
func (p *Pool) openDefaultSession(ctx *callContext) error {
	cacheKey := formCacheKey(ctx.endpoint, ctx.key)

	tok := p.cache.Get(cacheKey)
	if tok == nil {
		var err error
		// open new session
		tok, err = createSessionTokenForDuration(ctx, ctx.client, owner.NewIDFromPublicKey(&ctx.key.PublicKey), p.stokenDuration)
		if err != nil {
			return fmt.Errorf("session API client: %w", err)
		}

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
	_ = p.checkSessionTokenErr(err, ctx.endpoint)

	return err
}

// fillAppropriateKey use pool key if caller didn't specify its own.
func (p *Pool) fillAppropriateKey(prm *prmCommon) {
	if prm.key == nil {
		prm.key = p.key
	}
}

// PutObject writes an object through a remote server using NeoFS API protocol.
func (p *Pool) PutObject(ctx context.Context, prm PrmObjectPut) (*oid.ID, error) {
	var cIDp *cid.ID
	if cID, set := prm.hdr.ContainerID(); set {
		cIDp = &cID
	}

	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbPut)
	prmCtx.useAddress(newAddressFromCnrID(cIDp))

	p.fillAppropriateKey(&prm.prmCommon)

	var ctxCall callContext

	ctxCall.Context = ctx

	if err := p.initCallContext(&ctxCall, prm.prmCommon, prmCtx); err != nil {
		return nil, fmt.Errorf("init call context")
	}

	if ctxCall.sessionDefault {
		ctxCall.sessionTarget = prm.UseSession
		if err := p.openDefaultSession(&ctxCall); err != nil {
			return nil, fmt.Errorf("open default session: %w", err)
		}
	}

	id, err := ctxCall.client.objectPut(ctx, prm)
	if err != nil {
		// removes session token from cache in case of token error
		p.checkSessionTokenErr(err, ctxCall.endpoint)
		return nil, fmt.Errorf("init writing on API client: %w", err)
	}

	return id, nil
}

// DeleteObject marks an object for deletion from the container using NeoFS API protocol.
// As a marker, a special unit called a tombstone is placed in the container.
// It confirms the user's intent to delete the object, and is itself a container object.
// Explicit deletion is done asynchronously, and is generally not guaranteed.
func (p *Pool) DeleteObject(ctx context.Context, prm PrmObjectDelete) error {
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbDelete)
	prmCtx.useAddress(&prm.addr)

	p.fillAppropriateKey(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return err
	}

	return p.call(&cc, func() error {
		if err = cc.client.objectDelete(ctx, prm); err != nil {
			return fmt.Errorf("remove object via client: %w", err)
		}

		return nil
	})
}

type objectReadCloser sdkClient.ObjectReader

// Read implements io.Reader of the object payload.
func (x *objectReadCloser) Read(p []byte) (int, error) {
	return (*sdkClient.ObjectReader)(x).Read(p)
}

// Close implements io.Closer of the object payload.
func (x *objectReadCloser) Close() error {
	_, err := (*sdkClient.ObjectReader)(x).Close()
	return err
}

// ResGetObject is designed to provide object header nad read one object payload from NeoFS system.
type ResGetObject struct {
	Header object.Object

	Payload io.ReadCloser
}

// GetObject reads object header and initiates reading an object payload through a remote server using NeoFS API protocol.
func (p *Pool) GetObject(ctx context.Context, prm PrmObjectGet) (*ResGetObject, error) {
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbGet)
	prmCtx.useAddress(&prm.addr)

	p.fillAppropriateKey(&prm.prmCommon)

	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return nil, err
	}

	var res *ResGetObject
	return res, p.call(&cc, func() error {
		res, err = cc.client.objectGet(ctx, prm)
		return err
	})
}

// HeadObject reads object header through a remote server using NeoFS API protocol.
func (p *Pool) HeadObject(ctx context.Context, prm PrmObjectHead) (*object.Object, error) {
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbHead)
	prmCtx.useAddress(&prm.addr)

	p.fillAppropriateKey(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return nil, err
	}

	var obj *object.Object
	return obj, p.call(&cc, func() error {
		obj, err = cc.client.objectHead(ctx, prm)
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
	_, err := x.payload.Close()
	return err
}

// ObjectRange initiates reading an object's payload range through a remote
// server using NeoFS API protocol.
func (p *Pool) ObjectRange(ctx context.Context, prm PrmObjectRange) (*ResObjectRange, error) {
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbRange)
	prmCtx.useAddress(&prm.addr)

	p.fillAppropriateKey(&prm.prmCommon)

	var cc callContext
	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return nil, err
	}

	var res *ResObjectRange

	return res, p.call(&cc, func() error {
		res, err = cc.client.objectRange(ctx, prm)
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
	var prmCtx prmContext
	prmCtx.useDefaultSession()
	prmCtx.useVerb(sessionv2.ObjectVerbSearch)
	prmCtx.useAddress(newAddressFromCnrID(&prm.cnrID))

	p.fillAppropriateKey(&prm.prmCommon)

	var cc callContext

	cc.Context = ctx
	cc.sessionTarget = prm.UseSession

	err := p.initCallContext(&cc, prm.prmCommon, prmCtx)
	if err != nil {
		return nil, err
	}

	var res *ResObjectSearch

	return res, p.call(&cc, func() error {
		res, err = cc.client.objectSearch(ctx, prm)
		return err
	})
}

// PutContainer sends request to save container in NeoFS and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//   polling interval: 5s
//   waiting timeout: 120s
//
// Success can be verified by reading by identifier (see GetContainer).
func (p *Pool) PutContainer(ctx context.Context, prm PrmContainerPut) (*cid.ID, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.containerPut(ctx, prm)
}

// GetContainer reads NeoFS container by ID.
func (p *Pool) GetContainer(ctx context.Context, prm PrmContainerGet) (*container.Container, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.containerGet(ctx, prm)
}

// ListContainers requests identifiers of the account-owned containers.
func (p *Pool) ListContainers(ctx context.Context, prm PrmContainerList) ([]cid.ID, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.containerList(ctx, prm)
}

// DeleteContainer sends request to remove the NeoFS container and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//   polling interval: 5s
//   waiting timeout: 120s
//
// Success can be verified by reading by identifier (see GetContainer).
func (p *Pool) DeleteContainer(ctx context.Context, prm PrmContainerDelete) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	return cp.client.containerDelete(ctx, prm)
}

// GetEACL reads eACL table of the NeoFS container.
func (p *Pool) GetEACL(ctx context.Context, prm PrmContainerEACL) (*eacl.Table, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.containerEACL(ctx, prm)
}

// SetEACL sends request to update eACL table of the NeoFS container and waits for the operation to complete.
//
// Waiting parameters can be specified using SetWaitParams. If not called, defaults are used:
//   polling interval: 5s
//   waiting timeout: 120s
//
// Success can be verified by reading by identifier (see GetEACL).
func (p *Pool) SetEACL(ctx context.Context, prm PrmContainerSetEACL) error {
	cp, err := p.connection()
	if err != nil {
		return err
	}

	return cp.client.containerSetEACL(ctx, prm)
}

// Balance requests current balance of the NeoFS account.
func (p *Pool) Balance(ctx context.Context, prm PrmBalanceGet) (*accounting.Decimal, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.balanceGet(ctx, prm)
}

// waitForContainerPresence waits until the container is found on the NeoFS network.
func waitForContainerPresence(ctx context.Context, cli client, cnrID *cid.ID, waitParams *WaitParams) error {
	var prm PrmContainerGet
	if cnrID != nil {
		prm.SetContainerID(*cnrID)
	}

	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		_, err := cli.containerGet(ctx, prm)
		return err == nil
	})
}

// waitForEACLPresence waits until the container eacl is applied on the NeoFS network.
func waitForEACLPresence(ctx context.Context, cli client, cnrID *cid.ID, table *eacl.Table, waitParams *WaitParams) error {
	var prm PrmContainerEACL
	if cnrID != nil {
		prm.SetContainerID(*cnrID)
	}

	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		eaclTable, err := cli.containerEACL(ctx, prm)
		if err == nil {
			return eacl.EqualTables(*table, *eaclTable)
		}
		return false
	})
}

// waitForContainerRemoved waits until the container is removed from the NeoFS network.
func waitForContainerRemoved(ctx context.Context, cli client, cnrID *cid.ID, waitParams *WaitParams) error {
	var prm PrmContainerGet
	if cnrID != nil {
		prm.SetContainerID(*cnrID)
	}

	return waitFor(ctx, waitParams, func(ctx context.Context) bool {
		_, err := cli.containerGet(ctx, prm)
		return sdkClient.IsErrContainerNotFound(err) ||
			err != nil && strings.Contains(err.Error(), "not found")
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

// NetworkInfo requests information about the NeoFS network of which the remote server is a part.
func (p *Pool) NetworkInfo(ctx context.Context) (*netmap.NetworkInfo, error) {
	cp, err := p.connection()
	if err != nil {
		return nil, err
	}

	return cp.client.networkInfo(ctx, prmNetworkInfo{})
}

// Close closes the Pool and releases all the associated resources.
func (p *Pool) Close() {
	p.cancel()
	<-p.closedCh
}

// creates new session token with specified owner from SessionCreate call result.
func sessionTokenForOwner(id *owner.ID, cliRes *resCreateSession, exp uint64) *session.Token {
	st := session.NewToken()
	st.SetOwnerID(id)
	st.SetID(cliRes.id)
	st.SetSessionKey(cliRes.sessionKey)
	st.SetExp(exp)

	return st
}

func newAddressFromCnrID(cnrID *cid.ID) *address.Address {
	addr := address.NewAddress()
	if cnrID != nil {
		addr.SetContainerID(*cnrID)
	}
	return addr
}

func copySessionTokenWithoutSignatureAndContext(from session.Token) (to session.Token) {
	to.SetIat(from.Iat())
	to.SetExp(from.Exp())
	to.SetNbf(from.Nbf())

	sessionTokenID := make([]byte, len(from.ID()))
	copy(sessionTokenID, from.ID())
	to.SetID(sessionTokenID)

	sessionTokenKey := make([]byte, len(from.SessionKey()))
	copy(sessionTokenKey, from.SessionKey())
	to.SetSessionKey(sessionTokenKey)

	var sessionTokenOwner owner.ID
	buf, err := from.OwnerID().Marshal()
	if err != nil {
		panic(err) // should never happen
	}
	err = sessionTokenOwner.Unmarshal(buf)
	if err != nil {
		panic(err) // should never happen
	}
	to.SetOwnerID(&sessionTokenOwner)

	return to
}
