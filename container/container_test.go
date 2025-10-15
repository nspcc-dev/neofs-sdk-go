package container_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	containertest "github.com/nspcc-dev/neofs-sdk-go/container/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/stretchr/testify/require"
)

const (
	anyValidBackupFactor = 153493707
	anyValidName         = "any container name"
	anyValidDomainName   = "any domain name"
	anyValidDomainZone   = "any domain zone"
)

var (
	anyValidOwner = user.ID{53, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113, 133, 241,
		114, 112, 11, 234, 139}
	anyValidCreationTime = time.Unix(1727681164, 0)
	anyValidNonce        = uuid.UUID{229, 22, 237, 42, 123, 159, 78, 139, 136, 206, 237, 126, 224, 125, 147, 223}
	anyValidDomain       container.Domain       // set by init.
	anyValidBasicACL     acl.Basic              // set by init.
	anyValidPolicy       netmap.PlacementPolicy // set by init.
)

var validContainer container.Container // set by init.

func init() {
	anyValidBasicACL.FromBits(1043832770)

	anyValidDomain.SetName(anyValidDomainName)
	anyValidDomain.SetZone(anyValidDomainZone)

	anyValidPolicy.SetContainerBackupFactor(anyValidBackupFactor)
	rs := make([]netmap.ReplicaDescriptor, 2)
	rs[0].SetSelectorName("selector_0")
	rs[0].SetNumberOfObjects(2583748530)
	rs[1].SetSelectorName("selector_1")
	rs[1].SetNumberOfObjects(358755354)
	anyValidPolicy.SetReplicas(rs)
	ss := make([]netmap.Selector, 2)
	ss[0].SetName("selector_0")
	ss[0].SetNumberOfNodes(1814781076)
	ss[0].SelectSame()
	ss[0].SetFilterName("filter_0")
	ss[0].SelectByBucketAttribute("attribute_0")
	ss[1].SetName("selector_1")
	ss[1].SetNumberOfNodes(1505136737)
	ss[1].SelectDistinct()
	ss[1].SetFilterName("filter_1")
	ss[1].SelectByBucketAttribute("attribute_1")
	anyValidPolicy.SetSelectors(ss)
	// filters
	fs := make([]netmap.Filter, 2)
	subs := make([]netmap.Filter, 2)
	subs[0].SetName("filter_0_0")
	subs[0].Equal("key_0_0", "val_0_0")
	subs[1].SetName("filter_0_1")
	subs[1].NotEqual("key_0_1", "val_0_1")
	fs[0].SetName("filter_0")
	fs[0].LogicalAND(subs...)
	subs = make([]netmap.Filter, 4)
	subs[0].SetName("filter_1_0")
	subs[0].NumericGT("key_1_0", 1889407708985023116)
	subs[1].SetName("filter_1_1")
	subs[1].NumericGE("key_1_1", 1429243097315344888)
	subs[2].SetName("filter_1_2")
	subs[2].NumericLT("key_1_2", 3722656060317482335)
	subs[3].SetName("filter_1_3")
	subs[3].NumericLE("key_1_3", 1950504987705284805)
	fs[1].SetName("filter_1")
	fs[1].LogicalOR(subs...)
	anyValidPolicy.SetFilters(fs)

	validContainer.Init()
	validContainer.SetOwner(anyValidOwner)
	validContainer.SetBasicACL(anyValidBasicACL)
	validContainer.SetPlacementPolicy(anyValidPolicy)
	validContainer.SetAttribute("k1", "v1")
	validContainer.SetAttribute("k2", "v2")
	validContainer.SetName(anyValidName)
	validContainer.SetCreationTime(anyValidCreationTime)
	validContainer.WriteDomain(anyValidDomain)
	validContainer.DisableHomomorphicHashing()

	// init sets random nonce, we need a fixed one. There is no setter for it
	m := validContainer.ProtoMessage()
	m.Nonce = anyValidNonce[:]
	m.Version.Major = 2
	m.Version.Minor = 16
	if err := validContainer.FromProtoMessage(m); err != nil {
		panic(fmt.Errorf("unexpected encode-decode failure: %w", err))
	}
}

var (
	anyECDSAPrivateKey = ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(),
			X: new(big.Int).SetBytes([]byte{244, 235, 150, 254, 16, 223, 121, 92, 82, 95, 93, 0, 218, 75, 97,
				182, 224, 29, 29, 126, 136, 127, 95, 227, 148, 120, 101, 174, 116, 191, 113, 56}),
			Y: new(big.Int).SetBytes([]byte{162, 142, 254, 167, 43, 228, 23, 134, 112, 148, 125, 252, 40, 205,
				120, 74, 50, 155, 194, 180, 37, 229, 18, 105, 143, 250, 110, 254, 3, 20, 159, 152}),
		},
		D: new(big.Int).SetBytes([]byte{37, 38, 152, 197, 254, 145, 122, 170, 199, 181, 85, 225, 135, 215,
			58, 94, 65, 111, 216, 11, 91, 240, 13, 191, 233, 192, 59, 95, 242, 32, 142, 145}),
	}
	// corresponds to anyECDSAPrivateKey.
	anyBinECDSAPublicKey = []byte{2, 244, 235, 150, 254, 16, 223, 121, 92, 82, 95, 93, 0, 218, 75, 97, 182, 224, 29, 29,
		126, 136, 127, 95, 227, 148, 120, 101, 174, 116, 191, 113, 56}
	// corresponds to validContainer and anyECDSAPrivateKey.
	validContainerSignatureBytes = []byte{121, 154, 72, 98, 128, 173, 181, 250, 139, 192, 159, 13, 60, 18, 52, 8, 16, 32, 210, 214, 8,
		97, 254, 186, 154, 104, 129, 255, 11, 159, 173, 245, 131, 68, 211, 110, 221, 145, 7, 28, 101, 235, 25, 31, 215, 129, 233,
		30, 109, 18, 65, 2, 159, 243, 111, 157, 159, 241, 148, 9, 123, 145, 222, 151}
	validContainerSignature = neofscrypto.NewSignatureFromRawKey(neofscrypto.ECDSA_DETERMINISTIC_SHA256,
		anyBinECDSAPublicKey, validContainerSignatureBytes)
)

