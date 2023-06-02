package container

import (
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// ApplyNetworkConfig applies network configuration to the
// container. Changes the container if it does not satisfy
// network configuration.
func (x *Container) ApplyNetworkConfig(cfg netmap.NetworkInfo) {
	if cfg.HomomorphicHashingDisabled() {
		x.DisableHomomorphicHashing()
	}
}

// AssertNetworkConfig checks if a container matches passed
// network configuration.
func (x Container) AssertNetworkConfig(cfg netmap.NetworkInfo) bool {
	return x.IsHomomorphicHashingDisabled() == cfg.HomomorphicHashingDisabled()
}
