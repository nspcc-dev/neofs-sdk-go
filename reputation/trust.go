package reputation

import (
	"errors"
	"fmt"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protoreputation "github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Trust represents quantitative assessment of the trust of a participant in the
// NeoFS reputation system.
//
// Trust is mutually compatible with [protoreputation.Trust] message. See
// [Trust.FromProtoMessage] / [Trust.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type Trust struct {
	peer PeerID
	val  float64
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Table.ProtoMessage].
func (x *Trust) FromProtoMessage(m *protoreputation.Trust) error {
	if m.Value < 0 || m.Value > 1 {
		return fmt.Errorf("invalid trust value %v", m.Value)
	}

	if m.Peer == nil {
		return errors.New("missing peer field")
	}

	err := x.peer.FromProtoMessage(m.Peer)
	if err != nil {
		return fmt.Errorf("invalid peer field: %w", err)
	}

	x.val = m.Value

	return nil
}

// ProtoMessage converts t into message to transmit using the NeoFS API
// protocol.
//
// See also [Table.FromProtoMessage].
func (x Trust) ProtoMessage() *protoreputation.Trust {
	m := &protoreputation.Trust{
		Value: x.val,
	}
	if x.peer.key != nil {
		m.Peer = x.peer.ProtoMessage()
	}
	return m
}

// SetPeer specifies identifier of the participant of the NeoFS reputation system
// to which the Trust relates.
//
// See also Peer.
func (x *Trust) SetPeer(id PeerID) {
	x.peer = id
}

// Peer returns peer identifier set using SetPeer.
//
// Zero Trust returns zero PeerID which is incorrect according to the NeoFS API
// protocol.
func (x Trust) Peer() PeerID {
	return x.peer
}

// SetValue sets the Trust value. Value MUST be in range [0;1].
//
// See also Value.
func (x *Trust) SetValue(val float64) {
	if val < 0 || val > 1 {
		panic(fmt.Sprintf("trust value is out-of-range %v", val))
	}
	x.val = val
}

// Value returns value set using SetValue.
//
// Zero Trust has zero value.
func (x Trust) Value() float64 {
	return x.val
}

