package neofscrypto_test

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/api/accounting"
	"github.com/nspcc-dev/neofs-sdk-go/api/container"
	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/api/status"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	internalproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type request interface {
	proto.Message
	neofscrypto.Request
}

type response interface {
	proto.Message
	neofscrypto.Response
}

func verifyRequestFromFile(t testing.TB, file string, req request, getBody func(request) internalproto.Message) {
	b, err := os.ReadFile(filepath.Join(testDataDir, file))
	require.NoError(t, err)
	b, err = hex.DecodeString(string(b))
	require.NoError(t, err)
	require.NoError(t, proto.Unmarshal(b, req))
	require.NoError(t, neofscrypto.VerifyRequest(req, getBody(req)))
}

func verifyResponseFromFile(t testing.TB, file string, resp response, getBody func(response) internalproto.Message) {
	b, err := os.ReadFile(filepath.Join(testDataDir, file))
	require.NoError(t, err)
	b, err = hex.DecodeString(string(b))
	require.NoError(t, err)
	require.NoError(t, proto.Unmarshal(b, resp))
	require.NoError(t, neofscrypto.VerifyResponse(resp, getBody(resp)))
}

func testSignRequest(t testing.TB, req request, body internalproto.Message, meta **session.RequestMetaHeader, verif **session.RequestVerificationHeader) {
	require.Error(t, neofscrypto.VerifyRequest(req, body))

	key1, err := keys.NewPrivateKey()
	require.NoError(t, err)
	key2, err := keys.NewPrivateKey()
	require.NoError(t, err)
	key3, err := keys.NewPrivateKey()
	require.NoError(t, err)

	signers := []neofscrypto.Signer{
		neofsecdsa.Signer(key1.PrivateKey),
		neofsecdsa.SignerRFC6979(key2.PrivateKey),
		neofsecdsa.SignerWalletConnect(key3.PrivateKey),
	}

	for i := range signers {
		n := 100*i + 1
		*meta = &session.RequestMetaHeader{
			Version:     &refs.Version{Major: uint32(n), Minor: uint32(n + 1)},
			Epoch:       uint64(n + 3),
			Ttl:         uint32(n + 4),
			MagicNumber: uint64(n + 5),
			XHeaders: []*session.XHeader{
				{Key: "xheader_key" + strconv.Itoa(n), Value: "xheader_val" + strconv.Itoa(n)},
				{Key: "xheader_key" + strconv.Itoa(n+1), Value: "xheader_val" + strconv.Itoa(n+1)},
			},
			Origin: *meta,
		}

		*verif, err = neofscrypto.SignRequest(signers[i], req, body, nil)
		require.NoError(t, err)
	}

	require.NoError(t, neofscrypto.VerifyRequest(req, body))
}

func testSignResponse(t testing.TB, resp response, body internalproto.Message, meta **session.ResponseMetaHeader, verif **session.ResponseVerificationHeader) {
	require.Error(t, neofscrypto.VerifyResponse(resp, body))

	key1, err := keys.NewPrivateKey()
	require.NoError(t, err)
	key2, err := keys.NewPrivateKey()
	require.NoError(t, err)
	key3, err := keys.NewPrivateKey()
	require.NoError(t, err)

	signers := []neofscrypto.Signer{
		neofsecdsa.Signer(key1.PrivateKey),
		neofsecdsa.SignerRFC6979(key2.PrivateKey),
		neofsecdsa.SignerWalletConnect(key3.PrivateKey),
	}

	for i := range signers {
		n := 100*i + 1
		*meta = &session.ResponseMetaHeader{
			Version: &refs.Version{Major: uint32(n), Minor: uint32(n + 1)},
			Epoch:   uint64(n + 3),
			Ttl:     uint32(n + 4),
			XHeaders: []*session.XHeader{
				{Key: "xheader_key" + strconv.Itoa(n), Value: "xheader_val" + strconv.Itoa(n)},
				{Key: "xheader_key" + strconv.Itoa(n+1), Value: "xheader_val" + strconv.Itoa(n+1)},
			},
			Origin: *meta,
			Status: &status.Status{Code: uint32(n + 6)},
		}

		*verif, err = neofscrypto.SignResponse(signers[i], resp, body, nil)
		require.NoError(t, err)
	}

	require.NoError(t, neofscrypto.VerifyResponse(resp, body))
}

func TestAPIVerify(t *testing.T) {
	t.Run("accounting", func(t *testing.T) {
		t.Run("balance", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				verifyRequestFromFile(t, "accounting_balance_request", new(accounting.BalanceRequest), func(r request) internalproto.Message {
					return r.(*accounting.BalanceRequest).Body
				})
			})
			t.Run("response", func(t *testing.T) {
				verifyResponseFromFile(t, "accounting_balance_response", new(accounting.BalanceResponse), func(r response) internalproto.Message {
					return r.(*accounting.BalanceResponse).Body
				})
			})
		})
	})
	t.Run("container", func(t *testing.T) {
		t.Run("put", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				verifyRequestFromFile(t, "container_put_request", new(container.PutRequest), func(r request) internalproto.Message {
					return r.(*container.PutRequest).Body
				})
			})
			t.Run("response", func(t *testing.T) {
				verifyResponseFromFile(t, "container_put_response", new(container.PutResponse), func(r response) internalproto.Message {
					return r.(*container.PutResponse).Body
				})
			})
		})
		t.Run("get", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				verifyRequestFromFile(t, "container_get_request", new(container.GetRequest), func(r request) internalproto.Message {
					return r.(*container.GetRequest).Body
				})
			})
			t.Run("response", func(t *testing.T) {
				verifyResponseFromFile(t, "container_get_response", new(container.GetResponse), func(r response) internalproto.Message {
					return r.(*container.GetResponse).Body
				})
			})
		})
		t.Run("delete", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				verifyRequestFromFile(t, "container_delete_request", new(container.DeleteRequest), func(r request) internalproto.Message {
					return r.(*container.DeleteRequest).Body
				})
			})
			t.Run("response", func(t *testing.T) {
				verifyResponseFromFile(t, "container_delete_response", new(container.DeleteResponse), func(r response) internalproto.Message {
					return r.(*container.DeleteResponse).Body
				})
			})
		})
		t.Run("list", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				verifyRequestFromFile(t, "container_list_request", new(container.ListRequest), func(r request) internalproto.Message {
					return r.(*container.ListRequest).Body
				})
			})
			t.Run("response", func(t *testing.T) {
				verifyResponseFromFile(t, "container_list_response", new(container.ListResponse), func(r response) internalproto.Message {
					return r.(*container.ListResponse).Body
				})
			})
		})
	})
}

