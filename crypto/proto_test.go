package neofscrypto_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"math/rand/v2"
	"testing"

	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protoacl "github.com/nspcc-dev/neofs-sdk-go/proto/acl"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	protostatus "github.com/nspcc-dev/neofs-sdk-go/proto/status"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/encoding"
	eproto "google.golang.org/grpc/encoding/proto"
	"google.golang.org/protobuf/proto"
)

var corruptSigTestcases = []struct {
	name, msg string
	corrupt   func(valid *refs.Signature)
}{
	{name: "scheme/negative", msg: "negative scheme -1", corrupt: func(valid *refs.Signature) { valid.Scheme = -1 }},
	{name: "scheme/unsupported ", msg: "unsupported scheme 3", corrupt: func(valid *refs.Signature) { valid.Scheme = 3 }},
	{name: "scheme/other ", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		if valid.Scheme++; valid.Scheme >= 3 {
			valid.Scheme = 0
		}
	}},
	{name: "public key/nil", msg: "missing public key", corrupt: func(valid *refs.Signature) { valid.Key = nil }},
	{name: "public key/empty", msg: "missing public key", corrupt: func(valid *refs.Signature) { valid.Key = []byte{} }},
	{name: "public key/undersize", msg: "decode public key from binary: unexpected EOF", corrupt: func(valid *refs.Signature) {
		valid.Key = bytes.Clone(requestSignerECDSAPubBin)[:32]
	}},
	{name: "public key/oversize", msg: "decode public key from binary: extra data", corrupt: func(valid *refs.Signature) {
		valid.Key = append(bytes.Clone(requestSignerECDSAPubBin), 1)
	}},
	{name: "public key/prefix/zero", msg: "decode public key from binary: extra data", corrupt: func(valid *refs.Signature) {
		valid.Key[0] = 0x00
	}},
	{name: "public key/prefix/unsupported", msg: "decode public key from binary: invalid prefix 5", corrupt: func(valid *refs.Signature) {
		valid.Key[0] = 0x05
	}},
	{name: "public key/prefix/uncompressed in compressed form", msg: "decode public key from binary: EOF", corrupt: func(valid *refs.Signature) {
		valid.Key[0] = 0x04
	}},
	{name: "public key/prefix/other compressed", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		if valid.Key[0] == 0x02 {
			valid.Key[0] = 0x03
		} else {
			valid.Key[0] = 0x02
		}
	}},
	{name: "public key/wrong", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		valid.Key = neofscryptotest.Signer().PublicKeyBytes
	}},
	{name: "signature/nil", msg: "signature mismatch", corrupt: func(valid *refs.Signature) { valid.Sign = nil }},
	{name: "signature/empty", msg: "signature mismatch", corrupt: func(valid *refs.Signature) { valid.Sign = []byte{} }},
	{name: "signature/nil", msg: "signature mismatch", corrupt: func(valid *refs.Signature) { valid.Sign = nil }},
	{name: "signature/empty", msg: "signature mismatch", corrupt: func(valid *refs.Signature) { valid.Sign = []byte{} }},
	{name: "signature/undersize", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		valid.Sign = valid.Sign[:len(valid.Sign)-1]
	}},
	{name: "signature/oversize", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		valid.Sign = append(valid.Sign, 1)
	}},
	{name: "signature/one byte change", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
		valid.Sign[rand.IntN(len(valid.Sign))]++
	}},
	// TODO: uncomment after https://github.com/nspcc-dev/neofs-sdk-go/issues/673
	// {name: "public key/infinite", msg: "signature mismatch", corrupt: func(valid *refs.Signature) {
	// 	valid.Key = []byte{0x00}
	// }},
}

type invalidRequestVerificationHeaderTestcase = struct {
	name, msg  string
	apiVersion *refs.Version
	corrupt    func(valid *protosession.RequestVerificationHeader)
}

var (
	v225 = &refs.Version{Major: 2, Minor: 25}
	v226 = &refs.Version{Major: 2, Minor: 26}
)

// finalized in init.
var invalidOriginalRequestVerificationHeaderTestcases = []invalidRequestVerificationHeaderTestcase{
	{name: "body signature/missing", msg: "missing body signature", apiVersion: v225, corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.BodySignature = nil
	}},
	{name: "meta header signature/missing", msg: "missing meta header's signature", apiVersion: v225, corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.MetaSignature = nil
	}},
	{name: "full request signature/missing", msg: "missing request's signature", apiVersion: v226, corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.RequestSignature = nil
	}},
}

func init() {
	for _, tc := range corruptSigTestcases {
		invalidOriginalRequestVerificationHeaderTestcases = append(invalidOriginalRequestVerificationHeaderTestcases, invalidRequestVerificationHeaderTestcase{
			name: "body signature/" + tc.name, apiVersion: v225, msg: "invalid body signature: " + tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.BodySignature) },
		}, invalidRequestVerificationHeaderTestcase{
			name: "meta header signature/" + tc.name, apiVersion: v225, msg: "invalid meta header's signature: " + tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.MetaSignature) },
		}, invalidRequestVerificationHeaderTestcase{
			name: "full request header signature/" + tc.name, apiVersion: v226, msg: tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.RequestSignature) },
		})
	}
}

