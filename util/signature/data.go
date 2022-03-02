package signature

import (
	"crypto/ecdsa"
	"errors"

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

func SignData(key *ecdsa.PrivateKey, src DataSource, opts ...SignOption) (*signature.Signature, error) {
	if key == nil {
		return nil, ErrEmptyPrivateKey
	}

	data, err := dataForSignature(src)
	if err != nil {
		return nil, err
	}
	defer bytesPool.Put(&data)

	cfg := getConfig(opts...)

	sigData, err := sign(cfg.scheme, key, data)
	if err != nil {
		return nil, err
	}

	sig := signature.New()
	sig.SetKey((*keys.PublicKey)(&key.PublicKey).Bytes())
	sig.SetSign(sigData)
	sig.SetScheme(cfg.scheme)
	return sig, nil
}

func VerifyData(dataSrc DataSource, sig *signature.Signature, opts ...SignOption) error {
	data, err := dataForSignature(dataSrc)
	if err != nil {
		return err
	}
	defer bytesPool.Put(&data)

	cfg := getConfig(opts...)

	return verify(cfg, data, sig)
}
