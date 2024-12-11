package netmap_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/stretchr/testify/require"
)

const (
	anyValidCurrentEpoch           = uint64(10200868596141730080)
	anyValidMagicNumber            = uint64(4418809875917597199)
	anyValidMSPerBlock             = int64(6618240263362299360)
	anyValidAuditFee               = uint64(439242513058661347)
	anyValidStoragePrice           = uint64(17664525256563393128)
	anyValidNamedContainerFee      = uint64(8341161107066979354)
	anyValidContainerFee           = uint64(6947563960217184460)
	anyValidEigenTrustAlpha        = 0.0508058412675794
	anyValidEigenTrustIterations   = uint64(13357133751475762684)
	anyValidEpochDuration          = uint64(15624600918358785057)
	anyValidHomoHashDisabled       = true
	anyValidIRCandidateFee         = uint64(10380497368037027314)
	anyValidMaintenanceModeAllowed = true
	anyValidMaxObjectSize          = uint64(1434410563277613155)
	anyValidWithdrawalFee          = uint64(4895019021563966975)
)

var (
	anyValidBinAuditFee               = []byte{227, 131, 28, 9, 189, 128, 24, 6}
	anyValidBinStoragePrice           = []byte{104, 130, 96, 83, 33, 0, 37, 245}
	anyValidBinNamedContainerFee      = []byte{26, 4, 8, 97, 213, 193, 193, 115}
	anyValidBinContainerFee           = []byte{204, 40, 13, 175, 156, 180, 106, 96}
	anyValidBinEigenTrustAlpha        = []byte{101, 74, 97, 37, 57, 3, 170, 63}
	anyValidBinEigenTrustIterations   = []byte{252, 245, 23, 186, 96, 18, 94, 185}
	anyValidBinEpochDuration          = []byte{33, 124, 4, 168, 192, 187, 213, 216}
	anyValidBinHomoHashDisabled       = []byte{1}
	anyValidBinIRCandidateFee         = []byte{242, 109, 185, 165, 91, 239, 14, 144}
	anyValidBinMaintenanceModeAllowed = []byte{1}
	anyValidBinMaxObjectSize          = []byte{99, 232, 58, 182, 206, 12, 232, 19}
	anyValidBinWithdrawalFee          = []byte{255, 181, 25, 125, 221, 153, 238, 67}
)

// set by init.
var validNetworkInfo netmap.NetworkInfo

func init() {
	validNetworkInfo.SetCurrentEpoch(anyValidCurrentEpoch)
	validNetworkInfo.SetMagicNumber(anyValidMagicNumber)
	validNetworkInfo.SetMsPerBlock(anyValidMSPerBlock)
	validNetworkInfo.SetRawNetworkParameter("k1", []byte("v1"))
	validNetworkInfo.SetRawNetworkParameter("k2", []byte("v2"))
	validNetworkInfo.SetAuditFee(anyValidAuditFee)
	validNetworkInfo.SetStoragePrice(anyValidStoragePrice)
	validNetworkInfo.SetNamedContainerFee(anyValidNamedContainerFee)
	validNetworkInfo.SetContainerFee(anyValidContainerFee)
	validNetworkInfo.SetEigenTrustAlpha(anyValidEigenTrustAlpha)
	validNetworkInfo.SetNumberOfEigenTrustIterations(anyValidEigenTrustIterations)
	validNetworkInfo.SetEpochDuration(anyValidEpochDuration)
	validNetworkInfo.DisableHomomorphicHashing()
	validNetworkInfo.SetIRCandidateFee(anyValidIRCandidateFee)
	validNetworkInfo.AllowMaintenanceMode()
	validNetworkInfo.SetMaxObjectSize(anyValidMaxObjectSize)
	validNetworkInfo.SetWithdrawalFee(anyValidWithdrawalFee)
}

