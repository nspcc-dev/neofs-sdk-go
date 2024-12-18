package prototest

import (
	"encoding/base64"
	"math/rand"

	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	"github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/proto/status"
)

// RandBytes random non-empty []byte up to 1024 len.
func RandBytes() []byte {
	ln := 1 + rand.Uint32()%1024
	b := make([]byte, ln)
	//nolint:staticcheck
	rand.Read(b)
	return b
}

// RandRepeatedBytes returns non-empty [][]byte up to 10 elements. Each element
// may be nil and empty.
func RandRepeatedBytes() [][]byte {
	vs := make([][]byte, 1+rand.Uint32()%10)
	for i := range vs {
		switch rand.Uint32() % 3 {
		case 0: // nil
		case 1:
			vs[i] = []byte{}
		case 2:
			vs[i] = RandBytes()
		}
	}
	return vs
}

// RandString random non-empty string up to 1024 len.
func RandString() string {
	s := base64.StdEncoding.EncodeToString(RandBytes())
	if len(s) > 1024 {
		return s[:1024]
	}
	return s
}

// RandStrings returns non-empty []string up to 10 elements. Each element may be
// empty.
func RandStrings() []string {
	vs := make([]string, 1+rand.Uint32()%10)
	for i := range vs {
		switch rand.Uint32() % 2 {
		case 0: // empty
		case 1:
			vs[i] = RandString()
		}
	}
	return vs
}

// RandInteger returns random non-zero integer.
func RandInteger[T proto.Varint]() T {
	for {
		if i := rand.Uint64(); i != 0 {
			return T(i)
		}
	}
}

// RandUint64 returns random positive uint64.
func RandUint64() uint64 { return RandInteger[uint64]() }

// RandUint32 returns random positive uint32.
func RandUint32() uint32 { return RandInteger[uint32]() }

// RandInt64 returns random non-zero int64.
func RandInt64() int64 { return RandInteger[int64]() }

// RandFloat64 returns random non-zero float64.
func RandFloat64() float64 {
	for {
		if f := rand.NormFloat64(); f != 0 {
			return f
		}
	}
}

// RandRepeated returns non-empty list of *T from 2 to 10 elements. First
// element is always nil, the second one is a pointer to zero.
func RandRepeated[T any](randFunc func() *T) []*T {
	vs := make([]*T, 2+rand.Uint32()%10)
	vs[0] = nil
	vs[1] = new(T)
	for i := range vs[2:] {
		vs[2+i] = randFunc()
	}
	return vs
}

// RandVersion returns random refs.Version with all non-zero fields.
func RandVersion() *refs.Version { return &refs.Version{Major: RandUint32(), Minor: RandUint32()} }

// RandOwnerID returns random refs.OwnerID with all non-zero fields.
func RandOwnerID() *refs.OwnerID { return &refs.OwnerID{Value: RandBytes()} }

// RandContainerID returns random refs.ContainerID with all non-zero fields.
func RandContainerID() *refs.ContainerID { return &refs.ContainerID{Value: RandBytes()} }

// RandContainerIDs returns non-empty list of refs.ContainerID up to 10
// elements. Each element may be nil and pointer to zero.
func RandContainerIDs() []*refs.ContainerID { return RandRepeated(RandContainerID) }

// RandObjectID returns random refs.ObjectID with all non-zero fields.
func RandObjectID() *refs.ObjectID { return &refs.ObjectID{Value: RandBytes()} }

// RandObjectIDs returns non-empty list of refs.ObjectID up to 10 elements. Each
// element may be nil and pointer to zero.
func RandObjectIDs() []*refs.ObjectID { return RandRepeated(RandObjectID) }

// RandObjectAddress returns random refs.Address with all non-zero fields.
func RandObjectAddress() *refs.Address {
	return &refs.Address{ContainerId: RandContainerID(), ObjectId: RandObjectID()}
}

// RandObjectAddresses returns non-empty list of refs.Address up to 10 elements.
// Each element may be nil and pointer to zero.
func RandObjectAddresses() []*refs.Address { return RandRepeated(RandObjectAddress) }

// RandChecksum returns random refs.Checksum with all non-zero fields.
func RandChecksum() *refs.Checksum {
	return &refs.Checksum{Type: RandInteger[refs.ChecksumType](), Sum: RandBytes()}
}

// RandSignature returns random refs.Signature with all non-zero fields.
func RandSignature() *refs.Signature {
	return &refs.Signature{
		Key:    RandBytes(),
		Sign:   RandBytes(),
		Scheme: RandInteger[refs.SignatureScheme](),
	}
}

