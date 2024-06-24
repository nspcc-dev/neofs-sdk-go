package netmap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"google.golang.org/protobuf/proto"
)

// NetworkInfo groups information about the NeoFS network state. Mainly used to
// describe the current state of the network.
//
// NetworkInfo is mutually compatible with [netmap.NetworkInfo] message. See
// [NetworkInfo.ReadFromV2] / [NetworkInfo.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type NetworkInfo struct {
	curEpoch    uint64
	magicNumber uint64
	msPerBlock  int64
	prms        []*netmap.NetworkConfig_Parameter
}

func (x *NetworkInfo) readFromV2(m *netmap.NetworkInfo) error {
	var err error
	ps := m.GetNetworkConfig().GetParameters()
	for i := range ps {
		k := ps[i].GetKey()
		if len(k) == 0 {
			return fmt.Errorf("invalid network parameter #%d: missing name", i)
		}
		// further NPE are prevented by condition above
		if len(ps[i].Value) == 0 {
			return fmt.Errorf("invalid network parameter #%d: missing value", i)
		}

		for j := 0; j < i; j++ {
			if bytes.Equal(ps[j].Key, k) {
				return fmt.Errorf("multiple network parameters with name=%s", k)
			}
		}

		switch {
		case bytes.Equal(k, configEigenTrustAlpha):
			var num uint64

			num, err = decodeConfigValueUint64(ps[i].Value)
			if err == nil {
				if alpha := math.Float64frombits(num); alpha < 0 || alpha > 1 {
					err = fmt.Errorf("EigenTrust alpha value %0.2f is out of range [0, 1]", alpha)
				}
			}
		case
			bytes.Equal(k, configAuditFee),
			bytes.Equal(k, configStoragePrice),
			bytes.Equal(k, configContainerFee),
			bytes.Equal(k, configNamedContainerFee),
			bytes.Equal(k, configEigenTrustNumberOfIterations),
			bytes.Equal(k, configEpochDuration),
			bytes.Equal(k, configIRCandidateFee),
			bytes.Equal(k, configMaxObjSize),
			bytes.Equal(k, configWithdrawalFee):
			_, err = decodeConfigValueUint64(ps[i].Value)
		case bytes.Equal(k, configHomomorphicHashingDisabled),
			bytes.Equal(k, configMaintenanceModeAllowed):
			_, err = decodeConfigValueBool(ps[i].Value)
		}

		if err != nil {
			return fmt.Errorf("invalid network parameter #%d (%s): %w", i, k, err)
		}
	}

	x.curEpoch = m.GetCurrentEpoch()
	x.magicNumber = m.GetMagicNumber()
	x.msPerBlock = m.GetMsPerBlock()
	x.prms = ps

	return nil
}

// ReadFromV2 reads NetworkInfo from the [netmap.NetworkInfo] message. Returns
// an error if the message is malformed according to the NeoFS API V2 protocol.
// The message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [NetworkInfo.WriteToV2].
func (x *NetworkInfo) ReadFromV2(m *netmap.NetworkInfo) error {
	return x.readFromV2(m)
}

// WriteToV2 writes NetworkInfo to the [netmap.NetworkInfo] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [NetworkInfo.ReadFromV2].
func (x NetworkInfo) WriteToV2(m *netmap.NetworkInfo) {
	if x.prms != nil {
		m.NetworkConfig = &netmap.NetworkConfig{Parameters: x.prms}
	} else {
		m.NetworkConfig = nil
	}

	m.CurrentEpoch = x.curEpoch
	m.MagicNumber = x.magicNumber
	m.MsPerBlock = x.msPerBlock
}

