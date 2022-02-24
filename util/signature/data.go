package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
)

type DataSource interface {
	ReadSignedData([]byte) ([]byte, error)
	SignedDataSize() int
}

type DataWithSignature interface {
	DataSource
	GetSignature() *signature.Signature
	SetSignature(*signature.Signature)
}

type SignOption func(*cfg)

type KeySignatureHandler func(*signature.Signature)

type KeySignatureSource func() *signature.Signature

const (
	// PrivateKeyCompressedSize is constant with compressed size of private key (SK).
	// D coordinate stored, recover PK by formula x, y = curve.ScalarBaseMul(d,bytes).
	PrivateKeyCompressedSize = 32

	// PublicKeyCompressedSize is constant with compressed size of public key (PK).
	PublicKeyCompressedSize = 33

	// PublicKeyUncompressedSize is constant with uncompressed size of public key (PK).
	// First byte always should be 0x4 other 64 bytes is X and Y (32 bytes per coordinate).
	// 2 * 32 + 1.
	PublicKeyUncompressedSize = 65
)

var (
	// ErrEmptyPrivateKey is returned when used private key is empty.
	ErrEmptyPrivateKey = errors.New("empty private key")
	// ErrInvalidPublicKey is returned when public key cannot be unmarshalled.
	ErrInvalidPublicKey = errors.New("invalid public key")
	// ErrInvalidSignature is returned if signature cannot be verified.
	ErrInvalidSignature = errors.New("invalid signature")
)

func SignDataWithHandler(key *ecdsa.PrivateKey, src DataSource, handler KeySignatureHandler, opts ...SignOption) error {
	if key == nil {
		return ErrEmptyPrivateKey
	}

	data, err := dataForSignature(src)
	if err != nil {
		return err
	}
	defer bytesPool.Put(&data)

	cfg := getConfig(opts...)

	sigData, err := sign(cfg.defaultScheme, key, data)
	if err != nil {
		return err
	}

	sig := signature.New()
	sig.SetKey((*keys.PublicKey)(&key.PublicKey).Bytes())
	sig.SetSign(sigData)
	sig.SetScheme(cfg.defaultScheme)

	handler(sig)

	return nil
}

func VerifyDataWithSource(dataSrc DataSource, sigSrc KeySignatureSource, opts ...SignOption) error {
	data, err := dataForSignature(dataSrc)
	if err != nil {
		return err
	}
	defer bytesPool.Put(&data)

	cfg := getConfig(opts...)

	sig := sigSrc()

	var pub *keys.PublicKey
	if len(sig.Key()) != 0 {
		pub, err = keys.NewPublicKeyFromBytes(sig.Key(), elliptic.P256())
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
		}
	}

	scheme := sig.Scheme()
	if scheme == signature.Unspecified {
		scheme = cfg.defaultScheme
	}
	if cfg.restrictScheme != signature.Unspecified && scheme != cfg.restrictScheme {
		return fmt.Errorf("%w: unexpected signature scheme", ErrInvalidSignature)
	}

	return verify(
		scheme,
		(*ecdsa.PublicKey)(pub),
		data,
		sig.Sign(),
	)
}

func SignData(key *ecdsa.PrivateKey, v DataWithSignature, opts ...SignOption) error {
	return SignDataWithHandler(key, v, v.SetSignature, opts...)
}

func VerifyData(src DataWithSignature, opts ...SignOption) error {
	return VerifyDataWithSource(src, src.GetSignature, opts...)
}
