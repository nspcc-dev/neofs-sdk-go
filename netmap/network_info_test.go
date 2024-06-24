package netmap_test

import (
	"encoding/binary"
	"math"
	"math/rand"
	"reflect"
	"testing"

	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	netmaptest "github.com/nspcc-dev/neofs-sdk-go/netmap/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNetworkInfo_ReadFromV2(t *testing.T) {
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("network parameter", func(t *testing.T) {
			testCases := []struct {
				name string
				err  string
				prm  apinetmap.NetworkConfig_Parameter
			}{
				{name: "nil key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
					Key: nil, Value: []byte("any"),
				}},
				{name: "empty key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte{}, Value: []byte("any"),
				}},
				{name: "nil value", err: "invalid network parameter #1: missing value", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("any"), Value: nil,
				}},
				{name: "repeated keys", err: "multiple network parameters with name=any_key", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("any_key"), Value: []byte("any"),
				}},
				{name: "audit fee format", err: "invalid network parameter #1 (AuditFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("AuditFee"), Value: []byte("Hello, world!"),
				}},
				{name: "storage price format", err: "invalid network parameter #1 (BasicIncomeRate): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("BasicIncomeRate"), Value: []byte("Hello, world!"),
				}},
				{name: "container fee format", err: "invalid network parameter #1 (ContainerFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("ContainerFee"), Value: []byte("Hello, world!"),
				}},
				{name: "named container fee format", err: "invalid network parameter #1 (ContainerAliasFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("ContainerAliasFee"), Value: []byte("Hello, world!"),
				}},
				{name: "num of EigenTrust iterations format", err: "invalid network parameter #1 (EigenTrustIterations): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustIterations"), Value: []byte("Hello, world!"),
				}},
				{name: "epoch duration format", err: "invalid network parameter #1 (EpochDuration): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EpochDuration"), Value: []byte("Hello, world!"),
				}},
				{name: "IR candidate fee format", err: "invalid network parameter #1 (InnerRingCandidateFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("InnerRingCandidateFee"), Value: []byte("Hello, world!"),
				}},
				{name: "max object size format", err: "invalid network parameter #1 (MaxObjectSize): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("MaxObjectSize"), Value: []byte("Hello, world!"),
				}},
				{name: "withdrawal fee format", err: "invalid network parameter #1 (WithdrawFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("WithdrawFee"), Value: []byte("Hello, world!"),
				}},
				{name: "EigenTrust alpha format", err: "invalid network parameter #1 (EigenTrustAlpha): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte("Hello, world!"),
				}},
				{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value -3.14 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte{31, 133, 235, 81, 184, 30, 9, 192},
				}},
				{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value 1.10 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte{154, 153, 153, 153, 153, 153, 241, 63},
				}},
				{name: "disable homomorphic hashing format", err: "invalid network parameter #1 (HomomorphicHashingDisabled): invalid bool parameter", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("HomomorphicHashingDisabled"), Value: make([]byte, 32+1), // max 32
				}},
				{name: "allow maintenance mode format", err: "invalid network parameter #1 (MaintenanceModeAllowed): invalid bool parameter", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("MaintenanceModeAllowed"), Value: make([]byte, 32+1), // max 32
				}},
			}

			for i := range testCases {
				n := netmaptest.NetworkInfo()
				var m apinetmap.NetworkInfo

				n.WriteToV2(&m)
				m.NetworkConfig.Parameters = []*apinetmap.NetworkConfig_Parameter{
					{Key: []byte("any_key"), Value: []byte("any_val")},
					&testCases[i].prm,
				}

				require.ErrorContains(t, n.ReadFromV2(&m), testCases[i].err)
			}
		})
	})
}

