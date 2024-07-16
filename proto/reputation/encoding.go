package reputation

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldPeerPubKey
)

// MarshaledSize returns size of the PeerID in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *PeerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldPeerPubKey, x.PublicKey)
	}
	return sz
}

// MarshalStable writes the PeerID in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [PeerID.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PeerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToBytes(b, fieldPeerPubKey, x.PublicKey)
	}
}

const (
	_ = iota
	fieldTrustPeer
	fieldTrustValue
)

// MarshaledSize returns size of the Trust in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Trust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldTrustPeer, x.Peer) +
			proto.SizeDouble(fieldTrustValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Trust in Protocol Buffers V3 format with ascending
// order of fields by number into b. MarshalStable uses exactly
// [Trust.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Trust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldTrustPeer, x.Peer)
		proto.MarshalToDouble(b[off:], fieldTrustValue, x.Value)
	}
}

const (
	_ = iota
	fieldP2PTrustPeer
	fieldP2PTrustValue
)

// MarshaledSize returns size of the PeerToPeerTrust in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PeerToPeerTrust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldP2PTrustPeer, x.TrustingPeer) +
			proto.SizeEmbedded(fieldP2PTrustValue, x.Trust)
	}
	return sz
}

// MarshalStable writes the PeerToPeerTrust in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PeerToPeerTrust.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PeerToPeerTrust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldP2PTrustPeer, x.TrustingPeer)
		proto.MarshalToEmbedded(b[off:], fieldP2PTrustValue, x.Trust)
	}
}

const (
	_ = iota
	fieldGlobalTrustBodyManager
	fieldGlobalTrustBodyValue
)

// MarshaledSize returns size of the GlobalTrust_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GlobalTrust_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGlobalTrustBodyManager, x.Manager) +
			proto.SizeEmbedded(fieldGlobalTrustBodyValue, x.Trust)
	}
	return sz
}

// MarshalStable writes the GlobalTrust_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GlobalTrust_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GlobalTrust_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGlobalTrustBodyManager, x.Manager)
		proto.MarshalToEmbedded(b[off:], fieldGlobalTrustBodyValue, x.Trust)
	}
}

const (
	_ = iota
	fieldAnnounceLocalReqEpoch
	fieldAnnounceLocalReqTrusts
)

// MarshaledSize returns size of the AnnounceLocalTrustRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceLocalTrustRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldAnnounceLocalReqEpoch, x.Epoch)
		for i := range x.Trusts {
			sz += proto.SizeEmbedded(fieldAnnounceLocalReqTrusts, x.Trusts[i])
		}
	}
	return sz
}

// MarshalStable writes the AnnounceLocalTrustRequest_Body in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceLocalTrustRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceLocalTrustRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, fieldAnnounceLocalReqEpoch, x.Epoch)
		for i := range x.Trusts {
			off += proto.MarshalToEmbedded(b[off:], fieldAnnounceLocalReqTrusts, x.Trusts[i])
		}
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

const (
	_ = iota
	fieldAnnounceIntermediateReqEpoch
	fieldAnnounceIntermediateReqIter
	fieldAnnounceIntermediateReqTrust
)

// MarshaledSize returns size of the AnnounceIntermediateResultRequest_Body in
// Protocol Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceIntermediateResultRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldAnnounceIntermediateReqEpoch, x.Epoch) +
			proto.SizeVarint(fieldAnnounceIntermediateReqIter, x.Iteration) +
			proto.SizeEmbedded(fieldAnnounceIntermediateReqTrust, x.Trust)
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
		off := proto.MarshalToVarint(b, fieldAnnounceIntermediateReqEpoch, x.Epoch)
		off += proto.MarshalToVarint(b[off:], fieldAnnounceIntermediateReqIter, x.Iteration)
		proto.MarshalToEmbedded(b[off:], fieldAnnounceIntermediateReqTrust, x.Trust)
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