// corresponds to validContainer.
var validBinContainer = []byte{
	10, 4, 8, 2, 16, 16, 18, 27, 10, 25, 53, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113,
	133, 241, 114, 112, 11, 234, 139, 26, 16, 229, 22, 237, 42, 123, 159, 78, 139, 136, 206, 237, 126, 224, 125, 147, 223,
	32, 194, 191, 222, 241, 3, 42, 8, 10, 2, 107, 49, 18, 2, 118, 49, 42, 8, 10, 2, 107, 50, 18, 2, 118, 50, 42, 26, 10, 4,
	78, 97, 109, 101, 18, 18, 97, 110, 121, 32, 99, 111, 110, 116, 97, 105, 110, 101, 114, 32, 110, 97, 109, 101, 42, 23, 10, 9, 84,
	105, 109, 101, 115, 116, 97, 109, 112, 18, 10, 49, 55, 50, 55, 54, 56, 49, 49, 54, 52, 42, 32, 10, 13, 95, 95, 78, 69, 79, 70,
	83, 95, 95, 78, 65, 77, 69, 18, 15, 97, 110, 121, 32, 100, 111, 109, 97, 105, 110, 32, 110, 97, 109, 101, 42, 32, 10, 13, 95,
	95, 78, 69, 79, 70, 83, 95, 95, 90, 79, 78, 69, 18, 15, 97, 110, 121, 32, 100, 111, 109, 97, 105, 110, 32, 122, 111, 110, 101,
	42, 44, 10, 36, 95, 95, 78, 69, 79, 70, 83, 95, 95, 68, 73, 83, 65, 66, 76, 69, 95, 72, 79, 77, 79, 77, 79, 82, 80, 72, 73,
	67, 95, 72, 65, 83, 72, 73, 78, 71, 18, 4, 116, 114, 117, 101, 50, 160, 3, 10, 18, 8, 178, 191, 131, 208, 9, 18, 10, 115, 101,
	108, 101, 99, 116, 111, 114, 95, 48, 10, 18, 8, 154, 216, 136, 171, 1, 18, 10, 115, 101, 108, 101, 99, 116, 111, 114, 95, 49, 16, 203,
	193, 152, 73, 26, 43, 10, 10, 115, 101, 108, 101, 99, 116, 111, 114, 95, 48, 16, 148, 185, 173, 225, 6, 24, 1, 34, 11, 97, 116,
	116, 114, 105, 98, 117, 116, 101, 95, 48, 42, 8, 102, 105, 108, 116, 101, 114, 95, 48, 26, 43, 10, 10, 115, 101, 108, 101, 99, 116,
	111, 114, 95, 49, 16, 225, 160, 218, 205, 5, 24, 2, 34, 11, 97, 116, 116, 114, 105, 98, 117, 116, 101, 95, 49, 42, 8, 102, 105,
	108, 116, 101, 114, 95, 49, 34, 80, 10, 8, 102, 105, 108, 116, 101, 114, 95, 48, 24, 8, 42, 32, 10, 10, 102, 105, 108, 116, 101,
	114, 95, 48, 95, 48, 18, 7, 107, 101, 121, 95, 48, 95, 48, 24, 1, 34, 7, 118, 97, 108, 95, 48, 95, 48, 42, 32, 10, 10, 102,
	105, 108, 116, 101, 114, 95, 48, 95, 49, 18, 7, 107, 101, 121, 95, 48, 95, 49, 24, 2, 34, 7, 118, 97, 108, 95, 48, 95, 49, 34,
	196, 1, 10, 8, 102, 105, 108, 116, 101, 114, 95, 49, 24, 7, 42, 44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 48, 18, 7,
	107, 101, 121, 95, 49, 95, 48, 24, 3, 34, 19, 49, 56, 56, 57, 52, 48, 55, 55, 48, 56, 57, 56, 53, 48, 50, 51, 49, 49, 54, 42,
	44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 49, 18, 7, 107, 101, 121, 95, 49, 95, 49, 24, 4, 34, 19, 49, 52, 50, 57,
	50, 52, 51, 48, 57, 55, 51, 49, 53, 51, 52, 52, 56, 56, 56, 42, 44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 50, 18,
	7, 107, 101, 121, 95, 49, 95, 50, 24, 5, 34, 19, 51, 55, 50, 50, 54, 53, 54, 48, 54, 48, 51, 49, 55, 52, 56, 50, 51, 51, 53,
	42, 44, 10, 10, 102, 105, 108, 116, 101, 114, 95, 49, 95, 51, 18, 7, 107, 101, 121, 95, 49, 95, 51, 24, 6, 34, 19, 49, 57, 53,
	48, 53, 48, 52, 57, 56, 55, 55, 48, 53, 50, 56, 52, 56, 48, 53,
}