func TestNetworkInfo_Unmarshal(t *testing.T) {
	t.Run("invalid binary", func(t *testing.T) {
		var n netmap.NetworkInfo
		msg := []byte("definitely_not_protobuf")
		err := n.Unmarshal(msg)
		require.ErrorContains(t, err, "decode protobuf")
	})
	t.Run("invalid fields", func(t *testing.T) {
		t.Run("network parameter", func(t *testing.T) {
			testCases := []struct {
				name string
				err  string
				prm  apinetmap.NetworkConfig_Parameter
			}{
				{name: "nil key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
					Key: nil, Value: []byte("any"),
				}},
				{name: "empty key", err: "invalid network parameter #1: missing name", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte{}, Value: []byte("any"),
				}},
				{name: "nil value", err: "invalid network parameter #1: missing value", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("any"), Value: nil,
				}},
				{name: "repeated keys", err: "multiple network parameters with name=any_key", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("any_key"), Value: []byte("any"),
				}},
				{name: "audit fee format", err: "invalid network parameter #1 (AuditFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("AuditFee"), Value: []byte("Hello, world!"),
				}},
				{name: "storage price format", err: "invalid network parameter #1 (BasicIncomeRate): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("BasicIncomeRate"), Value: []byte("Hello, world!"),
				}},
				{name: "container fee format", err: "invalid network parameter #1 (ContainerFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("ContainerFee"), Value: []byte("Hello, world!"),
				}},
				{name: "named container fee format", err: "invalid network parameter #1 (ContainerAliasFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("ContainerAliasFee"), Value: []byte("Hello, world!"),
				}},
				{name: "num of EigenTrust iterations format", err: "invalid network parameter #1 (EigenTrustIterations): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustIterations"), Value: []byte("Hello, world!"),
				}},
				{name: "epoch duration format", err: "invalid network parameter #1 (EpochDuration): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EpochDuration"), Value: []byte("Hello, world!"),
				}},
				{name: "IR candidate fee format", err: "invalid network parameter #1 (InnerRingCandidateFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("InnerRingCandidateFee"), Value: []byte("Hello, world!"),
				}},
				{name: "max object size format", err: "invalid network parameter #1 (MaxObjectSize): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("MaxObjectSize"), Value: []byte("Hello, world!"),
				}},
				{name: "withdrawal fee format", err: "invalid network parameter #1 (WithdrawFee): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("WithdrawFee"), Value: []byte("Hello, world!"),
				}},
				{name: "EigenTrust alpha format", err: "invalid network parameter #1 (EigenTrustAlpha): invalid numeric parameter length 13", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte("Hello, world!"),
				}},
				{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value -3.14 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte{31, 133, 235, 81, 184, 30, 9, 192},
				}},
				{name: "negative EigenTrust alpha", err: "invalid network parameter #1 (EigenTrustAlpha): EigenTrust alpha value 1.10 is out of range [0, 1]", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("EigenTrustAlpha"), Value: []byte{154, 153, 153, 153, 153, 153, 241, 63},
				}},
				{name: "disable homomorphic hashing format", err: "invalid network parameter #1 (HomomorphicHashingDisabled): invalid bool parameter", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("HomomorphicHashingDisabled"), Value: make([]byte, 32+1), // max 32
				}},
				{name: "allow maintenance mode format", err: "invalid network parameter #1 (MaintenanceModeAllowed): invalid bool parameter", prm: apinetmap.NetworkConfig_Parameter{
					Key: []byte("MaintenanceModeAllowed"), Value: make([]byte, 32+1), // max 32
				}},
			}

			for i := range testCases {
				n := netmaptest.NetworkInfo()
				var m apinetmap.NetworkInfo

				n.WriteToV2(&m)
				m.NetworkConfig.Parameters = []*apinetmap.NetworkConfig_Parameter{
					{Key: []byte("any_key"), Value: []byte("any_val")},
					&testCases[i].prm,
				}

				b, err := proto.Marshal(&m)
				require.NoError(t, err)
				require.ErrorContains(t, n.Unmarshal(b), testCases[i].err)
			}
		})
	})
}

