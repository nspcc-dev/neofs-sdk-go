package reputation

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

// Field numbers of [PeerID] message.
const (
	_ = iota
	FieldPeerIDPublicKey
)

// MarshaledSize returns size of the PeerID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *PeerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldPeerIDPublicKey, x.PublicKey)
	}
	return sz
}

// MarshalStable writes the PeerID in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [PeerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PeerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, FieldPeerIDPublicKey, x.PublicKey)
	}
}

// Field numbers of [Trust] message.
const (
	_ = iota
	FieldTrustPeer
	FieldTrustValue
)

// MarshaledSize returns size of the Trust in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Trust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldTrustPeer, x.Peer) +
			proto.SizeDouble(FieldTrustValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Trust in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Trust.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Trust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldTrustPeer, x.Peer)
		proto.MarshalToDouble(b[off:], FieldTrustValue, x.Value)
	}
}

// Field numbers of [PeerToPeerTrust] message.
const (
	_ = iota
	FieldPeerToPeerTrustTrustingPeer
	FieldPeerToPeerTrustTrust
)

// MarshaledSize returns size of the PeerToPeerTrust in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PeerToPeerTrust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldPeerToPeerTrustTrustingPeer, x.TrustingPeer) +
			proto.SizeEmbedded(FieldPeerToPeerTrustTrust, x.Trust)
	}
	return sz
}

// MarshalStable writes the PeerToPeerTrust in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PeerToPeerTrust.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PeerToPeerTrust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldPeerToPeerTrustTrustingPeer, x.TrustingPeer)
		proto.MarshalToEmbedded(b[off:], FieldPeerToPeerTrustTrust, x.Trust)
	}
}

// Field numbers of [GlobalTrust_Body] message.
const (
	_ = iota
	FieldGlobalTrustBodyManager
	FieldGlobalTrustBodyTrust
)

// MarshaledSize returns size of the GlobalTrust_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GlobalTrust_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGlobalTrustBodyManager, x.Manager) +
			proto.SizeEmbedded(FieldGlobalTrustBodyTrust, x.Trust)
	}
	return sz
}

// MarshalStable writes the GlobalTrust_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GlobalTrust_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GlobalTrust_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGlobalTrustBodyManager, x.Manager)
		proto.MarshalToEmbedded(b[off:], FieldGlobalTrustBodyTrust, x.Trust)
	}
}

// Field numbers of [GlobalTrust] message.
const (
	_ = iota
	FieldGlobalTrustVersion
	FieldGlobalTrustBody
	FieldGlobalTrustSignature
)

// MarshaledSize returns size of the GlobalTrust in Protocol Buffers V3 format
// in bytes. MarshaledSize is NPE-safe.
func (x *GlobalTrust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGlobalTrustVersion, x.Version) +
			proto.SizeEmbedded(FieldGlobalTrustBody, x.Body) +
			proto.SizeEmbedded(FieldGlobalTrustSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the GlobalTrust in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GlobalTrust.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GlobalTrust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGlobalTrustVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], FieldGlobalTrustBody, x.Body)
		proto.MarshalToEmbedded(b[off:], FieldGlobalTrustSignature, x.Signature)
	}
}

// Field numbers of [AnnounceLocalTrustRequest_Body] message.
const (
	_ = iota
	FieldAnnounceLocalTrustRequestBodyEpoch
	FieldAnnounceLocalTrustRequestBodyTrusts
)

// MarshaledSize returns size of the AnnounceLocalTrustRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceLocalTrustRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeVarint(FieldAnnounceLocalTrustRequestBodyEpoch, x.Epoch) +
			proto.SizeRepeatedMessages(FieldAnnounceLocalTrustRequestBodyTrusts, x.Trusts)
	}
	return 0
}

// MarshalStable writes the AnnounceLocalTrustRequest_Body in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceLocalTrustRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceLocalTrustRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldAnnounceLocalTrustRequestBodyEpoch, x.Epoch)
		proto.MarshalToRepeatedMessages(b[off:], FieldAnnounceLocalTrustRequestBodyTrusts, x.Trusts)
	}
}

// MarshaledSize returns size of the AnnounceLocalTrustResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceLocalTrustResponse_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the AnnounceLocalTrustResponse_Body in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceLocalTrustResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceLocalTrustResponse_Body) MarshalStable([]byte) {}

// Field numbers of [AnnounceIntermediateResultRequest_Body] message.
const (
	_ = iota
	FieldAnnounceIntermediateResultRequestBodyEpoch
	FieldAnnounceIntermediateResultRequestBodyIteration
	FieldAnnounceIntermediateResultRequestBodyTrust
)

// MarshaledSize returns size of the AnnounceIntermediateResultRequest_Body in
// Protocol Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceIntermediateResultRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldAnnounceIntermediateResultRequestBodyEpoch, x.Epoch) +
			proto.SizeVarint(FieldAnnounceIntermediateResultRequestBodyIteration, x.Iteration) +
			proto.SizeEmbedded(FieldAnnounceIntermediateResultRequestBodyTrust, x.Trust)
	}
	return sz
}

// MarshalStable writes the AnnounceIntermediateResultRequest_Body in Protocol
// Buffers V3 format with ascending order of fields by number into b.
// MarshalStable uses exactly
// [AnnounceIntermediateResultRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceIntermediateResultRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldAnnounceIntermediateResultRequestBodyEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], FieldAnnounceIntermediateResultRequestBodyIteration, x.Iteration)
		proto.MarshalToEmbedded(b[off:], FieldAnnounceIntermediateResultRequestBodyTrust, x.Trust)
	}
}

// MarshaledSize returns size of the AnnounceIntermediateResultResponse_Body in
// Protocol Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceIntermediateResultResponse_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the AnnounceIntermediateResultResponse_Body in Protocol
// Buffers V3 format with ascending order of fields by number into b.
// MarshalStable uses exactly
// [AnnounceIntermediateResultResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceIntermediateResultResponse_Body) MarshalStable([]byte) {}