// corresponds to validContainer.
var validJSONContainer = `
{
 "version": {
  "major": 2,
  "minor": 16
 },
 "ownerID": {
  "value": "NcTZXOr0cNe+eVf8j/MkjihxhfFycAvqiw=="
 },
 "nonce": "5RbtKnufTouIzu1+4H2T3w==",
 "basicACL": 1043832770,
 "attributes": [
  {
   "key": "k1",
   "value": "v1"
  },
  {
   "key": "k2",
   "value": "v2"
  },
  {
   "key": "Name",
   "value": "any container name"
  },
  {
   "key": "Timestamp",
   "value": "1727681164"
  },
  {
   "key": "__NEOFS__NAME",
   "value": "any domain name"
  },
  {
   "key": "__NEOFS__ZONE",
   "value": "any domain zone"
  },
  {
   "key": "__NEOFS__DISABLE_HOMOMORPHIC_HASHING",
   "value": "true"
  }
 ],
 "placementPolicy": {
  "replicas": [
   {
    "count": 2583748530,
    "selector": "selector_0"
   },
   {
    "count": 358755354,
    "selector": "selector_1"
   }
  ],
  "containerBackupFactor": 153493707,
  "selectors": [
   {
    "name": "selector_0",
    "count": 1814781076,
    "clause": "SAME",
    "attribute": "attribute_0",
    "filter": "filter_0"
   },
   {
    "name": "selector_1",
    "count": 1505136737,
    "clause": "DISTINCT",
    "attribute": "attribute_1",
    "filter": "filter_1"
   }
  ],
  "filters": [
   {
    "name": "filter_0",
    "key": "",
    "op": "AND",
    "value": "",
    "filters": [
     {
      "name": "filter_0_0",
      "key": "key_0_0",
      "op": "EQ",
      "value": "val_0_0",
      "filters": []
     },
     {
      "name": "filter_0_1",
      "key": "key_0_1",
      "op": "NE",
      "value": "val_0_1",
      "filters": []
     }
    ]
   },
   {
    "name": "filter_1",
    "key": "",
    "op": "OR",
    "value": "",
    "filters": [
     {
      "name": "filter_1_0",
      "key": "key_1_0",
      "op": "GT",
      "value": "1889407708985023116",
      "filters": []
     },
     {
      "name": "filter_1_1",
      "key": "key_1_1",
      "op": "GE",
      "value": "1429243097315344888",
      "filters": []
     },
     {
      "name": "filter_1_2",
      "key": "key_1_2",
      "op": "LT",
      "value": "3722656060317482335",
      "filters": []
     },
     {
      "name": "filter_1_3",
      "key": "key_1_3",
      "op": "LE",
      "value": "1950504987705284805",
      "filters": []
     }
    ]
   }
  ],
  "subnetId": null,
  "ecRules": []
 }
}
`

// type does not provide getter, this is a helper.
func extractNonce(t testing.TB, c container.Container) uuid.UUID {
	m := c.ProtoMessage()
	var nonce uuid.UUID
	if b := m.GetNonce(); len(b) > 0 {
		require.NoError(t, nonce.UnmarshalBinary(b))
	}
	return nonce
}

func setContainerAttributes(m *protocontainer.Container, els ...string) {
	if len(els)%2 != 0 {
		panic("must be even")
	}
	m.Attributes = make([]*protocontainer.Container_Attribute, len(els)/2)
	for i := range len(els) / 2 {
		m.Attributes[i] = &protocontainer.Container_Attribute{Key: els[2*i], Value: els[2*i+1]}
	}
}