// Marshal encodes NetworkInfo into a binary format of the NeoFS API
// protocol (Protocol Buffers V3 with direct field order).
//
// See also [NetworkInfo.Unmarshal].
func (x NetworkInfo) Marshal() []byte {
	var m netmap.NetworkInfo
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the NetworkInfo.
// Returns an error describing a format violation of the specified fields.
// Unmarshal does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [NetworkInfo.Marshal].
func (x *NetworkInfo) Unmarshal(data []byte) error {
	var m netmap.NetworkInfo
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return x.readFromV2(&m)
}

// CurrentEpoch returns epoch set using [NetworkInfo.SetCurrentEpoch].
//
// Zero NetworkInfo has zero current epoch.
func (x NetworkInfo) CurrentEpoch() uint64 {
	return x.curEpoch
}

// SetCurrentEpoch sets current epoch of the NeoFS network.
//
// See also [NetworkInfo.CurrentEpoch].
func (x *NetworkInfo) SetCurrentEpoch(epoch uint64) {
	x.curEpoch = epoch
}

// MagicNumber returns magic number set using [NetworkInfo.SetMagicNumber].
//
// Zero NetworkInfo has zero magic.
func (x NetworkInfo) MagicNumber() uint64 {
	return x.magicNumber
}

// SetMagicNumber sets magic number of the NeoFS Sidechain.
//
// See also [NetworkInfo.MagicNumber].
func (x *NetworkInfo) SetMagicNumber(magic uint64) {
	x.magicNumber = magic
}

// MsPerBlock returns network parameter set using [NetworkInfo.SetMsPerBlock].
func (x NetworkInfo) MsPerBlock() int64 {
	return x.msPerBlock
}

// SetMsPerBlock sets MillisecondsPerBlock network parameter of the NeoFS Sidechain.
//
// See also [NetworkInfo.MsPerBlock].
func (x *NetworkInfo) SetMsPerBlock(v int64) {
	x.msPerBlock = v
}

func (x *NetworkInfo) setConfig(name, val []byte) {
	for i := range x.prms {
		if bytes.Equal(x.prms[i].GetKey(), name) {
			x.prms[i].Value = val
			return
		}
	}
	x.prms = append(x.prms, &netmap.NetworkConfig_Parameter{
		Key:   name,
		Value: val,
	})
}

func (x *NetworkInfo) resetConfig(name []byte) {
	for i := 0; i < len(x.prms); i++ { // do not use range, slice is changed inside
		if bytes.Equal(x.prms[i].GetKey(), name) {
			x.prms = append(x.prms[:i], x.prms[i+1:]...)
			i--
		}
	}
}

func (x NetworkInfo) configValue(name []byte) []byte {
	for i := range x.prms {
		if bytes.Equal(x.prms[i].GetKey(), name) {
			return x.prms[i].Value
		}
	}
	return nil
}

// SetRawNetworkParameter sets named NeoFS network parameter whose value is
// transmitted but not interpreted by the NeoFS API protocol.
//
// Argument MUST NOT be mutated, make a copy first.
//
// See also [NetworkInfo.RawNetworkParameter],
// [NetworkInfo.IterateRawNetworkParameters].
func (x *NetworkInfo) SetRawNetworkParameter(name string, value []byte) {
	x.setConfig([]byte(name), value)
}

// RawNetworkParameter reads raw network parameter set using [NetworkInfo.SetRawNetworkParameter]
// by its name. Returns nil if value is missing.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// Zero NetworkInfo has no network parameters.
func (x NetworkInfo) RawNetworkParameter(name string) []byte {
	return x.configValue([]byte(name))
}

// IterateRawNetworkParameters iterates over all raw networks parameters set
// using [NetworkInfo.SetRawNetworkParameter] and passes them into f.
//
// Handler MUST NOT be nil. Handler MUST NOT mutate value parameter.
//
// Zero NetworkInfo has no network parameters.
func (x NetworkInfo) IterateRawNetworkParameters(f func(name string, value []byte)) {
	for i := range x.prms {
		key := x.prms[i].GetKey()
		switch {
		default:
			f(string(key), x.prms[i].GetValue())
		case
			bytes.Equal(key, configEigenTrustAlpha),
			bytes.Equal(key, configAuditFee),
			bytes.Equal(key, configStoragePrice),
			bytes.Equal(key, configContainerFee),
			bytes.Equal(key, configNamedContainerFee),
			bytes.Equal(key, configEigenTrustNumberOfIterations),
			bytes.Equal(key, configEpochDuration),
			bytes.Equal(key, configIRCandidateFee),
			bytes.Equal(key, configMaxObjSize),
			bytes.Equal(key, configWithdrawalFee),
			bytes.Equal(key, configHomomorphicHashingDisabled),
			bytes.Equal(key, configMaintenanceModeAllowed):
		}
	}
}

func (x *NetworkInfo) setConfigUint64(name []byte, num uint64) {
	if num == 0 {
		x.resetConfig(name)
		return
	}

	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, num)

	x.setConfig(name, val)
}

func (x *NetworkInfo) setConfigBool(name []byte, val bool) {
	if !val {
		x.resetConfig(name)
		return
	}
	x.setConfig(name, []byte{1})
}

// decodeConfigValueUint64 parses val as little-endian uint64.
// val must be less than 8 bytes in size.
func decodeConfigValueUint64(val []byte) (uint64, error) {
	if ln := len(val); ln > 8 {
		return 0, fmt.Errorf("invalid numeric parameter length %d", ln)
	}

	res := uint64(0)
	for i := len(val) - 1; i >= 0; i-- {
		res = res*256 + uint64(val[i])
	}

	return res, nil
}

// decodeConfigValueBool parses val as boolean contract storage value.
func decodeConfigValueBool(val []byte) (bool, error) {
	arr := stackitem.NewByteArray(val)

	res, err := arr.TryBool()
	if err != nil {
		return false, fmt.Errorf("invalid bool parameter: %w", err)
	}

	return res, nil
}