func testNetworkInfoField[Type uint64 | int64](t *testing.T, get func(netmap.NetworkInfo) Type, set func(*netmap.NetworkInfo, Type),
	getAPI func(info *apinetmap.NetworkInfo) Type) {
	var n netmap.NetworkInfo

	require.Zero(t, get(n))

	const val = 13
	set(&n, val)
	require.EqualValues(t, val, get(n))

	const valOther = 42
	set(&n, valOther)
	require.EqualValues(t, valOther, get(n))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NetworkInfo

			set(&dst, val)

			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.Zero(t, get(dst))

			set(&src, val)

			require.NoError(t, dst.Unmarshal(src.Marshal()))
			require.EqualValues(t, val, get(dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NetworkInfo
			var msg apinetmap.NetworkInfo

			// set required data just to satisfy decoder
			src.SetRawNetworkParameter("any", []byte("any"))

			set(&dst, val)

			src.WriteToV2(&msg)
			require.Zero(t, getAPI(&msg))
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)

			src.WriteToV2(&msg)
			require.EqualValues(t, val, getAPI(&msg))
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
	})
}

func TestNetworkInfo_CurrentEpoch(t *testing.T) {
	testNetworkInfoField(t, netmap.NetworkInfo.CurrentEpoch, (*netmap.NetworkInfo).SetCurrentEpoch, (*apinetmap.NetworkInfo).GetCurrentEpoch)
}

func TestNetworkInfo_MagicNumber(t *testing.T) {
	testNetworkInfoField(t, netmap.NetworkInfo.MagicNumber, (*netmap.NetworkInfo).SetMagicNumber, (*apinetmap.NetworkInfo).GetMagicNumber)
}

func TestNetworkInfo_MsPerBlock(t *testing.T) {
	testNetworkInfoField(t, netmap.NetworkInfo.MsPerBlock, (*netmap.NetworkInfo).SetMsPerBlock, (*apinetmap.NetworkInfo).GetMsPerBlock)
}

func testNetworkConfig[Type uint64 | float64 | bool](t *testing.T, get func(netmap.NetworkInfo) Type, set func(*netmap.NetworkInfo, Type), apiPrm string,
	rand func() (_ Type, api []byte)) {
	var n netmap.NetworkInfo

	require.Zero(t, get(n))

	val, apiVal := rand()
	set(&n, val)
	require.EqualValues(t, val, get(n))

	valOther, _ := rand()
	set(&n, valOther)
	require.EqualValues(t, valOther, get(n))

	t.Run("encoding", func(t *testing.T) {
		t.Run("binary", func(t *testing.T) {
			var src, dst netmap.NetworkInfo

			set(&dst, val)

			err := dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.Zero(t, get(dst))

			set(&src, val)

			err = dst.Unmarshal(src.Marshal())
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))
		})
		t.Run("api", func(t *testing.T) {
			var src, dst netmap.NetworkInfo
			var msg apinetmap.NetworkInfo

			set(&dst, val)

			src.WriteToV2(&msg)
			require.Zero(t, msg.GetNetworkConfig().GetParameters())
			msg.NetworkConfig = &apinetmap.NetworkConfig{
				Parameters: []*apinetmap.NetworkConfig_Parameter{
					{Key: []byte("unique_parameter_unlikely_to_be_set"), Value: []byte("any")},
				},
			}
			require.NoError(t, dst.ReadFromV2(&msg))
			require.Zero(t, get(dst))

			set(&src, val)

			src.WriteToV2(&msg)

			zero := reflect.Zero(reflect.TypeOf(val)).Interface()
			if val == zero {
				require.Zero(t, msg.GetNetworkConfig().GetParameters())
			} else {
				require.Equal(t, []*apinetmap.NetworkConfig_Parameter{
					{Key: []byte(apiPrm), Value: apiVal},
				}, msg.NetworkConfig.Parameters)
			}
			err := dst.ReadFromV2(&msg)
			require.NoError(t, err)
			require.EqualValues(t, val, get(dst))

			if val != zero {
				set(&src, zero.(Type))
				src.WriteToV2(&msg)
				require.Empty(t, msg.GetNetworkConfig().GetParameters())
				err := dst.ReadFromV2(&msg)
				require.NoError(t, err)
				require.Zero(t, get(dst))
			}
		})
	})
}