func TestContainer_FromProtoMessage(t *testing.T) {
	m := &protocontainer.Container{
		Version:  &refs.Version{Major: 2526956385, Minor: 95168785},
		OwnerId:  &refs.OwnerID{Value: anyValidOwner[:]},
		Nonce:    anyValidNonce[:],
		BasicAcl: anyValidBasicACL.Bits(),
		Attributes: []*protocontainer.Container_Attribute{
			{Key: "k1", Value: "v1"},
			{Key: "k2", Value: "v2"},
			{Key: "Name", Value: anyValidName},
			{Key: "Timestamp", Value: "1727681164"},
			{Key: "__NEOFS__NAME", Value: anyValidDomainName},
			{Key: "__NEOFS__ZONE", Value: anyValidDomainZone},
			{Key: "__NEOFS__DISABLE_HOMOMORPHIC_HASHING", Value: "true"},
		},
		PlacementPolicy: &protonetmap.PlacementPolicy{
			Replicas: []*protonetmap.Replica{
				{Count: 2583748530, Selector: "selector_0"},
				{Count: 358755354, Selector: "selector_1"},
			},
			ContainerBackupFactor: anyValidBackupFactor,
			Selectors: []*protonetmap.Selector{
				{Name: "selector_0", Count: 1814781076, Clause: protonetmap.Clause_SAME, Attribute: "attribute_0", Filter: "filter_0"},
				{Name: "selector_1", Count: 1814781076, Clause: protonetmap.Clause_DISTINCT, Attribute: "attribute_1", Filter: "filter_1"},
			},
			Filters: []*protonetmap.Filter{
				{Name: "filter_0", Op: protonetmap.Operation_AND, Filters: []*protonetmap.Filter{
					{Name: "filter_0_0", Key: "key_0_0", Op: protonetmap.Operation_EQ, Value: "val_0_0"},
					{Name: "filter_0_1", Key: "key_0_1", Op: protonetmap.Operation_NE, Value: "val_0_1"},
				}},
				{Name: "filter_1", Op: protonetmap.Operation_OR, Filters: []*protonetmap.Filter{
					{Name: "filter_1_0", Key: "key_1_0", Op: protonetmap.Operation_GT, Value: "1889407708985023116"},
					{Name: "filter_1_1", Key: "key_1_1", Op: protonetmap.Operation_GE, Value: "1429243097315344888"},
					{Name: "filter_1_2", Key: "key_1_2", Op: protonetmap.Operation_LT, Value: "3722656060317482335"},
					{Name: "filter_1_3", Key: "key_1_3", Op: protonetmap.Operation_LE, Value: "1950504987705284805"},
				}},
			},
		},
	}

	var val container.Container
	require.NoError(t, val.FromProtoMessage(m))
	ver := val.Version()
	require.EqualValues(t, 2526956385, ver.Major())
	require.EqualValues(t, 95168785, ver.Minor())
	require.Equal(t, anyValidOwner, val.Owner())
	require.EqualValues(t, 1043832770, val.BasicACL())
	require.Equal(t, "v1", val.Attribute("k1"))
	require.Equal(t, "v2", val.Attribute("k2"))
	require.Equal(t, anyValidName, val.Name())
	require.Equal(t, anyValidCreationTime.Unix(), val.CreatedAt().Unix())
	require.Equal(t, anyValidDomainName, val.ReadDomain().Name())
	require.Equal(t, anyValidDomainZone, val.ReadDomain().Zone())
	require.True(t, val.IsHomomorphicHashingDisabled())

	pp := val.PlacementPolicy()
	require.EqualValues(t, anyValidBackupFactor, pp.ContainerBackupFactor())
	rs := pp.Replicas()
	require.Len(t, rs, 2)
	require.Equal(t, "selector_0", rs[0].SelectorName())
	require.EqualValues(t, 2583748530, rs[0].NumberOfObjects())
	require.Equal(t, "selector_1", rs[1].SelectorName())
	require.EqualValues(t, 358755354, rs[1].NumberOfObjects())
	ss := pp.Selectors()
	require.Len(t, ss, 2)
	require.Equal(t, "selector_0", ss[0].Name())
	require.EqualValues(t, 1814781076, ss[0].NumberOfNodes())
	require.True(t, ss[0].IsSame())
	require.Equal(t, "filter_0", ss[0].FilterName())
	require.Equal(t, "attribute_0", ss[0].BucketAttribute())
	require.Equal(t, "selector_1", ss[1].Name())
	require.EqualValues(t, 1814781076, ss[1].NumberOfNodes())
	require.True(t, ss[1].IsDistinct())
	require.Equal(t, "filter_1", ss[1].FilterName())
	require.Equal(t, "attribute_1", ss[1].BucketAttribute())
	fs := pp.Filters()
	require.Len(t, fs, 2)
	require.Equal(t, "filter_0", fs[0].Name())
	require.Zero(t, fs[0].Key())
	require.Equal(t, netmap.FilterOpAND, fs[0].Op())
	require.Zero(t, fs[0].Value())
	subs := fs[0].SubFilters()
	require.Equal(t, "filter_0_0", subs[0].Name())
	require.Equal(t, "key_0_0", subs[0].Key())
	require.Equal(t, netmap.FilterOpEQ, subs[0].Op())
	require.Equal(t, "val_0_0", subs[0].Value())
	require.Empty(t, subs[0].SubFilters())
	require.Equal(t, "filter_0_1", subs[1].Name())
	require.Equal(t, "key_0_1", subs[1].Key())
	require.Equal(t, netmap.FilterOpNE, subs[1].Op())
	require.Equal(t, "val_0_1", subs[1].Value())
	require.Empty(t, subs[1].SubFilters())
	require.Equal(t, "filter_1", fs[1].Name())
	require.Zero(t, fs[1].Key())
	require.Equal(t, netmap.FilterOpOR, fs[1].Op())
	require.Zero(t, fs[1].Value())
	subs = fs[1].SubFilters()
	require.Equal(t, "filter_1_0", subs[0].Name())
	require.Equal(t, "key_1_0", subs[0].Key())
	require.Equal(t, netmap.FilterOpGT, subs[0].Op())
	require.Equal(t, "1889407708985023116", subs[0].Value())
	require.Empty(t, subs[0].SubFilters())
	require.Equal(t, "filter_1_1", subs[1].Name())
	require.Equal(t, "key_1_1", subs[1].Key())
	require.Equal(t, netmap.FilterOpGE, subs[1].Op())
	require.Equal(t, "1429243097315344888", subs[1].Value())
	require.Empty(t, subs[1].SubFilters())
	require.Equal(t, "filter_1_2", subs[2].Name())
	require.Equal(t, "key_1_2", subs[2].Key())
	require.Equal(t, netmap.FilterOpLT, subs[2].Op())
	require.Equal(t, "3722656060317482335", subs[2].Value())
	require.Empty(t, subs[2].SubFilters())
	require.Equal(t, "filter_1_3", subs[3].Name())
	require.Equal(t, "key_1_3", subs[3].Key())
	require.Equal(t, netmap.FilterOpLE, subs[3].Op())
	require.Equal(t, "1950504987705284805", subs[3].Value())
	require.Empty(t, subs[3].SubFilters())

	// reset optional fields
	m.BasicAcl = 0
	m.Attributes = []*protocontainer.Container_Attribute{{Key: "__NEOFS__DISABLE_HOMOMORPHIC_HASHING", Value: "anything not true"}}
	val2 := val
	require.NoError(t, val2.FromProtoMessage(m))
	require.Zero(t, val2.BasicACL())
	require.False(t, val2.IsHomomorphicHashingDisabled())

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*protocontainer.Container)
		}{
			{name: "version/missing", err: "missing version",
				corrupt: func(m *protocontainer.Container) { m.Version = nil }},
			{name: "owner/missing", err: "missing owner",
				corrupt: func(m *protocontainer.Container) { m.OwnerId = nil }},
			{name: "owner/value/nil", err: "invalid owner: invalid length 0, expected 25",
				corrupt: func(m *protocontainer.Container) { m.OwnerId.Value = nil }},
			{name: "owner/value/empty", err: "invalid owner: invalid length 0, expected 25",
				corrupt: func(m *protocontainer.Container) { m.OwnerId.Value = []byte{} }},
			{name: "owner/value/undersize", err: "invalid owner: invalid length 24, expected 25",
				corrupt: func(m *protocontainer.Container) { m.OwnerId.Value = make([]byte, 24) }},
			{name: "owner/value/oversize", err: "invalid owner: invalid length 26, expected 25",
				corrupt: func(m *protocontainer.Container) { m.OwnerId.Value = make([]byte, 26) }},
			{name: "owner/value/wrong prefix", err: "invalid owner: invalid prefix byte 0x42, expected 0x35",
				corrupt: func(m *protocontainer.Container) {
					m.OwnerId.Value = bytes.Clone(anyValidOwner[:])
					m.OwnerId.Value[0] = 0x42
				}},
			{name: "owner/value/checksum mismatch", err: "invalid owner: checksum mismatch",
				corrupt: func(m *protocontainer.Container) {
					m.OwnerId.Value = bytes.Clone(anyValidOwner[:])
					m.OwnerId.Value[24]++
				}},
			{name: "nonce/nil", err: "missing nonce",
				corrupt: func(m *protocontainer.Container) { m.Nonce = nil }},
			{name: "nonce/empty", err: "missing nonce",
				corrupt: func(m *protocontainer.Container) { m.Nonce = []byte{} }},
			{name: "nonce/undersize", err: "invalid nonce: invalid UUID (got 15 bytes)",
				corrupt: func(m *protocontainer.Container) { m.Nonce = anyValidNonce[:15] }},
			{name: "nonce/oversize", err: "invalid nonce: invalid UUID (got 17 bytes)",
				corrupt: func(m *protocontainer.Container) { m.Nonce = append(anyValidNonce[:], 1) }},
			{name: "nonce/wrong version", err: "invalid nonce: wrong UUID version 3, expected 4",
				corrupt: func(m *protocontainer.Container) {
					m.Nonce = bytes.Clone(anyValidNonce[:])
					m.Nonce[6] = 3 << 4
				}},
			{name: "policy/rules/nil", err: "invalid placement policy: missing both REP and EC rules",
				corrupt: func(m *protocontainer.Container) { m.PlacementPolicy.Replicas = nil }},
			{name: "policy/rules/empty", err: "invalid placement policy: missing both REP and EC rules",
				corrupt: func(m *protocontainer.Container) {
					m.PlacementPolicy.Replicas, m.PlacementPolicy.EcRules = []*protonetmap.Replica{}, []*protonetmap.PlacementPolicy_ECRule{}
				}},
			{name: "attributes/no key", err: "empty attribute key",
				corrupt: func(m *protocontainer.Container) { setContainerAttributes(m, "k1", "v1", "", "v2") }},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				corrupt: func(m *protocontainer.Container) { setContainerAttributes(m, "k1", "v1", "k2", "") }},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				corrupt: func(m *protocontainer.Container) { setContainerAttributes(m, "k1", "v1", "k2", "v2", "k1", "v3") }},
			{name: "attributes/timestamp", err: "invalid attribute value Timestamp: foo (strconv.ParseInt: parsing \"foo\": invalid syntax)",
				corrupt: func(m *protocontainer.Container) { setContainerAttributes(m, "Timestamp", "foo") }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				val2 := val
				m := val2.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(container.Container).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestContainer_ProtoMessage(t *testing.T) {
	var val container.Container

	// zero
	m := val.ProtoMessage()
	require.Zero(t, m.GetVersion())
	require.Zero(t, m.GetOwnerId())
	require.Zero(t, m.GetNonce())
	require.Zero(t, m.GetBasicAcl())
	require.Zero(t, m.GetPlacementPolicy())
	require.Zero(t, m.GetAttributes())

	// filled
	m = validContainer.ProtoMessage()
	require.EqualValues(t, 2, m.GetVersion().GetMajor())
	require.EqualValues(t, 16, m.GetVersion().GetMinor())
	require.Len(t, m.GetNonce(), 16)
	require.EqualValues(t, 4, m.GetNonce()[6]>>4)
	require.EqualValues(t, 1043832770, m.GetBasicAcl())
	mas := m.GetAttributes()
	require.Len(t, mas, 7)
	for i, pair := range [][2]string{
		{"k1", "v1"},
		{"k2", "v2"},
		{"Name", anyValidName},
		{"Timestamp", "1727681164"},
		{"__NEOFS__NAME", anyValidDomainName},
		{"__NEOFS__ZONE", anyValidDomainZone},
		{"__NEOFS__DISABLE_HOMOMORPHIC_HASHING", "true"},
	} {
		require.EqualValues(t, pair[0], mas[i].GetKey())
		require.EqualValues(t, pair[1], mas[i].GetValue())
	}

	mp := m.GetPlacementPolicy()
	require.EqualValues(t, anyValidBackupFactor, mp.GetContainerBackupFactor())

	mrs := mp.GetReplicas()
	require.Len(t, mrs, 2)
	require.Equal(t, "selector_0", mrs[0].GetSelector())
	require.EqualValues(t, 2583748530, mrs[0].GetCount())
	require.Equal(t, "selector_1", mrs[1].GetSelector())
	require.EqualValues(t, 358755354, mrs[1].GetCount())

	mss := mp.GetSelectors()
	require.Len(t, mss, 2)
	require.Equal(t, "selector_0", mss[0].GetName())
	require.EqualValues(t, 1814781076, mss[0].GetCount())
	require.Equal(t, protonetmap.Clause_SAME, mss[0].GetClause())
	require.Equal(t, "filter_0", mss[0].GetFilter())
	require.Equal(t, "attribute_0", mss[0].GetAttribute())
	require.Equal(t, "selector_1", mss[1].GetName())
	require.EqualValues(t, 1505136737, mss[1].GetCount())
	require.Equal(t, protonetmap.Clause_DISTINCT, mss[1].GetClause())
	require.Equal(t, "filter_1", mss[1].GetFilter())
	require.Equal(t, "attribute_1", mss[1].GetAttribute())

	mfs := mp.GetFilters()
	require.Len(t, mfs, 2)
	// filter#0
	require.Equal(t, "filter_0", mfs[0].GetName())
	require.Zero(t, mfs[0].GetKey())
	require.Equal(t, protonetmap.Operation_AND, mfs[0].GetOp())
	require.Zero(t, mfs[0].GetValue())
	msubs := mfs[0].GetFilters()
	require.Len(t, msubs, 2)
	// sub#0
	require.Equal(t, "filter_0_0", msubs[0].GetName())
	require.Equal(t, "key_0_0", msubs[0].GetKey())
	require.Equal(t, protonetmap.Operation_EQ, msubs[0].GetOp())
	require.Equal(t, "val_0_0", msubs[0].GetValue())
	require.Zero(t, msubs[0].GetFilters())
	// sub#1
	require.Equal(t, "filter_0_1", msubs[1].GetName())
	require.Equal(t, "key_0_1", msubs[1].GetKey())
	require.Equal(t, protonetmap.Operation_NE, msubs[1].GetOp())
	require.Equal(t, "val_0_1", msubs[1].GetValue())
	require.Zero(t, msubs[1].GetFilters())
	// filter#1
	require.Equal(t, "filter_1", mfs[1].GetName())
	require.Zero(t, mfs[1].GetKey())
	require.Equal(t, protonetmap.Operation_OR, mfs[1].GetOp())
	require.Zero(t, mfs[1].GetValue())
	msubs = mfs[1].GetFilters()
	require.Len(t, msubs, 4)
	// sub#0
	require.Equal(t, "filter_1_0", msubs[0].GetName())
	require.Equal(t, "key_1_0", msubs[0].GetKey())
	require.Equal(t, protonetmap.Operation_GT, msubs[0].GetOp())
	require.Equal(t, "1889407708985023116", msubs[0].GetValue())
	require.Zero(t, msubs[0].GetFilters())
	// sub#1
	require.Equal(t, "filter_1_1", msubs[1].GetName())
	require.Equal(t, "key_1_1", msubs[1].GetKey())
	require.Equal(t, protonetmap.Operation_GE, msubs[1].GetOp())
	require.Equal(t, "1429243097315344888", msubs[1].GetValue())
	require.Zero(t, msubs[1].GetFilters())
	// sub#2
	require.Equal(t, "filter_1_2", msubs[2].GetName())
	require.Equal(t, "key_1_2", msubs[2].GetKey())
	require.Equal(t, protonetmap.Operation_LT, msubs[2].GetOp())
	require.Equal(t, "3722656060317482335", msubs[2].GetValue())
	require.Zero(t, msubs[2].GetFilters())
	// sub#3
	require.Equal(t, "filter_1_3", msubs[3].GetName())
	require.Equal(t, "key_1_3", msubs[3].GetKey())
	require.Equal(t, protonetmap.Operation_LE, msubs[3].GetOp())
	require.Equal(t, "1950504987705284805", msubs[3].GetValue())
	require.Zero(t, msubs[3].GetFilters())
}

func TestContainer_SignedData(t *testing.T) {
	require.Equal(t, validBinContainer, validContainer.SignedData())
}

func TestContainer_Marshal(t *testing.T) {
	require.Equal(t, validBinContainer, validContainer.Marshal())
}

func TestContainer_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(container.Container).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "owner/value/empty", err: "invalid owner: invalid length 0, expected 25",
				b: []byte{18, 0}},
			{name: "owner/value/undersize", err: "invalid owner: invalid length 24, expected 25",
				b: []byte{18, 26, 10, 24, 53, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113,
					133, 241, 114, 112, 11, 234}},
			{name: "owner/value/oversize", err: "invalid owner: invalid length 26, expected 25",
				b: []byte{18, 28, 10, 26, 53, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113,
					133, 241, 114, 112, 11, 234, 139, 1}},
			{name: "owner/value/wrong prefix", err: "invalid owner: invalid prefix byte 0x42, expected 0x35",
				b: []byte{18, 27, 10, 25, 66, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113,
					133, 241, 114, 112, 11, 234, 139}},
			{name: "owner/value/checksum mismatch", err: "invalid owner: checksum mismatch",
				b: []byte{18, 27, 10, 25, 53, 196, 217, 92, 234, 244, 112, 215, 190, 121, 87, 252, 143, 243, 36, 142, 40, 113,
					133, 241, 114, 112, 11, 234, 140}},
			{name: "nonce/undersize", err: "invalid nonce: invalid UUID (got 15 bytes)",
				b: []byte{26, 15, 229, 22, 237, 42, 123, 159, 78, 139, 136, 206, 237, 126, 224, 125, 147}},
			{name: "nonce/oversize", err: "invalid nonce: invalid UUID (got 17 bytes)",
				b: []byte{26, 17, 229, 22, 237, 42, 123, 159, 78, 139, 136, 206, 237, 126, 224, 125, 147, 223, 1}},
			{name: "nonce/wrong version", err: "invalid nonce: wrong UUID version 3, expected 4",
				b: []byte{26, 16, 229, 22, 237, 42, 123, 159, 48, 139, 136, 206, 237, 126, 224, 125, 147, 223}},
			{name: "policy/rules/missing", err: "invalid placement policy: missing both REP and EC rules",
				b: []byte{50, 0}},
			{name: "attributes/no key", err: "empty attribute key",
				b: []byte{42, 8, 10, 2, 107, 49, 18, 2, 118, 49, 42, 4, 18, 2, 118, 50}},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				b: []byte{42, 8, 10, 2, 107, 49, 18, 2, 118, 49, 42, 4, 10, 2, 107, 50}},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				b: []byte{42, 8, 10, 2, 107, 49, 18, 2, 118, 49, 42, 8, 10, 2, 107, 50, 18, 2, 118, 50, 42, 8, 10, 2, 107, 49,
					18, 2, 118, 51}},
			{name: "attributes/timestamp", err: "invalid attribute value Timestamp: foo (strconv.ParseInt: parsing \"foo\": invalid syntax)",
				b: []byte{42, 16, 10, 9, 84, 105, 109, 101, 115, 116, 97, 109, 112, 18, 3, 102, 111, 111}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(container.Container).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var val container.Container
	// zero
	require.NoError(t, val.Unmarshal(nil))
	require.Zero(t, val)

	// filled
	require.NoError(t, val.Unmarshal(validBinContainer))
	require.Equal(t, validContainer, val)
}

