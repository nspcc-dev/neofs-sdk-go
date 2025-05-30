package netmap

import (
	"encoding/binary"
	"errors"
	"fmt"
	"iter"
	"math"

	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

// NetworkInfo groups information about the NeoFS network state. Mainly used to
// describe the current state of the network.
//
// NetworkInfo is mutually compatible with [protonetmap.NetworkInfo]
// message. See [NetworkInfo.FromProtoMessage] / [NetworkInfo.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type NetworkInfo struct {
	curEpoch   uint64
	netMagic   uint64
	msPerBlock int64
	prms       [][2][]byte
}

// reads NetworkInfo from netmap.NetworkInfo message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field. Verifies format of any
// presented field according to NeoFS API V2 protocol.
func (x *NetworkInfo) fromProtoMessage(m *protonetmap.NetworkInfo, checkFieldPresence bool) error {
	c := m.GetNetworkConfig()
	if checkFieldPresence && c == nil {
		return errors.New("missing network config")
	}

	if checkFieldPresence && len(c.Parameters) == 0 {
		return errors.New("missing network parameters")
	}

	if c == nil {
		x.curEpoch = m.CurrentEpoch
		x.netMagic = m.MagicNumber
		x.msPerBlock = m.MsPerBlock
		return nil
	}

	var err error
	mNames := make(map[string]struct{}, len(c.Parameters))
	prms := make([][2][]byte, len(c.Parameters))

	for i, prm := range c.Parameters {
		if prm == nil {
			return fmt.Errorf("nil parameter #%d", i)
		}
		name := string(prm.GetKey())

		_, was := mNames[name]
		if was {
			return fmt.Errorf("duplicated parameter name: %s", name)
		}

		mNames[name] = struct{}{}

		switch name {
		default:
			if len(prm.GetValue()) == 0 {
				return fmt.Errorf("empty %q parameter value", name)
			}
		case configEigenTrustAlpha:
			var num uint64

			num, err = decodeConfigValueUint64(prm.GetValue())
			if err == nil {
				if alpha := math.Float64frombits(num); alpha < 0 || alpha > 1 {
					err = fmt.Errorf("EigenTrust alpha value %0.2f is out of range [0, 1]", alpha)
				}
			}
		case
			configAuditFee,
			configStoragePrice,
			configContainerFee,
			configNamedContainerFee,
			configEigenTrustNumberOfIterations,
			configEpochDuration,
			configIRCandidateFee,
			configMaxObjSize,
			configWithdrawalFee:
			_, err = decodeConfigValueUint64(prm.GetValue())
		case configHomomorphicHashingDisabled,
			configMaintenanceModeAllowed:
			_, err = decodeConfigValueBool(prm.GetValue())
		}

		if err != nil {
			return fmt.Errorf("invalid %s parameter: %w", name, err)
		}

		prms[i] = [2][]byte{prm.Key, prm.Value}
	}

	x.curEpoch = m.CurrentEpoch
	x.netMagic = m.MagicNumber
	x.msPerBlock = m.MsPerBlock
	x.prms = prms

	return nil
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [NetworkInfo.ProtoMessage].
func (x *NetworkInfo) FromProtoMessage(m *protonetmap.NetworkInfo) error {
	return x.fromProtoMessage(m, true)
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [NetworkInfo.FromProtoMessage].
func (x NetworkInfo) ProtoMessage() *protonetmap.NetworkInfo {
	m := &protonetmap.NetworkInfo{
		CurrentEpoch: x.curEpoch,
		MagicNumber:  x.netMagic,
		MsPerBlock:   x.msPerBlock,
	}
	if len(x.prms) > 0 {
		m.NetworkConfig = &protonetmap.NetworkConfig{
			Parameters: make([]*protonetmap.NetworkConfig_Parameter, len(x.prms)),
		}
		for i := range x.prms {
			m.NetworkConfig.Parameters[i] = &protonetmap.NetworkConfig_Parameter{Key: x.prms[i][0], Value: x.prms[i][1]}
		}
	}
	return m
}

// Marshal encodes NetworkInfo into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x NetworkInfo) Marshal() []byte {
	return neofsproto.Marshal(x)
}

// Unmarshal decodes NeoFS API protocol binary format into the NetworkInfo
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *NetworkInfo) Unmarshal(data []byte) error {
	return neofsproto.UnmarshalOptional(data, x, (*NetworkInfo).fromProtoMessage)
}

