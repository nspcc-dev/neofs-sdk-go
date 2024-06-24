package client

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultSignBufferSize = 4 << 20 // 4MB, max GRPC message size.

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
	endpoint string
	dial     func(context.Context, string) (net.Conn, error)
	withTLS  bool

	signer      neofscrypto.Signer
	signBuffers *sync.Pool

	interceptAPIRespInfo func(ResponseMetaInfo) error
	handleAPIOpResult    stat.OperationCallback

	// set on dial
	conn         *grpc.ClientConn
	transport    grpcTransport // based on conn
	serverPubKey []byte
}

// parses s into a URI and returns host:port and a flag indicating enabled TLS.
func parseURI(s string) (string, bool, error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		return s, false, err
	}

	const grpcScheme = "grpc"
	const grpcTLSScheme = "grpcs"
	// check if passed string was parsed correctly
	// URIs that do not start with a slash after the scheme are interpreted as:
	// `scheme:opaque` => if `opaque` is not empty, then it is supposed that URI
	// is in `host:port` format
	if uri.Host == "" {
		uri.Host = uri.Scheme
		uri.Scheme = grpcScheme
		if uri.Opaque != "" {
			uri.Host = net.JoinHostPort(uri.Host, uri.Opaque)
		}
	}

	if uri.Scheme != grpcScheme && uri.Scheme != grpcTLSScheme {
		return "", false, fmt.Errorf("unsupported URI scheme: %s", uri.Scheme)
	}

	return uri.Host, uri.Scheme == grpcTLSScheme, nil
}

// New initializes new Client to connect to NeoFS API server at the specified
// network endpoint with options. New does not dial the server: [Client.Dial]
// must be done after. Use [Dial] to open instant connection.
//
// URI format:
//
//	[scheme://]host:port
//
// with one of supported schemes:
//
//	grpc
//	grpcs
func New(uri string, opts Options) (*Client, error) {
	endpoint, withTLS, err := parseURI(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	pk, err := keys.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("randomize private key: %w", err)
	}

	return &Client{
		endpoint: endpoint,
		withTLS:  withTLS,
		signer:   neofsecdsa.Signer(pk.PrivateKey),
		signBuffers: &sync.Pool{New: func() any {
			b := make([]byte, defaultSignBufferSize)
			return &b
		}},
		interceptAPIRespInfo: opts.interceptAPIRespInfo,
		handleAPIOpResult:    opts.handleAPIReqResult,
	}, nil
}

// Dial establishes connection to the NeoFS API server by its parameterized
// network URI and options. After use, the connection must be closed using
// [Client.Close]. If Dial fails, the Client must no longer be used.
//
// If operation result handler is specified, Dial also requests server info
// required for it. See [Options.SetOpResultHandler].
//
// Dial does not modify context whose deadline may interrupt the connection. Use
// [context.WithTimeout] to prevent potential hangup.
func (c *Client) Dial(ctx context.Context) error {
	var creds credentials.TransportCredentials
	if c.withTLS {
		creds = credentials.NewTLS(nil)
	} else {
		creds = insecure.NewCredentials()
	}

	var err error
	c.conn, err = grpc.DialContext(ctx, c.endpoint,
		grpc.WithContextDialer(c.dial),
		grpc.WithTransportCredentials(creds),
		grpc.WithReturnConnectionError(),
		grpc.FailOnNonTempDialError(true),
	)
	if err != nil {
		return fmt.Errorf("gRPC dial %s: %w", c.endpoint, err)
	}

	c.transport = newGRPCTransport(c.conn)

	if c.handleAPIOpResult != nil {
		endpointInfo, err := c.GetEndpointInfo(ctx, GetEndpointInfoOptions{})
		if err != nil {
			return fmt.Errorf("request node info from the server for stat tracking: %w", err)
		}
		c.serverPubKey = endpointInfo.Node.PublicKey()
	}

	return nil
}

// Dial connects to the NeoFS API server by its URI with options and returns
// ready-to-go Client. After use, the connection must be closed via
// [Client.Close]. If application needs delayed dial, use [New] + [Client.Dial]
// combo.
func Dial(ctx context.Context, uri string, opts Options) (*Client, error) {
	c, err := New(uri, opts)
	if err != nil {
		return nil, err
	}
	return c, c.Dial(ctx)
}

// Close closes underlying connection to the NeoFS server. Implements
// [io.Closer]. Close MUST NOT be called before successful [Client.Dial]. Close
// can be called concurrently with server operations processing on running
// goroutines: in this case they are likely to fail due to a connection error.
//
// One-time method call during application shutdown stage is expected. Calling
// multiple times leads to undefined behavior.
func (c *Client) Close() error {
	return c.conn.Close()
}

// // SetSignMessageBuffers sets buffers which are using in GRPC message signing
// // process and helps to reduce memory allocations.
// func (x *Options) SetSignMessageBuffers(buffers *sync.Pool) {
// 	x.signBuffers = buffers
// }

// Options groups optional Client parameters.
//
// See also [New].
type Options struct {
	netMagic uint64

	handleAPIReqResult   stat.OperationCallback
	interceptAPIRespInfo func(ResponseMetaInfo) error

	signBuffersSize uint64
	signBuffers     *sync.Pool
}

// SetAPIResponseInfoInterceptor allows to intercept meta information from each
// NeoFS server response before its processing. If f returns an error, [Client]
// immediately returns it from the method. Nil (default) means ignore response
// meta info.
func (x *Options) SetAPIResponseInfoInterceptor(f func(ResponseMetaInfo) error) {
	x.interceptAPIRespInfo = f
}

// SetAPIRequestResultHandler makes the [Client] to pass result of each
// performed NeoFS API operation to the specified handler. Nil (default)
// disables handling.
func (x *Options) SetAPIRequestResultHandler(f stat.OperationCallback) {
	x.handleAPIReqResult = f
}
