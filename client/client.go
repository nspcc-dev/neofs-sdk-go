package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
)

const (
	// max GRPC message size.
	defaultBufferSize = 4194304 // 4MB
)

// Client represents virtual connection to the NeoFS network to communicate
// with NeoFS server using NeoFS API protocol. It is designed to provide
// an abstraction interface from the protocol details of data transfer over
// a network in NeoFS.
//
// Client can be created using [New].
// Before executing the NeoFS operations using the Client, connection to the
// server MUST BE correctly established (see Dial method and pay attention
// to the mandatory parameters). Using the Client before connecting have
// been established can lead to a panic. After the work, the Client SHOULD BE
// closed (see Close method): it frees internal and system resources which were
// allocated for the period of work of the Client. Calling [Client.Dial]/[Client.Close] method
// during the communication process step strongly discouraged as it leads to
// undefined behavior.
//
// Each method which produces a NeoFS API call may return a server response.
// Status responses are returned in the result structure, and can be cast
// to built-in error instance (or in the returned error if the client is
// configured accordingly). Certain statuses can be checked using [apistatus]
// and standard [errors] packages.
// All possible responses are documented in methods, however, some may be
// returned from all of them (pay attention to the presence of the pointer sign):
//   - *[apistatus.ServerInternal] on internal server error;
//   - *[apistatus.NodeUnderMaintenance] if a server is under maintenance;
//   - *[apistatus.SuccessDefaultV2] on default success.
//
// Client MUST NOT be copied by value: use pointer to Client instead.
//
// See client package overview to get some examples.
type Client struct {
	prm PrmInit

	c client.Client

	server neoFSAPIServer

	endpoint string
	nodeKey  []byte

	buffers *sync.Pool
}

// New creates an instance of Client initialized with the given parameters.
//
// See docs of [PrmInit] methods for details. See also [Client.Dial]/[Client.Close].
func New(prm PrmInit) (*Client, error) {
	var c = new(Client)
	pk, err := keys.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}

	prm.signer = neofsecdsa.SignerRFC6979(pk.PrivateKey)

	if prm.buffers != nil {
		c.buffers = prm.buffers
	} else {
		size := prm.signMessageBufferSizes
		if size == 0 {
			size = defaultBufferSize
		}

		c.buffers = &sync.Pool{}
		c.buffers.New = func() any {
			b := make([]byte, size)
			return &b
		}
	}

	c.prm = prm
	return c, nil
}

// Dial establishes a connection to the server from the NeoFS network.
// Returns an error describing failure reason. If failed, the Client
// SHOULD NOT be used.
//
// Uses the context specified by SetContext if it was called with non-nil
// argument, otherwise context.Background() is used. Dial returns context
// errors, see context package docs for details.
//
// Panics if required parameters are set incorrectly, look carefully
// at the method documentation.
//
// One-time method call during application start-up stage is expected.
// Calling multiple times leads to undefined behavior.
//
// Return client errors:
//   - [ErrMissingServer]
//   - [ErrNonPositiveTimeout]
//
// See also [Client.Close].
func (c *Client) Dial(prm PrmDial) error {
	if prm.endpoint == "" {
		return ErrMissingServer
	}
	c.endpoint = prm.endpoint

	if prm.timeoutDialSet {
		if prm.timeoutDial <= 0 {
			return ErrNonPositiveTimeout
		}
	} else {
		prm.timeoutDial = 5 * time.Second
	}

	if prm.streamTimeoutSet {
		if prm.streamTimeout <= 0 {
			return ErrNonPositiveTimeout
		}
	} else {
		prm.streamTimeout = 10 * time.Second
	}

	addr, withTLS, err := client.ParseURI(prm.endpoint)
	if err != nil {
		return fmt.Errorf("invalid server URI: %w", err)
	}

	var tlsCfg *tls.Config
	if withTLS {
		if tlsCfg = prm.tlsConfig; tlsCfg == nil {
			tlsCfg = new(tls.Config)
		}
	}

	c.c = *client.New(
		client.WithNetworkAddress(addr),
		client.WithTLSCfg(tlsCfg),
		client.WithDialTimeout(prm.timeoutDial),
		client.WithRWTimeout(prm.streamTimeout),
	)

	c.setNeoFSAPIServer((*coreServer)(&c.c))

	if prm.parentCtx == nil {
		prm.parentCtx = context.Background()
	}

	endpointInfo, err := c.EndpointInfo(prm.parentCtx, PrmEndpointInfo{})
	if err != nil {
		return err
	}

	c.nodeKey = endpointInfo.NodeInfo().PublicKey()

	return nil
}

