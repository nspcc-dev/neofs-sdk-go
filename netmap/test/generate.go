package test

import "github.com/nspcc-dev/neofs-sdk-go/netmap"

// NetworkParameter returns random netmap.NetworkParameter.
func NetworkParameter() *netmap.NetworkParameter {
	x := netmap.NewNetworkParameter()

	x.SetKey([]byte("key"))
	x.SetValue([]byte("value"))

	return x
}

// NetworkConfig returns random netmap.NetworkConfig.
func NetworkConfig() *netmap.NetworkConfig {
	x := netmap.NewNetworkConfig()

	x.SetParameters(
		NetworkParameter(),
		NetworkParameter(),
	)

	return x
}

// NetworkInfo returns random netmap.NetworkInfo.
func NetworkInfo() *netmap.NetworkInfo {
	x := netmap.NewNetworkInfo()

	x.SetCurrentEpoch(21)
	x.SetMagicNumber(32)
	x.SetMsPerBlock(43)
	x.SetNetworkConfig(NetworkConfig())

	return x
}
