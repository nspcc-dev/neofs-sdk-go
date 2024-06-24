package reputation

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldPeerPubKey
)

func (x *PeerID) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldPeerPubKey, x.PublicKey)
	}
	return sz
}

func (x *PeerID) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalBytes(b, fieldPeerPubKey, x.PublicKey)
	}
}

const (
	_ = iota
	fieldTrustPeer
	fieldTrustValue
)

func (x *Trust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldTrustPeer, x.Peer) +
			proto.SizeFloat64(fieldTrustValue, x.Value)
	}
	return sz
}

func (x *Trust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldTrustPeer, x.Peer)
		proto.MarshalFloat64(b[off:], fieldTrustValue, x.Value)
	}
}

const (
	_ = iota
	fieldP2PTrustPeer
	fieldP2PTrustValue
)

func (x *PeerToPeerTrust) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldP2PTrustPeer, x.TrustingPeer) +
			proto.SizeNested(fieldP2PTrustValue, x.Trust)
	}
	return sz
}

func (x *PeerToPeerTrust) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldP2PTrustPeer, x.TrustingPeer)
		proto.MarshalNested(b[off:], fieldP2PTrustValue, x.Trust)
	}
}

const (
	_ = iota
	fieldGlobalTrustBodyManager
	fieldGlobalTrustBodyValue
)

func (x *GlobalTrust_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGlobalTrustBodyManager, x.Manager) +
			proto.SizeNested(fieldGlobalTrustBodyValue, x.Trust)
	}
	return sz
}

func (x *GlobalTrust_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldGlobalTrustBodyManager, x.Manager)
		proto.MarshalNested(b[off:], fieldGlobalTrustBodyValue, x.Trust)
	}
}

const (
	_ = iota
	fieldAnnounceLocalReqEpoch
	fieldAnnounceLocalReqTrusts
)

func (x *AnnounceLocalTrustRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldAnnounceLocalReqEpoch, x.Epoch)
		for i := range x.Trusts {
			sz += proto.SizeNested(fieldAnnounceLocalReqTrusts, x.Trusts[i])
		}
	}
	return sz
}

func (x *AnnounceLocalTrustRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldAnnounceLocalReqEpoch, x.Epoch)
		for i := range x.Trusts {
			off += proto.MarshalNested(b[off:], fieldAnnounceLocalReqTrusts, x.Trusts[i])
		}
	}
}

func (x *AnnounceLocalTrustResponse_Body) MarshaledSize() int   { return 0 }
func (x *AnnounceLocalTrustResponse_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldAnnounceIntermediateReqEpoch
	fieldAnnounceIntermediateReqIter
	fieldAnnounceIntermediateReqTrust
)

func (x *AnnounceIntermediateResultRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldAnnounceIntermediateReqEpoch, x.Epoch) +
			proto.SizeVarint(fieldAnnounceIntermediateReqIter, x.Iteration) +
			proto.SizeNested(fieldAnnounceIntermediateReqTrust, x.Trust)
	}
	return sz
}

func (x *AnnounceIntermediateResultRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldAnnounceIntermediateReqEpoch, x.Epoch)
		off += proto.MarshalVarint(b[off:], fieldAnnounceIntermediateReqIter, x.Iteration)
		proto.MarshalNested(b[off:], fieldAnnounceIntermediateReqIter, x.Trust)
	}
}

func (x *AnnounceIntermediateResultResponse_Body) MarshaledSize() int   { return 0 }
func (x *AnnounceIntermediateResultResponse_Body) MarshalStable([]byte) {}
