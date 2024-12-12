package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/io"
	protorefs "github.com/nspcc-dev/neofs-api-go/v2/refs/grpc"
	apigrpc "github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

var p256Curve = elliptic.P256()

type signedMessageV2 interface {
	FromGRPCMessage(apigrpc.Message) error
	StableMarshal([]byte) []byte
}

// represents tested NeoFS authentication credentials.
type authCredentials struct {
	scheme protorefs.SignatureScheme
	pub    []byte
}

func authCredentialsFromSigner(s neofscrypto.Signer) authCredentials {
	var res authCredentials
	res.pub = neofscrypto.PublicKeyBytes(s.Public())
	switch scheme := s.Scheme(); scheme {
	default:
		res.scheme = protorefs.SignatureScheme(scheme)
	case neofscrypto.ECDSA_SHA512:
		res.scheme = protorefs.SignatureScheme_ECDSA_SHA512
	case neofscrypto.ECDSA_DETERMINISTIC_SHA256:
		res.scheme = protorefs.SignatureScheme_ECDSA_RFC6979_SHA256
	case neofscrypto.ECDSA_WALLETCONNECT:
		res.scheme = protorefs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT
	}
	return res
}

func checkAuthCredendials(exp, act authCredentials) error {
	if exp.scheme != act.scheme {
		return fmt.Errorf("unexpected scheme (client: %v, message: %v)", exp.scheme, act.scheme)
	}
	if !bytes.Equal(exp.pub, act.pub) {
		return fmt.Errorf("unexpected public key (client: %x, message: %x)", exp.pub, act.pub)
	}
	return nil
}

func signMessage[MESSAGE apigrpc.Message, MESSAGEV2 any, MESSAGEV2PTR interface {
	*MESSAGEV2
	signedMessageV2
}](key ecdsa.PrivateKey, m MESSAGE, _ MESSAGEV2PTR) (*protorefs.Signature, error) {
	mV2 := MESSAGEV2PTR(new(MESSAGEV2))
	if err := mV2.FromGRPCMessage(m); err != nil {
		panic(err)
	}
	b := mV2.StableMarshal(nil)
	h := sha512.Sum512(b)
	r, s, err := ecdsa.Sign(rand.Reader, &key, h[:])
	if err != nil {
		return nil, fmt.Errorf("sign ECDSA: %w", err)
	}
	sig := make([]byte, 1+64)
	sig[0] = 4
	r.FillBytes(sig[1:33])
	s.FillBytes(sig[33:])
	return &protorefs.Signature{Key: elliptic.MarshalCompressed(p256Curve, key.X, key.Y), Sign: sig}, nil
}

func verifyMessageSignature[MESSAGE apigrpc.Message, MESSAGEV2 any, MESSAGEV2PTR interface {
	*MESSAGEV2
	signedMessageV2
}](m MESSAGE, s *protorefs.Signature, expectedCreds *authCredentials) error {
	mV2 := MESSAGEV2PTR(new(MESSAGEV2))
	if err := mV2.FromGRPCMessage(m); err != nil {
		panic(err)
	}
	return verifyDataSignature(mV2.StableMarshal(nil), s, expectedCreds)
}

func verifyDataSignature(data []byte, s *protorefs.Signature, expectedCreds *authCredentials) error {
	if s == nil {
		return errors.New("missing")
	}
	creds := authCredentials{scheme: s.Scheme, pub: s.Key}
	if err := verifyProtoSignature(creds, s.Sign, data); err != nil {
		return err
	}
	if expectedCreds != nil {
		if err := checkAuthCredendials(*expectedCreds, creds); err != nil {
			return fmt.Errorf("unexpected credendials: %w", err)
		}
	}
	return nil
}

func verifyProtoSignature(creds authCredentials, sig, data []byte) error {
	switch creds.scheme {
	default:
		return fmt.Errorf("unsupported scheme: %v", creds.scheme)
	case protorefs.SignatureScheme_ECDSA_SHA512:
		if len(sig) != keys.SignatureLen+1 {
			return fmt.Errorf("invalid signature length %d, should be %d", len(sig), keys.SignatureLen+1)
		}
		r, s := unmarshalECP256Point([keys.SignatureLen + 1]byte(sig))
		if r == nil {
			return fmt.Errorf("invalid signature format %x", sig)
		}
		x, y := elliptic.UnmarshalCompressed(p256Curve, creds.pub)
		if x == nil {
			return fmt.Errorf("invalid public key: %x", sig)
		}
		h := sha512.Sum512(data)
		if !ecdsa.Verify(&ecdsa.PublicKey{Curve: p256Curve, X: x, Y: y}, h[:], r, s) {
			return errors.New("signature mismatch")
		}
	case protorefs.SignatureScheme_ECDSA_RFC6979_SHA256:
		if len(sig) != keys.SignatureLen {
			return fmt.Errorf("invalid signature length %d, should be %d", len(sig), keys.SignatureLen)
		}
		x, y := elliptic.UnmarshalCompressed(p256Curve, creds.pub)
		if x == nil {
			return fmt.Errorf("invalid signature's public key: %x", sig)
		}
		h := sha256.Sum256(data)
		r, s := ecP256PointFromBytes([keys.SignatureLen]byte(sig))
		if !ecdsa.Verify(&ecdsa.PublicKey{Curve: p256Curve, X: x, Y: y}, h[:], r, s) {
			return errors.New("signature mismatch")
		}
	case protorefs.SignatureScheme_ECDSA_RFC6979_SHA256_WALLET_CONNECT:
		const saltLen = 16
		if len(sig) != keys.SignatureLen+saltLen {
			return fmt.Errorf("invalid signature length %d, should be %d",
				len(sig), keys.SignatureLen)
		}
		x, y := elliptic.UnmarshalCompressed(p256Curve, creds.pub)
		if x == nil {
			return fmt.Errorf("invalid public key: %x", creds.pub)
		}

		b64 := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(b64, data)
		payloadLen := 2*saltLen + len(b64)
		b := make([]byte, 4+io.GetVarSize(payloadLen)+payloadLen+2)
		n := copy(b, []byte{0x01, 0x00, 0x01, 0xf0})
		n += io.PutVarUint(b[n:], uint64(payloadLen))
		n += hex.Encode(b[n:], sig[keys.SignatureLen:])
		n += copy(b[n:], b64)
		copy(b[n:], []byte{0x00, 0x00})

		h := sha256.Sum256(b)
		r, s := ecP256PointFromBytes([keys.SignatureLen]byte(sig))
		if !ecdsa.Verify(&ecdsa.PublicKey{Curve: p256Curve, X: x, Y: y}, h[:], r, s) {
			return errors.New("signature mismatch")
		}
	}
	return nil
}

func ecP256PointFromBytes(b [keys.SignatureLen]byte) (*big.Int, *big.Int) {
	return new(big.Int).SetBytes(b[:32]), new(big.Int).SetBytes(b[32:])
}

// decodes a serialized [elliptic.P256] point. It is an error if the point is
// not in uncompressed form, or is the point at infinity. On error, x = nil.
func unmarshalECP256Point(b [keys.SignatureLen + 1]byte) (x, y *big.Int) {
	if b[0] != 4 { // uncompressed form
		return
	}
	p := p256Curve.Params().P
	x, y = ecP256PointFromBytes([keys.SignatureLen]byte(b[1:]))
	if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
		return nil, nil
	}
	return x, y
}