var (
	reqMetaHdrV225 = &protosession.RequestMetaHeader{
		Version: &refs.Version{Major: 2, Minor: 25},
		Epoch:   18426399493784435637, Ttl: 360369950,
		XHeaders: []*protosession.XHeader{
			{Key: "x-header-1-key", Value: "x-header-1-val"},
			{Key: "x-header-2-key", Value: "x-header-2-val"},
		},
		SessionToken: &protosession.SessionToken{
			Body: &protosession.SessionToken_Body{
				Id:      []byte("any_ID"),
				OwnerId: &refs.OwnerID{Value: []byte("any_session_owner")},
				Lifetime: &protosession.SessionToken_Body_TokenLifetime{
					Exp: 9296388864757340046, Nbf: 7616299382059580946, Iat: 7881369180031591601,
				},
				SessionKey: []byte("any_session_key"),
				Context: &protosession.SessionToken_Body_Object{
					Object: &protosession.ObjectSessionContext{
						Verb: 598965377,
						Target: &protosession.ObjectSessionContext_Target{
							Container: &refs.ContainerID{Value: []byte("any_target_container")},
							Objects: []*refs.ObjectID{
								{Value: []byte("any_target_object_1")},
								{Value: []byte("any_target_object_2")},
							},
						},
					},
				},
			},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 598965377},
		},
		BearerToken: &protoacl.BearerToken{
			Body: &protoacl.BearerToken_Body{
				EaclTable: &protoacl.EACLTable{
					Version:     &refs.Version{Major: 318436066, Minor: 2840436841},
					ContainerId: &refs.ContainerID{Value: []byte("any_eACL_container")},
					Records: []*protoacl.EACLRecord{
						{Operation: 1119884853, Action: 62729415, Filters: []*protoacl.EACLRecord_Filter{
							{HeaderType: 623516729, MatchType: 1738829273, Key: "filter-1-1-key", Value: "filter-1-1-val"},
							{HeaderType: 1607116959, MatchType: 1367966035, Key: "filter-1-2-key", Value: "filter-1-2-val"},
						}, Targets: []*protoacl.EACLRecord_Target{
							{Role: 611878932, Keys: [][]byte{[]byte("subj-1-1-1"), []byte("subj-1-1-2")}},
							{Role: 1862775306, Keys: [][]byte{[]byte("subj-1-2-1"), []byte("subj-1-2-2")}},
						}},
						{Operation: 1240073398, Action: 1717003574, Filters: []*protoacl.EACLRecord_Filter{
							{HeaderType: 623516729, MatchType: 1738829273, Key: "filter-2-1-key", Value: "filter-2-1-val"},
							{HeaderType: 1607116959, MatchType: 1367966035, Key: "filter-2-2-key", Value: "filter-2-2-val"},
						}, Targets: []*protoacl.EACLRecord_Target{
							{Role: 611878932, Keys: [][]byte{[]byte("subj-2-1-1"), []byte("subj-2-1-2")}},
							{Role: 1862775306, Keys: [][]byte{[]byte("subj-2-2-1"), []byte("subj-2-2-2")}},
						}},
					},
				},
				OwnerId: &refs.OwnerID{Value: []byte("any_bearer_user")},
				Lifetime: &protoacl.BearerToken_Body_TokenLifetime{
					Exp: 13260042237062625207, Nbf: 8718573876473538197, Iat: 2028326755325539864},
				Issuer: &refs.OwnerID{Value: []byte("any_bearer_issuer")},
			},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 1375722142},
		},
		MagicNumber: 14001122173143970642,
	}
	reqMetaHdrBinV225 = []byte{10, 4, 8, 2, 16, 25, 16, 181, 247, 213, 227, 229, 150, 238, 219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10, 14, 120, 45, 104,
		101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118, 97, 108, 34, 32, 10,
		14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 118, 97,
		108, 42, 188, 1, 10, 159, 1, 10, 6, 97, 110, 121, 95, 73, 68, 18, 19, 10, 17, 97, 110, 121, 95, 115, 101, 115, 115, 105, 111, 110,
		95, 111, 119, 110, 101, 114, 26, 31, 8, 142, 175, 136, 206, 176, 141, 218, 129, 129, 1, 16, 146, 252, 192, 149, 246, 253, 161, 217, 105, 24,
		177, 201, 250, 251, 176, 240, 143, 176, 109, 34, 15, 97, 110, 121, 95, 115, 101, 115, 115, 105, 111, 110, 95, 107, 101, 121, 42, 78, 8, 129,
		249, 205, 157, 2, 18, 70, 10, 22, 10, 20, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114,
		18, 21, 10, 19, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95, 111, 98, 106, 101, 99, 116, 95, 49, 18, 21, 10, 19, 97, 110, 121,
		95, 116, 97, 114, 103, 101, 116, 95, 111, 98, 106, 101, 99, 116, 95, 50, 18, 24, 10, 7, 97, 110, 121, 95, 112, 117, 98, 18, 7, 97,
		110, 121, 95, 115, 105, 103, 24, 129, 249, 205, 157, 2, 50, 226, 3, 10, 197, 3, 10, 249, 2, 10, 12, 8, 226, 229, 235, 151, 1, 16,
		233, 192, 182, 202, 10, 18, 20, 10, 18, 97, 110, 121, 95, 101, 65, 67, 76, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 26, 167, 1,
		8, 181, 172, 128, 150, 4, 16, 199, 217, 244, 29, 26, 44, 8, 185, 184, 168, 169, 2, 16, 217, 219, 145, 189, 6, 26, 14, 102, 105, 108,
		116, 101, 114, 45, 49, 45, 49, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 49, 45, 118, 97, 108, 26, 44, 8,
		159, 209, 170, 254, 5, 16, 211, 130, 166, 140, 5, 26, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 50, 45, 107, 101, 121, 34, 14, 102,
		105, 108, 116, 101, 114, 45, 49, 45, 50, 45, 118, 97, 108, 34, 30, 8, 148, 144, 226, 163, 2, 18, 10, 115, 117, 98, 106, 45, 49, 45,
		49, 45, 49, 18, 10, 115, 117, 98, 106, 45, 49, 45, 49, 45, 50, 34, 30, 8, 138, 228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45,
		49, 45, 50, 45, 49, 18, 10, 115, 117, 98, 106, 45, 49, 45, 50, 45, 50, 26, 168, 1, 8, 182, 137, 168, 207, 4, 16, 182, 202, 221,
		178, 6, 26, 44, 8, 185, 184, 168, 169, 2, 16, 217, 219, 145, 189, 6, 26, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 49, 45, 107,
		101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 49, 45, 118, 97, 108, 26, 44, 8, 159, 209, 170, 254, 5, 16, 211, 130, 166,
		140, 5, 26, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 50, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 50,
		45, 118, 97, 108, 34, 30, 8, 148, 144, 226, 163, 2, 18, 10, 115, 117, 98, 106, 45, 50, 45, 49, 45, 49, 18, 10, 115, 117, 98, 106,
		45, 50, 45, 49, 45, 50, 34, 30, 8, 138, 228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45, 50, 45, 50, 45, 49, 18, 10, 115, 117,
		98, 106, 45, 50, 45, 50, 45, 50, 18, 17, 10, 15, 97, 110, 121, 95, 98, 101, 97, 114, 101, 114, 95, 117, 115, 101, 114, 26, 31, 8,
		183, 239, 172, 246, 142, 197, 200, 130, 184, 1, 16, 149, 205, 210, 185, 246, 151, 166, 255, 120, 24, 152, 236, 229, 220, 255, 141, 132, 147, 28,
		34, 19, 10, 17, 97, 110, 121, 95, 98, 101, 97, 114, 101, 114, 95, 105, 115, 115, 117, 101, 114, 18, 24, 10, 7, 97, 110, 121, 95, 112,
		117, 98, 18, 7, 97, 110, 121, 95, 115, 105, 103, 24, 158, 181, 255, 143, 5, 64, 210, 230, 221, 152, 247, 205, 254, 166, 194, 1}

	reqMetaHdr = &protosession.RequestMetaHeader{
		Version: &refs.Version{Major: 4012726028, Minor: 3480185720},
		Epoch:   18426399493784435637, Ttl: 360369950,
		XHeaders: []*protosession.XHeader{
			{Key: "x-header-1-key", Value: "x-header-1-val"},
			{Key: "x-header-2-key", Value: "x-header-2-val"},
		},
		SessionToken: &protosession.SessionToken{
			Body: &protosession.SessionToken_Body{
				Id:      []byte("any_ID"),
				OwnerId: &refs.OwnerID{Value: []byte("any_session_owner")},
				Lifetime: &protosession.SessionToken_Body_TokenLifetime{
					Exp: 9296388864757340046, Nbf: 7616299382059580946, Iat: 7881369180031591601,
				},
				SessionKey: []byte("any_session_key"),
				Context: &protosession.SessionToken_Body_Object{
					Object: &protosession.ObjectSessionContext{
						Verb: 598965377,
						Target: &protosession.ObjectSessionContext_Target{
							Container: &refs.ContainerID{Value: []byte("any_target_container")},
							Objects: []*refs.ObjectID{
								{Value: []byte("any_target_object_1")},
								{Value: []byte("any_target_object_2")},
							},
						},
					},
				},
			},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 598965377},
		},
		BearerToken: &protoacl.BearerToken{
			Body: &protoacl.BearerToken_Body{
				EaclTable: &protoacl.EACLTable{
					Version:     &refs.Version{Major: 318436066, Minor: 2840436841},
					ContainerId: &refs.ContainerID{Value: []byte("any_eACL_container")},
					Records: []*protoacl.EACLRecord{
						{Operation: 1119884853, Action: 62729415, Filters: []*protoacl.EACLRecord_Filter{
							{HeaderType: 623516729, MatchType: 1738829273, Key: "filter-1-1-key", Value: "filter-1-1-val"},
							{HeaderType: 1607116959, MatchType: 1367966035, Key: "filter-1-2-key", Value: "filter-1-2-val"},
						}, Targets: []*protoacl.EACLRecord_Target{
							{Role: 611878932, Keys: [][]byte{[]byte("subj-1-1-1"), []byte("subj-1-1-2")}},
							{Role: 1862775306, Keys: [][]byte{[]byte("subj-1-2-1"), []byte("subj-1-2-2")}},
						}},
						{Operation: 1240073398, Action: 1717003574, Filters: []*protoacl.EACLRecord_Filter{
							{HeaderType: 623516729, MatchType: 1738829273, Key: "filter-2-1-key", Value: "filter-2-1-val"},
							{HeaderType: 1607116959, MatchType: 1367966035, Key: "filter-2-2-key", Value: "filter-2-2-val"},
						}, Targets: []*protoacl.EACLRecord_Target{
							{Role: 611878932, Keys: [][]byte{[]byte("subj-2-1-1"), []byte("subj-2-1-2")}},
							{Role: 1862775306, Keys: [][]byte{[]byte("subj-2-2-1"), []byte("subj-2-2-2")}},
						}},
					},
				},
				OwnerId: &refs.OwnerID{Value: []byte("any_bearer_user")},
				Lifetime: &protoacl.BearerToken_Body_TokenLifetime{
					Exp: 13260042237062625207, Nbf: 8718573876473538197, Iat: 2028326755325539864},
				Issuer: &refs.OwnerID{Value: []byte("any_bearer_issuer")},
			},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 1375722142},
		},
		MagicNumber: 14001122173143970642,
	}
)

