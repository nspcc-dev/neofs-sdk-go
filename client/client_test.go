package client

import (
	"context"
	"net"
	"testing"

	apistatus "github.com/nspcc-dev/neofs-sdk-go/client/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

/*
File contains common functionality used for client package testing.
*/

var statusErr apistatus.ServerInternal

func init() {
	statusErr.SetMessage("test status error")
}

// pairs service spec and implementation to-be-registered in some [grpc.Server].
type testService struct {
	desc *grpc.ServiceDesc
	impl any
}

// returns ready-to-go [Client] of provided optional services. By default, any
// other service is unsupported.
//
// If caller registers stat callback (like [PrmInit.SetStatisticCallback] does)
// processing nodeKey, it must include NetmapService with implemented
// LocalNodeInfo method.
func newClient(t testing.TB, svcs ...testService) *Client {
	var prm PrmInit

	c, err := New(prm)
	require.NoError(t, err)

	srv := grpc.NewServer()
	for _, svc := range svcs {
		srv.RegisterService(svc.desc, svc.impl)
	}

	lis := bufconn.Listen(10 << 10)
	go func() { _ = srv.Serve(lis) }()

	var dialPrm PrmDial
	dialPrm.SetServerURI("grpc://localhost:8080") // any valid
	dialPrm.setDialFunc(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) })
	err = c.Dial(dialPrm)
	if err != nil {
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.Unimplemented, st.Code())
	}

	return c
}

func TestClient_Dial(t *testing.T) {
	var prmInit PrmInit

	c, err := New(prmInit)
	require.NoError(t, err)

	t.Run("failure", func(t *testing.T) {
		t.Run("endpoint", func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				s      string
				assert func(t testing.TB, err error)
			}{
				{name: "missing", s: "", assert: func(t testing.TB, err error) {
					require.ErrorIs(t, c.Dial(PrmDial{}), ErrMissingServer)
				}},
				{name: "contains control char", s: "grpc://st1.storage.fs.neo.org:8080" + string(rune(0x7f)), assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "net/url: invalid control character in URL")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "missing port", s: "grpc://st1.storage.fs.neo.org", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "invalid port", s: "grpc://st1.storage.fs.neo.org:foo", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, `invalid port ":foo" after host`)
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "unsupported scheme", s: "unknown://st1.storage.fs.neo.org:8080", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "unsupported scheme: unknown")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "multiaddr", s: "/ip4/st1.storage.fs.neo.org/tcp/8080", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "host only", s: "st1.storage.fs.neo.org", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
				{name: "invalid port without scheme", s: "st1.storage.fs.neo.org:foo", assert: func(t testing.TB, err error) {
					require.ErrorContains(t, err, "missing port in address")
					require.ErrorContains(t, err, "invalid server URI")
				}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var p PrmDial
					p.SetServerURI(tc.s)
					tc.assert(t, c.Dial(p))
				})
			}
		})
		t.Run("dial timeout", func(t *testing.T) {
			var p PrmDial
			p.SetServerURI("grpc://localhost:8080")
			p.SetTimeout(0)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
			p.SetTimeout(-1)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
		})
		t.Run("stream timeout", func(t *testing.T) {
			var p PrmDial
			p.SetServerURI("grpc://localhost:8080")
			p.SetStreamTimeout(0)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
			p.SetStreamTimeout(-1)
			require.ErrorIs(t, c.Dial(p), ErrNonPositiveTimeout)
		})
		t.Run("context", func(t *testing.T) {
			var anyValidPrm PrmDial
			anyValidPrm.SetServerURI("localhost:8080")
			t.Run("cancelled", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				p := anyValidPrm
				p.SetContext(ctx)
				err := c.Dial(p)
				require.ErrorIs(t, err, context.Canceled)
			})
			t.Run("deadline", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				cancel()

				p := anyValidPrm
				p.SetContext(ctx)
				err := c.Dial(p)
				require.ErrorIs(t, err, context.DeadlineExceeded)
			})
		})
	})
}

type nopPublicKey struct{}

func (x nopPublicKey) MaxEncodedSize() int     { return 10 }
func (x nopPublicKey) Encode(buf []byte) int   { return copy(buf, "public_key") }
func (x nopPublicKey) Decode([]byte) error     { return nil }
func (x nopPublicKey) Verify(_, _ []byte) bool { return true }

type nopSigner struct{}

func (nopSigner) Scheme() neofscrypto.Scheme      { return neofscrypto.ECDSA_SHA512 }
func (nopSigner) Sign([]byte) ([]byte, error)     { return []byte("signature"), nil }
func (x nopSigner) Public() neofscrypto.PublicKey { return nopPublicKey{} }
