package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
)

var curve = elliptic.P256()

type cfg struct {
	signFunc   func(key *ecdsa.PrivateKey, msg []byte) ([]byte, error)
	verifyFunc func(key *ecdsa.PublicKey, msg []byte, sig []byte) error
}

func defaultCfg() *cfg {
	return &cfg{
		signFunc:   sign,
		verifyFunc: verify,
	}
}

func sign(key *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	h := sha512.Sum512(msg)
	x, y, err := ecdsa.Sign(rand.Reader, key, h[:])
	if err != nil {
		return nil, err
	}
	return elliptic.Marshal(elliptic.P256(), x, y), nil
}

func verify(key *ecdsa.PublicKey, msg []byte, sig []byte) error {
	h := sha512.Sum512(msg)
	r, s := unmarshalXY(sig)
	if r != nil && s != nil && ecdsa.Verify(key, h[:], r, s) {
		return nil
	}
	return ErrInvalidSignature
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
		c.signFunc = signRFC6979
		c.verifyFunc = verifyRFC6979
	}
}

func signRFC6979(key *ecdsa.PrivateKey, msg []byte) ([]byte, error) {
	p := &keys.PrivateKey{PrivateKey: *key}
	return p.Sign(msg), nil
}

func verifyRFC6979(key *ecdsa.PublicKey, msg []byte, sig []byte) error {
	p := (*keys.PublicKey)(key)
	h := sha256.Sum256(msg)
	if p.Verify(sig, h[:]) {
		return nil
	}
	return ErrInvalidSignature
}