var (
	// private ECDSA key used:
	//
	//     key, _ := ecdsa.ParseRawPrivateKey(elliptic.P256(), []byte{115, 251, 74, 3, 148, 247, 192, 250, 55, 137, 206, 15, 96, 123, 7, 30, 131, 109, 111, 204, 217, 16, 8, 120, 214, 46, 197, 74, 178, 162, 158, 116})
	requestSignerECDSAPubBin = []byte{2, 29, 218, 78, 114, 240, 147, 52, 92, 4, 205, 181, 108, 203, 212, 126, 115, 177,
		76, 164, 84, 5, 184, 82, 114, 239, 112, 251, 29, 206, 161, 106, 71}
	getObjectRequestBody = &protoobject.GetRequest_Body{
		Address: &refs.Address{
			ContainerId: &refs.ContainerID{Value: []byte("any_container")},
			ObjectId:    &refs.ObjectID{Value: []byte("any_object")},
		},
		Raw: true,
	}
	getObjectRequestBodyBin = []byte{10, 31, 10, 15, 10, 13, 97, 110, 121, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 18, 12, 10, 10, 97,
		110, 121, 95, 111, 98, 106, 101, 99, 116, 16, 1}
	// clone to use.
	getObjectUnsignedRequestV225 = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdrV225,
	}
	// clone to use.
	getObjectSignedRequestV225 = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdrV225,
		VerifyHeader: &protosession.RequestVerificationHeader{
			BodySignature: &refs.Signature{
				Key: bytes.Clone(requestSignerECDSAPubBin),
				Sign: []byte{4, 106, 206, 116, 145, 253, 85, 235, 105, 14, 111, 114, 90, 136, 212, 224, 184, 114, 108,
					121, 166, 194, 248, 232, 131, 99, 4, 139, 28, 57, 232, 249, 7, 144, 17, 142, 101, 173, 11, 143, 193,
					36, 146, 132, 233, 183, 208, 93, 122, 192, 107, 6, 241, 115, 29, 246, 9, 234, 141, 252, 60, 216,
					115, 57, 253},
				Scheme: refs.SignatureScheme_ECDSA_SHA512,
			},
			MetaSignature: &refs.Signature{
				Key: bytes.Clone(requestSignerECDSAPubBin),
				Sign: []byte{201, 185, 147, 183, 96, 28, 9, 31, 222, 247, 63, 145, 160, 85, 82, 110, 100, 118, 32, 122,
					67, 61, 9, 46, 48, 111, 32, 204, 97, 251, 216, 18, 208, 94, 218, 120, 169, 144, 80, 168, 17, 121,
					182, 202, 178, 16, 242, 154, 181, 188, 199, 84, 95, 7, 88, 241, 117, 59, 148, 58, 188, 148, 250, 199},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
			},
		},
	}

	// clone to use.
	getObjectUnsignedRequest = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdr,
	}
	// clone to use.
	getObjectSignedRequest = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdr,
		VerifyHeader: &protosession.RequestVerificationHeader{
			RequestSignature: &refs.Signature{
				Key: bytes.Clone(requestSignerECDSAPubBin),
				Sign: []byte{208, 214, 230, 27, 94, 165, 84, 242, 123, 94, 75, 196, 247, 85, 109, 188, 98, 148, 45, 242,
					172, 5, 179, 169, 97, 76, 0, 95, 207, 87, 110, 203, 3, 190, 84, 75, 164, 25, 247, 99, 144, 203, 182,
					84, 16, 27, 176, 139, 76, 160, 23, 193, 254, 92, 93, 210, 12, 247, 232, 74, 43, 60, 69, 210},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
			},
		},
	}
)

