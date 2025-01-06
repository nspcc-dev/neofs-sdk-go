package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/eacl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoaccounting "github.com/nspcc-dev/neofs-sdk-go/proto/accounting"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	protorefs "github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protoreputation "github.com/nspcc-dev/neofs-sdk-go/proto/reputation"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/proto"
)

/*
Various NeoFS protocol messages. Any message (incl. each group element) must be
cloned via [proto.Clone] for the network transmission.
*/

// Cross-service.
var (
	// set of correct object IDs.
	validProtoObjectIDs = []*protorefs.ObjectID{
		{Value: []byte{218, 203, 9, 142, 129, 249, 13, 159, 198, 60, 153, 148, 70, 216, 50, 17, 15, 87, 47, 104, 143, 0, 187, 211, 120, 105, 250, 170, 220, 36, 108, 171}},
		{Value: []byte{28, 74, 243, 168, 65, 185, 194, 228, 239, 47, 76, 99, 131, 154, 18, 4, 91, 243, 28, 47, 183, 252, 203, 17, 32, 194, 193, 55, 213, 43, 15, 157}},
		{Value: []byte{64, 228, 234, 193, 115, 188, 136, 160, 127, 238, 221, 164, 4, 75, 158, 61, 82, 183, 241, 130, 189, 122, 192, 191, 244, 181, 98, 91, 179, 36, 197, 47}},
	}
	// correct object address with all fields.
	validFullProtoObjectAddress = &protorefs.Address{
		ContainerId: proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID),
		ObjectId:    proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
	}
	// correct signature with required fields only.
	validMinProtoSignature = &protorefs.Signature{}
	// correct signature with all fields.
	validFullProtoSignature = &protorefs.Signature{
		Key:    []byte("any_key"),
		Sign:   []byte("any_signature"),
		Scheme: protorefs.SignatureScheme(rand.Int31()),
	}
)

// Accounting service.
var (
	// correct balance with required fields only.
	validMinProtoBalance = &protoaccounting.Decimal{}
	// correct balance with all fields.
	validFullProtoBalance = &protoaccounting.Decimal{Value: 1609926665709559552, Precision: 2322521745}
	// correct AccountingService.Balance response payload with required fields only.
	validMinBalanceResponseBody = &protoaccounting.BalanceResponse_Body{
		Balance: proto.Clone(validMinProtoBalance).(*protoaccounting.Decimal),
	}
	// correct AccountingService.Balance response payload with all fields.
	validFullBalanceResponseBody = &protoaccounting.BalanceResponse_Body{
		Balance: proto.Clone(validFullProtoBalance).(*protoaccounting.Decimal),
	}
)

// Container service.
var (
	// correct container with required fields only.
	validMinProtoContainer = &protocontainer.Container{
		Version: proto.Clone(validMinProtoVersion).(*protorefs.Version),
		OwnerId: &protorefs.OwnerID{Value: []byte{53, 233, 31, 174, 37, 64, 241, 22, 182, 130, 7, 210, 222, 150, 85, 18, 106, 4,
			253, 122, 191, 90, 168, 187, 245}},
		Nonce: []byte{207, 5, 57, 28, 224, 103, 76, 207, 133, 186, 108, 96, 185, 52, 37, 205},
		PlacementPolicy: &protonetmap.PlacementPolicy{
			Replicas: make([]*protonetmap.Replica, 1),
		},
	}
	// correct container with all fields.
	validFullProtoContainer = &protocontainer.Container{
		Version:  proto.Clone(validFullProtoVersion).(*protorefs.Version),
		OwnerId:  proto.Clone(validMinProtoContainer.OwnerId).(*protorefs.OwnerID),
		Nonce:    bytes.Clone(validMinProtoContainer.Nonce),
		BasicAcl: 1043832770,
		Attributes: []*protocontainer.Container_Attribute{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: "Name", Value: "any container name"},
			{Key: "Timestamp", Value: "1732577694"},
			{Key: "__NEOFS__NAME", Value: "any domain name"},
			{Key: "__NEOFS__ZONE", Value: "any domain zone"},
			{Key: "__NEOFS__DISABLE_HOMOMORPHIC_HASHING", Value: "true"},
		},
		PlacementPolicy: &protonetmap.PlacementPolicy{
			Replicas: []*protonetmap.Replica{
				{Count: 3060437, Selector: "selector1"},
				{Count: 156936495, Selector: "selector2"},
			},
			ContainerBackupFactor: 920231904,
			Selectors: []*protonetmap.Selector{
				{Name: "selector1", Count: 1663184999, Clause: 1, Attribute: "attribute1", Filter: "filter1"},
				{Name: "selector2", Count: 2649065896, Clause: 2, Attribute: "attribute2", Filter: "filter2"},
				{Name: "selector_max", Count: 2649065896, Clause: math.MaxInt32, Attribute: "attribute_max", Filter: "filter_max"},
			},
			Filters: []*protonetmap.Filter{
				{Name: "filter1", Key: "key1", Op: 0, Value: "value1", Filters: []*protonetmap.Filter{
					{},
					{},
				}},
				{Op: 1},
				{Op: 2},
				{Op: 3},
				{Op: 4},
				{Op: 5},
				{Op: 6},
				{Op: 7},
				{Op: 8},
				{Op: math.MaxInt32},
			},
			SubnetId: &protorefs.SubnetID{Value: 987533317},
		},
	}
	// correct eACL with required fields only.
	validMinEACL = &protoacl.EACLTable{}
	// correct eACL with required all fields.
	validFullEACL = &protoacl.EACLTable{
		Version:     &protorefs.Version{Major: 538919038, Minor: 3957317479},
		ContainerId: proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID),
		Records: []*protoacl.EACLRecord{
			{},
			{Operation: 1, Action: 1},
			{Operation: 2, Action: 2},
			{Operation: 3, Action: 3},
			{Operation: 4, Action: math.MaxInt32},
			{Operation: 5},
			{Operation: 6},
			{Operation: 7},
			{Operation: math.MaxInt32},
			{Filters: []*protoacl.EACLRecord_Filter{
				{HeaderType: 0, MatchType: 0, Key: "key1", Value: "val1"},
				{HeaderType: 1, MatchType: 1},
				{HeaderType: 2, MatchType: 2},
				{HeaderType: 3, MatchType: 3},
				{HeaderType: math.MaxInt32, MatchType: 4},
				{MatchType: 5},
				{MatchType: 6},
				{MatchType: 7},
				{MatchType: math.MaxInt32},
			}},
			{Targets: []*protoacl.EACLRecord_Target{
				{Role: 0, Keys: [][]byte{[]byte("key1"), []byte("key2")}},
				{Role: 1},
				{Role: 2},
				{Role: 3},
				{Role: math.MaxInt32},
			}},
		},
	}
	// correct ContainerService.Put response payload with required fields only.
	validMinPutContainerResponseBody = &protocontainer.PutResponse_Body{
		ContainerId: proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID),
	}
	// correct ContainerService.Put response payload with all fields.
	validFullPutContainerResponseBody = &protocontainer.PutResponse_Body{
		ContainerId: proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID),
	}
	// correct ContainerService.Get response payload with required fields only.
	validMinGetContainerResponseBody = &protocontainer.GetResponse_Body{
		Container: proto.Clone(validMinProtoContainer).(*protocontainer.Container),
	}
	// correct ContainerService.Get response payload with all fields.
	validFullGetContainerResponseBody = &protocontainer.GetResponse_Body{
		Container: proto.Clone(validFullProtoContainer).(*protocontainer.Container),
		Signature: &protorefs.SignatureRFC6979{Key: []byte("any_key"), Sign: []byte("any_signature")},
		SessionToken: &protosession.SessionToken{
			Body: &protosession.SessionToken_Body{
				Id:         []byte("any_ID"),
				OwnerId:    &protorefs.OwnerID{Value: []byte("any_user")},
				Lifetime:   &protosession.SessionToken_Body_TokenLifetime{Exp: 1, Nbf: 2, Iat: 3},
				SessionKey: []byte("any_session_key"),
			},
			Signature: proto.Clone(validFullProtoSignature).(*protorefs.Signature),
		},
	}
	// correct ContainerService.List response payload with required fields only.
	validMinListContainersResponseBody = (*protocontainer.ListResponse_Body)(nil)
	// correct ContainerService.List response payload with all fields.
	validFullListContainersResponseBody = &protocontainer.ListResponse_Body{
		ContainerIds: []*protorefs.ContainerID{
			proto.Clone(validProtoContainerIDs[0]).(*protorefs.ContainerID),
			proto.Clone(validProtoContainerIDs[1]).(*protorefs.ContainerID),
			proto.Clone(validProtoContainerIDs[2]).(*protorefs.ContainerID),
		},
	}
	// correct ContainerService.Delete response payload with required fields only.
	validMinDeleteContainerResponseBody = (*protocontainer.DeleteResponse_Body)(nil)
	// correct ContainerService.Delete response payload with all fields.
	validFullDeleteContainerResponseBody = &protocontainer.DeleteResponse_Body{}
	// correct ContainerService.GetExtendedACL response payload with required fields only.
	validMinEACLResponseBody = &protocontainer.GetExtendedACLResponse_Body{
		Eacl: proto.Clone(validMinEACL).(*protoacl.EACLTable),
	}
	// correct ContainerService.GetExtendedACL response payload with all fields.
	validFullEACLResponseBody = &protocontainer.GetExtendedACLResponse_Body{
		Eacl:         proto.Clone(validFullEACL).(*protoacl.EACLTable),
		Signature:    proto.Clone(validFullGetContainerResponseBody.Signature).(*protorefs.SignatureRFC6979),
		SessionToken: proto.Clone(validFullGetContainerResponseBody.SessionToken).(*protosession.SessionToken),
	}
	// correct ContainerService.SetExtendedACL response payload with required fields only.
	validMinSetEACLResponseBody = (*protocontainer.SetExtendedACLResponse_Body)(nil)
	// correct ContainerService.SetExtendedACL response payload with all fields.
	validFullSetEACLResponseBody = &protocontainer.SetExtendedACLResponse_Body{}
	// correct ContainerService.AnnounceUsedSpace response payload with required fields only.
	validMinUsedSpaceResponseBody = (*protocontainer.AnnounceUsedSpaceResponse_Body)(nil)
	// correct ContainerService.AnnounceUsedSpace response payload with all fields.
	validFullUsedSpaceResponseBody = &protocontainer.AnnounceUsedSpaceResponse_Body{}
)

