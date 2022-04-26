package container

import (
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
)

// ApplyNetworkConfig applies network configuration to the
// container. Changes the container if it does not satisfy
// network configuration.
//
// Network config must contain unique parameters. Duplicated
// network parameters lead to undefined behaviour.
//
// Returns any network config parsing errors.
func (c *Container) ApplyNetworkConfig(cfg netmap.NetworkConfig) error {
	v, err := homomorphicHashDisabled(cfg)
	if err != nil {
		return err
	}

	c.v2.SetHomomorphicHashingDisabled(v)

	return nil
}

// AssertNetworkConfig checks if a container matches passed
// network configuration.
//
// Network config must contain unique parameters. Duplicated
// network parameters lead to undefined behaviour.
//
// Returns any network config parsing errors.
func (c Container) AssertNetworkConfig(cfg netmap.NetworkConfig) (res bool, err error) {
	v, err := homomorphicHashDisabled(cfg)
	if err != nil {
		return false, nil
	}

	return c.HomomorphicHashingDisabled() == v, nil
}

const HomomorphicHashingDisabledKey = "HomomorphicHashingDisabled"

func homomorphicHashDisabled(cfg netmap.NetworkConfig) (res bool, err error) {
	cfg.IterateParameters(func(prm *netmap.NetworkParameter) bool {
		if string(prm.Key()) == HomomorphicHashingDisabledKey {
			arr := stackitem.NewByteArray(prm.Value())

			res, err = arr.TryBool()
			if err != nil {
				err = fmt.Errorf("could not parse %s config value: %w",
					HomomorphicHashingDisabledKey, err,
				)
			}

			return true
		}

		return false
	})

	return
}
