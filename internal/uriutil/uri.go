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

	// httpPort and tlsPort are the well-known ports used as defaults for the
	// grpc and grpcs schemes respectively when no port is specified explicitly.
	httpPort = "80"
	tlsPort  = "443"
)

// Parse parses URI and returns a host and a flag indicating that TLS is
// enabled.
//
// If scheme is specified without a port, port 80 (grpc) or 443 (grpcs) is
// used by default.
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

	if uri.Host == "" {
		return "", false, errors.New("missing host in address")
	}

	if uri.Port() == "" {
		if _, _, splitErr := net.SplitHostPort(uri.Host); splitErr == nil {
			// port is present but invalid (e.g. non-numeric)
			return "", false, errors.New("missing port in address")
		}
		// no port specified explicitly, use the default one for the scheme
		port := httpPort
		if uri.Scheme == grpcTLSScheme {
			port = tlsPort
		}
		uri.Host = net.JoinHostPort(uri.Host, port)
	}

	return uri.Host, uri.Scheme == grpcTLSScheme, nil
}