// Netmap service.
var (
	// correct node info with required fields only.
	validMinNodeInfo = &protonetmap.NodeInfo{
		PublicKey: []byte("any_pub"),
		Addresses: []string{"any_endpoint"},
	}
	// correct node info with all fields.
	validFullNodeInfo = newValidFullNodeInfo(0)
	// correct network map with required fields only.
	validMinProtoNetmap = &protonetmap.Netmap{}
	// correct network map with all fields.
	validFullProtoNetmap = &protonetmap.Netmap{
		Epoch: 17416815529850981458,
		Nodes: []*protonetmap.NodeInfo{newValidFullNodeInfo(0), newValidFullNodeInfo(1), newValidFullNodeInfo(2)},
	}
	// correct network info with required fields only.
	validMinProtoNetInfo = &protonetmap.NetworkInfo{
		NetworkConfig: &protonetmap.NetworkConfig{
			Parameters: []*protonetmap.NetworkConfig_Parameter{
				{Value: []byte("any")},
			},
		},
	}
	// correct network info with all fields.
	validFullProtoNetInfo = &protonetmap.NetworkInfo{
		CurrentEpoch: 17416815529850981458,
		MagicNumber:  8576993077569092248,
		MsPerBlock:   9059417785180743518,
		NetworkConfig: &protonetmap.NetworkConfig{
			Parameters: []*protonetmap.NetworkConfig_Parameter{
				{Key: []byte("k1"), Value: []byte("v1")},
				{Key: []byte("k2"), Value: []byte("v2")},
				{Key: []byte("AuditFee"), Value: []byte{148, 103, 221, 13, 230, 131, 76, 41}},        // 2975898477883385748
				{Key: []byte("BasicIncomeRate"), Value: []byte{75, 10, 132, 219, 93, 88, 10, 159}},   // 11460069361935714891
				{Key: []byte("ContainerFee"), Value: []byte{138, 229, 49, 0, 30, 129, 67, 130}},      // 9386488014222517642
				{Key: []byte("ContainerAliasFee"), Value: []byte{138, 229, 49, 0, 30, 129, 67, 130}}, // 9386488014222517642
				// TODO: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/653
				// {Key: []byte("EigenTrustAlpha"), Value: []byte("5.551764501727871")},
				{Key: []byte("EigenTrustIterations"), Value: []byte{130, 92, 74, 224, 95, 59, 146, 249}}, // 17983501545014713474
				{Key: []byte("EpochDuration"), Value: []byte{161, 231, 2, 119, 184, 52, 66, 217}},        // 15655133221568571297
				{Key: []byte("HomomorphicHashingDisabled"), Value: []byte("any")},
				{Key: []byte("InnerRingCandidateFee"), Value: []byte{0, 11, 236, 200, 112, 164, 1, 217}}, // 15636960185521277696
				{Key: []byte("MaintenanceModeAllowed"), Value: []byte("any")},
				{Key: []byte("MaxObjectSize"), Value: []byte{109, 133, 46, 32, 118, 66, 240, 72}}, // 5255773840254862701
				{Key: []byte("WithdrawFee"), Value: []byte{216, 63, 55, 77, 56, 24, 171, 101}},    // 7325975848940945368
			},
		},
	}
	// correct NetmapService.LocalNodeInfo response payload with required fields only.
	validMinNodeInfoResponseBody = &protonetmap.LocalNodeInfoResponse_Body{
		Version:  proto.Clone(validMinProtoVersion).(*protorefs.Version),
		NodeInfo: proto.Clone(validMinNodeInfo).(*protonetmap.NodeInfo),
	}
	// correct NetmapService.LocalNodeInfo response payload with all fields.
	validFullNodeInfoResponseBody = &protonetmap.LocalNodeInfoResponse_Body{
		Version:  proto.Clone(validFullProtoVersion).(*protorefs.Version),
		NodeInfo: proto.Clone(validFullNodeInfo).(*protonetmap.NodeInfo),
	}
	// correct NetmapService.NetmapSnapshot response payload with required fields only.
	validMinNetmapResponseBody = &protonetmap.NetmapSnapshotResponse_Body{
		Netmap: proto.Clone(validMinProtoNetmap).(*protonetmap.Netmap),
	}
	// correct NetmapService.NetmapSnapshot response payload with all fields.
	validFullNetmapResponseBody = &protonetmap.NetmapSnapshotResponse_Body{
		Netmap: proto.Clone(validFullProtoNetmap).(*protonetmap.Netmap),
	}
	// correct NetmapService.NetworkInfo response payload with required fields only.
	validMinNetInfoResponseBody = &protonetmap.NetworkInfoResponse_Body{
		NetworkInfo: proto.Clone(validMinProtoNetInfo).(*protonetmap.NetworkInfo),
	}
	// correct NetmapService.NetworkInfo response payload with all fields.
	validFullNetInfoResponseBody = &protonetmap.NetworkInfoResponse_Body{
		NetworkInfo: proto.Clone(validFullProtoNetInfo).(*protonetmap.NetworkInfo),
	}
)

