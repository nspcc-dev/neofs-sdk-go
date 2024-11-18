package uriutil_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/internal/uriutil"
	"github.com/stretchr/testify/require"
)

func TestParseURI(t *testing.T) {
	for _, tc := range []struct {
		s       string
		host    string
		withTLS bool
	}{
		{s: "not a URI", host: "not a URI", withTLS: false},
		{s: "8080", host: "8080", withTLS: false},
		{s: "127.0.0.1", host: "127.0.0.1", withTLS: false},
		{s: "st1.storage.fs.neo.org", host: "st1.storage.fs.neo.org", withTLS: false},
		// multiaddr
		{s: "/ip4/127.0.0.1/tcp/8080", host: "", withTLS: false},
		// no scheme (TCP)
		{s: "127.0.0.1:8080", host: "127.0.0.1:8080", withTLS: false},
		{s: "st1.storage.fs.neo.org:8080", host: "st1.storage.fs.neo.org:8080", withTLS: false},
		// with scheme, no port
		{s: "grpc://127.0.0.1", host: "127.0.0.1", withTLS: false},
		{s: "grpc://st1.storage.fs.neo.org", host: "st1.storage.fs.neo.org", withTLS: false},
		{s: "grpcs://127.0.0.1", host: "127.0.0.1", withTLS: true},
		{s: "grpcs://st1.storage.fs.neo.org", host: "st1.storage.fs.neo.org", withTLS: true},
		// with scheme and port
		{s: "grpc://127.0.0.1:8080", host: "127.0.0.1:8080", withTLS: false},
		{s: "grpc://st1.storage.fs.neo.org:8080", host: "st1.storage.fs.neo.org:8080", withTLS: false},
		{s: "grpcs://127.0.0.1:8082", host: "127.0.0.1:8082", withTLS: true},
		{s: "grpcs://st1.storage.fs.neo.org:8082", host: "st1.storage.fs.neo.org:8082", withTLS: true},
	} {
		host, withTLS, err := uriutil.Parse(tc.s)
		require.NoError(t, err, tc.s)
		require.Equal(t, tc.host, host, tc.s)
		require.Equal(t, tc.withTLS, withTLS, tc.s)
	}

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, s, err string
		}{
			{name: "unsupported scheme", s: "unknown://st1.storage.fs.neo.org:8082", err: "unsupported scheme: unknown"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := uriutil.Parse(tc.s)
				require.EqualError(t, err, tc.err)
			})
		}
	})
}