// sets underlying provider of neoFSAPIServer. The method is used for testing as an approach
// to skip Dial stage and override NeoFS API server. MUST NOT be used outside test code.
// In real applications wrapper over github.com/nspcc-dev/neofs-api-go/v2/rpc/client
// is statically used.
func (c *Client) setNeoFSAPIServer(server neoFSAPIServer) {
	c.server = server
}

// Close closes underlying connection to the NeoFS server. Implements io.Closer.
// MUST NOT be called before successful Dial. Can be called concurrently
// with server operations processing on running goroutines: in this case
// they are likely to fail due to a connection error.
//
// One-time method call during application shutdown stage (after [Client.Dial])
// is expected. Calling multiple times leads to undefined behavior.
//
// See also [Client.Dial].
func (c *Client) Close() error {
	return c.c.Conn().Close()
}

func (c *Client) sendStatistic(m stat.Method, err error) func() {
	if c.prm.statisticCallback == nil {
		return func() {}
	}

	ts := time.Now()
	return func() {
		c.prm.statisticCallback(c.nodeKey, c.endpoint, m, time.Since(ts), err)
	}
}

// PrmInit groups initialization parameters of Client instances.
//
// See also [New].
type PrmInit struct {
	signer neofscrypto.Signer

	cbRespInfo func(ResponseMetaInfo) error

	netMagic uint64

	statisticCallback stat.OperationCallback

	signMessageBufferSizes uint64
	buffers                *sync.Pool
}

// SetSignMessageBufferSizes sets single buffer size to the buffers pool inside client.
// This pool are using in GRPC message signing process and helps to reduce memory allocations.
func (x *PrmInit) SetSignMessageBufferSizes(size uint64) {
	x.signMessageBufferSizes = size
}

// SetSignMessageBuffers sets buffers which are using in GRPC message signing process and helps to reduce memory allocations.
func (x *PrmInit) SetSignMessageBuffers(buffers *sync.Pool) {
	x.buffers = buffers
}

// SetResponseInfoCallback makes the Client to pass ResponseMetaInfo from each
// NeoFS server response to f. Nil (default) means ignore response meta info.
func (x *PrmInit) SetResponseInfoCallback(f func(ResponseMetaInfo) error) {
	x.cbRespInfo = f
}

// SetStatisticCallback makes the Client to pass [stat.OperationCallback] for the external statistic.
func (x *PrmInit) SetStatisticCallback(statisticCallback stat.OperationCallback) {
	x.statisticCallback = statisticCallback
}

// PrmDial groups connection parameters for the Client.
//
// See also Dial.
type PrmDial struct {
	endpoint string

	tlsConfig *tls.Config

	timeoutDialSet bool
	timeoutDial    time.Duration

	streamTimeoutSet bool
	streamTimeout    time.Duration

	parentCtx context.Context
}

// SetServerURI sets server URI in the NeoFS network.
// Required parameter.
//
// Format of the URI:
//
//	[scheme://]host:port
//
// Supported schemes:
//
//	grpc
//	grpcs
//
// See also SetTLSConfig.
func (x *PrmDial) SetServerURI(endpoint string) {
	x.endpoint = endpoint
}

// SetTLSConfig sets tls.Config to open TLS client connection
// to the NeoFS server. Nil (default) means insecure connection.
//
// See also SetServerURI.
func (x *PrmDial) SetTLSConfig(tlsConfig *tls.Config) {
	x.tlsConfig = tlsConfig
}

// SetTimeout sets the timeout for connection to be established.
// MUST BE positive. If not called, 5s timeout will be used by default.
func (x *PrmDial) SetTimeout(timeout time.Duration) {
	x.timeoutDialSet = true
	x.timeoutDial = timeout
}

// SetStreamTimeout sets the timeout for individual operations in streaming RPC.
// MUST BE positive. If not called, 10s timeout will be used by default.
func (x *PrmDial) SetStreamTimeout(timeout time.Duration) {
	x.streamTimeoutSet = true
	x.streamTimeout = timeout
}

// SetContext allows to specify optional base context within which connection
// should be established.
//
// Context SHOULD NOT be nil.
func (x *PrmDial) SetContext(ctx context.Context) {
	x.parentCtx = ctx
}