// Object service.
var (
	// valid object header with required fields only.
	validMinObjectHeader = &protoobject.Header{}
	// correct object header with all fields.
	validFullObjectHeader = &protoobject.Header{
		Version: &protorefs.Version{Major: 2551725017, Minor: 2526948189},
		ContainerId: &protorefs.ContainerID{Value: []byte{80, 212, 0, 200, 84, 144, 252, 77, 205, 169, 28, 36, 61, 25, 4, 32,
			182, 161, 107, 148, 193, 86, 1, 252, 224, 65, 204, 176, 27, 189, 63, 198}},
		OwnerId: &protorefs.OwnerID{Value: []byte{53, 36, 208, 131, 238, 151, 230, 27, 245, 87, 156, 55, 90, 144, 192, 82,
			205, 97, 243, 240, 98, 0, 4, 202, 190}},
		CreationEpoch:   535166283637641128,
		PayloadLength:   7493095166286485665,
		PayloadHash:     &protorefs.Checksum{Type: 745469659, Sum: []byte("payload_checksum")},
		ObjectType:      1336146323,
		HomomorphicHash: &protorefs.Checksum{Type: 56973732, Sum: []byte("homomorphic_checksum")},
		SessionToken: &protosession.SessionToken{
			Body: &protosession.SessionToken_Body{
				Id: []byte{219, 53, 231, 42, 56, 82, 65, 196, 175, 34, 22, 36, 170, 248, 64, 45},
				OwnerId: &protorefs.OwnerID{Value: []byte{53, 79, 105, 50, 97, 214, 227, 217, 243, 111, 24, 28, 164, 116, 174, 36,
					217, 111, 165, 197, 109, 225, 168, 165, 133}},
				Lifetime: &protosession.SessionToken_Body_TokenLifetime{
					Exp: 2306780414485650416, Nbf: 17091941679101563337, Iat: 10428481937388069414,
				},
				SessionKey: []byte{3, 47, 174, 204, 218, 71, 223, 103, 27, 142, 185, 141, 190, 177, 199, 235, 100, 168, 68, 216, 253,
					4, 124, 162, 237, 187, 141, 28, 109, 121, 22, 77, 77},
				Context: &protosession.SessionToken_Body_Object{
					Object: &protosession.ObjectSessionContext{
						Verb: 1849442930,
						Target: &protosession.ObjectSessionContext_Target{
							Container: &protorefs.ContainerID{Value: []byte{43, 155, 220, 2, 70, 86, 249, 4, 211, 12, 14, 152, 15, 165,
								141, 240, 15, 199, 82, 245, 32, 86, 49, 60, 3, 15, 235, 107, 227, 21, 201, 226}},
							Objects: []*protorefs.ObjectID{
								{Value: []byte{168, 182, 85, 123, 227, 177, 127, 228, 62, 192, 73, 61, 38, 102, 136, 138, 20, 155, 175,
									89, 95, 241, 200, 148, 156, 142, 215, 78, 34, 223, 238, 62}},
								{Value: []byte{104, 187, 144, 239, 201, 242, 213, 136, 32, 1, 74, 125, 157, 143, 114, 57, 57, 182, 218,
									172, 126, 69, 157, 62, 119, 45, 116, 152, 225, 222, 16, 243}},
								{Value: []byte{106, 193, 15, 88, 111, 154, 77, 182, 11, 190, 3, 154, 84, 249, 1, 165, 220, 23, 234, 101,
									210, 105, 114, 230, 251, 102, 164, 142, 128, 6, 35, 131}},
							},
						}},
				},
			},
			Signature: &protorefs.Signature{Key: []byte("any_public_key"), Sign: []byte("any_signature"), Scheme: 343874216},
		},
		Attributes: []*protoobject.Header_Attribute{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: "__NEOFS__EXPIRATION_EPOCH", Value: "15108052785492221606"},
		},
		Split: &protoobject.Header_Split{
			Parent: &protorefs.ObjectID{Value: []byte{136, 16, 11, 39, 44, 190, 117, 150, 28, 108, 97, 182, 137, 71, 116, 141,
				39, 3, 240, 58, 177, 143, 185, 171, 139, 189, 87, 178, 168, 91, 108, 49}},
			Previous: &protorefs.ObjectID{Value: []byte{70, 184, 70, 223, 213, 136, 169, 221, 63, 103, 244, 43, 109, 226, 9,
				243, 154, 177, 74, 6, 128, 100, 237, 126, 81, 203, 210, 206, 97, 16, 12, 145}},
			ParentSignature: &protorefs.Signature{Key: []byte("any_parent_key"), Sign: []byte("any_parent_signature"), Scheme: 343874216},
			ParentHeader: &protoobject.Header{
				Version: &protorefs.Version{Major: 1650885558, Minor: 1215827697},
				ContainerId: &protorefs.ContainerID{Value: []byte{180, 73, 166, 38, 121, 174, 19, 54, 183, 40, 110, 62, 221, 124, 243,
					108, 222, 97, 21, 41, 154, 159, 92, 217, 99, 136, 75, 2, 71, 243, 230, 33}},
				OwnerId:         &protorefs.OwnerID{Value: []byte{53, 147, 252, 32, 131, 247, 225, 223, 238, 111, 227, 232, 235, 86, 220, 225, 95, 68, 242, 143, 250, 19, 209, 207, 137}},
				CreationEpoch:   13908636632389871906,
				PayloadLength:   9446280261481989231,
				PayloadHash:     &protorefs.Checksum{Type: 1764227836, Sum: []byte("parent_payload_checksum")},
				ObjectType:      950142306,
				HomomorphicHash: &protorefs.Checksum{Type: 2086030953, Sum: []byte("parent_homomorphic_checksum")},
				Attributes: []*protoobject.Header_Attribute{
					{Key: "parent_k1", Value: "parent_v1"},
					{Key: "parent_k2", Value: "parent_v2"},
					{Key: "__NEOFS__EXPIRATION_EPOCH", Value: "5546294308840974481"},
				},
			},
			Children: []*protorefs.ObjectID{
				{Value: []byte{62, 123, 103, 12, 105, 55, 53, 123, 78, 108, 241, 217, 90, 252, 200, 18, 237, 194, 154, 76, 101, 254,
					10, 80, 245, 97, 195, 227, 184, 247, 23, 2}},
				{Value: []byte{127, 105, 152, 33, 27, 219, 170, 156, 77, 47, 133, 82, 253, 100, 203, 229, 12, 231, 39, 223, 155, 199,
					124, 164, 78, 208, 243, 23, 220, 13, 101, 91}},
				{Value: []byte{232, 111, 102, 246, 179, 18, 108, 53, 36, 150, 64, 248, 108, 100, 161, 85, 82, 27, 39, 90, 97, 184, 146,
					230, 139, 162, 43, 171, 65, 184, 255, 238}},
			},
			SplitId: []byte{161, 132, 100, 12, 194, 100, 65, 179, 165, 156, 156, 2, 173, 208, 33, 45},
			First: &protorefs.ObjectID{Value: []byte{43, 82, 110, 195, 252, 103, 56, 184, 106, 229, 94, 136, 213, 63, 133,
				47, 174, 125, 1, 181, 102, 158, 110, 102, 115, 41, 204, 232, 44, 176, 233, 78}},
		},
	}
	// correct split info with required fields only.
	validMinSplitInfo = &protoobject.SplitInfo{
		LastPart: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
	}
	// correct split info with all fields.
	validFullSplitInfo = &protoobject.SplitInfo{
		SplitId:   []byte{181, 76, 71, 204, 73, 230, 65, 146, 156, 76, 98, 233, 55, 162, 45, 223},
		LastPart:  proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
		Link:      proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID),
		FirstPart: proto.Clone(validProtoObjectIDs[2]).(*protorefs.ObjectID),
	}
	// correct ObjectService.Put response payload with required fields only.
	validMinPutObjectResponseBody = &protoobject.PutResponse_Body{
		ObjectId: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
	}
	// correct ObjectService.Put response payload with all fields.
	validFullPutObjectResponseBody = &protoobject.PutResponse_Body{
		ObjectId: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
	}
	// correct ObjectService.Delete response payload with required fields only.
	validMinDeleteObjectResponseBody = &protoobject.DeleteResponse_Body{
		Tombstone: &protorefs.Address{
			ObjectId: proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
		},
	}
	// correct ObjectService.Delete response payload with all fields.
	validFullDeleteObjectResponseBody = &protoobject.DeleteResponse_Body{
		Tombstone: proto.Clone(validFullProtoObjectAddress).(*protorefs.Address),
	}
	// correct ObjectService.GetRangeHash response payload with required fields only.
	validMinObjectHashResponseBody = &protoobject.GetRangeHashResponse_Body{
		HashList: [][]byte{[]byte("one")},
	}
	// correct ObjectService.GetRangeHash response payload with all fields.
	validFullObjectHashResponseBody = &protoobject.GetRangeHashResponse_Body{
		Type:     protorefs.ChecksumType(rand.Int31()),
		HashList: [][]byte{[]byte("one"), []byte("two")},
	}
	// correct ObjectService.Head split info response payload with required fields only.
	validMinObjectSplitInfoHeadResponseBody = &protoobject.HeadResponse_Body{
		Head: &protoobject.HeadResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validMinSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.Head split info response payload with all fields.
	validFullObjectSplitInfoHeadResponseBody = &protoobject.HeadResponse_Body{
		Head: &protoobject.HeadResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validFullSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.Head response payload with required fields only.
	validMinObjectHeadResponseBody = &protoobject.HeadResponse_Body{
		Head: &protoobject.HeadResponse_Body_Header{
			Header: &protoobject.HeaderWithSignature{
				Header:    proto.Clone(validMinObjectHeader).(*protoobject.Header),
				Signature: proto.Clone(validMinProtoSignature).(*protorefs.Signature),
			},
		},
	}
	// correct ObjectService.Head response payload with all fields.
	validFullObjectHeadResponseBody = &protoobject.HeadResponse_Body{
		Head: &protoobject.HeadResponse_Body_Header{
			Header: &protoobject.HeaderWithSignature{
				Header:    proto.Clone(validFullObjectHeader).(*protoobject.Header),
				Signature: proto.Clone(validFullProtoSignature).(*protorefs.Signature),
			},
		},
	}
	// correct ObjectService.Get heading response payload with required fields only.
	validMinHeadingObjectGetResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_Init_{
			Init: &protoobject.GetResponse_Body_Init{
				ObjectId:  proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
				Signature: proto.Clone(validMinProtoSignature).(*protorefs.Signature),
				Header:    proto.Clone(validMinObjectHeader).(*protoobject.Header),
			},
		},
	}
	// correct ObjectService.Get heading response payload with all fields.
	validFullHeadingObjectGetResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_Init_{
			Init: &protoobject.GetResponse_Body_Init{
				ObjectId:  proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
				Signature: proto.Clone(validFullProtoSignature).(*protorefs.Signature),
				Header:    proto.Clone(validFullObjectHeader).(*protoobject.Header),
			},
		},
	}
	// correct ObjectService.Get chunk response payload with all fields.
	validFullChunkObjectGetResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_Chunk{
			Chunk: []byte("Hello, world!"),
		},
	}
	// correct ObjectService.Get split info response payload with required fields only.
	validMinObjectSplitInfoGetResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validMinSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.Get split info response payload with all fields.
	validFullObjectSplitInfoGetResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validFullSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.GetRange chunk response payload with all fields.
	validFullChunkObjectRangeResponseBody = &protoobject.GetRangeResponse_Body{
		RangePart: &protoobject.GetRangeResponse_Body_Chunk{
			Chunk: []byte("Hello, world!"),
		},
	}
	// correct ObjectService.GetRange split info response payload with required fields only.
	validMinObjectSplitInfoRangeResponseBody = &protoobject.GetRangeResponse_Body{
		RangePart: &protoobject.GetRangeResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validMinSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.GetRange split info response payload with all fields.
	validFullObjectSplitInfoRangeResponseBody = &protoobject.GetRangeResponse_Body{
		RangePart: &protoobject.GetRangeResponse_Body_SplitInfo{
			SplitInfo: proto.Clone(validFullSplitInfo).(*protoobject.SplitInfo),
		},
	}
	// correct ObjectService.Search response payload with required fields only.
	validMinSearchResponseBody = &protoobject.SearchResponse_Body{}
	// correct ObjectService.Search response payload with all fields.
	validFullSearchResponseBody = &protoobject.SearchResponse_Body{
		IdList: []*protorefs.ObjectID{
			proto.Clone(validProtoObjectIDs[0]).(*protorefs.ObjectID),
			proto.Clone(validProtoObjectIDs[1]).(*protorefs.ObjectID),
			proto.Clone(validProtoObjectIDs[2]).(*protorefs.ObjectID),
		},
	}
)

// Reputation service.
var (
	// correct ReputationService.AnnounceIntermediateResult response payload with
	// required fields only.
	validMinAnnounceIntermediateRepResponseBody = (*protoreputation.AnnounceIntermediateResultResponse_Body)(nil)
	// correct ReputationService.AnnounceIntermediateResult response payload with
	// all fields.
	validFullAnnounceIntermediateRepResponseBody = &protoreputation.AnnounceIntermediateResultResponse_Body{}
	// correct ReputationService.AnnounceLocalTrust response payload with required
	// fields only.
	validMinAnnounceLocalTrustResponseBody = (*protoreputation.AnnounceLocalTrustResponse_Body)(nil)
	// correct ReputationService.AnnounceLocalTrust response payload with all
	// fields.
	validFullAnnounceLocalTrustRepResponseBody = &protoreputation.AnnounceLocalTrustResponse_Body{}
)

// Session service.
var (
	// correct SessionService.Create response payload with required fields
	// only.
	validMinCreateSessionResponseBody = &protosession.CreateResponse_Body{
		Id:         []byte("any_ID"),
		SessionKey: []byte("any_pub"),
	}
	// correct SessionService.Create response payload with all fields.
	validFullCreateSessionResponseBody = proto.Clone(validMinCreateSessionResponseBody).(*protosession.CreateResponse_Body)
)

