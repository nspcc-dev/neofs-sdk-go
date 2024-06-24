package container

import (
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// ApplyNetworkConfig applies network configuration to the
// container. Changes the container if it does not satisfy
// network configuration.
//
// See also [Container.AssertNetworkConfig].
func (x *Container) ApplyNetworkConfig(cfg netmap.NetworkInfo) {
	x.SetHomomorphicHashingDisabled(cfg.HomomorphicHashingDisabled())
}

// AssertNetworkConfig checks if a container matches passed
// network configuration.
//
// See also [Container.ApplyNetworkConfig].
func (x Container) AssertNetworkConfig(cfg netmap.NetworkInfo) bool {
	return x.HomomorphicHashingDisabled() == cfg.HomomorphicHashingDisabled()
}