// CurrentEpoch returns epoch set using SetCurrentEpoch.
//
// Zero NetworkInfo has zero current epoch.
func (x NetworkInfo) CurrentEpoch() uint64 {
	return x.curEpoch
}

// SetCurrentEpoch sets current epoch of the NeoFS network.
func (x *NetworkInfo) SetCurrentEpoch(epoch uint64) {
	x.curEpoch = epoch
}

// MagicNumber returns magic number set using SetMagicNumber.
//
// Zero NetworkInfo has zero magic.
func (x NetworkInfo) MagicNumber() uint64 {
	return x.netMagic
}

// SetMagicNumber sets magic number of the NeoFS Sidechain.
//
// See also MagicNumber.
func (x *NetworkInfo) SetMagicNumber(epoch uint64) {
	x.netMagic = epoch
}

// MsPerBlock returns network parameter set using SetMsPerBlock.
func (x NetworkInfo) MsPerBlock() int64 {
	return x.msPerBlock
}

// SetMsPerBlock sets MillisecondsPerBlock network parameter of the NeoFS Sidechain.
//
// See also MsPerBlock.
func (x *NetworkInfo) SetMsPerBlock(v int64) {
	x.msPerBlock = v
}

func (x *NetworkInfo) setConfig(name string, val []byte) {
	for i := range x.prms {
		if string(x.prms[i][0]) == name {
			x.prms[i][1] = val
			return
		}
	}

	x.prms = append(x.prms, [2][]byte{[]byte(name), val})
}

func (x NetworkInfo) configValue(name string) []byte {
	for i := range x.prms {
		if string(x.prms[i][0]) == name {
			return x.prms[i][1]
		}
	}
	return nil
}

// SetRawNetworkParameter sets named NeoFS network parameter whose value is
// transmitted but not interpreted by the NeoFS API protocol.
//
// Argument MUST NOT be mutated, make a copy first.
//
// See also RawNetworkParameter.
func (x *NetworkInfo) SetRawNetworkParameter(name string, value []byte) {
	x.setConfig(name, value)
}

// RawNetworkParameter reads raw network parameter set using [NetworkInfo.SetRawNetworkParameter]
// by its name. Returns nil if value is missing.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
//
// Zero NetworkInfo has no network parameters.
func (x *NetworkInfo) RawNetworkParameter(name string) []byte {
	return x.configValue(name)
}

// RawNetworkParameters returns an iterator that yields the non-system network
// parameters.
func (x NetworkInfo) RawNetworkParameters() iter.Seq2[string, []byte] {
	return func(yield func(string, []byte) bool) {
		for i := range x.prms {
			name := string(x.prms[i][0])
			switch name {
			default:
				if !yield(name, x.prms[i][1]) {
					return
				}
			case
				configEigenTrustAlpha,
				configAuditFee,
				configStoragePrice,
				configContainerFee,
				configNamedContainerFee,
				configEigenTrustNumberOfIterations,
				configEpochDuration,
				configIRCandidateFee,
				configMaxObjSize,
				configWithdrawalFee,
				configHomomorphicHashingDisabled,
				configMaintenanceModeAllowed:
			}
		}
	}
}

// IterateRawNetworkParameters iterates over all raw networks parameters set
// using SetRawNetworkParameter and passes them into f.
//
// Handler MUST NOT be nil. Handler MUST NOT mutate value parameter.
//
// Zero NetworkInfo has no network parameters.
// Deprecated: use [NetworkInfo.RawNetworkParameters] instead.
func (x *NetworkInfo) IterateRawNetworkParameters(f func(name string, value []byte)) {
	for i := range x.prms {
		name := string(x.prms[i][0])
		switch name {
		default:
			f(name, x.prms[i][1])
		case
			configEigenTrustAlpha,
			configAuditFee,
			configStoragePrice,
			configContainerFee,
			configNamedContainerFee,
			configEigenTrustNumberOfIterations,
			configEpochDuration,
			configIRCandidateFee,
			configMaxObjSize,
			configWithdrawalFee,
			configHomomorphicHashingDisabled,
			configMaintenanceModeAllowed:
		}
	}
}

func (x *NetworkInfo) setConfigUint64(name string, num uint64) {
	val := make([]byte, 8)
	binary.LittleEndian.PutUint64(val, num)

	x.setConfig(name, val)
}

func (x *NetworkInfo) setConfigBool(name string, val bool) {
	v := stackitem.NewBool(val)
	x.setConfig(name, v.Bytes())
}