func TestAPISign(t *testing.T) {
	t.Run("accounting", func(t *testing.T) {
		t.Run("balance", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				req := &accounting.BalanceRequest{Body: &accounting.BalanceRequest_Body{
					OwnerId: &refs.OwnerID{Value: []byte("any_user")},
				}}
				testSignRequest(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
			t.Run("response", func(t *testing.T) {
				req := &accounting.BalanceResponse{Body: &accounting.BalanceResponse_Body{
					Balance: &accounting.Decimal{Value: 1, Precision: 2},
				}}
				testSignResponse(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
		})
	})
	t.Run("container", func(t *testing.T) {
		t.Run("put", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				req := &container.PutRequest{Body: &container.PutRequest_Body{
					Container: &container.Container{
						Version:  &refs.Version{Major: 1, Minor: 2},
						OwnerId:  &refs.OwnerID{Value: []byte("any_user")},
						Nonce:    []byte("any_nonce"),
						BasicAcl: 3,
						Attributes: []*container.Container_Attribute{
							{Key: "attr_key1", Value: "attr_val1"},
							{Key: "attr_key2", Value: "attr_val2"},
						},
						PlacementPolicy: &netmap.PlacementPolicy{
							Replicas:              []*netmap.Replica{{Count: 4}},
							ContainerBackupFactor: 5,
						},
					},
					Signature: &refs.SignatureRFC6979{
						Key:  []byte("any_public_key"),
						Sign: []byte("any_signature"),
					},
				}}
				testSignRequest(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
			t.Run("response", func(t *testing.T) {
				req := &container.PutResponse{Body: &container.PutResponse_Body{
					ContainerId: &refs.ContainerID{Value: []byte("any_container")},
				}}
				testSignResponse(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
		})
		t.Run("get", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				req := &container.GetRequest{Body: &container.GetRequest_Body{
					ContainerId: &refs.ContainerID{Value: []byte("any_container_id")},
				}}
				testSignRequest(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
			t.Run("response", func(t *testing.T) {
				req := &container.GetResponse{Body: &container.GetResponse_Body{
					Container: &container.Container{
						Version:  &refs.Version{Major: 1, Minor: 2},
						OwnerId:  &refs.OwnerID{Value: []byte("any_user")},
						Nonce:    []byte("any_nonce"),
						BasicAcl: 3,
						Attributes: []*container.Container_Attribute{
							{Key: "attr_key1", Value: "attr_val1"},
							{Key: "attr_key2", Value: "attr_val2"},
						},
						PlacementPolicy: &netmap.PlacementPolicy{
							Replicas:              []*netmap.Replica{{Count: 4}},
							ContainerBackupFactor: 5,
						},
					},
					Signature: &refs.SignatureRFC6979{
						Key:  []byte("any_public_key"),
						Sign: []byte("any_signature"),
					},
					SessionToken: &session.SessionToken{
						Body: &session.SessionToken_Body{
							Id:         []byte("any_ID"),
							OwnerId:    &refs.OwnerID{Value: []byte("any_user")},
							Lifetime:   &session.SessionToken_Body_TokenLifetime{Exp: 101, Nbf: 102, Iat: 103},
							SessionKey: []byte("any_session_key"),
							Context: &session.SessionToken_Body_Container{Container: &session.ContainerSessionContext{
								Verb:        200,
								Wildcard:    true,
								ContainerId: &refs.ContainerID{Value: []byte("any_container")},
							}},
						},
						Signature: &refs.Signature{
							Key:    []byte("any_public_key"),
							Sign:   []byte("any_signature"),
							Scheme: 123,
						},
					},
				}}
				testSignResponse(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
		})
		t.Run("delete", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				req := &container.DeleteRequest{Body: &container.DeleteRequest_Body{
					ContainerId: &refs.ContainerID{Value: []byte("any_container_id")},
					Signature: &refs.SignatureRFC6979{
						Key:  []byte("any_public_key"),
						Sign: []byte("any_signature"),
					},
				}}
				testSignRequest(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
			t.Run("response", func(t *testing.T) {
				req := &container.DeleteResponse{Body: &container.DeleteResponse_Body{}}
				testSignResponse(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
		})
		t.Run("list", func(t *testing.T) {
			t.Run("request", func(t *testing.T) {
				req := &container.ListRequest{Body: &container.ListRequest_Body{
					OwnerId: &refs.OwnerID{Value: []byte("any_user")},
				}}
				testSignRequest(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
			t.Run("response", func(t *testing.T) {
				req := &container.ListResponse{Body: &container.ListResponse_Body{
					ContainerIds: []*refs.ContainerID{
						{Value: []byte("any_container1")},
						{Value: []byte("any_container2")},
					},
				}}
				testSignResponse(t, req, req.Body, &req.MetaHeader, &req.VerifyHeader)
			})
		})
	})
}
