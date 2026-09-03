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
	// TLS/SSL. When the scheme is not specified explicitly, addresses with
	// this port are assumed to require a TLS connection.
	tlsPort = "443"
)

// Parse parses URI and returns a host and a flag indicating that TLS is
// enabled.
//
// If scheme is not specified explicitly, port 443 is treated as an
// indication that TLS connection is required. If grpcs scheme is specified
// without a port, port 443 is used by default.
func Parse(s string) (string, bool, error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		if !strings.Contains(s, "/") {
			_, port, err := net.SplitHostPort(s)
			if err != nil {
				return s, false, err
			}
			return s, port == tlsPort, nil
		}
		return s, false, err
	}

	// check if passed string was parsed correctly
	// URIs that do not start with a slash after the scheme are interpreted as:
	// `scheme:opaque` => if `opaque` is not empty, then it is supposed that URI
	// is in `host:port` format
	schemeMissing := uri.Host == ""
	if schemeMissing {
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

	port := uri.Port()
	if port == "" {
		if uri.Scheme != grpcTLSScheme {
			return "", false, errors.New("missing port in address")
		}
		// default TLS port for grpcs scheme without explicit port
		port = tlsPort
		uri.Host = net.JoinHostPort(uri.Host, port)
	}

	withTLS := uri.Scheme == grpcTLSScheme || (schemeMissing && port == tlsPort)

	return uri.Host, withTLS, nil
}
