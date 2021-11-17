package resolver

import (
	"fmt"
	"net"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

const (
	testnetDomain = ".containers.testnet.fs.neo.org"
	mainnetDomain = ".containers.fs.neo.org"
)

// ResolveContainerTestnet request txt record name + '.containers.testnet.fs.neo.org' to default dns server.
func ResolveContainerTestnet(name string) (*cid.ID, error) {
	return ResolveContainerDomainName(name + testnetDomain)
}

// ResolveContainerMainnet request txt record name + '.containers.fs.neo.org' to default dns server.
func ResolveContainerMainnet(name string) (*cid.ID, error) {
	return ResolveContainerDomainName(name + mainnetDomain)
}

// ResolveContainerDomainName trys to resolve container domain name to container ID using system dns server.
func ResolveContainerDomainName(domain string) (*cid.ID, error) {
	results, err := net.LookupTXT(domain)
	if err != nil {
		return nil, err
	}

	cnrID := cid.New()
	for _, res := range results {
		if err = cnrID.Parse(res); err != nil {
			continue
		}
		return cnrID, nil
	}

	return nil, fmt.Errorf("not found")
}