var validBinNetworkInfo = []byte{
	8, 160, 210, 220, 139, 145, 254, 176, 200, 141, 1, 16, 143, 212, 185, 192, 185, 133, 177, 169, 61, 24, 224, 139, 151, 255,
	197, 206, 173, 236, 91, 34, 239, 2, 10, 8, 10, 2, 107, 49, 18, 2, 118, 49, 10, 8, 10, 2, 107, 50, 18, 2, 118, 50, 10, 20,
	10, 8, 65, 117, 100, 105, 116, 70, 101, 101, 18, 8, 227, 131, 28, 9, 189, 128, 24, 6, 10, 27, 10, 15, 66, 97, 115, 105, 99, 73,
	110, 99, 111, 109, 101, 82, 97, 116, 101, 18, 8, 104, 130, 96, 83, 33, 0, 37, 245, 10, 29, 10, 17, 67, 111, 110, 116, 97, 105, 110,
	101, 114, 65, 108, 105, 97, 115, 70, 101, 101, 18, 8, 26, 4, 8, 97, 213, 193, 193, 115, 10, 24, 10, 12, 67, 111, 110, 116, 97, 105,
	110, 101, 114, 70, 101, 101, 18, 8, 204, 40, 13, 175, 156, 180, 106, 96, 10, 27, 10, 15, 69, 105, 103, 101, 110, 84, 114, 117, 115,
	116, 65, 108, 112, 104, 97, 18, 8, 101, 74, 97, 37, 57, 3, 170, 63, 10, 32, 10, 20, 69, 105, 103, 101, 110, 84, 114, 117, 115, 116,
	73, 116, 101, 114, 97, 116, 105, 111, 110, 115, 18, 8, 252, 245, 23, 186, 96, 18, 94, 185, 10, 25, 10, 13, 69, 112, 111, 99, 104,
	68, 117, 114, 97, 116, 105, 111, 110, 18, 8, 33, 124, 4, 168, 192, 187, 213, 216, 10, 31, 10, 26, 72, 111, 109, 111, 109, 111, 114,
	112, 104, 105, 99, 72, 97, 115, 104, 105, 110, 103, 68, 105, 115, 97, 98, 108, 101, 100, 18, 1, 1, 10, 33, 10, 21, 73, 110, 110,
	101, 114, 82, 105, 110, 103, 67, 97, 110, 100, 105, 100, 97, 116, 101, 70, 101, 101, 18, 8, 242, 109, 185, 165, 91, 239, 14, 144,
	10, 27, 10, 22, 77, 97, 105, 110, 116, 101, 110, 97, 110, 99, 101, 77, 111, 100, 101, 65, 108, 108, 111, 119, 101, 100, 18, 1, 1, 10,
	25, 10, 13, 77, 97, 120, 79, 98, 106, 101, 99, 116, 83, 105, 122, 101, 18, 8, 99, 232, 58, 182, 206, 12, 232, 19, 10, 23,
	10, 11, 87, 105, 116, 104, 100, 114, 97, 119, 70, 101, 101, 18, 8, 255, 181, 25, 125, 221, 153, 238, 67,
}

func TestNetworkInfo_CurrentEpoch(t *testing.T) {
	var x netmap.NetworkInfo
	require.Zero(t, x.CurrentEpoch())

	const e = 13
	x.SetCurrentEpoch(e)
	require.EqualValues(t, e, x.CurrentEpoch())

	const e2 = e + 1
	x.SetCurrentEpoch(e2)
	require.EqualValues(t, e2, x.CurrentEpoch())
}

func TestNetworkInfo_MagicNumber(t *testing.T) {
	var x netmap.NetworkInfo
	require.Zero(t, x.MagicNumber())

	const magic = 321
	x.SetMagicNumber(magic)
	require.EqualValues(t, magic, x.MagicNumber())

	const magic2 = magic + 1
	x.SetMagicNumber(magic2)
	require.EqualValues(t, magic2, x.MagicNumber())
}

func TestNetworkInfo_MsPerBlock(t *testing.T) {
	var x netmap.NetworkInfo
	require.Zero(t, x.MsPerBlock())

	const ms = 789
	x.SetMsPerBlock(ms)
	require.EqualValues(t, ms, x.MsPerBlock())

	const ms2 = ms + 1
	x.SetMsPerBlock(ms2)
	require.EqualValues(t, ms2, x.MsPerBlock())
}