func TestEncodeRequest(t *testing.T) {
	getR := proto.Clone(getObjectUnsignedRequest).(*protoobject.GetRequest)

	c := encoding.GetCodecV2(eproto.Name)
	bufs, err := c.Marshal(getR)
	require.NoError(t, err)

	var (
		protoLen = bufs.Len()
		protoRes = make([]byte, bufs.Len())
	)
	bufs.CopyTo(protoRes)

	var (
		buf                   = make([]byte, protoLen)
		packageRes, packageLn = neofsproto.EncodeRequest(buf, getR.GetBody(), getR.GetMetaHeader())
	)

	require.Equal(t, protoLen, packageLn)
	require.EqualValues(t, protoRes, packageRes[:packageLn])
}

func TestSignRequestWithBuffer(t *testing.T) {
	anySigner := neofscryptotest.Signer()
	pub := &anySigner.ECDSAPrivateKey.PublicKey
	checkSignerCreds := func(scheme neofscrypto.Scheme, sigs ...*refs.Signature) {
		for i, sig := range sigs {
			require.NotNil(t, sig, i)
			require.EqualValues(t, scheme, sig.Scheme, i)
			require.Equal(t, anySigner.PublicKeyBytes, sig.Key, i)
		}
	}

	t.Run("signer failure", func(t *testing.T) {
		for i, part := range []string{
			"body",
			"meta header",
			"verification header's origin",
		} {
			t.Run(part, func(t *testing.T) {
				var req protoobject.GetRequest
				signer := newNFailedSigner(anySigner, i+1)
				_, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](signer, &req, nil)
				require.ErrorContains(t, err, "sign "+part+":")
			})
		}
	})

	for _, multipleSigs := range []bool{true, false} {
		t.Run(fmt.Sprintf("multiple request signature=%t", multipleSigs), func(t *testing.T) {
			var req *protoobject.GetRequest
			if multipleSigs {
				req = getObjectUnsignedRequestV225
			} else {
				req = getObjectUnsignedRequest
			}

			for _, tc := range []struct {
				name       string
				signer     neofscrypto.Signer
				hashFunc   func([]byte) []byte
				verifyFunc func(t testing.TB, pub *ecdsa.PublicKey, hash, sig []byte)
			}{
				{
					name:       "ECDSA_SHA512",
					signer:     anySigner,
					hashFunc:   func(b []byte) []byte { h := sha512.Sum512(b); return h[:] },
					verifyFunc: verifyECDSAWithSHA512Signature,
				},
				{
					name:       "ECDSA_SHA256_RFC6979",
					signer:     anySigner.RFC6979,
					hashFunc:   func(b []byte) []byte { h := sha256.Sum256(b); return h[:] },
					verifyFunc: verifyECDSAWithSHA256RFC6979Signature,
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					req := proto.Clone(req).(*protoobject.GetRequest)

					vh, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](tc.signer, req, nil)
					require.NoError(t, err)
					require.NotNil(t, vh)
					require.Nil(t, vh.Origin)

					if multipleSigs {
						checkSignerCreds(tc.signer.Scheme(), vh.BodySignature, vh.MetaSignature)
						tc.verifyFunc(t, pub, tc.hashFunc(getObjectRequestBodyBin), vh.BodySignature.Sign)
						tc.verifyFunc(t, pub, tc.hashFunc(reqMetaHdrBinV225), vh.MetaSignature.Sign)
					} else {
						checkSignerCreds(tc.signer.Scheme(), vh.RequestSignature)
						reqMarshaled, _ := neofsproto.EncodeRequest(nil, req.GetBody(), req.GetMetaHeader())
						tc.verifyFunc(t, pub, tc.hashFunc(reqMarshaled), vh.RequestSignature.Sign)
					}

					req.VerifyHeader = vh
					err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
					require.NoError(t, err)
				})
			}
			t.Run("ECDSA_SHA256_WalletConnect", func(t *testing.T) {
				req := proto.Clone(req).(*protoobject.GetRequest)

				vh, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](anySigner.WalletConnect, req, nil)
				require.NoError(t, err)
				require.NotNil(t, vh)
				require.Nil(t, vh.Origin)

				if multipleSigs {
					checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vh.BodySignature, vh.MetaSignature)
					verifyWalletConnectSignature(t, pub, getObjectRequestBodyBin, vh.BodySignature.Sign)
					verifyWalletConnectSignature(t, pub, reqMetaHdrBinV225, vh.MetaSignature.Sign)
				} else {
					checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vh.RequestSignature)
					reqMarshaled, _ := neofsproto.EncodeRequest(nil, req.GetBody(), req.GetMetaHeader())
					verifyWalletConnectSignature(t, pub, reqMarshaled, vh.RequestSignature.Sign)
				}

				req.VerifyHeader = vh
				err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
				require.NoError(t, err)
			})
		})
	}
}

