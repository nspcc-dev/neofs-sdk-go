package container

import (
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// ApplyNetworkConfig applies network configuration to the
// container. Changes the container if it does not satisfy
// network configuration.
//
// Deprecated: network settings should not affect containers.
func (x *Container) ApplyNetworkConfig(cfg netmap.NetworkInfo) {
	//nolint:staticcheck // compatibility
	if cfg.HomomorphicHashingDisabled() {
		//nolint:staticcheck
		x.DisableHomomorphicHashing()
	}
}

// AssertNetworkConfig checks if a container matches passed
// network configuration.
//
// Deprecated: network settings should not affect containers.
func (x Container) AssertNetworkConfig(cfg netmap.NetworkInfo) bool {
	//nolint:staticcheck // compatibility
	return x.IsHomomorphicHashingDisabled() == cfg.HomomorphicHashingDisabled()
}
