package reputation

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/reputation"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/proto"
)

// Trust represents quantitative assessment of the trust of a participant in the
// NeoFS reputation system.
//
// Trust is mutually compatible with [reputation.Trust] message. See
// [Trust.ReadFromV2] / [Trust.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type Trust struct {
	peerSet bool
	peer    PeerID

	val float64
}

// ReadFromV2 reads Trust from the reputation.Trust message. Returns an error if
// the message is malformed according to the NeoFS API V2 protocol. The message
// must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Trust.WriteToV2].
func (x *Trust) ReadFromV2(m *reputation.Trust) error {
	if m.Value < 0 || m.Value > 1 {
		return fmt.Errorf("invalid trust value %v", m.Value)
	}

	x.peerSet = m.Peer != nil
	if !x.peerSet {
		return errors.New("missing peer")
	}

	err := x.peer.ReadFromV2(m.Peer)
	if err != nil {
		return fmt.Errorf("invalid peer: %w", err)
	}

	x.val = m.Value

	return nil
}

// WriteToV2 writes Trust to the reputation.Trust message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Trust.ReadFromV2].
func (x Trust) WriteToV2(m *reputation.Trust) {
	if x.peerSet {
		m.Peer = new(reputation.PeerID)
		x.peer.WriteToV2(m.Peer)
	} else {
		m.Peer = nil
	}
	m.Value = x.val
}

// SetPeer specifies identifier of the participant of the NeoFS reputation
// system to which the Trust relates.
//
// See also [Trust.Peer].
func (x *Trust) SetPeer(id PeerID) {
	x.peer, x.peerSet = id, true
}

// Peer returns identifier of the participant of the NeoFS reputation system to
// which the Trust relates.
//
// Zero Trust returns zero PeerID which is incorrect according to the NeoFS API
// protocol.
//
// See also [Trust.SetPeer].
func (x Trust) Peer() PeerID {
	if x.peerSet {
		return x.peer
	}
	return PeerID{}
}

// SetValue sets the Trust value. Value MUST be in range [0;1].
//
// See also [Trust.Value].
func (x *Trust) SetValue(val float64) {
	if val < 0 || val > 1 {
		panic(fmt.Sprintf("trust value is out-of-range %v", val))
	}
	x.val = val
}

// Value returns the Trust value.
//
// Zero Trust has zero value.
//
// See also [Trust.SetValue].
func (x Trust) Value() float64 {
	return x.val
}

// PeerToPeerTrust represents trust of one participant of the NeoFS reputation
// system to another one.
//
// Trust is mutually compatible with [reputation.PeerToPeerTrust] message. See
// [PeerToPeerTrust.ReadFromV2] / [PeerToPeerTrust.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type PeerToPeerTrust struct {
	trustingPeerSet bool
	trustingPeer    PeerID

	valSet bool
	val    Trust
}

// ReadFromV2 reads PeerToPeerTrust from the reputation.PeerToPeerTrust message.
// Returns an error if the message is malformed according to the NeoFS API V2
// protocol. The message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PeerToPeerTrust.WriteToV2].
func (x *PeerToPeerTrust) ReadFromV2(m *reputation.PeerToPeerTrust) error {
	if x.trustingPeerSet = m.TrustingPeer != nil; x.trustingPeerSet {
		err := x.trustingPeer.ReadFromV2(m.TrustingPeer)
		if err != nil {
			return fmt.Errorf("invalid trusting peer: %w", err)
		}
	} else {
		return errors.New("missing trusting peer")
	}

	if x.valSet = m.Trust != nil; x.valSet {
		err := x.val.ReadFromV2(m.Trust)
		if err != nil {
			return fmt.Errorf("invalid trust: %w", err)
		}
	} else {
		return errors.New("missing trust")
	}

	return nil
}

// WriteToV2 writes PeerToPeerTrust to the reputation.PeerToPeerTrust message of
// the NeoFS API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [PeerToPeerTrust.ReadFromV2].
func (x PeerToPeerTrust) WriteToV2(m *reputation.PeerToPeerTrust) {
	if x.trustingPeerSet {
		m.TrustingPeer = new(reputation.PeerID)
		x.trustingPeer.WriteToV2(m.TrustingPeer)
	} else {
		m.TrustingPeer = nil
	}

	if x.valSet {
		m.Trust = new(reputation.Trust)
		x.val.WriteToV2(m.Trust)
	} else {
		m.Trust = nil
	}
}

// SetTrustingPeer specifies the peer from which trust comes in terms of the
// NeoFS reputation system.
//
// See also [PeerToPeerTrust.TrustingPeer].
func (x *PeerToPeerTrust) SetTrustingPeer(id PeerID) {
	x.trustingPeer, x.trustingPeerSet = id, true
}

// TrustingPeer returns the peer from which trust comes in terms of the NeoFS
// reputation system.
//
// Zero PeerToPeerTrust has no trusting peer which is incorrect according
// to the NeoFS API protocol.
//
// See also [PeerToPeerTrust.SetTrustingPeer].
func (x PeerToPeerTrust) TrustingPeer() PeerID {
	if x.trustingPeerSet {
		return x.trustingPeer
	}
	return PeerID{}
}

// SetTrust sets trust value of the trusting peer to another participant
// of the NeoFS reputation system.
//
// See also [PeerToPeerTrust.Trust].
func (x *PeerToPeerTrust) SetTrust(t Trust) {
	x.val, x.valSet = t, true
}

// Trust returns trust value of the trusting peer to another participant of the
// NeoFS reputation system.
//
// Zero PeerToPeerTrust returns zero Trust which is incorrect according to the
// NeoFS API protocol.
//
// See also [PeerToPeerTrust.SetTrust].
func (x PeerToPeerTrust) Trust() Trust {
	if x.valSet {
		return x.val
	}
	return Trust{}
}