func TestVerifyRequestWithBuffer(t *testing.T) {
	t.Run("correctly signed", func(t *testing.T) {
		err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](getObjectSignedRequest, nil)
		require.NoError(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			t.Run("untyped", func(t *testing.T) {
				require.Panics(t, func() {
					_ = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](nil, nil)
				})
			})
			t.Run("typed", func(t *testing.T) {
				err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body]((*protoobject.GetRequest)(nil), nil)
				require.EqualError(t, err, "missing verification header")
			})
		})
		t.Run("without verification header", func(t *testing.T) {
			req := proto.Clone(getObjectSignedRequest).(*protoobject.GetRequest)
			req.VerifyHeader = nil
			err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
			require.EqualError(t, err, "missing verification header")
		})
		for _, tc := range invalidOriginalRequestVerificationHeaderTestcases {
			t.Run(tc.name, func(t *testing.T) {
				var req *protoobject.GetRequest
				switch tc.apiVersion {
				case v225:
					req = proto.Clone(getObjectSignedRequestV225).(*protoobject.GetRequest)
					tc.corrupt(req.VerifyHeader)
					err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
					require.EqualError(t, err, "invalid verification header at depth 0: "+tc.msg)
				case v226:
					req = proto.Clone(getObjectSignedRequest).(*protoobject.GetRequest)
					tc.corrupt(req.VerifyHeader)
					err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
					require.EqualError(t, err, tc.msg)
				default:
					t.Fatalf("unknown test API version %s", tc.apiVersion.String())
				}
			})
		}
		t.Run("resigned", func(t *testing.T) {
			for _, tc := range []struct {
				name, msg string
				corrupt   func(valid *protoobject.GetRequest)
			}{
				{name: "redundant verification header", msg: "invalid verification header at depth 0: missing meta header's signature",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.VerifyHeader = &protosession.RequestVerificationHeader{Origin: valid.VerifyHeader}
					},
				},
				{name: "lacking verification header", msg: "incorrect number of verification headers",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.MetaHeader = &protosession.RequestMetaHeader{Origin: valid.MetaHeader}
					},
				},
				{name: "with body signature", msg: "invalid verification header at depth 0: invalid body signature: missing public key",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.VerifyHeader.BodySignature = new(refs.Signature)
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					req := proto.Clone(getObjectSignedRequestV225).(*protoobject.GetRequest)
					tc.corrupt(req)
					err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
					require.EqualError(t, err, tc.msg)
				})
			}
		})
	})
}

type invalidResponseVerificationHeaderTestcase = struct {
	name, msg string
	corrupt   func(valid *protosession.ResponseVerificationHeader)
}

// finalized in init.
var invalidOriginalResponseVerificationHeaderTestcases = []invalidResponseVerificationHeaderTestcase{
	{name: "body signature/missing", msg: "missing body signature", corrupt: func(valid *protosession.ResponseVerificationHeader) {
		valid.BodySignature = nil
	}},
	{name: "meta header signature/missing", msg: "missing meta header's signature", corrupt: func(valid *protosession.ResponseVerificationHeader) {
		valid.MetaSignature = nil
	}},
	{name: "verification header's origin signature/missing", msg: "missing verification header's origin signature", corrupt: func(valid *protosession.ResponseVerificationHeader) {
		valid.OriginSignature = nil
	}},
}

func init() {
	for _, tc := range corruptSigTestcases {
		invalidOriginalResponseVerificationHeaderTestcases = append(invalidOriginalResponseVerificationHeaderTestcases, invalidResponseVerificationHeaderTestcase{
			name: "body signature/" + tc.name, msg: "invalid body signature: " + tc.msg,
			corrupt: func(valid *protosession.ResponseVerificationHeader) { tc.corrupt(valid.BodySignature) },
		}, invalidResponseVerificationHeaderTestcase{
			name: "meta header signature/" + tc.name, msg: "invalid meta header's signature: " + tc.msg,
			corrupt: func(valid *protosession.ResponseVerificationHeader) { tc.corrupt(valid.MetaSignature) },
		}, invalidResponseVerificationHeaderTestcase{
			name: "verification header's origin signature/" + tc.name, msg: "invalid verification header's origin signature: " + tc.msg,
			corrupt: func(valid *protosession.ResponseVerificationHeader) { tc.corrupt(valid.OriginSignature) },
		})
	}
}

