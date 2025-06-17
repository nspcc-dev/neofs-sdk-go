package neofscrypto_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/sha512"
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
	name, msg string
	corrupt   func(valid *protosession.RequestVerificationHeader)
}

// finalized in init.
var invalidOriginalRequestVerificationHeaderTestcases = []invalidRequestVerificationHeaderTestcase{
	{name: "body signature/missing", msg: "missing body signature", corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.BodySignature = nil
	}},
	{name: "meta header signature/missing", msg: "missing meta header's signature", corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.MetaSignature = nil
	}},
	{name: "verification header's origin signature/missing", msg: "missing verification header's origin signature", corrupt: func(valid *protosession.RequestVerificationHeader) {
		valid.OriginSignature = nil
	}},
}

func init() {
	for _, tc := range corruptSigTestcases {
		invalidOriginalRequestVerificationHeaderTestcases = append(invalidOriginalRequestVerificationHeaderTestcases, invalidRequestVerificationHeaderTestcase{
			name: "body signature/" + tc.name, msg: "invalid body signature: " + tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.BodySignature) },
		}, invalidRequestVerificationHeaderTestcase{
			name: "meta header signature/" + tc.name, msg: "invalid meta header's signature: " + tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.MetaSignature) },
		}, invalidRequestVerificationHeaderTestcase{
			name: "verification header's origin signature/" + tc.name, msg: "invalid verification header's origin signature: " + tc.msg,
			corrupt: func(valid *protosession.RequestVerificationHeader) { tc.corrupt(valid.OriginSignature) },
		})
	}
}