// decodeConfigValueUint64 parses val as little-endian uint64.
// val must be less than 8 bytes in size.
func decodeConfigValueUint64(val []byte) (uint64, error) {
	if ln := len(val); ln > 8 {
		return 0, fmt.Errorf("invalid uint64 parameter length %d", ln)
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
		return false, fmt.Errorf("invalid bool parameter contract format %w", err)
	}

	return res, nil
}

func (x NetworkInfo) configUint64(name string) uint64 {
	val := x.configValue(name)
	if val == nil {
		return 0
	}

	res, err := decodeConfigValueUint64(val)
	if err != nil {
		// potential panic is OK since value MUST be correct since it is
		// verified in FromProtoMessage or set by provided method.
		panic(err)
	}

	return res
}

func (x NetworkInfo) configBool(name string) bool {
	val := x.configValue(name)
	if val == nil {
		return false
	}

	res, err := decodeConfigValueBool(val)
	if err != nil {
		// potential panic is OK since value MUST be correct since it is
		// verified in FromProtoMessage or set by provided method.
		panic(err)
	}

	return res
}

const configAuditFee = "AuditFee"

// SetAuditFee sets the configuration value of the audit fee for the Inner Ring.
//
// See also AuditFee.
func (x *NetworkInfo) SetAuditFee(fee uint64) {
	x.setConfigUint64(configAuditFee, fee)
}

// AuditFee returns audit fee set using SetAuditFee.
//
// Zero NetworkInfo has zero audit fee.
func (x NetworkInfo) AuditFee() uint64 {
	return x.configUint64(configAuditFee)
}

const configStoragePrice = "BasicIncomeRate"

// SetStoragePrice sets the price per gigabyte of data storage that data owners
// pay to storage nodes.
//
// See also StoragePrice.
func (x *NetworkInfo) SetStoragePrice(price uint64) {
	x.setConfigUint64(configStoragePrice, price)
}

// StoragePrice returns storage price set using SetStoragePrice.
//
// Zero NetworkInfo has zero storage price.
func (x NetworkInfo) StoragePrice() uint64 {
	return x.configUint64(configStoragePrice)
}

const configContainerFee = "ContainerFee"

// SetContainerFee sets fee for the container creation that creator pays to
// each Alphabet node.
//
// See also ContainerFee.
func (x *NetworkInfo) SetContainerFee(fee uint64) {
	x.setConfigUint64(configContainerFee, fee)
}

// ContainerFee returns container fee set using SetContainerFee.
//
// Zero NetworkInfo has zero container fee.
func (x NetworkInfo) ContainerFee() uint64 {
	return x.configUint64(configContainerFee)
}

const configNamedContainerFee = "ContainerAliasFee"

// SetNamedContainerFee sets fee for creation of the named container creation
// that creator pays to each Alphabet node.
//
// See also NamedContainerFee.
func (x *NetworkInfo) SetNamedContainerFee(fee uint64) {
	x.setConfigUint64(configNamedContainerFee, fee)
}

// NamedContainerFee returns container fee set using SetNamedContainerFee.
//
// Zero NetworkInfo has zero container fee.
func (x NetworkInfo) NamedContainerFee() uint64 {
	return x.configUint64(configNamedContainerFee)
}

const configEigenTrustAlpha = "EigenTrustAlpha"

// SetEigenTrustAlpha sets alpha parameter for EigenTrust algorithm used in
// reputation system of the storage nodes. Value MUST be in range [0, 1].
//
// See also EigenTrustAlpha.
func (x *NetworkInfo) SetEigenTrustAlpha(alpha float64) {
	if alpha < 0 || alpha > 1 {
		panic(fmt.Sprintf("EigenTrust alpha parameter MUST be in range [0, 1], got %.2f", alpha))
	}

	x.setConfigUint64(configEigenTrustAlpha, math.Float64bits(alpha))
}

// EigenTrustAlpha returns EigenTrust parameter set using SetEigenTrustAlpha.
//
// Zero NetworkInfo has zero alpha parameter.
func (x NetworkInfo) EigenTrustAlpha() float64 {
	alpha := math.Float64frombits(x.configUint64(configEigenTrustAlpha))
	if alpha < 0 || alpha > 1 {
		panic(fmt.Sprintf("unexpected invalid %s parameter value %.2f", configEigenTrustAlpha, alpha))
	}

	return alpha
}

const configEigenTrustNumberOfIterations = "EigenTrustIterations"