var (
	respMetaHdr = &protosession.ResponseMetaHeader{
		Version: &refs.Version{Major: 4012726028, Minor: 3480185720},
		Epoch:   18426399493784435637,
		Ttl:     360369950,
		XHeaders: []*protosession.XHeader{
			{Key: "x-header-1-key", Value: "x-header-1-val"},
			{Key: "x-header-2-key", Value: "x-header-2-val"},
		},
		Status: &protostatus.Status{
			Code:    2013711884,
			Message: "any status message",
			Details: []*protostatus.Status_Detail{
				{Id: 673818269, Value: []byte("detail_1")},
				{Id: 1795152762, Value: []byte("detail_2")},
			},
		},
	}
	respMetaHdrBin = []byte{10, 12, 8, 140, 174, 181, 249, 14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238,
		219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14,
		120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118, 97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45,
		107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 118, 97, 108, 50, 62, 8, 140, 156, 155, 192, 7, 18, 18, 97,
		110, 121, 32, 115, 116, 97, 116, 117, 115, 32, 109, 101, 115, 115, 97, 103, 101, 26, 16, 8, 157, 205, 166, 193, 2, 18, 8, 100, 101,
		116, 97, 105, 108, 95, 49, 26, 16, 8, 250, 182, 255, 215, 6, 18, 8, 100, 101, 116, 97, 105, 108, 95, 50}

	respMetaHdrL2 = &protosession.ResponseMetaHeader{
		Version: &refs.Version{Major: 4012726028, Minor: 3480185720},
		Epoch:   18426399493784435637,
		Ttl:     360369950,
		XHeaders: []*protosession.XHeader{
			{Key: "x-header-1-key", Value: "x-header-1-val"},
			{Key: "x-header-2-key", Value: "x-header-2-val"},
		},
		Origin: respMetaHdr,
		Status: &protostatus.Status{
			Code:    1472978490,
			Message: "any status message",
			Details: []*protostatus.Status_Detail{
				{Id: 542687564, Value: []byte("detail_1")},
				{Id: 789115882, Value: []byte("detail_2")},
			},
		},
	}
	respMetaHdrL2Bin = []byte{10, 12, 8, 140, 174, 181, 249, 14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238,
		219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14,
		120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118, 97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45,
		107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 118, 97, 108, 42, 163, 1, 10, 12, 8, 140, 174, 181, 249,
		14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238, 219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10,
		14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118,
		97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101,
		114, 45, 50, 45, 118, 97, 108, 50, 62, 8, 140, 156, 155, 192, 7, 18, 18, 97, 110, 121, 32, 115, 116, 97, 116, 117, 115, 32, 109, 101,
		115, 115, 97, 103, 101, 26, 16, 8, 157, 205, 166, 193, 2, 18, 8, 100, 101, 116, 97, 105, 108, 95, 49, 26, 16, 8, 250, 182, 255,
		215, 6, 18, 8, 100, 101, 116, 97, 105, 108, 95, 50, 50, 62, 8, 186, 188, 175, 190, 5, 18, 18, 97, 110, 121, 32, 115, 116, 97, 116,
		117, 115, 32, 109, 101, 115, 115, 97, 103, 101, 26, 16, 8, 204, 130, 227, 130, 2, 18, 8, 100, 101, 116, 97, 105, 108, 95, 49, 26,
		16, 8, 234, 231, 163, 248, 2, 18, 8, 100, 101, 116, 97, 105, 108, 95, 50}
)