var (
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
	reqMetaHdrBin = []byte{10, 12, 8, 140, 174, 181, 249, 14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238, 219,
		255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14, 120,
		45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118, 97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 107,
		101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 118, 97, 108, 42, 188, 1, 10, 159, 1, 10, 6, 97, 110, 121, 95,
		73, 68, 18, 19, 10, 17, 97, 110, 121, 95, 115, 101, 115, 115, 105, 111, 110, 95, 111, 119, 110, 101, 114, 26, 31, 8, 142, 175, 136, 206,
		176, 141, 218, 129, 129, 1, 16, 146, 252, 192, 149, 246, 253, 161, 217, 105, 24, 177, 201, 250, 251, 176, 240, 143, 176, 109,
		34, 15, 97, 110, 121, 95, 115, 101, 115, 115, 105, 111, 110, 95, 107, 101, 121, 42, 78, 8, 129, 249, 205, 157, 2, 18, 70, 10, 22,
		10, 20, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 18, 21, 10, 19, 97, 110, 121, 95,
		116, 97, 114, 103, 101, 116, 95, 111, 98, 106, 101, 99, 116, 95, 49, 18, 21, 10, 19, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95,
		111, 98, 106, 101, 99, 116, 95, 50, 18, 24, 10, 7, 97, 110, 121, 95, 112, 117, 98, 18, 7, 97, 110, 121, 95, 115, 105, 103, 24, 129,
		249, 205, 157, 2, 50, 226, 3, 10, 197, 3, 10, 249, 2, 10, 12, 8, 226, 229, 235, 151, 1, 16, 233, 192, 182, 202, 10, 18,
		20, 10, 18, 97, 110, 121, 95, 101, 65, 67, 76, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 26, 167, 1, 8, 181, 172, 128, 150, 4,
		16, 199, 217, 244, 29, 26, 44, 8, 185, 184, 168, 169, 2, 16, 217, 219, 145, 189, 6, 26, 14, 102, 105, 108, 116, 101, 114, 45, 49,
		45, 49, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 49, 45, 118, 97, 108, 26, 44, 8, 159, 209, 170, 254,
		5, 16, 211, 130, 166, 140, 5, 26, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 50, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116,
		101, 114, 45, 49, 45, 50, 45, 118, 97, 108, 34, 30, 8, 148, 144, 226, 163, 2, 18, 10, 115, 117, 98, 106, 45, 49, 45, 49, 45, 49,
		18, 10, 115, 117, 98, 106, 45, 49, 45, 49, 45, 50, 34, 30, 8, 138, 228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45, 49, 45, 50,
		45, 49, 18, 10, 115, 117, 98, 106, 45, 49, 45, 50, 45, 50, 26, 168, 1, 8, 182, 137, 168, 207, 4, 16, 182, 202, 221, 178, 6, 26,
		44, 8, 185, 184, 168, 169, 2, 16, 217, 219, 145, 189, 6, 26, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 49, 45, 107, 101, 121,
		34, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 49, 45, 118, 97, 108, 26, 44, 8, 159, 209, 170, 254, 5, 16, 211, 130, 166, 140,
		5, 26, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 50, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 50,
		45, 118, 97, 108, 34, 30, 8, 148, 144, 226, 163, 2, 18, 10, 115, 117, 98, 106, 45, 50, 45, 49, 45, 49, 18, 10, 115, 117, 98, 106,
		45, 50, 45, 49, 45, 50, 34, 30, 8, 138, 228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45, 50, 45, 50, 45, 49, 18, 10, 115, 117,
		98, 106, 45, 50, 45, 50, 45, 50, 18, 17, 10, 15, 97, 110, 121, 95, 98, 101, 97, 114, 101, 114, 95, 117, 115, 101, 114, 26, 31, 8, 183,
		239, 172, 246, 142, 197, 200, 130, 184, 1, 16, 149, 205, 210, 185, 246, 151, 166, 255, 120, 24, 152, 236, 229, 220, 255,
		141, 132, 147, 28, 34, 19, 10, 17, 97, 110, 121, 95, 98, 101, 97, 114, 101, 114, 95, 105, 115, 115, 117, 101, 114, 18, 24, 10, 7, 97,
		110, 121, 95, 112, 117, 98, 18, 7, 97, 110, 121, 95, 115, 105, 103, 24, 158, 181, 255, 143, 5, 64, 210, 230, 221, 152, 247, 205,
		254, 166, 194, 1}

	reqMetaHdrL2 = &protosession.RequestMetaHeader{
		Version: &refs.Version{Major: 4012726028, Minor: 3480185720},
		Epoch:   18426399493784435637, Ttl: 360369950,
		XHeaders: []*protosession.XHeader{
			{Key: "x-header-1-key", Value: "x-header-1-val"},
			{Key: "x-header-2-key", Value: "x-header-2-val"},
		},
		// tokens unset to reduce the code, they are checked at L1
		Origin:      reqMetaHdr,
		MagicNumber: 14001122173143970642,
	}
	reqMetaHdrL2Bin = []byte{10, 12, 8, 140, 174, 181, 249, 14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238,
		219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14,
		120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118, 97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45,
		107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 118, 97, 108, 58, 146, 6, 10, 12, 8, 140, 174, 181, 249,
		14, 16, 248, 214, 189, 251, 12, 16, 181, 247, 213, 227, 229, 150, 238, 219, 255, 1, 24, 158, 158, 235, 171, 1, 34, 32, 10,
		14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 49, 45, 118,
		97, 108, 34, 32, 10, 14, 120, 45, 104, 101, 97, 100, 101, 114, 45, 50, 45, 107, 101, 121, 18, 14, 120, 45, 104, 101, 97, 100, 101,
		114, 45, 50, 45, 118, 97, 108, 42, 188, 1, 10, 159, 1, 10, 6, 97, 110, 121, 95, 73, 68, 18, 19, 10, 17, 97, 110, 121, 95, 115, 101,
		115, 115, 105, 111, 110, 95, 111, 119, 110, 101, 114, 26, 31, 8, 142, 175, 136, 206, 176, 141, 218, 129, 129, 1, 16, 146, 252, 192,
		149, 246, 253, 161, 217, 105, 24, 177, 201, 250, 251, 176, 240, 143, 176, 109, 34, 15, 97, 110, 121, 95, 115, 101, 115, 115, 105,
		111, 110, 95, 107, 101, 121, 42, 78, 8, 129, 249, 205, 157, 2, 18, 70, 10, 22, 10, 20, 97, 110, 121, 95, 116, 97, 114, 103, 101,
		116, 95, 99, 111, 110, 116, 97, 105, 110, 101, 114, 18, 21, 10, 19, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95, 111, 98, 106, 101,
		99, 116, 95, 49, 18, 21, 10, 19, 97, 110, 121, 95, 116, 97, 114, 103, 101, 116, 95, 111, 98, 106, 101, 99, 116, 95, 50, 18, 24, 10, 7,
		97, 110, 121, 95, 112, 117, 98, 18, 7, 97, 110, 121, 95, 115, 105, 103, 24, 129, 249, 205, 157, 2, 50, 226, 3, 10, 197, 3, 10,
		249, 2, 10, 12, 8, 226, 229, 235, 151, 1, 16, 233, 192, 182, 202, 10, 18, 20, 10, 18, 97, 110, 121, 95, 101, 65, 67, 76, 95,
		99, 111, 110, 116, 97, 105, 110, 101, 114, 26, 167, 1, 8, 181, 172, 128, 150, 4, 16, 199, 217, 244, 29, 26, 44, 8, 185, 184, 168,
		169, 2, 16, 217, 219, 145, 189, 6, 26, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 49, 45, 107, 101, 121, 34, 14, 102, 105, 108,
		116, 101, 114, 45, 49, 45, 49, 45, 118, 97, 108, 26, 44, 8, 159, 209, 170, 254, 5, 16, 211, 130, 166, 140, 5, 26, 14, 102, 105,
		108, 116, 101, 114, 45, 49, 45, 50, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 49, 45, 50, 45, 118, 97, 108, 34,
		30, 8, 148, 144, 226, 163, 2, 18, 10, 115, 117, 98, 106, 45, 49, 45, 49, 45, 49, 18, 10, 115, 117, 98, 106, 45, 49, 45, 49, 45,
		50, 34, 30, 8, 138, 228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45, 49, 45, 50, 45, 49, 18, 10, 115, 117, 98, 106, 45, 49, 45,
		50, 45, 50, 26, 168, 1, 8, 182, 137, 168, 207, 4, 16, 182, 202, 221, 178, 6, 26, 44, 8, 185, 184, 168, 169, 2, 16, 217, 219,
		145, 189, 6, 26, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 49, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 50,
		45, 49, 45, 118, 97, 108, 26, 44, 8, 159, 209, 170, 254, 5, 16, 211, 130, 166, 140, 5, 26, 14, 102, 105, 108, 116, 101, 114, 45,
		50, 45, 50, 45, 107, 101, 121, 34, 14, 102, 105, 108, 116, 101, 114, 45, 50, 45, 50, 45, 118, 97, 108, 34, 30, 8, 148, 144, 226,
		163, 2, 18, 10, 115, 117, 98, 106, 45, 50, 45, 49, 45, 49, 18, 10, 115, 117, 98, 106, 45, 50, 45, 49, 45, 50, 34, 30, 8, 138,
		228, 158, 248, 6, 18, 10, 115, 117, 98, 106, 45, 50, 45, 50, 45, 49, 18, 10, 115, 117, 98, 106, 45, 50, 45, 50, 45, 50, 18, 17,
		10, 15, 97, 110, 121, 95, 98, 101, 97, 114, 101, 114, 95, 117, 115, 101, 114, 26, 31, 8, 183, 239, 172, 246, 142, 197, 200, 130, 184,
		1, 16, 149, 205, 210, 185, 246, 151, 166, 255, 120, 24, 152, 236, 229, 220, 255, 141, 132, 147, 28, 34, 19, 10, 17, 97, 110,
		121, 95, 98, 101, 97, 114, 101, 114, 95, 105, 115, 115, 117, 101, 114, 18, 24, 10, 7, 97, 110, 121, 95, 112, 117, 98, 18, 7, 97, 110,
		121, 95, 115, 105, 103, 24, 158, 181, 255, 143, 5, 64, 210, 230, 221, 152, 247, 205, 254, 166, 194, 1, 64, 210, 230, 221,
		152, 247, 205, 254, 166, 194, 1}
)