func newValidFullNodeInfo(ind int) *protonetmap.NodeInfo {
	si := strconv.Itoa(ind)
	return &protonetmap.NodeInfo{
		PublicKey: []byte("pub_" + si),
		Addresses: []string{"endpoint_" + si + "_0", "endpoint_" + si + "_1"},
		Attributes: []*protonetmap.NodeInfo_Attribute{
			{Key: "attr_key_" + si + "_0", Value: "attr_val_" + si + "_0"},
			{Key: "attr_key_" + si + "_1", Value: "attr_val_" + si + "_1"},
		},
		State: protonetmap.NodeInfo_State(ind),
	}
}

func checkContainerIDTransport(id cid.ID, m *protorefs.ContainerID) error {
	if v1, v2 := id[:], m.GetValue(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

func checkObjectIDTransport(id oid.ID, m *protorefs.ObjectID) error {
	if v1, v2 := id[:], m.GetValue(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

func checkUserIDTransport(id user.ID, m *protorefs.OwnerID) error {
	if v1, v2 := id[:], m.GetValue(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

func checkSignatureTransport(sig neofscrypto.Signature, m *protorefs.Signature) error {
	scheme := sig.Scheme()
	var expScheme protorefs.SignatureScheme
	switch scheme {
	default:
		expScheme = protorefs.SignatureScheme(scheme)
	case neofscrypto.ECDSA_SHA512:
		expScheme = protorefs.SignatureScheme_ECDSA_SHA512
	case neofscrypto.ECDSA_DETERMINISTIC_SHA256:
		expScheme = protorefs.SignatureScheme_ECDSA_RFC6979_SHA256
	case neofscrypto.ECDSA_WALLETCONNECT:
		expScheme = protorefs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT
	}
	if actScheme := m.GetScheme(); actScheme != expScheme {
		return fmt.Errorf("scheme field (client: %v, message: %v)", actScheme, expScheme)
	}
	if v1, v2 := sig.PublicKeyBytes(), m.GetKey(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("public key field (client: %x, message: %x)", v1, v2)
	}
	if v1, v2 := sig.Value(), m.GetSign(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

func checkSignatureRFC6979Transport(sig neofscrypto.Signature, m *protorefs.SignatureRFC6979) error {
	if v1, v2 := sig.PublicKeyBytes(), m.GetKey(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("public key field (client: %x, message: %x)", v1, v2)
	}
	if v1, v2 := sig.Value(), m.GetSign(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

// returns context oneof field of unexported type.
func checkCommonSessionTransport(t interface {
	ID() uuid.UUID
	Issuer() user.ID
	Exp() uint64
	Nbf() uint64
	Iat() uint64
	IssuerPublicKeyBytes() []byte
	AssertAuthKey(neofscrypto.PublicKey) bool
	Signature() (neofscrypto.Signature, bool)
}, m *protosession.SessionToken) (any, error) {
	body := m.GetBody()
	if body == nil {
		return nil, errors.New("missing body field in the message")
	}
	// 1. ID
	id := t.ID()
	if v1, v2 := id[:], body.GetId(); !bytes.Equal(v1, v2) {
		return nil, fmt.Errorf("ID field (client: %x, message: %x)", v1, v2)
	}
	// 2. issuer
	if err := checkUserIDTransport(t.Issuer(), body.GetOwnerId()); err != nil {
		return nil, fmt.Errorf("issuer field: %w", err)
	}
	// 3. lifetime
	lt := body.GetLifetime()
	if v1, v2 := t.Exp(), lt.GetExp(); v1 != v2 {
		return nil, fmt.Errorf("exp lifetime field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := t.Nbf(), lt.GetNbf(); v1 != v2 {
		return nil, fmt.Errorf("nbf lifetime field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := t.Iat(), lt.GetIat(); v1 != v2 {
		return nil, fmt.Errorf("iat lifetime field (client: %d, message: %d)", v1, v2)
	}
	// 4. session key
	var k neofsecdsa.PublicKey
	kb := body.GetSessionKey()
	if err := k.Decode(kb); err != nil {
		return nil, fmt.Errorf("invalid session key in the message: %w", err)
	}
	if !t.AssertAuthKey(&k) {
		return nil, errors.New("session key mismatch")
	}
	// 5+. context
	c := body.GetContext()
	if c == nil {
		return nil, errors.New("missing context")
	}

	msig := m.GetSignature()
	if sig, ok := t.Signature(); ok {
		if msig == nil {
			return nil, errors.New("missing signature field in the message")
		}
		if err := checkSignatureTransport(sig, msig); err != nil {
			return nil, fmt.Errorf("signature field: %w", err)
		}
	} else if msig != nil {
		return nil, errors.New("signature field is set while should not be")
	}
	return c, nil
}

func checkContainerSessionTransport(t session.Container, m *protosession.SessionToken) error {
	c, err := checkCommonSessionTransport(&t, m)
	if err != nil {
		return err
	}
	cb, ok := c.(*protosession.SessionToken_Body_Container)
	if !ok {
		return fmt.Errorf("wrong oneof context field type (client: %T, message: %T)", cb, c)
	}
	cc := cb.Container
	// 1. verb
	var expVerb session.ContainerVerb
	actVerb := cc.GetVerb()
	switch actVerb {
	default:
		expVerb = session.ContainerVerb(actVerb)
	case protosession.ContainerSessionContext_PUT:
		expVerb = session.VerbContainerPut
	case protosession.ContainerSessionContext_DELETE:
		expVerb = session.VerbContainerDelete
	case protosession.ContainerSessionContext_SETEACL:
		expVerb = session.VerbContainerSetEACL
	}
	if !t.AssertVerb(expVerb) {
		return fmt.Errorf("wrong verb in the context field: %v", actVerb)
	}
	// 1.2, 1.3 container(s)
	wc := cc.GetWildcard()
	mc := cc.GetContainerId()
	if mc == nil != wc {
		return errors.New("wildcard flag conflicts with container ID in the context")
	}
	if wc {
		if !t.AppliedTo(cidtest.ID()) {
			return errors.New("wildcard flag is set while should not be")
		}
	} else {
		var expCnr cid.ID
		actCnr := mc.GetValue()
		if copy(expCnr[:], actCnr); !t.AppliedTo(expCnr) {
			return fmt.Errorf("wrong container in the context field: %x", actCnr)
		}
	}
	return nil
}

func checkObjectSessionTransport(t session.Object, m *protosession.SessionToken) error {
	c, err := checkCommonSessionTransport(&t, m)
	if err != nil {
		return err
	}
	co, ok := c.(*protosession.SessionToken_Body_Object)
	if !ok {
		return fmt.Errorf("wrong oneof context field type (client: %T, message: %T)", co, c)
	}
	oo := co.Object
	// 1. verb
	var expVerb session.ObjectVerb
	actVerb := oo.GetVerb()
	switch actVerb {
	default:
		expVerb = session.ObjectVerb(actVerb)
	case protosession.ObjectSessionContext_PUT:
		expVerb = session.VerbObjectPut
	case protosession.ObjectSessionContext_GET:
		expVerb = session.VerbObjectGet
	case protosession.ObjectSessionContext_HEAD:
		expVerb = session.VerbObjectHead
	case protosession.ObjectSessionContext_SEARCH:
		expVerb = session.VerbObjectSearch
	case protosession.ObjectSessionContext_DELETE:
		expVerb = session.VerbObjectDelete
	case protosession.ObjectSessionContext_RANGE:
		expVerb = session.VerbObjectRange
	case protosession.ObjectSessionContext_RANGEHASH:
		expVerb = session.VerbObjectRangeHash
	}
	if !t.AssertVerb(expVerb) {
		return fmt.Errorf("wrong verb in the context field: %v", actVerb)
	}
	// 2. target
	// 2.1. container
	mtgt := oo.GetTarget()
	mc := mtgt.GetContainer()
	if mc == nil {
		return errors.New("missing container in the context field")
	}
	var expCnr cid.ID
	actCnr := mc.GetValue()
	if copy(expCnr[:], actCnr); !t.AssertContainer(expCnr) {
		return fmt.Errorf("wrong container in the context field: %x", actCnr)
	}
	// 2.2. objects
	mo := mtgt.GetObjects()
	var expObj oid.ID
	for i := range mo {
		actObj := mo[i].GetValue()
		if copy(expObj[:], actObj); !t.AssertObject(expObj) {
			return fmt.Errorf("wrong object #%d in the context field: %x", i, actObj)
		}
	}
	// FIXME: t can have more objects, this is wrong but won't be detected. Full
	//  list should be accessible to verify.
	return nil
}

func checkBearerTokenTransport(b bearer.Token, m *protoacl.BearerToken) error {
	body := m.GetBody()
	if body == nil {
		return errors.New("missing body field in the message")
	}
	// 1. eACL
	me := body.GetEaclTable()
	if me == nil {
		return errors.New("missing eACL in the message")
	}
	if err := checkEACLTransport(b.EACLTable(), me); err != nil {
		return fmt.Errorf("eACL field: %w", err)
	}
	// 2. owner
	mo := body.GetOwnerId()
	if mo == nil {
		return errors.New("missing owner field")
	}
	var expUsr user.ID
	actUsr := mo.GetValue()
	if copy(expUsr[:], actUsr); !b.AssertUser(expUsr) {
		return fmt.Errorf("wrong owner: %x", actUsr)
	}
	// 3. lifetime
	lt := body.GetLifetime()
	if v1, v2 := b.Exp(), lt.GetExp(); v1 != v2 {
		return fmt.Errorf("exp lifetime field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := b.Nbf(), lt.GetNbf(); v1 != v2 {
		return fmt.Errorf("nbf lifetime field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := b.Iat(), lt.GetIat(); v1 != v2 {
		return fmt.Errorf("iat lifetime field (client: %d, message: %d)", v1, v2)
	}
	// 4. issuer
	if err := checkUserIDTransport(b.Issuer(), body.GetIssuer()); err != nil {
		return fmt.Errorf("issuer field: %w", err)
	}
	return nil
}

func checkVersionTransport(v version.Version, m *protorefs.Version) error {
	if v1, v2 := v.Major(), m.GetMajor(); v1 != v2 {
		return fmt.Errorf("major field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := v.Minor(), m.GetMinor(); v1 != v2 {
		return fmt.Errorf("minor field (client: %d, message: %d)", v1, v2)
	}
	return nil
}

func checkBalanceTransport(b accounting.Decimal, m *protoaccounting.Decimal) error {
	if v1, v2 := b.Value(), m.GetValue(); v1 != v2 {
		return fmt.Errorf("value field (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := b.Precision(), m.GetPrecision(); v1 != v2 {
		return fmt.Errorf("precision field (client: %d, message: %d)", v1, v2)
	}
	return nil
}

func checkStoragePolicyFilterTransport(f netmap.Filter, m *protonetmap.Filter) error {
	// 1. name
	if v1, v2 := f.Name(), m.GetName(); v1 != v2 {
		return fmt.Errorf("name (client: %q, message: %q)", v1, v2)
	}
	// 2. key
	if v1, v2 := f.Key(), m.GetKey(); v1 != v2 {
		return fmt.Errorf("key (client: %q, message: %q)", v1, v2)
	}
	// 3. op
	var expOp protonetmap.Operation
	switch op := f.Op(); op {
	default:
		expOp = protonetmap.Operation(op)
	case netmap.FilterOpEQ:
		expOp = protonetmap.Operation_EQ
	case netmap.FilterOpNE:
		expOp = protonetmap.Operation_NE
	case netmap.FilterOpGT:
		expOp = protonetmap.Operation_GT
	case netmap.FilterOpGE:
		expOp = protonetmap.Operation_GE
	case netmap.FilterOpLT:
		expOp = protonetmap.Operation_LT
	case netmap.FilterOpLE:
		expOp = protonetmap.Operation_LE
	case netmap.FilterOpOR:
		expOp = protonetmap.Operation_OR
	case netmap.FilterOpAND:
		expOp = protonetmap.Operation_AND
	}
	if actOp := m.GetOp(); actOp != expOp {
		return fmt.Errorf("op (client: %v, message: %v)", expOp, actOp)
	}
	// 4. value
	if v1, v2 := f.Value(), m.GetValue(); v1 != v2 {
		return fmt.Errorf("value (client: %q, message: %q)", v1, v2)
	}
	// 5. sub-filters
	cfs, mfs := f.SubFilters(), m.GetFilters()
	if v1, v2 := len(cfs), len(mfs); v1 != v2 {
		return fmt.Errorf("number of sub-filters (client: %d, message: %d)", v1, v2)
	}
	for i := range cfs {
		if err := checkStoragePolicyFilterTransport(cfs[i], mfs[i]); err != nil {
			return fmt.Errorf("sub-filter#%d: %w", i, err)
		}
	}
	return nil
}

func checkStoragePolicyTransport(p netmap.PlacementPolicy, m *protonetmap.PlacementPolicy) error {
	// 1. replicas
	crs, mrs := p.Replicas(), m.GetReplicas()
	if v1, v2 := len(crs), len(mrs); v1 != v2 {
		return fmt.Errorf("number of replicas (client: %d, message: %d)", v1, v2)
	}
	for i, cr := range crs {
		mr := mrs[i]
		if v1, v2 := cr.NumberOfObjects(), mr.GetCount(); v1 != v2 {
			return fmt.Errorf("replica#%d field: object count (client: %d, message: %d)", i, v1, v2)
		}
		if v1, v2 := cr.SelectorName(), mr.GetSelector(); v1 != v2 {
			return fmt.Errorf("replica#%d field: selector (client: %v, message: %v)", i, v1, v2)
		}
	}
	// 2. backup factor
	if v1, v2 := p.ContainerBackupFactor(), m.GetContainerBackupFactor(); v1 != v2 {
		return fmt.Errorf("backup factor (client: %d, message: %d)", v1, v2)
	}
	// 3. selectors
	css, mss := p.Selectors(), m.GetSelectors()
	if v1, v2 := len(css), len(mss); v1 != v2 {
		return fmt.Errorf("number of selectors (client: %d, message: %d)", v1, v2)
	}
	for i, cs := range css {
		ms := mss[i]
		// 1. name
		if v1, v2 := cs.Name(), ms.GetName(); v1 != v2 {
			return fmt.Errorf("selector#%d field: name (client: %q, message: %q)", i, v1, v2)
		}
		// 2. count
		if v1, v2 := cs.NumberOfNodes(), ms.GetCount(); v1 != v2 {
			return fmt.Errorf("selector#%d field: node count (client: %d, message: %d)", i, v1, v2)
		}
		// 3. clause
		var expClause protonetmap.Clause
		actClause := ms.GetClause()
		switch {
		default:
			expClause = p.ProtoMessage().Selectors[i].Clause
		case cs.IsSame():
			expClause = protonetmap.Clause_SAME
		case cs.IsDistinct():
			expClause = protonetmap.Clause_DISTINCT
		}
		if actClause != expClause {
			return fmt.Errorf("selector#%d field: clause (client: %v, message: %v)", i, expClause, actClause)
		}
		// 4. attribute
		if v1, v2 := cs.BucketAttribute(), ms.GetAttribute(); v1 != v2 {
			return fmt.Errorf("selector#%d field: attribute (client: %q, message: %q)", i, v1, v2)
		}
		// 5. filter
		if v1, v2 := cs.FilterName(), ms.GetFilter(); v1 != v2 {
			return fmt.Errorf("selector#%d field: filter (client: %q, message: %q)", i, v1, v2)
		}
	}
	// filters
	cfs, mfs := p.Filters(), m.GetFilters()
	if v1, v2 := len(cfs), len(mfs); v1 != v2 {
		return fmt.Errorf("number of filters (client: %d, message: %d)", v1, v2)
	}
	for i, mf := range mfs {
		if err := checkStoragePolicyFilterTransport(cfs[i], mf); err != nil {
			return fmt.Errorf("filter#%d field: %w", i, err)
		}
	}
	return nil
}

func checkContainerTransport(c container.Container, m *protocontainer.Container) error {
	// 1. version
	if err := checkVersionTransport(c.Version(), m.GetVersion()); err != nil {
		return fmt.Errorf("version field: %w", err)
	}
	// 2. owner
	if err := checkUserIDTransport(c.Owner(), m.GetOwnerId()); err != nil {
		return fmt.Errorf("owner field: %w", err)
	}
	// 3. nonce
	// TODO(https://github.com/nspcc-dev/neofs-sdk-go/issues/664): access nonce from c directly
	mc := c.ProtoMessage()
	if v1, v2 := mc.GetNonce(), m.GetNonce(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("nonce field (client: %x, message: %x)", v1, v2)
	}
	// 4. basic ACL
	if v1, v2 := c.BasicACL().Bits(), m.GetBasicAcl(); v1 != v2 {
		return fmt.Errorf("basic ACL field (client: %d, message: %d)", v1, v2)
	}
	// 5. attributes
	var mas [][2]string
	var name, dmn, zone string
	var disableHomoHash bool
	var timestamp int64
	for _, ma := range m.GetAttributes() {
		k, v := ma.GetKey(), ma.GetValue()
		mas = append(mas, [2]string{k, v})
		switch k {
		case "Name":
			name = v
		case "Timestamp":
			var err error
			if timestamp, err = strconv.ParseInt(v, 10, 64); err != nil {
				return fmt.Errorf("invalid timestamp attribute value %q in the message: %w", v, err)
			}
		}
		if tail, ok := strings.CutPrefix(k, "__NEOFS__"); ok {
			switch tail {
			case "NAME":
				dmn = v
			case "ZONE":
				zone = v
			case "DISABLE_HOMOMORPHIC_HASHING":
				disableHomoHash = v == "true"
			}
		}
	}
	var cas [][2]string
	c.IterateAttributes(func(k, v string) { cas = append(cas, [2]string{k, v}) })
	if v1, v2 := len(cas), len(mas); v1 != v2 {
		return fmt.Errorf("number of attributes (client: %d, message: %d)", v1, v2)
	}
	for i, ca := range cas {
		if ma := mas[i]; ca != ma {
			return fmt.Errorf("attribute #%d (client: %v, message: %v)", i, ca, ma)
		}
	}
	if v1, v2 := c.Name(), name; v1 != v2 {
		return fmt.Errorf("name attribute (client: %q, message: %q)", v1, v2)
	}
	if v1, v2 := c.IsHomomorphicHashingDisabled(), disableHomoHash; v2 != v1 {
		return fmt.Errorf("homomorphic hashing flag attribute (client: %t, message: %t)", v1, v2)
	}
	if v1, v2 := c.CreatedAt().Unix(), timestamp; v2 != v1 {
		return fmt.Errorf("timestamp attribute (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := c.ReadDomain().Name(), dmn; v1 != v2 {
		return fmt.Errorf("domain name attribute (client: %q, message: %q)", v1, v2)
	}
	if zone == "" {
		zone = "container"
	}
	if v1, v2 := c.ReadDomain().Zone(), zone; v2 != v1 {
		return fmt.Errorf("domain zone attribute (client: %q, message: %q)", v1, v2)
	}
	// 6. policy
	mp := m.GetPlacementPolicy()
	if mp == nil {
		return errors.New("missing storage policy field in the message")
	}
	if err := checkStoragePolicyTransport(c.PlacementPolicy(), mp); err != nil {
		return fmt.Errorf("storage policy field: %w", err)
	}
	return nil
}

func checkEACLFilterTransport(f eacl.Filter, m *protoacl.EACLRecord_Filter) error {
	// 1. header type
	var expHdr protoacl.HeaderType
	switch ht := f.From(); ht {
	default:
		expHdr = protoacl.HeaderType(ht)
	case eacl.HeaderFromRequest:
		expHdr = protoacl.HeaderType_REQUEST
	case eacl.HeaderFromObject:
		expHdr = protoacl.HeaderType_OBJECT
	case eacl.HeaderFromService:
		expHdr = protoacl.HeaderType_SERVICE
	}
	if act := m.GetHeaderType(); act != expHdr {
		return fmt.Errorf("header type (client: %v, message: %v)", act, expHdr)
	}
	// matcher
	var expMatcher protoacl.MatchType
	switch m := f.Matcher(); m {
	default:
		expMatcher = protoacl.MatchType(m)
	case eacl.MatchStringEqual:
		expMatcher = protoacl.MatchType_STRING_EQUAL
	case eacl.MatchStringNotEqual:
		expMatcher = protoacl.MatchType_STRING_NOT_EQUAL
	case eacl.MatchNotPresent:
		expMatcher = protoacl.MatchType_NOT_PRESENT
	case eacl.MatchNumGT:
		expMatcher = protoacl.MatchType_NUM_GT
	case eacl.MatchNumGE:
		expMatcher = protoacl.MatchType_NUM_GE
	case eacl.MatchNumLT:
		expMatcher = protoacl.MatchType_NUM_LT
	case eacl.MatchNumLE:
		expMatcher = protoacl.MatchType_NUM_LE
	}
	if act := m.GetMatchType(); act != expMatcher {
		return fmt.Errorf("match type (client: %v, message: %v)", act, expMatcher)
	}
	// 4. key
	if v1, v2 := f.Key(), m.GetKey(); v1 != v2 {
		return fmt.Errorf("key field (client: %q, message: %q)", v1, v2)
	}
	// 4. value
	if v1, v2 := f.Value(), m.GetValue(); v1 != v2 {
		return fmt.Errorf("value field (client: %q, message: %q)", v1, v2)
	}
	return nil
}

func checkEACLTargetTransport(t eacl.Target, m *protoacl.EACLRecord_Target) error {
	// role
	var expRole protoacl.Role
	switch r := t.Role(); r {
	default:
		expRole = protoacl.Role(r)
	case eacl.RoleUser:
		expRole = protoacl.Role_USER
	case eacl.RoleSystem:
		expRole = protoacl.Role_SYSTEM
	case eacl.RoleOthers:
		expRole = protoacl.Role_OTHERS
	}
	if act := m.GetRole(); act != expRole {
		return fmt.Errorf("role (client: %v, message: %v)", act, expRole)
	}
	// 2. subjects
	cks, mks := t.RawSubjects(), m.GetKeys()
	if v1, v2 := len(cks), len(mks); v1 != v2 {
		return fmt.Errorf("number of subjects (client: %d, message: %d)", v1, v2)
	}
	for i := range cks {
		if !bytes.Equal(cks[i], mks[i]) {
			return fmt.Errorf("subject#%d (client: %x, message: %x)", i, cks[i], mks[i])
		}
	}
	return nil
}

func checkEACLRecordTransport(r eacl.Record, m *protoacl.EACLRecord) error {
	// 1. op
	var expOp protoacl.Operation
	switch op := r.Operation(); op {
	default:
		expOp = protoacl.Operation(op)
	case eacl.OperationGet:
		expOp = protoacl.Operation_GET
	case eacl.OperationHead:
		expOp = protoacl.Operation_HEAD
	case eacl.OperationPut:
		expOp = protoacl.Operation_PUT
	case eacl.OperationDelete:
		expOp = protoacl.Operation_DELETE
	case eacl.OperationSearch:
		expOp = protoacl.Operation_SEARCH
	case eacl.OperationRange:
		expOp = protoacl.Operation_GETRANGE
	case eacl.OperationRangeHash:
		expOp = protoacl.Operation_GETRANGEHASH
	}
	if act := m.GetOperation(); act != expOp {
		return fmt.Errorf("op (client: %v, message: %v)", act, expOp)
	}
	// 2. action
	var expAction protoacl.Action
	switch a := r.Action(); a {
	default:
		expAction = protoacl.Action(a)
	case eacl.ActionAllow:
		expAction = protoacl.Action_ALLOW
	case eacl.ActionDeny:
		expAction = protoacl.Action_DENY
	}
	if act := m.GetAction(); act != expAction {
		return fmt.Errorf("action (client: %v, message: %v)", act, expAction)
	}
	// 3. filters
	mfs, cfs := m.GetFilters(), r.Filters()
	if v1, v2 := len(cfs), len(mfs); v1 != v2 {
		return fmt.Errorf("number of filters (client: %d, message: %d)", v1, v2)
	}
	for i := range cfs {
		if err := checkEACLFilterTransport(cfs[i], mfs[i]); err != nil {
			return fmt.Errorf("filter#%d field: %w", i, err)
		}
	}
	// 4. targets
	mts, cts := m.GetTargets(), r.Targets()
	if v1, v2 := len(cfs), len(mfs); v1 != v2 {
		return fmt.Errorf("number of targets (client: %d, message: %d)", v1, v2)
	}
	for i := range mts {
		if err := checkEACLTargetTransport(cts[i], mts[i]); err != nil {
			return fmt.Errorf("target#%d field: %w", i, err)
		}
	}
	return nil
}

func checkEACLTransport(e eacl.Table, m *protoacl.EACLTable) error {
	// 1. version
	if err := checkVersionTransport(e.Version(), m.GetVersion()); err != nil {
		return fmt.Errorf("version field: %w", err)
	}
	// 2. container ID
	mc := m.GetContainerId()
	if c := e.GetCID(); c.IsZero() {
		if mc != nil {
			return errors.New("container ID field is set while should not be")
		}
	} else {
		if mc == nil {
			return errors.New("missing container ID field")
		}
		if err := checkContainerIDTransport(c, mc); err != nil {
			return fmt.Errorf("container ID field: %w", err)
		}
	}
	// 3. records
	mrs, crs := m.GetRecords(), e.Records()
	if v1, v2 := len(crs), len(mrs); v1 != v2 {
		return fmt.Errorf("number of records (client: %d, message: %d)", v1, v2)
	}
	for i := range mrs {
		if err := checkEACLRecordTransport(crs[i], mrs[i]); err != nil {
			return fmt.Errorf("record#%d field: %w", i, err)
		}
	}
	return nil
}

func checkContainerSizeEstimationTransport(e container.SizeEstimation, m *protocontainer.AnnounceUsedSpaceRequest_Body_Announcement) error {
	// 1. epoch
	if v1, v2 := e.Epoch(), m.GetEpoch(); v1 != v2 {
		return fmt.Errorf("epoch field (client: %d, message: %d)", v1, v2)
	}
	// 1. container ID
	mc := m.GetContainerId()
	if mc == nil {
		return newErrMissingRequestBodyField("container ID")
	}
	if err := checkContainerIDTransport(e.Container(), mc); err != nil {
		return fmt.Errorf("container ID field: %w", err)
	}
	// 3. value
	if v1, v2 := e.Value(), m.GetUsedSpace(); v1 != v2 {
		return fmt.Errorf("value field (client: %d, message: %d)", v1, v2)
	}
	return nil
}

func checkNodeInfoTransport(n netmap.NodeInfo, m *protonetmap.NodeInfo) error {
	// 1. public key
	if v1, v2 := n.PublicKey(), m.GetPublicKey(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("public key field (client: %x, message: %x)", v1, v2)
	}
	// 2. addresses
	maddrs := m.GetAddresses()
	var caddrs []string
	netmap.IterateNetworkEndpoints(n, func(e string) { caddrs = append(caddrs, e) })
	if v1, v2 := len(caddrs), len(maddrs); v1 != v2 {
		return fmt.Errorf("number of addresses (client: %d, message: %d)", v1, v2)
	}
	for i := range caddrs {
		if v1, v2 := caddrs[i], maddrs[i]; v1 != v2 {
			return fmt.Errorf("name (client: %q, message: %q)", v1, v2)
		}
	}
	// 3. attributes
	attrs, mattrs := n.GetAttributes(), m.GetAttributes()
	if v1, v2 := len(attrs), len(mattrs); v1 != v2 {
		return fmt.Errorf("number of attributes (client: %d, message: %d)", v1, v2)
	}
	for i, ma := range mattrs {
		a := attrs[i]
		if v1, v2 := a[0], ma.GetKey(); v1 != v2 {
			return fmt.Errorf("attribute#%d field: key (client: %q, message: %q)", i, v1, v2)
		}
		if v1, v2 := a[1], ma.GetValue(); v1 != v2 {
			return fmt.Errorf("attribute#%d field: value (client: %q, message: %q)", i, v1, v2)
		}
		if len(ma.GetParents()) > 0 {
			return fmt.Errorf("attribute#%d field: parents field is set while should not be", i)
		}
	}
	// 4. state
	var expState protonetmap.NodeInfo_State
	switch {
	default:
		expState = n.ProtoMessage().State
	case n.IsOnline():
		expState = protonetmap.NodeInfo_ONLINE
	case n.IsOffline():
		expState = protonetmap.NodeInfo_OFFLINE
	case n.IsMaintenance():
		expState = protonetmap.NodeInfo_MAINTENANCE
	}
	if act := m.GetState(); act != expState {
		return fmt.Errorf("state field (client: %v, message: %v)", expState, act)
	}
	return nil
}

func checkNetmapTransport(n netmap.NetMap, m *protonetmap.Netmap) error {
	// 1. epoch
	if v1, v2 := n.Epoch(), m.GetEpoch(); v1 != v2 {
		return fmt.Errorf("epoch field (client: %d, message: %d)", v1, v2)
	}
	// 2. nodes
	cns, mns := n.Nodes(), m.GetNodes()
	if v1, v2 := len(cns), len(mns); v1 != v2 {
		return fmt.Errorf("number of nodes (client: %d, message: %d)", v1, v2)
	}
	for i := range cns {
		if err := checkNodeInfoTransport(cns[i], mns[i]); err != nil {
			return fmt.Errorf("node#%d field: %w", i, err)
		}
	}
	return nil
}

func checkNetInfoTransport(n netmap.NetworkInfo, m *protonetmap.NetworkInfo) error {
	// 1. current epoch
	if v1, v2 := n.CurrentEpoch(), m.GetCurrentEpoch(); v1 != v2 {
		return fmt.Errorf("current epoch field (client: %d, message: %d)", v1, v2)
	}
	// 2. magic
	if v1, v2 := n.MagicNumber(), m.GetMagicNumber(); v1 != v2 {
		return fmt.Errorf("network magic field (client: %d, message: %d)", v1, v2)
	}
	// 3. ms per block
	if v1, v2 := n.MsPerBlock(), m.GetMsPerBlock(); v1 != v2 {
		return fmt.Errorf("ms per block field (client: %d, message: %d)", v1, v2)
	}
	// 4. config
	mps := m.GetNetworkConfig().GetParameters()
	var raw []string
	n.IterateRawNetworkParameters(func(name string, value []byte) { raw = append(raw, name, string(value)) })

	var mraw []string
	var auditFee, storagePrice, cnrDmnFee, cnrFee, etIters, epochDur, irFee, maxObjSize, withdrawFee uint64
	var etAlpha float64
	var homoHashDisabled, maintenanceAllowed bool
	for _, mp := range mps {
		k, v := mp.GetKey(), mp.GetValue()
		switch string(k) {
		default:
			mraw = append(mraw, string(k), string(v))
		case "AuditFee":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			auditFee = binary.LittleEndian.Uint64(v)
		case "BasicIncomeRate":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			storagePrice = binary.LittleEndian.Uint64(v)
		case "ContainerAliasFee":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			cnrDmnFee = binary.LittleEndian.Uint64(v)
		case "ContainerFee":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			cnrFee = binary.LittleEndian.Uint64(v)
		case "EigenTrustAlpha":
			var err error
			if etAlpha, err = strconv.ParseFloat(string(v), 64); err != nil {
				return fmt.Errorf("invalid parameter %q value %q: %w", k, v, err)
			}
		case "EigenTrustIterations":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			etIters = binary.LittleEndian.Uint64(v)
		case "EpochDuration":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			epochDur = binary.LittleEndian.Uint64(v)
		case "HomomorphicHashingDisabled":
			for _, b := range v {
				if homoHashDisabled = b != 0; homoHashDisabled {
					break
				}
			}
		case "InnerRingCandidateFee":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			irFee = binary.LittleEndian.Uint64(v)
		case "MaintenanceModeAllowed":
			for _, b := range v {
				if maintenanceAllowed = b != 0; maintenanceAllowed {
					break
				}
			}
		case "MaxObjectSize":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			maxObjSize = binary.LittleEndian.Uint64(v)
		case "WithdrawFee":
			if l := len(v); l < 8 {
				return fmt.Errorf("too short parameter %q value: %d bytes", k, l)
			}
			withdrawFee = binary.LittleEndian.Uint64(v)
		}
	}
	if v1, v2 := len(raw), len(mraw); v1 != v2 {
		return fmt.Errorf("number of raw config values: (client: %d, message: %d)", v1, v2)
	}
	for i := 0; i < len(raw); i += 2 {
		if raw[i] != mraw[i] {
			return fmt.Errorf("raw config #%d: key (client: %q, message: %q)", i, raw[i], mraw[i])
		}
		if raw[i+1] != mraw[i+1] {
			return fmt.Errorf("raw config #%d: value (client: %q, message: %q)", i, raw[i+1], mraw[i+1])
		}
	}
	if v1, v2 := n.AuditFee(), auditFee; v1 != v2 {
		return fmt.Errorf("audit fee parameter (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.StoragePrice(), storagePrice; v1 != v2 {
		return fmt.Errorf("storage price parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.ContainerFee(), cnrFee; v1 != v2 {
		return fmt.Errorf("container fee parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.NamedContainerFee(), cnrDmnFee; v1 != v2 {
		return fmt.Errorf("container domain fee parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.NumberOfEigenTrustIterations(), etIters; v1 != v2 {
		return fmt.Errorf("number of Eigen-Trust iterations parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.EigenTrustAlpha(), etAlpha; v1 != v2 {
		return fmt.Errorf("Eigen-Trust alpha parameter value (client: %v, message: %v)", v1, v2)
	}
	if v1, v2 := n.EpochDuration(), epochDur; v1 != v2 {
		return fmt.Errorf("epoch duration parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.IRCandidateFee(), irFee; v1 != v2 {
		return fmt.Errorf("IR candidate fee parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.MaintenanceModeAllowed(), maintenanceAllowed; v1 != v2 {
		return fmt.Errorf("maintenance mode allowance parameter value (client: %t, message: %t)", v1, v2)
	}
	if v1, v2 := n.MaxObjectSize(), maxObjSize; v1 != v2 {
		return fmt.Errorf("max object size parameter value (client: %d, message: %d)", v1, v2)
	}
	if v1, v2 := n.WithdrawalFee(), withdrawFee; v1 != v2 {
		return fmt.Errorf("withdrawal fee parameter value (client: %d, message: %d)", v1, v2)
	}
	return nil
}

func checkReputationPeerTransport(p reputation.PeerID, m *protoreputation.PeerID) error {
	if m == nil {
		return errors.New("missing peer field")
	}
	if v1, v2 := p.PublicKey(), m.GetPublicKey(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("peer field (client: %x, message: %x)", v1, v2)
	}
	return nil
}

func checkTrustTransport(t reputation.Trust, m *protoreputation.Trust) error {
	if err := checkReputationPeerTransport(t.Peer(), m.GetPeer()); err != nil {
		return fmt.Errorf("peer field: %w", err)
	}
	if v1, v2 := t.Value(), m.GetValue(); v1 != v2 {
		return fmt.Errorf("value field (client: %v, message: %v)", v1, v2)
	}
	return nil
}

func checkP2PTrustTransport(t reputation.PeerToPeerTrust, m *protoreputation.PeerToPeerTrust) error {
	if err := checkReputationPeerTransport(t.TrustingPeer(), m.GetTrustingPeer()); err != nil {
		return fmt.Errorf("trusting peer field: %w", err)
	}
	if err := checkTrustTransport(t.Trust(), m.GetTrust()); err != nil {
		return fmt.Errorf("trust field: %w", err)
	}
	return nil
}

func checkHashTransport(c checksum.Checksum, m *protorefs.Checksum) error {
	var expType protorefs.ChecksumType
	switch typ := c.Type(); typ {
	default:
		expType = protorefs.ChecksumType(typ)
	case checksum.SHA256:
		expType = protorefs.ChecksumType_SHA256
	case checksum.TillichZemor:
		expType = protorefs.ChecksumType_TZ
	}
	if actType := m.GetType(); actType != expType {
		return fmt.Errorf("type field (client: %v, message %v)", expType, actType)
	}
	if v1, v2 := c.Value(), m.GetSum(); !bytes.Equal(v1, v2) {
		return fmt.Errorf("value field (client: %x, message %x)", v1, v2)
	}
	return nil
}

func checkObjectHeaderWithSignatureTransport(o object.Object, m *protoobject.HeaderWithSignature) error {
	if err := checkObjectHeaderTransport(o, m.GetHeader()); err != nil {
		return fmt.Errorf("header field: %w", err)
	}
	s := o.Signature()
	ms := m.GetSignature()
	if s != nil {
		if ms == nil {
			return errors.New("missing signature field")
		}
		if err := checkSignatureTransport(*s, ms); err != nil {
			return fmt.Errorf("signature field: %w", err)
		}
	} else {
		if ms != nil {
			return errors.New("signature field is set while should not be")
		}
	}
	return nil
}

func checkObjectHeaderTransport(h object.Object, m *protoobject.Header) error {
	// 1. version
	ver := h.Version()
	if ver != nil {
		if err := checkVersionTransport(*ver, m.GetVersion()); err != nil {
			return fmt.Errorf("version field: %w", err)
		}
	} else {
		if m.GetVersion() != nil {
			return errors.New("version field is set while should not")
		}
	}
	// 2. container
	cnr := h.GetContainerID()
	if cnr.IsZero() {
		if m.GetContainerId() != nil {
			return errors.New("container field is set while should not")
		}
	} else {
		if err := checkContainerIDTransport(cnr, m.GetContainerId()); err != nil {
			return fmt.Errorf("container field: %w", err)
		}
	}
	// 3. owner
	ownr := h.Owner()
	if ownr.IsZero() {
		if m.GetOwnerId() != nil {
			return errors.New("owner field is set while should not be")
		}
	} else {
		if err := checkUserIDTransport(ownr, m.GetOwnerId()); err != nil {
			return fmt.Errorf("owner field: %w", err)
		}
	}
	// 4. creation epoch
	if v1, v2 := h.CreationEpoch(), m.GetCreationEpoch(); v1 != v2 {
		return fmt.Errorf("creation epoch field (client: %d, message: %d)", v1, v2)
	}
	// 5. payload length
	if v1, v2 := h.PayloadSize(), m.GetPayloadLength(); v1 != v2 {
		return fmt.Errorf("payload length field (client: %d, message: %d)", v1, v2)
	}
	// 6. payload checksum
	cs, ok := h.PayloadChecksum()
	mcs := m.GetPayloadHash()
	if ok {
		if mcs == nil {
			return errors.New("missing payload checksum field")
		}
		if err := checkHashTransport(cs, mcs); err != nil {
			return fmt.Errorf("payload checksum field: %w", err)
		}
	} else {
		if mcs != nil {
			return errors.New("payload checksum field is set while should not be")
		}
	}
	// 7. type
	var expType protoobject.ObjectType
	switch typ := h.Type(); typ {
	default:
		expType = protoobject.ObjectType(typ)
	case object.TypeRegular:
		expType = protoobject.ObjectType_REGULAR
	case object.TypeTombstone:
		expType = protoobject.ObjectType_TOMBSTONE
	case object.TypeStorageGroup:
		expType = protoobject.ObjectType_STORAGE_GROUP
	case object.TypeLock:
		expType = protoobject.ObjectType_LOCK
	case object.TypeLink:
		expType = protoobject.ObjectType_LINK
	}
	if actType := m.GetObjectType(); actType != expType {
		return fmt.Errorf("type field (client: %v, message %v)", expType, actType)
	}
	// 8. payload homomorphic checksum
	cs, ok = h.PayloadHomomorphicHash()
	mcs = m.GetHomomorphicHash()
	if ok {
		if mcs == nil {
			return errors.New("missing payload homomorphic checksum field")
		}
		if err := checkHashTransport(cs, mcs); err != nil {
			return fmt.Errorf("payload homomorphic checksum field: %w", err)
		}
	} else {
		if mcs != nil {
			return errors.New("payload homomorphic checksum field is set while should not be")
		}
	}
	// 9. session token
	st := h.SessionToken()
	mst := m.GetSessionToken()
	if st != nil {
		if mst == nil {
			return errors.New("missing session token field")
		}
		if err := checkObjectSessionTransport(*st, mst); err != nil {
			return fmt.Errorf("session token field: %w", err)
		}
	} else {
		if mst != nil {
			return errors.New("session token field is set while should not be")
		}
	}
	// 10. attributes
	as := h.Attributes()
	mas := m.GetAttributes()
	if v1, v2 := len(as), len(mas); v1 != v2 {
		return fmt.Errorf("number of attributes (client: %d, message: %d)", v1, v2)
	}
	for i := range as {
		if v1, v2 := as[i].Key(), mas[i].GetKey(); v1 != v2 {
			return fmt.Errorf("attribute#%d: key (client: %q, message: %q)", i, v1, v2)
		}
		if v1, v2 := as[i].Value(), mas[i].GetValue(); v1 != v2 {
			return fmt.Errorf("attribute#%d: value (client: %q, message: %q)", i, v1, v2)
		}
	}
	// 11. split
	parID := h.GetParentID()
	prev := h.GetPreviousID()
	first := h.GetFirstID()
	children := h.Children()
	parHdr := h.Parent()
	splitID := h.SplitID()
	sh := m.GetSplit()
	if parID.IsZero() && parHdr == nil && prev.IsZero() && first.IsZero() && len(children) == 0 && splitID == nil {
		if sh != nil {
			return errors.New("split header field is set while should not be")
		}
	} else {
		if err := checkObjectSplitTransport(parID, prev, parHdr, children, splitID, first, sh); err != nil {
			return fmt.Errorf("split header field: %w", err)
		}
	}
	return nil
}

func checkObjectSplitTransport(parID oid.ID, prev oid.ID, parHdrSig *object.Object, children []oid.ID,
	splitID *object.SplitID, first oid.ID, m *protoobject.Header_Split) error {
	// 1. parent ID
	mid := m.GetParent()
	if parID.IsZero() {
		if mid != nil {
			return errors.New("parent ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing parent ID field")
		}
		if err := checkObjectIDTransport(parID, mid); err != nil {
			return fmt.Errorf("parent ID field: %w", err)
		}
	}
	// 2. previous ID
	mid = m.GetPrevious()
	if prev.IsZero() {
		if mid != nil {
			return errors.New("previous ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing previous ID field")
		}
		if err := checkObjectIDTransport(prev, mid); err != nil {
			return fmt.Errorf("previous ID field: %w", err)
		}
	}
	// 3,4. parent signature, header
	mph := m.GetParentHeader()
	mps := m.GetParentSignature()
	if parHdrSig != nil {
		if mph == nil && mps == nil {
			return errors.New("missing both parent header and signature")
		}
		if mph != nil {
			if err := checkObjectHeaderTransport(*parHdrSig, mph); err != nil {
				return fmt.Errorf("parent header field: %w", err)
			}
		}
		if ps := parHdrSig.Signature(); ps != nil {
			if mps == nil {
				return errors.New("missing parent header field")
			}
			if err := checkSignatureTransport(*ps, mps); err != nil {
				return fmt.Errorf("parent signature field: %w", err)
			}
		} else {
			if mps != nil {
				return errors.New("parent signature field is set while should not be")
			}
		}
	} else {
		if mph != nil {
			return errors.New("parent header field is set while should not be")
		}
		if mps != nil {
			return errors.New("parent signature field is set while should not be")
		}
	}
	// 5. children
	mc := m.GetChildren()
	if v1, v2 := len(children), len(mc); v1 != v2 {
		return fmt.Errorf("number of children (client: %d, message: %d)", v1, v2)
	}
	for i := range children {
		if mc[i] == nil {
			return fmt.Errorf("children field: nil element #%d", i)
		}
		if err := checkObjectIDTransport(children[i], mc[i]); err != nil {
			return fmt.Errorf("children field: child#%d: %w", i, err)
		}
	}
	// 6. split ID
	actSplitID := m.GetSplitId()
	if splitID != nil {
		if expSplitID := splitID.ToV2(); !bytes.Equal(actSplitID, expSplitID) {
			return fmt.Errorf("split ID field (client: %x, message: %x)", expSplitID, actSplitID)
		}
	} else {
		if len(actSplitID) > 0 {
			return errors.New("split ID field is set while should not be")
		}
	}
	// 7. first ID
	mid = m.GetFirst()
	if first.IsZero() {
		if mid != nil {
			return errors.New("first ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing first ID field")
		}
		if err := checkObjectIDTransport(first, mid); err != nil {
			return fmt.Errorf("first ID field: %w", err)
		}
	}
	return nil
}

func checkSplitInfoTransport(s object.SplitInfo, m *protoobject.SplitInfo) error {
	// 1. split ID
	splitID := s.SplitID()
	actSplitID := m.GetSplitId()
	if splitID != nil {
		if expSplitID := splitID.ToV2(); !bytes.Equal(actSplitID, expSplitID) {
			return fmt.Errorf("split ID field (client: %x, message: %x)", expSplitID, actSplitID)
		}
	} else {
		if len(actSplitID) > 0 {
			return errors.New("split ID field is set while should not be")
		}
	}
	// 2. last ID
	id := s.GetLastPart()
	mid := m.GetLastPart()
	if id.IsZero() {
		if mid != nil {
			return errors.New("last ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing last ID field")
		}
		if err := checkObjectIDTransport(id, mid); err != nil {
			return fmt.Errorf("last ID field: %w", err)
		}
	}
	// 3. linker
	id = s.GetLink()
	mid = m.GetLink()
	if id.IsZero() {
		if mid != nil {
			return errors.New("linker ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing linker ID field")
		}
		if err := checkObjectIDTransport(id, mid); err != nil {
			return fmt.Errorf("linker ID field: %w", err)
		}
	}
	// 4. first part
	id = s.GetFirstPart()
	mid = m.GetFirstPart()
	if id.IsZero() {
		if mid != nil {
			return errors.New("first ID field is set while should not be")
		}
	} else {
		if mid == nil {
			return errors.New("missing first ID field")
		}
		if err := checkObjectIDTransport(id, mid); err != nil {
			return fmt.Errorf("first ID field: %w", err)
		}
	}
	return nil
}

func checkObjectSearchFilterTransport(f object.SearchFilter, m *protoobject.SearchFilter) error {
	// 1. matcher
	var expMatcher protoobject.MatchType
	switch m := f.Operation(); m {
	default:
		expMatcher = protoobject.MatchType(m)
	case object.MatchStringEqual:
		expMatcher = protoobject.MatchType_STRING_EQUAL
	case object.MatchStringNotEqual:
		expMatcher = protoobject.MatchType_STRING_NOT_EQUAL
	case object.MatchNotPresent:
		expMatcher = protoobject.MatchType_NOT_PRESENT
	case object.MatchCommonPrefix:
		expMatcher = protoobject.MatchType_COMMON_PREFIX
	case object.MatchNumGT:
		expMatcher = protoobject.MatchType_NUM_GT
	case object.MatchNumGE:
		expMatcher = protoobject.MatchType_NUM_GE
	case object.MatchNumLT:
		expMatcher = protoobject.MatchType_NUM_LT
	case object.MatchNumLE:
		expMatcher = protoobject.MatchType_NUM_LE
	}
	if mtch := m.GetMatchType(); mtch != expMatcher {
		return fmt.Errorf("matcher (client: %v, message: %v)", expMatcher, mtch)
	}
	// 2. key
	if v1, v2 := f.Header(), m.GetKey(); v1 != v2 {
		return fmt.Errorf("key (client: %q, message: %q)", v1, v2)
	}
	// 3. value
	if v1, v2 := f.Value(), m.GetValue(); v1 != v2 {
		return fmt.Errorf("value (client: %q, message: %q)", v1, v2)
	}
	return nil
}

func checkObjectSearchFiltersTransport(fs []object.SearchFilter, ms []*protoobject.SearchFilter) error {
	if v1, v2 := len(fs), len(ms); v1 != v2 {
		return fmt.Errorf("number of attributes (client: %d, message: %d)", v1, v2)
	}
	for i := range fs {
		if err := checkObjectSearchFilterTransport(fs[i], ms[i]); err != nil {
			return fmt.Errorf("filter #%d: %w", i, err)
		}
	}
	return nil
}
