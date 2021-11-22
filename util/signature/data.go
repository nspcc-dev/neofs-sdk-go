package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
)

type DataSource interface {
	ReadSignedData([]byte) ([]byte, error)
	SignedDataSize() int
}

type DataWithSignature interface {
	DataSource
	GetSignatureWithKey() (key, sig []byte)
	SetSignatureWithKey(key, sig []byte)
}

type SignOption func(*cfg)

type KeySignatureHandler func(key []byte, sig []byte)

type KeySignatureSource func() (key, sig []byte)

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

func DataSignature(key *ecdsa.PrivateKey, src DataSource, opts ...SignOption) ([]byte, error) {
	if key == nil {
		return nil, ErrEmptyPrivateKey
	}

	data, err := dataForSignature(src)
	if err != nil {
		return nil, err
	}
	defer bytesPool.Put(&data)

	cfg := defaultCfg()

	for i := range opts {
		opts[i](cfg)
	}

	return cfg.signFunc(key, data)
}

func SignDataWithHandler(key *ecdsa.PrivateKey, src DataSource, handler KeySignatureHandler, opts ...SignOption) error {
	sig, err := DataSignature(key, src, opts...)
	if err != nil {
		return err
	}

	b := elliptic.MarshalCompressed(key.Curve, key.X, key.Y)
	handler(b, sig)

	return nil
}

func VerifyDataWithSource(dataSrc DataSource, sigSrc KeySignatureSource, opts ...SignOption) error {
	data, err := dataForSignature(dataSrc)
	if err != nil {
		return err
	}
	defer bytesPool.Put(&data)

	cfg := defaultCfg()

	for i := range opts {
		opts[i](cfg)
	}

	key, sig := sigSrc()

	var pub *ecdsa.PublicKey
	if len(key) != 0 {
		x, y := elliptic.UnmarshalCompressed(elliptic.P256(), key)
		if x == nil || y == nil {
			return ErrInvalidPublicKey
		}
		pub = &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	}

	return cfg.verifyFunc(pub, data, sig)
}

func SignData(key *ecdsa.PrivateKey, v DataWithSignature, opts ...SignOption) error {
	return SignDataWithHandler(key, v, v.SetSignatureWithKey, opts...)
}

func VerifyData(src DataWithSignature, opts ...SignOption) error {
	return VerifyDataWithSource(src, src.GetSignatureWithKey, opts...)
}
