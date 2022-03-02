package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
)

var curve = elliptic.P256()

type cfg struct {
	schemeFixed bool
	scheme      signature.Scheme
}

func getConfig(opts ...SignOption) *cfg {
	cfg := &cfg{
		scheme: signature.ECDSAWithSHA512,
	}

	for i := range opts {
		opts[i](cfg)
	}

	return cfg
}

func sign(scheme signature.Scheme, key *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	switch scheme {
	case signature.ECDSAWithSHA512:
		h := sha512.Sum512(msg)
		x, y, err := ecdsa.Sign(rand.Reader, key, h[:])
		if err != nil {
			return nil, err
		}
		return elliptic.Marshal(elliptic.P256(), x, y), nil
	case signature.RFC6979WithSHA256:
		p := &keys.PrivateKey{PrivateKey: *key}
		return p.Sign(msg), nil
	default:
		panic(fmt.Sprintf("unsupported scheme %s", scheme))
	}
}

func verify(cfg *cfg, msg []byte, f func() (key, sign []byte, scheme signature.Scheme)) error {
	key, sign, scheme := f()

	pub, err := keys.NewPublicKeyFromBytes(key, elliptic.P256())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPublicKey, err)
	}

	if !cfg.schemeFixed {
		cfg.scheme = scheme
	}

	switch cfg.scheme {
	case signature.ECDSAWithSHA512:
		h := sha512.Sum512(msg)
		r, s := unmarshalXY(sign)
		if r != nil && s != nil && ecdsa.Verify((*ecdsa.PublicKey)(pub), h[:], r, s) {
			return nil
		}
		return ErrInvalidSignature
	case signature.RFC6979WithSHA256:
		h := sha256.Sum256(msg)
		if pub.Verify(sign, h[:]) {
			return nil
		}
		return ErrInvalidSignature
	default:
		return fmt.Errorf("unsupported signature scheme %s", cfg.scheme)
	}
}

// unmarshalXY converts a point, serialized by Marshal, into an x, y pair.
// It is an error if the point is not in uncompressed form.
// On error, x,y = nil.
// Unlike the original version of the code, we ignore that x or y not on the curve
// --------------
// It's copy-paste elliptic.Unmarshal(curve, data) stdlib function, without last line
// of code.
// Link - https://golang.org/pkg/crypto/elliptic/#Unmarshal
func unmarshalXY(data []byte) (x *big.Int, y *big.Int) {
	if len(data) != PublicKeyUncompressedSize {
		return
	} else if data[0] != 4 { // uncompressed form
		return
	}

	p := curve.Params().P
	x = new(big.Int).SetBytes(data[1:PublicKeyCompressedSize])
	y = new(big.Int).SetBytes(data[PublicKeyCompressedSize:])

	if x.Cmp(p) >= 0 || y.Cmp(p) >= 0 {
		x, y = nil, nil
	}

	return
}

func SignWithRFC6979() SignOption {
	return func(c *cfg) {
		c.schemeFixed = true
		c.scheme = signature.RFC6979WithSHA256
	}
}