var (
	responseSignerECDSAPubBin = []byte{2, 233, 67, 160, 254, 231, 98, 137, 171, 220, 101, 138, 15, 186, 53, 234, 17, 18, 38, 245,
		80, 107, 40, 37, 164, 156, 142, 103, 157, 13, 253, 251, 6}
	responseSignerL2ECDSAPubBin = []byte{3, 154, 201, 144, 52, 75, 150, 123, 180, 230, 46, 67, 182, 66, 134, 3, 8, 227, 139, 137, 41,
		117, 235, 244, 250, 191, 92, 36, 38, 101, 142, 96, 47}
	getObjectResponseBody = &protoobject.GetResponse_Body{
		ObjectPart: &protoobject.GetResponse_Body_Init_{Init: &protoobject.GetResponse_Body_Init{
			ObjectId:  &refs.ObjectID{Value: []byte("any_ID")},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 2128773493},
			Header: &protoobject.Header{
				Version:       &refs.Version{Major: 1559619596, Minor: 436551331},
				ContainerId:   &refs.ContainerID{Value: []byte("any_container")},
				OwnerId:       &refs.OwnerID{Value: []byte("any_owner")},
				CreationEpoch: 10561284447300915844,
				PayloadLength: 766049361057238504,
			},
		}},
	}
	getObjectResponseBodyBin = []byte{10, 103, 10, 8, 10, 6, 97, 110, 121, 95, 73, 68, 18, 24, 10, 7, 97, 110, 121, 95, 112, 117, 98, 18,
		7, 97, 110, 121, 95, 115, 105, 103, 24, 245, 130, 138, 247, 7, 26, 65, 10, 12, 8, 140, 208, 215, 231, 5, 16, 163, 253, 148,
		208, 1, 18, 15, 10, 13, 97, 110, 121, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 26, 11, 10, 9, 97, 110, 121, 95, 111, 119, 110, 101,
		114, 32, 132, 165, 234, 233, 250, 135, 206, 200, 146, 1, 40, 232, 155, 237, 241, 220, 186, 227, 208, 10}
	// clone to use.
	getObjectUnsignedResponse = &protoobject.GetResponse{
		Body:       getObjectResponseBody,
		MetaHeader: respMetaHdr,
	}
	// clone to use.
	getObjectSignedResponse = &protoobject.GetResponse{
		Body:       getObjectResponseBody,
		MetaHeader: respMetaHdrL2,
		VerifyHeader: &protosession.ResponseVerificationHeader{
			BodySignature: nil,
			MetaSignature: &refs.Signature{
				Key: bytes.Clone(responseSignerL2ECDSAPubBin),
				Sign: []byte{163, 138, 107, 57, 226, 203, 104, 22, 98, 98, 154, 169, 227, 112, 3, 55, 162, 221, 244, 199, 195,
					216, 209, 202, 212, 243, 50, 72, 182, 18, 127, 57, 37, 49, 78, 5, 106, 149, 146, 166, 55, 44, 33, 68, 9, 60,
					65, 169, 33, 187, 65, 162, 142, 150, 252, 118, 125, 74, 248, 34, 78, 7, 173, 240},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
			},
			OriginSignature: &refs.Signature{
				Key: bytes.Clone(responseSignerL2ECDSAPubBin),
				Sign: []byte{35, 20, 219, 207, 205, 109, 68, 60, 253, 133, 135, 95, 96, 89, 130, 130, 166, 245, 61, 9, 119, 6, 155,
					185, 203, 202, 213, 19, 81, 248, 139, 17, 95, 180, 242, 115, 169, 254, 213, 162, 235, 166, 147, 69, 207, 221,
					32, 124, 246, 203, 254, 238, 152, 255, 162, 137, 1, 19, 51, 197, 43, 8, 61, 53, 203, 66, 71, 251, 161, 112, 24,
					55, 193, 198, 128, 208, 134, 151, 147, 79},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
			},
			Origin: &protosession.ResponseVerificationHeader{
				BodySignature: &refs.Signature{
					Key: bytes.Clone(responseSignerECDSAPubBin),
					Sign: []byte{4, 47, 78, 194, 50, 74, 38, 226, 116, 92, 209, 84, 150, 183, 182, 60, 89, 137, 211, 166, 28, 6,
						69, 228, 234, 249, 76, 229, 35, 189, 132, 18, 113, 55, 20, 148, 119, 161, 251, 206, 198, 13, 235, 106, 107,
						55, 61, 181, 42, 253, 212, 180, 57, 102, 139, 79, 194, 182, 148, 182, 8, 90, 153, 62, 21},
					Scheme: refs.SignatureScheme_ECDSA_SHA512,
				},
				MetaSignature: &refs.Signature{
					Key: bytes.Clone(responseSignerECDSAPubBin),
					Sign: []byte{194, 115, 78, 219, 234, 44, 29, 128, 18, 143, 78, 19, 10, 93, 38, 153, 190, 184, 145, 114, 36, 45,
						60, 89, 106, 245, 247, 129, 125, 156, 102, 143, 200, 55, 66, 203, 106, 47, 145, 53, 40, 161, 152, 35, 23,
						22, 31, 155, 178, 6, 195, 243, 249, 70, 220, 117, 127, 172, 232, 216, 214, 255, 126, 218},
					Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
				},
				OriginSignature: &refs.Signature{
					Key: bytes.Clone(responseSignerECDSAPubBin),
					Sign: []byte{64, 177, 241, 85, 198, 123, 114, 71, 253, 169, 228, 142, 139, 152, 102, 62, 51, 51, 124, 38, 184,
						105, 50, 147, 175, 126, 186, 191, 40, 60, 105, 76, 198, 104, 219, 130, 45, 27, 116, 43, 185, 193, 159, 63, 216,
						46, 140, 26, 149, 219, 236, 188, 19, 136, 32, 12, 102, 207, 87, 38, 159, 57, 85, 38, 175, 41, 150, 171, 42,
						233, 67, 111, 218, 149, 90, 74, 159, 142, 26, 211},
					Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
				},
			},
		},
	}
)