// PeerToPeerTrust represents trust of one participant of the NeoFS reputation
// system to another one.
//
// Trust is mutually compatible [protoreputation.PeerToPeerTrust] message. See
// [PeerToPeerTrust.FromProtoMessage] / [PeerToPeerTrust.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type PeerToPeerTrust struct {
	peer  PeerID
	trust *Trust
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [PeerToPeerTrust.ProtoMessage].
func (x *PeerToPeerTrust) FromProtoMessage(m *protoreputation.PeerToPeerTrust) error {
	if m.TrustingPeer == nil {
		return errors.New("missing trusting peer")
	}

	err := x.peer.FromProtoMessage(m.TrustingPeer)
	if err != nil {
		return fmt.Errorf("invalid trusting peer: %w", err)
	}

	if m.Trust == nil {
		return errors.New("missing trust")
	}

	if x.trust == nil {
		x.trust = new(Trust)
	}
	err = x.trust.FromProtoMessage(m.Trust)
	if err != nil {
		return fmt.Errorf("invalid trust: %w", err)
	}

	return nil
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [PeerToPeerTrust.FromProtoMessage].
func (x PeerToPeerTrust) ProtoMessage() *protoreputation.PeerToPeerTrust {
	var m protoreputation.PeerToPeerTrust
	if x.peer.key != nil {
		m.TrustingPeer = x.peer.ProtoMessage()
	}
	if x.trust != nil {
		m.Trust = x.trust.ProtoMessage()
	}
	return &m
}

// SetTrustingPeer specifies the peer from which trust comes in terms of the
// NeoFS reputation system.
//
// See also TrustingPeer.
func (x *PeerToPeerTrust) SetTrustingPeer(id PeerID) {
	x.peer = id
}

// TrustingPeer returns peer set using SetTrustingPeer.
//
// Zero PeerToPeerTrust has no trusting peer which is incorrect according
// to the NeoFS API protocol.
func (x PeerToPeerTrust) TrustingPeer() PeerID {
	return x.peer
}

// SetTrust sets trust value of the trusting peer to another participant
// of the NeoFS reputation system.
//
// See also Trust.
func (x *PeerToPeerTrust) SetTrust(t Trust) {
	x.trust = &t
}

// Trust returns trust set using SetTrust.
//
// Zero PeerToPeerTrust returns zero Trust which is incorrect according to the
// NeoFS API protocol.
func (x PeerToPeerTrust) Trust() Trust {
	if x.trust != nil {
		return *x.trust
	}
	return Trust{}
}

// GlobalTrust represents the final assessment of trust in the participant of
// the NeoFS reputation system obtained taking into account all other participants.
//
// GlobalTrust is mutually compatible with [protoreputation.GlobalTrust]
// message. See [GlobalTrust.FromProtoMessage] / [GlobalTrust.ProtoMessage] methods.
//
// To submit GlobalTrust value in NeoFS zero instance SHOULD be declared,
// initialized using Init method and filled using dedicated methods.
type GlobalTrust struct {
	version *version.Version
	manager PeerID
	trust   *Trust
	sig     *neofscrypto.Signature
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [PeerID.ProtoMessage].
func (x *GlobalTrust) FromProtoMessage(m *protoreputation.GlobalTrust) error {
	if m.Version == nil {
		return errors.New("missing version")
	}

	if x.version == nil {
		x.version = new(version.Version)
	}
	if err := x.version.FromProtoMessage(m.Version); err != nil {
		return fmt.Errorf("invalid version")
	}

	if m.Signature == nil {
		return errors.New("missing signature")
	}

	if x.sig == nil {
		x.sig = new(neofscrypto.Signature)
	}
	if err := x.sig.FromProtoMessage(m.Signature); err != nil {
		return fmt.Errorf("invalid signature")
	}

	if m.Body == nil {
		return errors.New("missing body")
	}

	if m.Body.Manager == nil {
		return errors.New("missing manager")
	}

	err := x.manager.FromProtoMessage(m.Body.Manager)
	if err != nil {
		return fmt.Errorf("invalid manager: %w", err)
	}

	if m.Body.Trust == nil {
		return errors.New("missing trust")
	}

	if x.trust == nil {
		x.trust = new(Trust)
	}
	err = x.trust.FromProtoMessage(m.Body.Trust)
	if err != nil {
		return fmt.Errorf("invalid trust: %w", err)
	}

	return nil
}

func (x GlobalTrust) protoBodyMessage() *protoreputation.GlobalTrust_Body {
	if x.trust == nil && x.manager.key == nil {
		return nil
	}
	var m protoreputation.GlobalTrust_Body
	if x.trust != nil {
		m.Trust = x.trust.ProtoMessage()
	}
	if x.manager.key != nil {
		m.Manager = x.manager.ProtoMessage()
	}
	return &m
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [GlobalTrust.FromProtoMessage].
func (x GlobalTrust) ProtoMessage() *protoreputation.GlobalTrust {
	m := &protoreputation.GlobalTrust{
		Body: x.protoBodyMessage(),
	}
	if x.version != nil {
		m.Version = x.version.ProtoMessage()
	}
	if x.sig != nil {
		m.Signature = x.sig.ProtoMessage()
	}
	return m
}

// Init initializes all internal data of the GlobalTrust required by NeoFS API
// protocol. Init MUST be called when creating a new global trust instance.
// Init SHOULD NOT be called multiple times. Init SHOULD NOT be called if
// the GlobalTrust instance is used for decoding only.
func (x *GlobalTrust) Init() {
	ver := version.Current()
	x.version = &ver
}

// SetManager sets identifier of the NeoFS reputation system's participant which
// performed trust estimation.
//
// See also Manager.
func (x *GlobalTrust) SetManager(id PeerID) {
	x.manager = id
}

// Manager returns peer set using SetManager.
//
// Zero GlobalTrust has zero manager which is incorrect according to the
// NeoFS API protocol.
func (x GlobalTrust) Manager() PeerID {
	return x.manager
}

// SetTrust sets the global trust score of the network to a specific network
// member.
//
// See also Trust.
func (x *GlobalTrust) SetTrust(trust Trust) {
	x.trust = &trust
}

// Trust returns trust set using SetTrust.
//
// Zero GlobalTrust return zero Trust which is incorrect according to the
// NeoFS API protocol.
func (x GlobalTrust) Trust() Trust {
	if x.trust != nil {
		return *x.trust
	}
	return Trust{}
}

// Sign calculates and writes signature of the [GlobalTrust] data. Returns
// signature calculation errors.
//
// Zero [GlobalTrust] is unsigned.
//
// Note that any [GlobalTrust] mutation is likely to break the signature, so it is
// expected to be calculated as a final stage of [GlobalTrust] formation.
//
// See also [GlobalTrust.VerifySignature], [GlobalTrust.SignedData].
func (x *GlobalTrust) Sign(signer neofscrypto.Signer) error {
	var sig neofscrypto.Signature

	err := sig.Calculate(signer, x.SignedData())
	if err != nil {
		return fmt.Errorf("calculate signature: %w", err)
	}

	x.sig = &sig

	return nil
}

// SignedData returns actual payload to sign.
//
// See also [GlobalTrust.Sign].
func (x *GlobalTrust) SignedData() []byte {
	return neofsproto.MarshalMessage(x.protoBodyMessage())
}

// VerifySignature checks if GlobalTrust signature is presented and valid.
//
// Zero GlobalTrust fails the check.
//
// See also Sign.
func (x GlobalTrust) VerifySignature() bool {
	return x.sig != nil && x.sig.Verify(x.SignedData())
}

// Marshal encodes GlobalTrust into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x GlobalTrust) Marshal() []byte {
	return neofsproto.Marshal(x)
}

// Unmarshal decodes NeoFS API protocol binary format into the GlobalTrust
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *GlobalTrust) Unmarshal(data []byte) error {
	return neofsproto.Unmarshal(data, x)
}