// RandSignatureRFC6979 returns random refs.SignatureRFC6979 with all non-zero
// fields.
func RandSignatureRFC6979() *refs.SignatureRFC6979 {
	return &refs.SignatureRFC6979{Key: RandBytes(), Sign: RandBytes()}
}

// RandSubnetID returns random refs.SubnetID with all non-zero fields.
func RandSubnetID() *refs.SubnetID {
	return &refs.SubnetID{Value: RandUint32()}
}

// RandPlacementReplica returns random netmap.Replica with all non-zero fields.
func RandPlacementReplica() *netmap.Replica {
	return &netmap.Replica{Count: RandUint32(), Selector: RandString()}
}

// RandPlacementReplicas returns non-empty list of netmap.Replica up to 10
// elements. Each element may be nil and pointer to zero.
func RandPlacementReplicas() []*netmap.Replica { return RandRepeated(RandPlacementReplica) }

// RandPlacementSelector returns random netmap.Selector with all non-zero
// fields.
func RandPlacementSelector() *netmap.Selector {
	return &netmap.Selector{
		Name:      RandString(),
		Count:     RandUint32(),
		Clause:    RandInteger[netmap.Clause](),
		Attribute: RandString(),
		Filter:    RandString(),
	}
}

// RandPlacementSelectors returns non-empty list of netmap.Selector up to 10
// elements. Each element may be nil and pointer to zero.
func RandPlacementSelectors() []*netmap.Selector { return RandRepeated(RandPlacementSelector) }

func randPlacementFilter(withSubs bool) *netmap.Filter {
	v := &netmap.Filter{
		Name:    RandString(),
		Key:     "",
		Op:      0,
		Value:   "",
		Filters: nil,
	}
	if withSubs {
		v.Filters = randPlacementFilters(false)
	}
	return v
}

func randPlacementFilters(withSubs bool) []*netmap.Filter {
	return RandRepeated(func() *netmap.Filter { return randPlacementFilter(withSubs) })
}

// RandPlacementFilter returns random netmap.Filter with all non-zero fields.
func RandPlacementFilter() *netmap.Filter { return randPlacementFilter(true) }

// RandPlacementFilters returns non-empty list of netmap.Filter up to 10
// elements. Each element may be nil and pointer to zero.
func RandPlacementFilters() []*netmap.Filter { return randPlacementFilters(true) }

// RandPlacementPolicy returns random netmap.PlacementPolicy with all non-zero
// fields.
func RandPlacementPolicy() *netmap.PlacementPolicy {
	return &netmap.PlacementPolicy{
		Replicas:              RandPlacementReplicas(),
		ContainerBackupFactor: RandUint32(),
		Selectors:             RandPlacementSelectors(),
		Filters:               RandPlacementFilters(),
		SubnetId:              RandSubnetID(),
	}
}

// RandSessionTokenLifetime returns random
// session.SessionToken_Body_TokenLifetime with all non-zero fields.
func RandSessionTokenLifetime() *session.SessionToken_Body_TokenLifetime {
	return &session.SessionToken_Body_TokenLifetime{
		Exp: RandUint64(),
		Nbf: RandUint64(),
		Iat: RandUint64(),
	}
}

// RandObjectSessionTarget returns random session.ObjectSessionContext_Target
// with all non-zero fields.
func RandObjectSessionTarget() *session.ObjectSessionContext_Target {
	return &session.ObjectSessionContext_Target{
		Container: RandContainerID(),
		Objects:   RandObjectIDs(),
	}
}

// RandObjectSessionContext returns random session.ObjectSessionContext with all
// non-zero fields.
func RandObjectSessionContext() *session.ObjectSessionContext {
	return &session.ObjectSessionContext{
		Verb:   RandInteger[session.ObjectSessionContext_Verb](),
		Target: RandObjectSessionTarget(),
	}
}

// RandContainerSessionContext returns random session.ContainerSessionContext
// with all non-zero fields.
func RandContainerSessionContext() *session.ContainerSessionContext {
	return &session.ContainerSessionContext{
		Verb:        RandInteger[session.ContainerSessionContext_Verb](),
		Wildcard:    true,
		ContainerId: RandContainerID(),
	}
}