func (x NetworkInfo) configUint64(name []byte) uint64 {
	val := x.configValue(name)
	if val == nil {
		return 0
	}

	res, err := decodeConfigValueUint64(val)
	if err != nil {
		// potential panic is OK since value MUST be correct since it is
		// verified in ReadFromV2 or set by provided method.
		panic(err)
	}

	return res
}

func (x NetworkInfo) configBool(name []byte) bool {
	val := x.configValue(name)
	if val == nil {
		return false
	}

	res, err := decodeConfigValueBool(val)
	if err != nil {
		// potential panic is OK since value MUST be correct since it is
		// verified in ReadFromV2 or set by provided method.
		panic(err)
	}

	return res
}

var configAuditFee = []byte("AuditFee")

// SetAuditFee sets the configuration value of the audit fee for the Inner Ring.
//
// See also [NetworkInfo.AuditFee].
func (x *NetworkInfo) SetAuditFee(fee uint64) {
	x.setConfigUint64(configAuditFee, fee)
}

// AuditFee returns audit fee set using [NetworkInfo.SetAuditFee].
//
// Zero NetworkInfo has zero audit fee.
func (x NetworkInfo) AuditFee() uint64 {
	return x.configUint64(configAuditFee)
}

var configStoragePrice = []byte("BasicIncomeRate")

// SetStoragePrice sets the price per gigabyte of data storage that data owners
// pay to storage nodes.
//
// See also [NetworkInfo.StoragePrice].
func (x *NetworkInfo) SetStoragePrice(price uint64) {
	x.setConfigUint64(configStoragePrice, price)
}

// StoragePrice returns storage price set using [NetworkInfo.SetStoragePrice].
//
// Zero NetworkInfo has zero storage price.
func (x NetworkInfo) StoragePrice() uint64 {
	return x.configUint64(configStoragePrice)
}

var configContainerFee = []byte("ContainerFee")

// SetContainerFee sets fee for the container creation that creator pays to
// each Alphabet node.
//
// See also [NetworkInfo.ContainerFee].
func (x *NetworkInfo) SetContainerFee(fee uint64) {
	x.setConfigUint64(configContainerFee, fee)
}

// ContainerFee returns container fee set using SetContainerFee.
//
// Zero NetworkInfo has zero container fee.
func (x NetworkInfo) ContainerFee() uint64 {
	return x.configUint64(configContainerFee)
}

var configNamedContainerFee = []byte("ContainerAliasFee")

// SetNamedContainerFee sets fee for creation of the named container creation
// that creator pays to each Alphabet node.
//
// See also [NetworkInfo.NamedContainerFee].
func (x *NetworkInfo) SetNamedContainerFee(fee uint64) {
	x.setConfigUint64(configNamedContainerFee, fee)
}

// NamedContainerFee returns container fee set using
// [NetworkInfo.SetNamedContainerFee].
//
// Zero NetworkInfo has zero container fee.
func (x NetworkInfo) NamedContainerFee() uint64 {
	return x.configUint64(configNamedContainerFee)
}

var configEigenTrustAlpha = []byte("EigenTrustAlpha")

// SetEigenTrustAlpha sets alpha parameter for EigenTrust algorithm used in
// reputation system of the storage nodes. Value MUST be in range [0, 1].
//
// See also [NetworkInfo.EigenTrustAlpha].
func (x *NetworkInfo) SetEigenTrustAlpha(alpha float64) {
	if alpha < 0 || alpha > 1 {
		panic(fmt.Sprintf("EigenTrust alpha parameter MUST be in range [0, 1], got %.2f", alpha))
	}

	x.setConfigUint64(configEigenTrustAlpha, math.Float64bits(alpha))
}

// EigenTrustAlpha returns EigenTrust parameter set using
// [NetworkInfo.SetEigenTrustAlpha].
//
// Zero NetworkInfo has zero alpha parameter.
func (x NetworkInfo) EigenTrustAlpha() float64 {
	alpha := math.Float64frombits(x.configUint64(configEigenTrustAlpha))
	if alpha < 0 || alpha > 1 {
		panic(fmt.Sprintf("unexpected invalid %s parameter value %.2f", configEigenTrustAlpha, alpha))
	}

	return alpha
}

var configEigenTrustNumberOfIterations = []byte("EigenTrustIterations")

// SetNumberOfEigenTrustIterations sets number of iterations of the EigenTrust
// algorithm to perform. The algorithm is used by the storage nodes for
// calculating the reputation values.
//
// See also [NetworkInfo.NumberOfEigenTrustIterations].
func (x *NetworkInfo) SetNumberOfEigenTrustIterations(num uint64) {
	x.setConfigUint64(configEigenTrustNumberOfIterations, num)
}

