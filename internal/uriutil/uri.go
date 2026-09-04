package uriutil

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	grpcScheme    = "grpc"
	grpcTLSScheme = "grpcs"

	// tlsPort is the well-known port reserved for protocols running over
	// TLS/SSL. It is used as the default port for the grpcs scheme when
	// no port is specified explicitly.
	tlsPort = "443"
)

// Parse parses URI and returns a host and a flag indicating that TLS is
// enabled.
//
// If grpcs scheme is specified without a port, port 443 is used by default.
func Parse(s string) (string, bool, error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		if !strings.Contains(s, "/") {
			_, _, err := net.SplitHostPort(s)
			return s, false, err
		}
		return s, false, err
	}

	// check if passed string was parsed correctly
	// URIs that do not start with a slash after the scheme are interpreted as:
	// `scheme:opaque` => if `opaque` is not empty, then it is supposed that URI
	// is in `host:port` format
	if uri.Host == "" {
		uri.Host = uri.Scheme
		uri.Scheme = grpcScheme // assume GRPC by default
		if uri.Opaque != "" {
			uri.Host = net.JoinHostPort(uri.Host, uri.Opaque)
		}
	}

	switch uri.Scheme {
	case grpcTLSScheme, grpcScheme:
	default:
		return "", false, fmt.Errorf("unsupported scheme: %s", uri.Scheme)
	}

	if uri.Port() == "" {
		if uri.Scheme != grpcTLSScheme {
			return "", false, errors.New("missing port in address")
		}
		// default TLS port for grpcs scheme without explicit port
		uri.Host = net.JoinHostPort(uri.Host, tlsPort)
	}

	return uri.Host, uri.Scheme == grpcTLSScheme, nil
}