func TestContainer_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validContainer, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONContainer, string(b))
}

func TestContainer_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(container.Container).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "owner/value/empty", err: "invalid owner: invalid length 0, expected 25",
				j: `{"ownerID":{}}`},
			{name: "owner/value/undersize", err: "invalid owner: invalid length 24, expected 25",
				j: `{"ownerID":{"value":"NcTZXOr0cNe+eVf8j/MkjihxhfFycAvq"}}`},
			{name: "owner/value/oversize", err: "invalid owner: invalid length 26, expected 25",
				j: `{"ownerID":{"value":"NcTZXOr0cNe+eVf8j/MkjihxhfFycAvqiwE="}}`},
			{name: "owner/value/wrong prefix", err: "invalid owner: invalid prefix byte 0x42, expected 0x35",
				j: `{"ownerID":{"value":"QsTZXOr0cNe+eVf8j/MkjihxhfFycAvqiw=="}}`},
			{name: "owner/value/checksum mismatch", err: "invalid owner: checksum mismatch",
				j: `{"ownerID":{"value":"NcTZXOr0cNe+eVf8j/MkjihxhfFycAvqjA=="}}`},
			{name: "nonce/undersize", err: "invalid nonce: invalid UUID (got 15 bytes)",
				j: `{"nonce":"5RbtKnufTouIzu1+4H2T"}`},
			{name: "nonce/oversize", err: "invalid nonce: invalid UUID (got 17 bytes)",
				j: `{"nonce":"5RbtKnufTouIzu1+4H2T3wE="}`},
			{name: "nonce/wrong version", err: "invalid nonce: wrong UUID version 3, expected 4",
				j: `{"nonce":"5RbtKnufMIuIzu1+4H2T3w=="}`},
			{name: "policy/rules/missing", err: "invalid placement policy: missing both REP and EC rules",
				j: `{"placementPolicy":{}}`},
			{name: "attributes/no key", err: "empty attribute key",
				j: `{"attributes":[{"key":"k1","value":"v1"},{"key":"","value":"v2"}]}`},
			{name: "attributes/no value", err: `empty "k2" attribute value`,
				j: `{"attributes":[{"key":"k1", "value":"v1"}, {"key":"k2", "value":""}]}`},
			{name: "attributes/duplicated", err: "duplicated attribute k1",
				j: `{"attributes":[{"key":"k1","value":"v1"},{"key":"k2","value":"v2"},{"key":"k1","value":"v3"}]}`},
			{name: "attributes/timestamp", err: "invalid attribute value Timestamp: foo (strconv.ParseInt: parsing \"foo\": invalid syntax)",
				j: `{"attributes":[{"key":"Timestamp", "value":"foo"}]}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(container.Container).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var val container.Container
	// zero
	require.NoError(t, val.UnmarshalJSON([]byte("{}")))
	require.Zero(t, val)

	// filled
	require.NoError(t, val.UnmarshalJSON([]byte(validJSONContainer)))
	require.Equal(t, validContainer, val)
}

func TestContainer_Init(t *testing.T) {
	var val container.Container
	require.Zero(t, val.Version())
	require.Zero(t, extractNonce(t, val))

	val.Init()
	require.Equal(t, version.Current(), val.Version())
	nonce := extractNonce(t, val)
	require.EqualValues(t, 4, nonce.Version())
}

func TestContainer_SetOwner(t *testing.T) {
	var val container.Container
	require.Zero(t, val.Owner())

	val.SetOwner(anyValidOwner)
	require.Equal(t, anyValidOwner, val.Owner())

	otherOwner := usertest.OtherID(anyValidOwner)
	val.SetOwner(otherOwner)
	require.Equal(t, otherOwner, val.Owner())
}

func TestContainer_SetBasicACL(t *testing.T) {
	var val container.Container
	require.Zero(t, val.BasicACL())

	val.SetBasicACL(anyValidBasicACL)
	require.Equal(t, anyValidBasicACL, val.BasicACL())

	var otherBasicACL acl.Basic
	otherBasicACL.FromBits(anyValidBasicACL.Bits() + 1)
	val.SetBasicACL(otherBasicACL)
	require.Equal(t, otherBasicACL, val.BasicACL())
}

func TestContainer_SetPlacementPolicy(t *testing.T) {
	var val container.Container
	require.Zero(t, val.PlacementPolicy())

	val.SetPlacementPolicy(anyValidPolicy)
	require.Equal(t, anyValidPolicy, val.PlacementPolicy())

	ppOther := anyValidPolicy
	ppOther.SetContainerBackupFactor(anyValidPolicy.ContainerBackupFactor() + 1)
	val.SetPlacementPolicy(ppOther)
	require.Equal(t, ppOther, val.PlacementPolicy())
}

func TestContainer_SetAttribute(t *testing.T) {
	var val container.Container
	require.Panics(t, func() { val.SetAttribute("", "v") })
	require.Panics(t, func() { val.SetAttribute("k", "") })
	for range val.Attributes() {
		t.Fatal("handler must not be called")
	}
	for range val.UserAttributes() {
		t.Fatal("handler must not be called")
	}

	const k1, v1 = "k1", "v1"
	const sk1, sv1 = "__NEOFS__sk1", "sv1"
	const k2, v2 = "k2", "v2"
	const sk2, sv2 = "__NEOFS__sk2", "sv2"

	require.Zero(t, val.Attribute(k1))
	require.Zero(t, val.Attribute(sk1))
	require.Zero(t, val.Attribute(k2))
	require.Zero(t, val.Attribute(sk2))

	val.SetAttribute(k1, v1)
	val.SetAttribute(sk1, sv1)
	val.SetAttribute(k2, v2)
	val.SetAttribute(sk2, sv2)
	require.Equal(t, v1, val.Attribute(k1))
	require.Equal(t, sv1, val.Attribute(sk1))
	require.Equal(t, v2, val.Attribute(k2))
	require.Equal(t, sv2, val.Attribute(sk2))

	var collected []string
	for k, v := range val.Attributes() {
		collected = append(collected, k, v)
	}
	require.Equal(t, []string{
		k1, v1,
		sk1, sv1,
		k2, v2,
		sk2, sv2,
	}, collected)

	collected = nil
	for k, v := range val.UserAttributes() {
		collected = append(collected, k, v)
	}
	require.Equal(t, []string{
		k1, v1,
		k2, v2,
	}, collected)

	val.SetAttribute(k1, v1+"_other")
	require.Equal(t, v1+"_other", val.Attribute(k1))
}

func TestContainer_Attributes(t *testing.T) {
	var cnr container.Container
	for range cnr.Attributes() {
		t.Fatal("handler must not be called")
	}

	exp := []string{
		"key1", "val1",
		"key2", "val2",
		"key3", "val3",
	}
	for i := 0; i < len(exp); i += 2 {
		cnr.SetAttribute(exp[i], exp[i+1])
	}

	var got []string
	for k, v := range cnr.Attributes() {
		got = append(got, k, v)
	}
	require.Equal(t, exp, got)

	require.NotPanics(t, func() {
		for range cnr.Attributes() {
			break
		}
	})
}

func TestContainer_UserAttributes(t *testing.T) {
	var cnr container.Container
	for range cnr.UserAttributes() {
		t.Fatal("handler must not be called")
	}

	exp := []string{
		"key1", "val1",
		"key2", "val2",
		"key3", "val3",
		"__NEOFS__ANY", "val4",
	}
	for i := 0; i < len(exp); i += 2 {
		cnr.SetAttribute(exp[i], exp[i+1])
	}

	var got []string
	for k, v := range cnr.UserAttributes() {
		got = append(got, k, v)
	}
	require.Equal(t, exp[:6], got)

	require.NotPanics(t, func() {
		for range cnr.UserAttributes() {
			break
		}
	})
}

func TestContainer_SetName(t *testing.T) {
	var val container.Container
	require.Panics(t, func() { val.SetName("") })

	val.SetName(anyValidName)
	require.Equal(t, anyValidName, val.Name())

	const otherName = anyValidName + "_other"
	val.SetName(otherName)
	require.Equal(t, otherName, val.Name())
}

func TestContainer_SetCreationTime(t *testing.T) {
	var val container.Container
	require.Zero(t, val.CreatedAt().Unix())

	val.SetCreationTime(anyValidCreationTime)
	require.Equal(t, anyValidCreationTime.Unix(), val.CreatedAt().Unix())

	otherTime := anyValidCreationTime.Add(time.Hour)
	val.SetCreationTime(otherTime)
	require.Equal(t, otherTime.Unix(), val.CreatedAt().Unix())
}

func TestContainer_DisableHomomorphicHashing(t *testing.T) {
	var val container.Container
	require.False(t, val.IsHomomorphicHashingDisabled())

	val.DisableHomomorphicHashing()
	require.True(t, val.IsHomomorphicHashingDisabled())
}

func TestContainer_WriteDomain(t *testing.T) {
	var val container.Container
	require.Zero(t, val.ReadDomain())

	var d container.Domain
	d.SetName(anyValidDomainName)
	d.SetZone(anyValidDomainZone)
	val.WriteDomain(d)
	require.Equal(t, d, val.ReadDomain())

	const otherName = anyValidDomainName + "_other"
	const otherZone = anyValidDomainZone + "_other"
	var dOther container.Domain
	dOther.SetName(otherName)
	dOther.SetZone(otherZone)
	val.WriteDomain(dOther)
	require.Equal(t, dOther, val.ReadDomain())
}

func TestAssertID(t *testing.T) {
	val := containertest.Container()

	require.False(t, val.AssertID(cidtest.ID()))

	h := sha256.Sum256(val.Marshal())
	require.True(t, val.AssertID(h))
}

func TestContainer_CalculateSignature(t *testing.T) {
	t.Run("invalid signer", func(t *testing.T) {
		err := new(container.Container).CalculateSignature(new(neofscrypto.Signature), usertest.User())
		require.EqualError(t, err, "incorrect signer: expected ECDSA_DETERMINISTIC_SHA256 scheme")
	})
	t.Run("failure", func(t *testing.T) {
		err := new(container.Container).CalculateSignature(new(neofscrypto.Signature), usertest.FailSigner(usertest.User()))
		require.Error(t, err)
	})

	var sig neofscrypto.Signature
	require.NoError(t, validContainer.CalculateSignature(&sig, neofsecdsa.SignerRFC6979(anyECDSAPrivateKey)))
	require.Equal(t, validContainerSignature, sig)
}

func TestContainer_VerifySignature(t *testing.T) {
	require.False(t, container.Container{}.VerifySignature(validContainerSignature))
	require.True(t, validContainer.VerifySignature(validContainerSignature))

	sig := validContainerSignature
	sig.SetScheme(sig.Scheme() + 1)
	require.False(t, validContainer.VerifySignature(sig))

	for i := range anyBinECDSAPublicKey {
		pubCp := bytes.Clone(anyBinECDSAPublicKey)
		pubCp[i]++
		sig.SetPublicKeyBytes(pubCp)
		require.False(t, validContainer.VerifySignature(sig), i)
	}

	for i := range validContainerSignatureBytes {
		sigBytesCp := bytes.Clone(validContainerSignatureBytes)
		sigBytesCp[i]++
		sig.SetValue(sigBytesCp)
		require.False(t, validContainer.VerifySignature(sig), i)
	}
}