func TestNetworkInfo_SetRawNetworkParameter(t *testing.T) {
	var x netmap.NetworkInfo
	const k1, v1 = "k1", "v1"
	const k2, v2 = "k2", "v2"

	require.Zero(t, x.RawNetworkParameter(k1))
	require.Zero(t, x.RawNetworkParameter(k2))
	x.IterateRawNetworkParameters(func(string, []byte) {
		t.Fatal("handler must not be called")
	})

	x.SetRawNetworkParameter(k1, []byte(v1))
	x.SetRawNetworkParameter(k2, []byte(v2))

	require.EqualValues(t, v1, x.RawNetworkParameter(k1))
	require.EqualValues(t, v2, x.RawNetworkParameter(k2))
	var collected [][2]string
	x.IterateRawNetworkParameters(func(name string, value []byte) {
		collected = append(collected, [2]string{name, string(value)})
	})
	require.ElementsMatch(t, [][2]string{{k1, v1}, {k2, v2}}, collected)
}

func testConfigValue[T comparable](t testing.TB,
	getter func(netmap.NetworkInfo) T,
	setter func(*netmap.NetworkInfo, T),
	val1, val2 T,
) {
	require.NotEqual(t, val1, val2)
	var x netmap.NetworkInfo
	require.Zero(t, getter(x))
	setter(&x, val1)
	require.Equal(t, val1, getter(x))
	setter(&x, val2)
	require.Equal(t, val2, getter(x))
}

func TestNetworkInfo_SetAuditFee(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.AuditFee, (*netmap.NetworkInfo).SetAuditFee, 1, 2)
}

func TestNetworkInfo_SetStoragePrice(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.StoragePrice, (*netmap.NetworkInfo).SetStoragePrice, 1, 2)
}

func TestNetworkInfo_SetContainerFee(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.ContainerFee, (*netmap.NetworkInfo).SetContainerFee, 1, 2)
}

func TestNetworkInfo_SetNamedContainerFee(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.NamedContainerFee, (*netmap.NetworkInfo).SetNamedContainerFee, 1, 2)
}

func TestNetworkInfo_SetEigenTrustAlpha(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.EigenTrustAlpha, (*netmap.NetworkInfo).SetEigenTrustAlpha, 0.1, 0.2)
	require.Panics(t, func() { new(netmap.NetworkInfo).SetEigenTrustAlpha(-0.5) })
	require.Panics(t, func() { new(netmap.NetworkInfo).SetEigenTrustAlpha(1.5) })
}

func TestNetworkInfo_SetNumberOfEigenTrustIterations(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.NumberOfEigenTrustIterations, (*netmap.NetworkInfo).SetNumberOfEigenTrustIterations, 1, 2)
}

func TestNetworkInfo_SetEpochDuration(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.EpochDuration, (*netmap.NetworkInfo).SetEpochDuration, 1, 2)
}

func TestNetworkInfo_SetIRCandidateFee(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.IRCandidateFee, (*netmap.NetworkInfo).SetIRCandidateFee, 1, 2)
}

func TestNetworkInfo_SetMaxObjectSize(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.IRCandidateFee, (*netmap.NetworkInfo).SetIRCandidateFee, 1, 2)
}

func TestNetworkInfo_SetWithdrawalFee(t *testing.T) {
	testConfigValue(t, netmap.NetworkInfo.WithdrawalFee, (*netmap.NetworkInfo).SetWithdrawalFee, 1, 2)
}

func TestNetworkInfo_DisableHomomorphicHashing(t *testing.T) {
	var x netmap.NetworkInfo
	require.False(t, x.HomomorphicHashingDisabled())
	x.DisableHomomorphicHashing()
	require.True(t, x.HomomorphicHashingDisabled())
}

func TestNetworkInfo_AllowMaintenanceMode(t *testing.T) {
	var x netmap.NetworkInfo
	require.False(t, x.MaintenanceModeAllowed())
	x.AllowMaintenanceMode()
	require.True(t, x.MaintenanceModeAllowed())
}

func setNetworkPrms[T string | []byte](ni *protonetmap.NetworkInfo, els ...T) {
	if len(els)%2 != 0 {
		panic("must be even")
	}
	ni.NetworkConfig.Parameters = make([]*protonetmap.NetworkConfig_Parameter, len(els)/2)
	for i := range len(els) / 2 {
		ni.NetworkConfig.Parameters[i] = &protonetmap.NetworkConfig_Parameter{
			Key:   []byte(els[2*i]),
			Value: []byte(els[2*i+1]),
		}
	}
}