// RandSessionTokenBody returns random session.SessionToken_Body with all
// non-zero fields.
func RandSessionTokenBody() *session.SessionToken_Body {
	v := &session.SessionToken_Body{
		Id:         RandBytes(),
		OwnerId:    RandOwnerID(),
		Lifetime:   RandSessionTokenLifetime(),
		SessionKey: RandBytes(),
	}
	switch rand.Uint32() % 3 {
	case 0: // nil
	case 1:
		v.Context = &session.SessionToken_Body_Object{Object: RandObjectSessionContext()}
	case 2:
		v.Context = &session.SessionToken_Body_Container{Container: RandContainerSessionContext()}
	}
	return v
}

// RandSessionToken returns random session.SessionToken with all non-zero
// fields.
func RandSessionToken() *session.SessionToken {
	return &session.SessionToken{
		Body:      RandSessionTokenBody(),
		Signature: RandSignature(),
	}
}

// RandEACLFilter returns random acl.EACLRecord_Filter with all non-zero fields.
func RandEACLFilter() *acl.EACLRecord_Filter {
	return &acl.EACLRecord_Filter{
		HeaderType: RandInteger[acl.HeaderType](),
		MatchType:  RandInteger[acl.MatchType](),
		Key:        RandString(),
		Value:      RandString(),
	}
}

// RandEACLFilters returns non-empty list of acl.EACLRecord_Filter up to 10
// elements. Each element may be nil and pointer to zero.
func RandEACLFilters() []*acl.EACLRecord_Filter { return RandRepeated(RandEACLFilter) }

// RandEACLTarget returns random acl.EACLRecord_Target with all non-zero fields.
func RandEACLTarget() *acl.EACLRecord_Target {
	return &acl.EACLRecord_Target{
		Role: RandInteger[acl.Role](),
		Keys: RandRepeatedBytes(),
	}
}

// RandEACLTargets returns non-empty list of acl.EACLRecord_Target up to 10
// elements. Each element may be nil and pointer to zero.
func RandEACLTargets() []*acl.EACLRecord_Target { return RandRepeated(RandEACLTarget) }

// RandEACLRecord returns random acl.EACLRecord with all non-zero fields.
func RandEACLRecord() *acl.EACLRecord {
	return &acl.EACLRecord{
		Operation: RandInteger[acl.Operation](),
		Action:    RandInteger[acl.Action](),
		Filters:   RandEACLFilters(),
		Targets:   RandEACLTargets(),
	}
}

// RandEACLRecords returns non-empty list of acl.EACLRecord up to 10 elements.
// Each element may be nil and pointer to zero.
func RandEACLRecords() []*acl.EACLRecord { return RandRepeated(RandEACLRecord) }

// RandEACL returns random acl.EACLTable with all non-zero fields.
func RandEACL() *acl.EACLTable {
	return &acl.EACLTable{
		Version:     RandVersion(),
		ContainerId: RandContainerID(),
		Records:     RandEACLRecords(),
	}
}

// RandBearerTokenLifetime returns random acl.BearerToken_Body_TokenLifetime
// with all non-zero fields.
func RandBearerTokenLifetime() *acl.BearerToken_Body_TokenLifetime {
	return &acl.BearerToken_Body_TokenLifetime{
		Exp: RandUint64(),
		Nbf: RandUint64(),
		Iat: RandUint64(),
	}
}

// RandBearerTokenBody returns random acl.BearerToken_Body with all non-zero
// fields.
func RandBearerTokenBody() *acl.BearerToken_Body {
	return &acl.BearerToken_Body{
		EaclTable: RandEACL(),
		OwnerId:   RandOwnerID(),
		Lifetime:  RandBearerTokenLifetime(),
		Issuer:    RandOwnerID(),
	}
}

// RandBearerToken returns random acl.BearerToken with all non-zero fields.
func RandBearerToken() *acl.BearerToken {
	return &acl.BearerToken{
		Body:      RandBearerTokenBody(),
		Signature: RandSignature(),
	}
}

// RandStatusDetail returns random status.Status_Detail with all non-zero
// fields.
func RandStatusDetail() *status.Status_Detail {
	return &status.Status_Detail{
		Id:    RandUint32(),
		Value: RandBytes(),
	}
}

// RandStatusDetails returns non-empty list of status.Status_Detail up to 10
// elements. Each element may be nil and pointer to zero.
func RandStatusDetails() []*status.Status_Detail { return RandRepeated(RandStatusDetail) }

// RandStatus returns random status.Status with all non-zero fields.
func RandStatus() *status.Status {
	return &status.Status{
		Code:    RandUint32(),
		Message: RandString(),
		Details: RandStatusDetails(),
	}
}