func testNetworkConfigUint(t *testing.T, get func(netmap.NetworkInfo) uint64, set func(*netmap.NetworkInfo, uint64), apiPrm string) {
	testNetworkConfig(t, get, set, apiPrm, func() (uint64, []byte) {
		n := rand.Uint64()
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, n)
		return n, b
	})
}

func testNetworkConfigFloat(t *testing.T, get func(netmap.NetworkInfo) float64, set func(*netmap.NetworkInfo, float64), apiPrm string) {
	testNetworkConfig(t, get, set, apiPrm, func() (float64, []byte) {
		n := rand.Float64()
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, math.Float64bits(n))
		return n, b
	})
}

func testNetworkConfigBool(t *testing.T, get func(netmap.NetworkInfo) bool, set func(*netmap.NetworkInfo, bool), apiPrm string) {
	testNetworkConfig(t, get, set, apiPrm, func() (bool, []byte) {
		if rand.Int()%2 == 0 {
			return true, []byte{1}
		}
		return false, []byte{0}
	})
}

func TestNetworkInfo_AuditFee(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.AuditFee, (*netmap.NetworkInfo).SetAuditFee, "AuditFee")
}

func TestNetworkInfo_StoragePrice(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.StoragePrice, (*netmap.NetworkInfo).SetStoragePrice, "BasicIncomeRate")
}

func TestNetworkInfo_ContainerFee(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.ContainerFee, (*netmap.NetworkInfo).SetContainerFee, "ContainerFee")
}

func TestNetworkInfo_NamedContainerFee(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.NamedContainerFee, (*netmap.NetworkInfo).SetNamedContainerFee, "ContainerAliasFee")
}

func TestNetworkInfo_EigenTrustAlpha(t *testing.T) {
	testNetworkConfigFloat(t, netmap.NetworkInfo.EigenTrustAlpha, (*netmap.NetworkInfo).SetEigenTrustAlpha, "EigenTrustAlpha")
}

func TestNetworkInfo_NumberOfEigenTrustIterations(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.NumberOfEigenTrustIterations, (*netmap.NetworkInfo).SetNumberOfEigenTrustIterations, "EigenTrustIterations")
}

func TestNetworkInfo_EpochDuration(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.EpochDuration, (*netmap.NetworkInfo).SetEpochDuration, "EpochDuration")
}

func TestNetworkInfo_IRCandidateFee(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.IRCandidateFee, (*netmap.NetworkInfo).SetIRCandidateFee, "InnerRingCandidateFee")
}

func TestNetworkInfo_MaxObjectSize(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.MaxObjectSize, (*netmap.NetworkInfo).SetMaxObjectSize, "MaxObjectSize")
}

func TestNetworkInfo_WithdrawalFee(t *testing.T) {
	testNetworkConfigUint(t, netmap.NetworkInfo.WithdrawalFee, (*netmap.NetworkInfo).SetWithdrawalFee, "WithdrawFee")
}

func TestNetworkInfo_HomomorphicHashingDisabled(t *testing.T) {
	testNetworkConfigBool(t, netmap.NetworkInfo.HomomorphicHashingDisabled, (*netmap.NetworkInfo).SetHomomorphicHashingDisabled, "HomomorphicHashingDisabled")
}

func TestNetworkInfo_MaintenanceModeAllowed(t *testing.T) {
	testNetworkConfigBool(t, netmap.NetworkInfo.MaintenanceModeAllowed, (*netmap.NetworkInfo).SetMaintenanceModeAllowed, "MaintenanceModeAllowed")
}

func TestNetworkInfo_Marshal(t *testing.T) {
	v := netmaptest.NetworkInfo()

	var v2 netmap.NetworkInfo
	require.NoError(t, v2.Unmarshal(v.Marshal()))

	require.Equal(t, v, v2)
}