// NumberOfEigenTrustIterations returns number of EigenTrust iterations set
// using [NetworkInfo.SetNumberOfEigenTrustIterations].
//
// Zero NetworkInfo has zero iteration number.
func (x NetworkInfo) NumberOfEigenTrustIterations() uint64 {
	return x.configUint64(configEigenTrustNumberOfIterations)
}

var configEpochDuration = []byte("EpochDuration")

// SetEpochDuration sets NeoFS epoch duration measured in number of blocks of
// the NeoFS Sidechain.
//
// See also [NetworkInfo.EpochDuration].
func (x *NetworkInfo) SetEpochDuration(blocks uint64) {
	x.setConfigUint64(configEpochDuration, blocks)
}

// EpochDuration returns epoch duration set using
// [NetworkInfo.SetEpochDuration].
//
// Zero NetworkInfo has zero iteration number.
func (x NetworkInfo) EpochDuration() uint64 {
	return x.configUint64(configEpochDuration)
}

var configIRCandidateFee = []byte("InnerRingCandidateFee")

// SetIRCandidateFee sets fee for Inner Ring entrance paid by a new member.
//
// See also [NetworkInfo.IRCandidateFee].
func (x *NetworkInfo) SetIRCandidateFee(fee uint64) {
	x.setConfigUint64(configIRCandidateFee, fee)
}

// IRCandidateFee returns IR entrance fee set using
// [NetworkInfo.SetIRCandidateFee].
//
// Zero NetworkInfo has zero fee.
func (x NetworkInfo) IRCandidateFee() uint64 {
	return x.configUint64(configIRCandidateFee)
}

var configMaxObjSize = []byte("MaxObjectSize")

// SetMaxObjectSize sets maximum size of the object stored locally on the
// storage nodes (physical objects). Binary representation of any physically
// stored object MUST NOT overflow the limit.
//
// See also [NetworkInfo.MaxObjectSize].
func (x *NetworkInfo) SetMaxObjectSize(sz uint64) {
	x.setConfigUint64(configMaxObjSize, sz)
}

// MaxObjectSize returns maximum object size set using
// [NetworkInfo.SetMaxObjectSize].
//
// Zero NetworkInfo has zero maximum size.
func (x NetworkInfo) MaxObjectSize() uint64 {
	return x.configUint64(configMaxObjSize)
}

var configWithdrawalFee = []byte("WithdrawFee")

// SetWithdrawalFee sets fee for withdrawals from the NeoFS accounts that
// account owners pay to each Alphabet node.
//
// See also [NetworkInfo.WithdrawalFee].
func (x *NetworkInfo) SetWithdrawalFee(fee uint64) {
	x.setConfigUint64(configWithdrawalFee, fee)
}

// WithdrawalFee returns withdrawal fee set using
// [NetworkInfo.SetWithdrawalFee].
//
// Zero NetworkInfo has zero fee.
func (x NetworkInfo) WithdrawalFee() uint64 {
	return x.configUint64(configWithdrawalFee)
}

var configHomomorphicHashingDisabled = []byte("HomomorphicHashingDisabled")

// SetHomomorphicHashingDisabled sets flag indicating whether homomorphic
// hashing of the containers' objects in the network is disabled.
//
// See also [NetworkInfo.HomomorphicHashingDisabled].
func (x *NetworkInfo) SetHomomorphicHashingDisabled(v bool) {
	x.setConfigBool(configHomomorphicHashingDisabled, v)
}

// HomomorphicHashingDisabled returns flag indicating whether homomorphic
// hashing of the containers' objects in the network is disabled.
//
// Zero NetworkInfo has enabled homomorphic hashing.
//
// See also [NetworkInfo.SetHomomorphicHashingDisabled].
func (x NetworkInfo) HomomorphicHashingDisabled() bool {
	return x.configBool(configHomomorphicHashingDisabled)
}

var configMaintenanceModeAllowed = []byte("MaintenanceModeAllowed")

// SetMaintenanceModeAllowed sets flag indicating whether storage nodes are
// allowed to go into the maintenance mode.
//
// See also [NetworkInfo.MaintenanceModeAllowed].
func (x *NetworkInfo) SetMaintenanceModeAllowed(v bool) {
	x.setConfigBool(configMaintenanceModeAllowed, v)
}

// MaintenanceModeAllowed returns flag indicating whether storage nodes are
// allowed to go into the maintenance mode.
//
// Zero NetworkInfo has disallows maintenance mode.
//
// See also [NetworkInfo.SetMaintenanceModeAllowed].
func (x NetworkInfo) MaintenanceModeAllowed() bool {
	return x.configBool(configMaintenanceModeAllowed)
}