func TestSignResponseWithBuffer(t *testing.T) {
	anySigner := neofscryptotest.Signer()
	pub := &anySigner.ECDSAPrivateKey.PublicKey
	checkSignerCreds := func(scheme neofscrypto.Scheme, sigs ...*refs.Signature) {
		for i, sig := range sigs {
			require.NotNil(t, sig, i)
			require.EqualValues(t, scheme, sig.Scheme, i)
			require.Equal(t, anySigner.PublicKeyBytes, sig.Key, i)
		}
	}

	t.Run("signer failure", func(t *testing.T) {
		for i, part := range []string{
			"body",
			"meta header",
			"verification header's origin",
		} {
			t.Run(part, func(t *testing.T) {
				var req protoobject.GetResponse
				signer := newNFailedSigner(anySigner, i+1)
				_, err := neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](signer, &req, nil)
				require.ErrorContains(t, err, "sign "+part+":")
			})
		}
	})

	for _, tc := range []struct {
		name       string
		signer     neofscrypto.Signer
		hashFunc   func([]byte) []byte
		verifyFunc func(t testing.TB, pub *ecdsa.PublicKey, hash, sig []byte)
	}{
		{
			name:       "ECDSA_SHA512",
			signer:     anySigner,
			hashFunc:   func(b []byte) []byte { h := sha512.Sum512(b); return h[:] },
			verifyFunc: verifyECDSAWithSHA512Signature,
		},
		{
			name:       "ECDSA_SHA256_RFC6979",
			signer:     anySigner.RFC6979,
			hashFunc:   func(b []byte) []byte { h := sha256.Sum256(b); return h[:] },
			verifyFunc: verifyECDSAWithSHA256RFC6979Signature,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := proto.Clone(getObjectUnsignedResponse).(*protoobject.GetResponse)

			vh, err := neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](tc.signer, r, nil)
			require.NoError(t, err)
			require.NotNil(t, vh)
			require.Nil(t, vh.Origin)

			checkSignerCreds(tc.signer.Scheme(), vh.BodySignature, vh.MetaSignature, vh.OriginSignature)

			tc.verifyFunc(t, pub, tc.hashFunc(getObjectResponseBodyBin), vh.BodySignature.Sign)
			tc.verifyFunc(t, pub, tc.hashFunc(respMetaHdrBin), vh.MetaSignature.Sign)
			tc.verifyFunc(t, pub, tc.hashFunc(nil), vh.OriginSignature.Sign)

			r.VerifyHeader = vh
			err = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
			require.NoError(t, err)

			t.Run("re-sign", func(t *testing.T) {
				r.MetaHeader = respMetaHdrL2

				vhL2, err := neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](tc.signer, r, nil)
				require.NoError(t, err)
				require.NotNil(t, vhL2)
				require.True(t, vhL2.Origin == vh) // as pointers

				checkSignerCreds(tc.signer.Scheme(), vhL2.MetaSignature, vhL2.OriginSignature)

				require.Nil(t, vhL2.BodySignature)
				tc.verifyFunc(t, pub, tc.hashFunc(respMetaHdrL2Bin), vhL2.MetaSignature.Sign)
				originHash := tc.hashFunc(neofsproto.MarshalMessage(vh))
				tc.verifyFunc(t, pub, originHash, vhL2.OriginSignature.Sign)

				r.VerifyHeader = vhL2
				err = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
				require.NoError(t, err)
			})
		})
	}
	t.Run("ECDSA_SHA256_WalletConnect", func(t *testing.T) {
		r := proto.Clone(getObjectUnsignedResponse).(*protoobject.GetResponse)

		vh, err := neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](anySigner.WalletConnect, r, nil)
		require.NoError(t, err)
		require.NotNil(t, vh)
		require.Nil(t, vh.Origin)

		checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vh.BodySignature, vh.MetaSignature, vh.OriginSignature)

		verifyWalletConnectSignature(t, pub, getObjectResponseBodyBin, vh.BodySignature.Sign)
		verifyWalletConnectSignature(t, pub, respMetaHdrBin, vh.MetaSignature.Sign)
		verifyWalletConnectSignature(t, pub, nil, vh.OriginSignature.Sign)

		r.VerifyHeader = vh
		err = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
		require.NoError(t, err)

		t.Run("re-sign", func(t *testing.T) {
			r.MetaHeader = respMetaHdrL2

			vhL2, err := neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](anySigner.WalletConnect, r, nil)
			require.NoError(t, err)
			require.NotNil(t, vhL2)
			require.True(t, vhL2.Origin == vh) // as pointers

			checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vhL2.MetaSignature, vhL2.OriginSignature)

			require.Nil(t, vhL2.BodySignature)
			verifyWalletConnectSignature(t, pub, respMetaHdrL2Bin, vhL2.MetaSignature.Sign)
			verifyWalletConnectSignature(t, pub, neofsproto.MarshalMessage(vh.Origin), vh.OriginSignature.Sign)

			r.VerifyHeader = vhL2
			err = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
			require.NoError(t, err)
		})
	})
}

func TestVerifyResponseWithBuffer(t *testing.T) {
	t.Run("correctly signed", func(t *testing.T) {
		err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](getObjectSignedResponse, nil)
		require.NoError(t, err)
	})
	t.Run("invalid", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			t.Run("untyped", func(t *testing.T) {
				require.Panics(t, func() {
					_ = neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](nil, nil)
				})
			})
			t.Run("typed", func(t *testing.T) {
				err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body]((*protoobject.GetResponse)(nil), nil)
				require.EqualError(t, err, "missing verification header")
			})
		})
		t.Run("without verification header", func(t *testing.T) {
			r := proto.Clone(getObjectSignedResponse).(*protoobject.GetResponse)
			r.VerifyHeader = nil
			err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
			require.EqualError(t, err, "missing verification header")
		})
		for _, tc := range invalidOriginalResponseVerificationHeaderTestcases {
			t.Run(tc.name, func(t *testing.T) {
				r := proto.Clone(getObjectSignedResponse).(*protoobject.GetResponse)
				r.MetaHeader = r.MetaHeader.Origin
				r.VerifyHeader = r.VerifyHeader.Origin
				tc.corrupt(r.VerifyHeader)
				err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
				require.EqualError(t, err, "invalid verification header at depth 0: "+tc.msg)

				t.Run("resigned", func(t *testing.T) {
					resp := &protoobject.GetResponse{
						Body:         r.Body,
						MetaHeader:   &protosession.ResponseMetaHeader{Origin: r.MetaHeader},
						VerifyHeader: r.VerifyHeader,
					}
					resp.VerifyHeader, err = neofscrypto.SignResponseWithBuffer[*protoobject.GetResponse_Body](neofscryptotest.Signer(), resp, nil)
					require.NoError(t, err)

					err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](resp, nil)
					require.EqualError(t, err, "invalid verification header at depth 1: "+tc.msg)
				})
			})
		}
		t.Run("resigned", func(t *testing.T) {
			for _, tc := range []struct {
				name, msg string
				corrupt   func(valid *protoobject.GetResponse)
			}{
				{name: "redundant verification header", msg: "incorrect number of verification headers",
					corrupt: func(valid *protoobject.GetResponse) {
						valid.VerifyHeader = &protosession.ResponseVerificationHeader{Origin: valid.VerifyHeader}
					},
				},
				{name: "lacking verification header", msg: "incorrect number of verification headers",
					corrupt: func(valid *protoobject.GetResponse) {
						valid.MetaHeader = &protosession.ResponseMetaHeader{Origin: valid.MetaHeader}
					},
				},
				{name: "with body signature", msg: "invalid verification header at depth 0: body signature is set in non-origin verification header",
					corrupt: func(valid *protoobject.GetResponse) {
						valid.VerifyHeader.BodySignature = new(refs.Signature)
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					r := proto.Clone(getObjectSignedResponse).(*protoobject.GetResponse)
					tc.corrupt(r)
					err := neofscrypto.VerifyResponseWithBuffer[*protoobject.GetResponse_Body](r, nil)
					require.EqualError(t, err, tc.msg)
				})
			}
		})
	})
}
