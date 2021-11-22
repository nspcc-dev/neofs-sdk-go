package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"

	"github.com/nspcc-dev/rfc6979"
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
	digest := sha256.Sum256(msg)
	r, s := rfc6979.SignECDSA(key, digest[:], sha256.New)
	return getSignatureSlice(key.Curve, r, s), nil
}

func verifyRFC6979(pub *ecdsa.PublicKey, msg []byte, sig []byte) error {
	h := sha256.Sum256(msg)
	if pub.X == nil || pub.Y == nil || len(sig) != 64 {
		return ErrInvalidSignature
	}

	rBytes := new(big.Int).SetBytes(sig[0:32])
	sBytes := new(big.Int).SetBytes(sig[32:64])
	if ecdsa.Verify(pub, h[:], rBytes, sBytes) {
		return nil
	}
	return ErrInvalidSignature
}

func getSignatureSlice(curve elliptic.Curve, r, s *big.Int) []byte {
	params := curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	signature := make([]byte, curveOrderByteSize*2)
	_ = r.FillBytes(signature[:curveOrderByteSize])
	_ = s.FillBytes(signature[curveOrderByteSize:])

	return signature
}