// GlobalTrust represents the final assessment of trust in the participant of
// the NeoFS reputation system obtained taking into account all other participants.
//
// GlobalTrust is mutually compatible with [reputation.GlobalTrust] message. See
// [GlobalTrust.ReadFromV2] / [GlobalTrust.WriteToV2] methods.
//
// To submit GlobalTrust value in NeoFS zero instance should be initialized via
// [NewGlobalTrust] and filled using dedicated methods.
type GlobalTrust struct {
	versionSet bool
	version    version.Version

	managerSet bool
	manager    PeerID

	trustSet bool
	trust    Trust

	sigSet bool
	sig    neofscrypto.Signature
}

// NewGlobalTrust constructs new GlobalTrust instance.
func NewGlobalTrust(manager PeerID, trust Trust) GlobalTrust {
	return GlobalTrust{
		versionSet: true,
		version:    version.Current,
		managerSet: true,
		manager:    manager,
		trustSet:   true,
		trust:      trust,
	}
}

func (x *GlobalTrust) readFromV2(m *reputation.GlobalTrust) error {
	if x.versionSet = m.Version != nil; x.versionSet {
		err := x.version.ReadFromV2(m.Version)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else {
		return errors.New("missing version")
	}

	if x.sigSet = m.Signature != nil; x.sigSet {
		err := x.sig.ReadFromV2(m.Signature)
		if err != nil {
			return fmt.Errorf("invalid signature: %w", err)
		}
	}

	if m.Body == nil {
		return errors.New("missing body")
	}

	if x.managerSet = m.Body.Manager != nil; x.managerSet {
		err := x.manager.ReadFromV2(m.Body.Manager)
		if err != nil {
			return fmt.Errorf("invalid manager: %w", err)
		}
	} else {
		return errors.New("missing manager")
	}

	if x.trustSet = m.Body.Trust != nil; x.trustSet {
		err := x.trust.ReadFromV2(m.Body.Trust)
		if err != nil {
			return fmt.Errorf("invalid trust: %w", err)
		}
	} else {
		return errors.New("missing trust")
	}

	return nil
}

func (x GlobalTrust) writeToV2(m *reputation.GlobalTrust) {
	if x.versionSet {
		m.Version = new(refs.Version)
		x.version.WriteToV2(m.Version)
	} else {
		m.Version = nil
	}

	if x.sigSet {
		m.Signature = new(refs.Signature)
		x.sig.WriteToV2(m.Signature)
	} else {
		m.Signature = nil
	}

	m.Body = x.fillBody()
}

func (x GlobalTrust) fillBody() *reputation.GlobalTrust_Body {
	if !x.managerSet && !x.trustSet {
		return nil
	}

	var body reputation.GlobalTrust_Body
	if x.managerSet {
		body.Manager = new(reputation.PeerID)
		x.manager.WriteToV2(body.Manager)
	}
	if x.trustSet {
		body.Trust = new(reputation.Trust)
		x.trust.WriteToV2(body.Trust)
	}

	return &body
}

// SetManager sets identifier of the NeoFS reputation system's participant which
// performed trust estimation.
//
// See also [GlobalTrust.Manager].
func (x *GlobalTrust) SetManager(id PeerID) {
	x.manager, x.managerSet = id, true
}

// Manager returns identifier of the NeoFS reputation system's participant which
// performed trust estimation.
//
// Zero GlobalTrust has zero manager which is incorrect according to the
// NeoFS API protocol.
//
// See also [GlobalTrust.SetManager].
func (x GlobalTrust) Manager() PeerID {
	if x.managerSet {
		return x.manager
	}
	return PeerID{}
}

// SetTrust sets the global trust score of the network to a specific network
// member.
//
// See also [GlobalTrust.Trust].
func (x *GlobalTrust) SetTrust(trust Trust) {
	x.trust, x.trustSet = trust, true
}

// Trust returns the global trust score of the network to a specific network
// member.
//
// Zero GlobalTrust return zero Trust which is incorrect according to the
// NeoFS API protocol.
//
// See also [GlobalTrust.Trust].
func (x GlobalTrust) Trust() Trust {
	if x.trustSet {
		return x.trust
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
// See also [GlobalTrust.VerifySignature].
func (x *GlobalTrust) Sign(signer neofscrypto.Signer) error {
	err := x.sig.Calculate(signer, x.signedData())
	if err != nil {
		return err
	}
	x.sigSet = true
	return nil
}

func (x *GlobalTrust) signedData() []byte {
	body := x.fillBody()
	b := make([]byte, body.MarshaledSize())
	body.MarshalStable(b)
	return b
}

// VerifySignature checks if GlobalTrust signature is presented and valid.
//
// Zero GlobalTrust fails the check.
//
// See also [GlobalTrust.Sign].
func (x GlobalTrust) VerifySignature() bool {
	return x.sigSet && x.sig.Verify(x.signedData())
}

// Marshal encodes GlobalTrust into a Protocol Buffers V3 binary format.
//
// See also [GlobalTrust.Unmarshal].
func (x GlobalTrust) Marshal() []byte {
	var m reputation.GlobalTrust
	x.writeToV2(&m)

	b, err := proto.Marshal(&m)
	if err != nil {
		// while it is bad to panic on external package return, we can do nothing better
		// for this case: how can a normal message not be encoded?
		panic(fmt.Errorf("unexpected marshal protobuf message failure: %w", err))
	}
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the GlobalTrust.
// Returns an error if the message is malformed according to the NeoFS API V2
// protocol.
//
// See also [GlobalTrust.Marshal].
func (x *GlobalTrust) Unmarshal(data []byte) error {
	var m reputation.GlobalTrust
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return x.readFromV2(&m)
}