// SetNumberOfEigenTrustIterations sets number of iterations of the EigenTrust
// algorithm to perform. The algorithm is used by the storage nodes for
// calculating the reputation values.
//
// See also NumberOfEigenTrustIterations.
func (x *NetworkInfo) SetNumberOfEigenTrustIterations(num uint64) {
	x.setConfigUint64(configEigenTrustNumberOfIterations, num)
}

// NumberOfEigenTrustIterations returns number of EigenTrust iterations set
// using SetNumberOfEigenTrustIterations.
//
// Zero NetworkInfo has zero iteration number.
func (x NetworkInfo) NumberOfEigenTrustIterations() uint64 {
	return x.configUint64(configEigenTrustNumberOfIterations)
}

const configEpochDuration = "EpochDuration"

// SetEpochDuration sets NeoFS epoch duration measured in number of blocks of
// the NeoFS Sidechain.
//
// See also EpochDuration.
func (x *NetworkInfo) SetEpochDuration(blocks uint64) {
	x.setConfigUint64(configEpochDuration, blocks)
}

// EpochDuration returns epoch duration set using SetEpochDuration.
//
// Zero NetworkInfo has zero iteration number.
func (x NetworkInfo) EpochDuration() uint64 {
	return x.configUint64(configEpochDuration)
}

const configIRCandidateFee = "InnerRingCandidateFee"

// SetIRCandidateFee sets fee for Inner Ring entrance paid by a new member.
//
// See also IRCandidateFee.
func (x *NetworkInfo) SetIRCandidateFee(fee uint64) {
	x.setConfigUint64(configIRCandidateFee, fee)
}

// IRCandidateFee returns IR entrance fee set using SetIRCandidateFee.
//
// Zero NetworkInfo has zero fee.
func (x NetworkInfo) IRCandidateFee() uint64 {
	return x.configUint64(configIRCandidateFee)
}

const configMaxObjSize = "MaxObjectSize"

// SetMaxObjectSize sets maximum size of the object stored locally on the
// storage nodes (physical objects). Binary representation of any physically
// stored object MUST NOT overflow the limit.
//
// See also MaxObjectSize.
func (x *NetworkInfo) SetMaxObjectSize(sz uint64) {
	x.setConfigUint64(configMaxObjSize, sz)
}

// MaxObjectSize returns maximum object size set using SetMaxObjectSize.
//
// Zero NetworkInfo has zero maximum size.
func (x NetworkInfo) MaxObjectSize() uint64 {
	return x.configUint64(configMaxObjSize)
}

const configWithdrawalFee = "WithdrawFee"

// SetWithdrawalFee sets fee for withdrawals from the NeoFS accounts that
// account owners pay to each Alphabet node.
//
// See also WithdrawalFee.
func (x *NetworkInfo) SetWithdrawalFee(sz uint64) {
	x.setConfigUint64(configWithdrawalFee, sz)
}

// WithdrawalFee returns withdrawal fee set using SetWithdrawalFee.
//
// Zero NetworkInfo has zero fee.
func (x NetworkInfo) WithdrawalFee() uint64 {
	return x.configUint64(configWithdrawalFee)
}

const configHomomorphicHashingDisabled = "HomomorphicHashingDisabled"

// DisableHomomorphicHashing sets flag requiring to disable homomorphic
// hashing of the containers in the network.
//
// See also HomomorphicHashingDisabled.
func (x *NetworkInfo) DisableHomomorphicHashing() {
	x.setConfigBool(configHomomorphicHashingDisabled, true)
}

// HomomorphicHashingDisabled returns the state of the homomorphic
// hashing network setting.
//
// Zero NetworkInfo has enabled homomorphic hashing.
func (x NetworkInfo) HomomorphicHashingDisabled() bool {
	return x.configBool(configHomomorphicHashingDisabled)
}

const configMaintenanceModeAllowed = "MaintenanceModeAllowed"

// AllowMaintenanceMode sets the flag allowing nodes to go into maintenance mode.
//
// See also MaintenanceModeAllowed.
func (x *NetworkInfo) AllowMaintenanceMode() {
	x.setConfigBool(configMaintenanceModeAllowed, true)
}

// MaintenanceModeAllowed returns true iff network config allows
// maintenance mode for storage nodes.
//
// Zero NetworkInfo has disallows maintenance mode.
func (x NetworkInfo) MaintenanceModeAllowed() bool {
	return x.configBool(configMaintenanceModeAllowed)
}
