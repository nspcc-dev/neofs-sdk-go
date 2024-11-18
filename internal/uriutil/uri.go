package uriutil

import (
	"fmt"
	"net"
	"net/url"
)

// Parse parses URI and returns a host and a flag indicating that TLS is
// enabled.
func Parse(s string) (string, bool, error) {
	uri, err := url.ParseRequestURI(s)
	if err != nil {
		return s, false, nil
	}

	const (
		grpcScheme    = "grpc"
		grpcTLSScheme = "grpcs"
	)

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

	return uri.Host, uri.Scheme == grpcTLSScheme, nil
}