var (
	requestSignerECDSAPubBin = []byte{3, 222, 100, 155, 214, 54, 45, 96, 2, 218, 144, 121, 166, 210, 58, 194, 143, 221, 111, 63, 87,
		254, 66, 2, 236, 94, 45, 93, 30, 39, 191, 127, 80}
	requestSignerL2ECDSAPubBin = []byte{3, 95, 195, 112, 130, 26, 227, 140, 73, 208, 191, 208, 134, 199, 189, 139, 238, 55, 22, 49,
		165, 67, 146, 187, 82, 232, 85, 95, 144, 75, 87, 243, 21}
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
	getObjectUnsignedRequest = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdr,
	}
	// clone to use.
	getObjectSignedRequest = &protoobject.GetRequest{
		Body:       getObjectRequestBody,
		MetaHeader: reqMetaHdrL2,
		VerifyHeader: &protosession.RequestVerificationHeader{
			BodySignature: nil,
			MetaSignature: &refs.Signature{
				Key:    bytes.Clone(requestSignerL2ECDSAPubBin),
				Sign:   []byte{26, 147, 47, 31, 10, 173, 115, 179, 126, 16, 132, 149, 125, 68, 153, 129, 254, 184, 34, 53, 155, 194, 128, 115, 88, 68, 158, 91, 45, 8, 91, 169, 125, 215, 202, 234, 142, 72, 14, 110, 222, 142, 124, 200, 53, 189, 217, 100, 254, 100, 13, 9, 66, 60, 188, 5, 167, 116, 215, 230, 34, 150, 203, 132},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
			},
			OriginSignature: &refs.Signature{
				Key:    bytes.Clone(requestSignerL2ECDSAPubBin),
				Sign:   []byte{175, 192, 13, 37, 185, 173, 75, 11, 49, 178, 102, 150, 37, 208, 1, 158, 69, 252, 242, 121, 204, 220, 170, 117, 103, 250, 194, 218, 212, 144, 245, 177, 56, 67, 189, 182, 12, 122, 241, 4, 187, 154, 253, 56, 24, 138, 16, 103, 143, 203, 29, 228, 136, 33, 49, 245, 30, 165, 111, 23, 117, 149, 149, 228, 242, 157, 202, 93, 66, 215, 69, 103, 197, 232, 107, 147, 246, 192, 177, 158},
				Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
			},
			Origin: &protosession.RequestVerificationHeader{
				BodySignature: &refs.Signature{
					Key: bytes.Clone(requestSignerECDSAPubBin),
					Sign: []byte{4, 54, 181, 48, 83, 197, 23, 131, 0, 233, 48, 96, 155, 28, 68, 0, 189, 120, 251, 60, 163, 5, 136, 106, 63,
						126, 99, 34, 198, 66, 247, 207, 135, 12, 130, 49, 130, 155, 236, 204, 71, 23, 33, 178, 163, 27, 28, 101, 33, 33,
						91, 229, 217, 170, 250, 226, 62, 93, 22, 3, 181, 81, 69, 9, 97},
					Scheme: refs.SignatureScheme_ECDSA_SHA512,
				},
				MetaSignature: &refs.Signature{
					Key: bytes.Clone(requestSignerECDSAPubBin),
					Sign: []byte{152, 135, 221, 72, 61, 96, 131, 169, 229, 9, 203, 210, 132, 62, 40, 1, 211, 63, 130, 4, 136, 199, 186,
						219, 104, 2, 50, 101, 89, 252, 144, 184, 28, 125, 230, 39, 128, 238, 210, 223, 69, 128, 164, 112, 218, 133,
						80, 96, 19, 169, 156, 125, 250, 99, 197, 152, 73, 74, 15, 152, 186, 168, 170, 189},
					Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256,
				},
				OriginSignature: &refs.Signature{
					Key: bytes.Clone(requestSignerECDSAPubBin),
					Sign: []byte{232, 128, 107, 75, 64, 63, 81, 149, 215, 6, 170, 132, 68, 181, 142, 100, 169, 242, 40, 227, 12, 103,
						202, 72, 190, 66, 240, 251, 115, 112, 36, 115, 169, 186, 16, 121, 153, 101, 206, 38, 156, 154, 69, 80, 198, 172, 125,
						115, 114, 54, 224, 44, 198, 137, 131, 236, 163, 209, 208, 136, 146, 184, 70, 136, 60, 200, 208, 106, 154, 206, 83,
						44, 222, 202, 169, 116, 157, 3, 5, 181},
					Scheme: refs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT,
				},
			},
		},
	}
)

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
			req := proto.Clone(getObjectUnsignedRequest).(*protoobject.GetRequest)

			vh, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](tc.signer, req, nil)
			require.NoError(t, err)
			require.NotNil(t, vh)
			require.Nil(t, vh.Origin)

			checkSignerCreds(tc.signer.Scheme(), vh.BodySignature, vh.MetaSignature, vh.OriginSignature)

			tc.verifyFunc(t, pub, tc.hashFunc(getObjectRequestBodyBin), vh.BodySignature.Sign)
			tc.verifyFunc(t, pub, tc.hashFunc(reqMetaHdrBin), vh.MetaSignature.Sign)
			tc.verifyFunc(t, pub, tc.hashFunc(nil), vh.OriginSignature.Sign)

			req.VerifyHeader = vh
			err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
			require.NoError(t, err)

			t.Run("re-sign", func(t *testing.T) {
				req.MetaHeader = reqMetaHdrL2

				vhL2, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](tc.signer, req, nil)
				require.NoError(t, err)
				require.NotNil(t, vhL2)
				require.True(t, vhL2.Origin == vh) // as pointers

				checkSignerCreds(tc.signer.Scheme(), vhL2.MetaSignature, vhL2.OriginSignature)

				require.Nil(t, vhL2.BodySignature)
				tc.verifyFunc(t, pub, tc.hashFunc(reqMetaHdrL2Bin), vhL2.MetaSignature.Sign)
				originHash := tc.hashFunc(neofsproto.MarshalMessage(vh))
				tc.verifyFunc(t, pub, originHash, vhL2.OriginSignature.Sign)

				req.VerifyHeader = vhL2
				err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
				require.NoError(t, err)
			})
		})
	}
	t.Run("ECDSA_SHA256_WalletConnect", func(t *testing.T) {
		req := proto.Clone(getObjectUnsignedRequest).(*protoobject.GetRequest)

		vh, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](anySigner.WalletConnect, req, nil)
		require.NoError(t, err)
		require.NotNil(t, vh)
		require.Nil(t, vh.Origin)

		checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vh.BodySignature, vh.MetaSignature, vh.OriginSignature)

		verifyWalletConnectSignature(t, pub, getObjectRequestBodyBin, vh.BodySignature.Sign)
		verifyWalletConnectSignature(t, pub, reqMetaHdrBin, vh.MetaSignature.Sign)
		verifyWalletConnectSignature(t, pub, nil, vh.OriginSignature.Sign)

		req.VerifyHeader = vh
		err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
		require.NoError(t, err)

		t.Run("re-sign", func(t *testing.T) {
			req.MetaHeader = reqMetaHdrL2

			vhL2, err := neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](anySigner.WalletConnect, req, nil)
			require.NoError(t, err)
			require.NotNil(t, vhL2)
			require.True(t, vhL2.Origin == vh) // as pointers

			checkSignerCreds(neofscrypto.ECDSA_WALLETCONNECT, vhL2.MetaSignature, vhL2.OriginSignature)

			require.Nil(t, vhL2.BodySignature)
			verifyWalletConnectSignature(t, pub, reqMetaHdrL2Bin, vhL2.MetaSignature.Sign)
			verifyWalletConnectSignature(t, pub, neofsproto.MarshalMessage(vh.Origin), vh.OriginSignature.Sign)

			req.VerifyHeader = vhL2
			err = neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
			require.NoError(t, err)
		})
	})
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
				req := proto.Clone(getObjectSignedRequest).(*protoobject.GetRequest)
				req.MetaHeader = req.MetaHeader.Origin
				req.VerifyHeader = req.VerifyHeader.Origin
				tc.corrupt(req.VerifyHeader)
				err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
				require.EqualError(t, err, "invalid verification header at depth 0: "+tc.msg)

				t.Run("resigned", func(t *testing.T) {
					req := &protoobject.GetRequest{
						Body:         req.Body,
						MetaHeader:   &protosession.RequestMetaHeader{Origin: req.MetaHeader},
						VerifyHeader: req.VerifyHeader,
					}
					req.VerifyHeader, err = neofscrypto.SignRequestWithBuffer[*protoobject.GetRequest_Body](neofscryptotest.Signer(), req, nil)
					require.NoError(t, err)

					err := neofscrypto.VerifyRequestWithBuffer[*protoobject.GetRequest_Body](req, nil)
					require.EqualError(t, err, "invalid verification header at depth 1: "+tc.msg)
				})
			})
		}
		t.Run("resigned", func(t *testing.T) {
			for _, tc := range []struct {
				name, msg string
				corrupt   func(valid *protoobject.GetRequest)
			}{
				{name: "redundant verification header", msg: "incorrect number of verification headers",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.VerifyHeader = &protosession.RequestVerificationHeader{Origin: valid.VerifyHeader}
					},
				},
				{name: "lacking verification header", msg: "incorrect number of verification headers",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.MetaHeader = &protosession.RequestMetaHeader{Origin: valid.MetaHeader}
					},
				},
				{name: "with body signature", msg: "invalid verification header at depth 0: body signature is set in non-origin verification header",
					corrupt: func(valid *protoobject.GetRequest) {
						valid.VerifyHeader.BodySignature = new(refs.Signature)
					},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					req := proto.Clone(getObjectSignedRequest).(*protoobject.GetRequest)
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
		Init: &protoobject.GetResponse_Body_Init{
			ObjectId:  &refs.ObjectID{Value: []byte("any_ID")},
			Signature: &refs.Signature{Key: []byte("any_pub"), Sign: []byte("any_sig"), Scheme: 2128773493},
			Header: &protoobject.Header{
				Version:       &refs.Version{Major: 1559619596, Minor: 436551331},
				ContainerId:   &refs.ContainerID{Value: []byte("any_container")},
				OwnerId:       &refs.OwnerID{Value: []byte("any_owner")},
				CreationEpoch: 10561284447300915844,
				PayloadLength: 766049361057238504,
			},
		},
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