func TestNetworkInfo_FromProtoMessage(t *testing.T) {
	m := &protonetmap.NetworkInfo{
		CurrentEpoch: anyValidCurrentEpoch,
		MagicNumber:  anyValidMagicNumber,
		MsPerBlock:   anyValidMSPerBlock,
		NetworkConfig: &protonetmap.NetworkConfig{
			Parameters: []*protonetmap.NetworkConfig_Parameter{
				{Key: []byte("k1"), Value: []byte("v1")},
				{Key: []byte("k2"), Value: []byte("v2")},
				{Key: []byte("AuditFee"), Value: anyValidBinAuditFee},
				{Key: []byte("BasicIncomeRate"), Value: anyValidBinStoragePrice},
				{Key: []byte("ContainerAliasFee"), Value: anyValidBinNamedContainerFee},
				{Key: []byte("ContainerFee"), Value: anyValidBinContainerFee},
				{Key: []byte("EigenTrustAlpha"), Value: anyValidBinEigenTrustAlpha},
				{Key: []byte("EigenTrustIterations"), Value: anyValidBinEigenTrustIterations},
				{Key: []byte("EpochDuration"), Value: anyValidBinEpochDuration},
				{Key: []byte("HomomorphicHashingDisabled"), Value: anyValidBinHomoHashDisabled},
				{Key: []byte("InnerRingCandidateFee"), Value: anyValidBinIRCandidateFee},
				{Key: []byte("MaintenanceModeAllowed"), Value: anyValidBinMaintenanceModeAllowed},
				{Key: []byte("MaxObjectSize"), Value: anyValidBinMaxObjectSize},
				{Key: []byte("WithdrawFee"), Value: anyValidBinWithdrawalFee},
			},
		},
	}

	var val netmap.NetworkInfo
	require.NoError(t, val.FromProtoMessage(m))
	require.EqualValues(t, "v1", val.RawNetworkParameter("k1"))
	require.EqualValues(t, "v2", val.RawNetworkParameter("k2"))
	require.Equal(t, anyValidCurrentEpoch, val.CurrentEpoch())
	require.Equal(t, anyValidMagicNumber, val.MagicNumber())
	require.Equal(t, anyValidMSPerBlock, val.MsPerBlock())
	require.Equal(t, anyValidAuditFee, val.AuditFee())
	require.Equal(t, anyValidStoragePrice, val.StoragePrice())
	require.Equal(t, anyValidNamedContainerFee, val.NamedContainerFee())
	require.Equal(t, anyValidContainerFee, val.ContainerFee())
	require.Equal(t, anyValidEigenTrustAlpha, val.EigenTrustAlpha())
	require.Equal(t, anyValidEigenTrustIterations, val.NumberOfEigenTrustIterations())
	require.Equal(t, anyValidEpochDuration, val.EpochDuration())
	require.Equal(t, anyValidHomoHashDisabled, val.HomomorphicHashingDisabled())
	require.Equal(t, anyValidIRCandidateFee, val.IRCandidateFee())
	require.Equal(t, anyValidMaintenanceModeAllowed, val.MaintenanceModeAllowed())
	require.Equal(t, anyValidMaxObjectSize, val.MaxObjectSize())
	require.Equal(t, anyValidWithdrawalFee, val.WithdrawalFee())

	// reset optional fields
	m.NetworkConfig.Parameters = m.NetworkConfig.Parameters[:1]
	m.CurrentEpoch = 0
	m.MagicNumber = 0
	m.MsPerBlock = 0
	val2 := val
	require.NoError(t, val2.FromProtoMessage(m))
	require.EqualValues(t, "v1", val.RawNetworkParameter("k1"))
	require.Zero(t, val2.RawNetworkParameter("k2"))
	require.Zero(t, val2.CurrentEpoch())
	require.Zero(t, val2.MagicNumber())
	require.Zero(t, val2.CurrentEpoch())
	require.Zero(t, val2.AuditFee())
	require.Zero(t, val2.StoragePrice())
	require.Zero(t, val2.NamedContainerFee())
	require.Zero(t, val2.ContainerFee())
	require.Zero(t, val2.EigenTrustAlpha())
	require.Zero(t, val2.NumberOfEigenTrustIterations())
	require.Zero(t, val2.EpochDuration())
	require.Zero(t, val2.HomomorphicHashingDisabled())
	require.Zero(t, val2.IRCandidateFee())
	require.Zero(t, val2.MaintenanceModeAllowed())
	require.Zero(t, val2.MaxObjectSize())
	require.Zero(t, val2.WithdrawalFee())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*protonetmap.NetworkInfo)
		}{
			{name: "netconfig/missing", err: "missing network config",
				corrupt: func(m *protonetmap.NetworkInfo) { m.NetworkConfig = nil }},
			{name: "netconfig/prms/missing", err: "missing network parameters",
				corrupt: func(m *protonetmap.NetworkInfo) { m.NetworkConfig = new(protonetmap.NetworkConfig) }},
			{name: "netconfig/prms/nil", err: "nil parameter #1",
				corrupt: func(m *protonetmap.NetworkInfo) {
					m.NetworkConfig.Parameters[1] = nil
				}},
			{name: "netconfig/prms/no value", err: `empty "k1" parameter value`,
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, "k1", "") }},
			{name: "netconfig/prms/duplicated", err: "duplicated parameter name: k1",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, "k1", "v1", "k2", "v2", "k1", "v3") }},
			{name: "netconfig/prms/eigen trust alpha/overflow", err: "invalid EigenTrustAlpha parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, "EigenTrustAlpha", "123456789") }},
			{name: "netconfig/prms/eigen trust alpha/negative", err: "invalid EigenTrustAlpha parameter: EigenTrust alpha value -0.50 is out of range [0, 1]",
				corrupt: func(m *protonetmap.NetworkInfo) {
					setNetworkPrms(m, []byte("EigenTrustAlpha"), []byte{0, 0, 0, 0, 0, 0, 224, 191})
				}},
			{name: "netconfig/prms/eigen trust alpha/too big", err: "invalid EigenTrustAlpha parameter: EigenTrust alpha value 1.50 is out of range [0, 1]",
				corrupt: func(m *protonetmap.NetworkInfo) {
					setNetworkPrms(m, []byte("EigenTrustAlpha"), []byte{0, 0, 0, 0, 0, 0, 248, 63})
				}},
			{name: "netconfig/prms/homo hash disabled/overflow", err: "invalid HomomorphicHashingDisabled parameter: invalid bool parameter contract format too big: integer",
				corrupt: func(m *protonetmap.NetworkInfo) {
					setNetworkPrms(m, []byte("HomomorphicHashingDisabled"), make([]byte, 33))
				}},
			{name: "netconfig/prms/maintenance allowed/overflow", err: "invalid MaintenanceModeAllowed parameter: invalid bool parameter contract format too big: integer",
				corrupt: func(m *protonetmap.NetworkInfo) {
					setNetworkPrms(m, []byte("MaintenanceModeAllowed"), make([]byte, 33))
				}},
			{name: "netconfig/prms/audit fee/overflow", err: "invalid AuditFee parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("AuditFee"), make([]byte, 9)) }},
			{name: "netconfig/prms/storage price/overflow", err: "invalid BasicIncomeRate parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("BasicIncomeRate"), make([]byte, 9)) }},
			{name: "netconfig/prms/container fee/overflow", err: "invalid ContainerFee parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("ContainerFee"), make([]byte, 9)) }},
			{name: "netconfig/prms/named container fee/overflow", err: "invalid ContainerAliasFee parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("ContainerAliasFee"), make([]byte, 9)) }},
			{name: "netconfig/prms/eigen trust iterations/overflow", err: "invalid EigenTrustIterations parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("EigenTrustIterations"), make([]byte, 9)) }},
			{name: "netconfig/prms/epoch duration/overflow", err: "invalid EpochDuration parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("EpochDuration"), make([]byte, 9)) }},
			{name: "netconfig/prms/ir candidate fee/overflow", err: "invalid InnerRingCandidateFee parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("InnerRingCandidateFee"), make([]byte, 9)) }},
			{name: "netconfig/prms/max object size/overflow", err: "invalid MaxObjectSize parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("MaxObjectSize"), make([]byte, 9)) }},
			{name: "netconfig/prms/withdrawal fee/overflow", err: "invalid WithdrawFee parameter: invalid uint64 parameter length 9",
				corrupt: func(m *protonetmap.NetworkInfo) { setNetworkPrms(m, []byte("WithdrawFee"), make([]byte, 9)) }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(netmap.NetworkInfo).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestNetworkInfo_ProtoMessage(t *testing.T) {
	var val netmap.NetworkInfo

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetCurrentEpoch())
	require.Zero(t, m.GetMagicNumber())
	require.Zero(t, m.GetMsPerBlock())
	require.Zero(t, m.GetNetworkConfig())

	// filled
	m = validNetworkInfo.ProtoMessage()
	require.Equal(t, anyValidCurrentEpoch, m.GetCurrentEpoch())
	require.Equal(t, anyValidMagicNumber, m.GetMagicNumber())
	require.Equal(t, anyValidMSPerBlock, m.GetMsPerBlock())
	mc := m.GetNetworkConfig()
	require.NotNil(t, mc)
	require.Len(t, mc.Parameters, 14)
	for i, pair := range [][2]any{
		{"k1", "v1"},
		{"k2", "v2"},
		{"AuditFee", anyValidBinAuditFee},
		{"BasicIncomeRate", anyValidBinStoragePrice},
		{"ContainerAliasFee", anyValidBinNamedContainerFee},
		{"ContainerFee", anyValidBinContainerFee},
		{"EigenTrustAlpha", anyValidBinEigenTrustAlpha},
		{"EigenTrustIterations", anyValidBinEigenTrustIterations},
		{"EpochDuration", anyValidBinEpochDuration},
		{"HomomorphicHashingDisabled", anyValidBinHomoHashDisabled},
		{"InnerRingCandidateFee", anyValidBinIRCandidateFee},
		{"MaintenanceModeAllowed", anyValidBinMaintenanceModeAllowed},
		{"MaxObjectSize", anyValidBinMaxObjectSize},
		{"WithdrawFee", anyValidBinWithdrawalFee},
	} {
		require.EqualValues(t, pair[0], mc.Parameters[i].Key)
		require.EqualValues(t, pair[1], mc.Parameters[i].Value)
	}
}

func TestNetworkInfo_Marshal(t *testing.T) {
	require.Equal(t, validBinNetworkInfo, validNetworkInfo.Marshal())
}

func TestNetworkInfo_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(netmap.NetworkInfo).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name string
			err  string
			b    []byte
		}{
			{name: "netconfig/prms/no value", err: `empty "k1" parameter value`,
				b: []byte{34, 6, 10, 4, 10, 2, 107, 49}},
			{name: "netconfig/prms/duplicated", err: "duplicated parameter name: k1",
				b: []byte{34, 30, 10, 8, 10, 2, 107, 49, 18, 2, 118, 49, 10, 8, 10, 2, 107, 50, 18, 2, 118, 50, 10, 8, 10, 2, 107,
					49, 18, 2, 118, 51}},
			{name: "netconfig/prms/eigen trust alpha/overflow", err: "invalid EigenTrustAlpha parameter: invalid uint64 parameter length 9",
				b: []byte{34, 30, 10, 28, 10, 15, 69, 105, 103, 101, 110, 84, 114, 117, 115, 116, 65, 108, 112, 104, 97, 18, 9, 49, 50,
					51, 52, 53, 54, 55, 56, 57}},
			{name: "netconfig/prms/eigen trust alpha/negative", err: "invalid EigenTrustAlpha parameter: EigenTrust alpha value -0.50 is out of range [0, 1]",
				b: []byte{34, 29, 10, 27, 10, 15, 69, 105, 103, 101, 110, 84, 114, 117, 115, 116, 65, 108, 112, 104, 97, 18, 8, 0, 0,
					0, 0, 0, 0, 224, 191}},
			{name: "netconfig/prms/eigen trust alpha/too big", err: "invalid EigenTrustAlpha parameter: EigenTrust alpha value 1.50 is out of range [0, 1]",
				b: []byte{34, 29, 10, 27, 10, 15, 69, 105, 103, 101, 110, 84, 114, 117, 115, 116, 65, 108, 112, 104, 97, 18, 8, 0, 0,
					0, 0, 0, 0, 248, 63}},
			{name: "netconfig/prms/homo hash disabled/overflow", err: "invalid HomomorphicHashingDisabled parameter: invalid bool parameter contract format too big: integer",
				b: []byte{34, 65, 10, 63, 10, 26, 72, 111, 109, 111, 109, 111, 114, 112, 104, 105, 99, 72, 97, 115, 104, 105, 110, 103,
					68, 105, 115, 97, 98, 108, 101, 100, 18, 33, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/maintenance allowed/overflow", err: "invalid MaintenanceModeAllowed parameter: invalid bool parameter contract format too big: integer",
				b: []byte{34, 61, 10, 59, 10, 22, 77, 97, 105, 110, 116, 101, 110, 97, 110, 99, 101, 77, 111, 100, 101, 65, 108, 108, 111,
					119, 101, 100, 18, 33, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
					0, 0, 0, 0}},
			{name: "netconfig/prms/audit fee/overflow", err: "invalid AuditFee parameter: invalid uint64 parameter length 9",
				b: []byte{34, 23, 10, 21, 10, 8, 65, 117, 100, 105, 116, 70, 101, 101, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/storage price/overflow", err: "invalid BasicIncomeRate parameter: invalid uint64 parameter length 9",
				b: []byte{34, 30, 10, 28, 10, 15, 66, 97, 115, 105, 99, 73, 110, 99, 111, 109, 101, 82, 97, 116, 101, 18, 9, 0, 0, 0, 0,
					0, 0, 0, 0, 0}},
			{name: "netconfig/prms/container fee/overflow", err: "invalid ContainerFee parameter: invalid uint64 parameter length 9",
				b: []byte{34, 27, 10, 25, 10, 12, 67, 111, 110, 116, 97, 105, 110, 101, 114, 70, 101, 101, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/named container fee/overflow", err: "invalid ContainerAliasFee parameter: invalid uint64 parameter length 9",
				b: []byte{34, 32, 10, 30, 10, 17, 67, 111, 110, 116, 97, 105, 110, 101, 114, 65, 108, 105, 97, 115, 70, 101, 101, 18, 9, 0,
					0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/eigen trust iterations/overflow", err: "invalid EigenTrustIterations parameter: invalid uint64 parameter length 9",
				b: []byte{34, 35, 10, 33, 10, 20, 69, 105, 103, 101, 110, 84, 114, 117, 115, 116, 73, 116, 101, 114, 97, 116, 105, 111, 110,
					115, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/epoch duration/overflow", err: "invalid EpochDuration parameter: invalid uint64 parameter length 9",
				b: []byte{34, 28, 10, 26, 10, 13, 69, 112, 111, 99, 104, 68, 117, 114, 97, 116, 105, 111, 110, 18, 9, 0, 0, 0, 0, 0, 0,
					0, 0, 0}},
			{name: "netconfig/prms/ir candidate fee/overflow", err: "invalid InnerRingCandidateFee parameter: invalid uint64 parameter length 9",
				b: []byte{34, 36, 10, 34, 10, 21, 73, 110, 110, 101, 114, 82, 105, 110, 103, 67, 97, 110, 100, 105, 100, 97, 116, 101,
					70, 101, 101, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
			{name: "netconfig/prms/max object size/overflow", err: "invalid MaxObjectSize parameter: invalid uint64 parameter length 9",
				b: []byte{34, 28, 10, 26, 10, 13, 77, 97, 120, 79, 98, 106, 101, 99, 116, 83, 105, 122, 101, 18, 9, 0, 0, 0, 0, 0,
					0, 0, 0, 0}},
			{name: "netconfig/prms/withdrawal fee/overflow", err: "invalid WithdrawFee parameter: invalid uint64 parameter length 9",
				b: []byte{34, 26, 10, 24, 10, 11, 87, 105, 116, 104, 100, 114, 97, 119, 70, 101, 101, 18, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(netmap.NetworkInfo).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var val netmap.NetworkInfo
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val.RawNetworkParameter("k1"))
	require.Zero(t, val.RawNetworkParameter("k2"))
	require.Zero(t, val.CurrentEpoch())
	require.Zero(t, val.MagicNumber())
	require.Zero(t, val.CurrentEpoch())
	require.Zero(t, val.AuditFee())
	require.Zero(t, val.StoragePrice())
	require.Zero(t, val.NamedContainerFee())
	require.Zero(t, val.ContainerFee())
	require.Zero(t, val.EigenTrustAlpha())
	require.Zero(t, val.NumberOfEigenTrustIterations())
	require.Zero(t, val.EpochDuration())
	require.Zero(t, val.HomomorphicHashingDisabled())
	require.Zero(t, val.IRCandidateFee())
	require.Zero(t, val.MaintenanceModeAllowed())
	require.Zero(t, val.MaxObjectSize())
	require.Zero(t, val.WithdrawalFee())

	// filled
	require.NoError(t, val.Unmarshal(validBinNetworkInfo))
	require.Equal(t, validNetworkInfo, val)
}
