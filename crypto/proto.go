package neofscrypto

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/proto/session"
)

var (
	errSignBody               = errors.New("sign body")
	errSignMeta               = errors.New("sign meta header")
	errSignVerifyOrigin       = errors.New("sign verification header's origin")
	errMissingVerifyHdr       = errors.New("missing verification header")
	errWrongVerifyHdrNum      = errors.New("incorrect number of verification headers")
	errMissingVerifyOriginSig = errors.New("missing verification header's origin signature")
	errInvalidVerifyOriginSig = errors.New("invalid verification header's origin signature")
	errMissingMetaSig         = errors.New("missing meta header's signature")
	errInvalidMetaSig         = errors.New("invalid meta header's signature")
	errMissingBodySig         = errors.New("missing body signature")
	errInvalidBodySig         = errors.New("invalid body signature")
	errNonOriginBodySig       = errors.New("body signature is set in non-origin verification header")
)

func newErrInvalidVerificationHeader(depth uint, cause error) error {
	return fmt.Errorf("invalid verification header at depth %d: %w", depth, cause)
}

// ProtoMessage is the marshaling interface provided by all NeoFS proto-level
// structures.
type ProtoMessage interface {
	MarshaledSize() int
	MarshalStable([]byte)
}

// SignedRequest is a generic interface of a signed NeoFS API request.
type SignedRequest[B ProtoMessage] interface {
	GetBody() B
	GetMetaHeader() *session.RequestMetaHeader
	GetVerifyHeader() *session.RequestVerificationHeader
}

// SignedResponse is a generic interface of a signed NeoFS API response.
type SignedResponse[B ProtoMessage] interface {
	GetBody() B
	GetMetaHeader() *session.ResponseMetaHeader
	GetVerifyHeader() *session.ResponseVerificationHeader
}

// SignRequestWithBuffer signs request parts using provided [neofscrypto.Signer]
// according to the NeoFS API protocol, and returns resulting verification
// header to attach to this request.
//
// If signer implements [SignerV2], signing is done using it. In this case,
// [Signer] methods are not called. [OverlapSigner] may be used to pass
// [SignerV2] when [Signer] is unimplemented.
//
// Buffer is optional and free after the call.
func SignRequestWithBuffer[B ProtoMessage](signer Signer, r SignedRequest[B], buf []byte) (*session.RequestVerificationHeader, error) {
	signerV2, signV2 := signer.(SignerV2)
	var ln int
	var err error
	vhOriginal := r.GetVerifyHeader()

	var bs []byte
	var bSig Signature
	signBody := vhOriginal == nil
	if signBody { // body should be signed by the original sender only
		buf, ln = encodeMessage(r.GetBody(), buf)
		if signV2 {
			bSig, err = signerV2.SignData(buf[:ln])
		} else {
			bs, err = signer.Sign(buf[:ln])
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errSignBody, err)
		}
	}

	buf, ln = encodeMessage(r.GetMetaHeader(), buf)
	var ms []byte
	var mSig Signature
	if signV2 {
		mSig, err = signerV2.SignData(buf[:ln])
	} else {
		ms, err = signer.Sign(buf[:ln])
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSignMeta, err)
	}

	buf, ln = encodeMessage(vhOriginal, buf)
	var vs []byte
	var vSig Signature
	if signV2 {
		vSig, err = signerV2.SignData(buf[:ln])
	} else {
		vs, err = signer.Sign(buf[:ln])
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSignVerifyOrigin, err)
	}

	res := &session.RequestVerificationHeader{
		Origin: vhOriginal,
	}
	var pub []byte
	var scheme refs.SignatureScheme
	if signV2 {
		res.MetaSignature = mSig.ProtoMessage()
		res.OriginSignature = vSig.ProtoMessage()
	} else {
		scheme = refs.SignatureScheme(signer.Scheme())
		pub = PublicKeyBytes(signer.Public())
		res.MetaSignature = &refs.Signature{Key: pub, Sign: ms, Scheme: scheme}
		res.OriginSignature = &refs.Signature{Key: pub, Sign: vs, Scheme: scheme}
	}
	if signBody {
		if signV2 {
			res.BodySignature = bSig.ProtoMessage()
		} else {
			res.BodySignature = &refs.Signature{Key: pub, Sign: bs, Scheme: scheme}
		}
	}
	return res, nil
}

// VerifyRequestWithBuffer checks whether verification header of the request is
// formed according to the NeoFS API protocol.
//
// Buffer is optional and free after the call.
func VerifyRequestWithBuffer[B ProtoMessage](r SignedRequest[B], buf []byte) error {
	v := r.GetVerifyHeader()
	if v == nil {
		return errMissingVerifyHdr
	}

	b := r.GetBody()
	m := r.GetMetaHeader()
	bs := maxEncodedSize(b, m, v)
	mo, vo := m.GetOrigin(), v.GetOrigin()
	for {
		if (mo == nil) != (vo == nil) {
			return errWrongVerifyHdrNum
		}
		if vo == nil {
			break
		}
		if s := maxEncodedSize(mo, vo); s > bs {
			bs = s
		}
		mo, vo = mo.GetOrigin(), vo.GetOrigin()
	}

	if len(buf) < bs {
		buf = make([]byte, bs)
	}

	for i := uint(0); ; m, v, i = m.Origin, v.Origin, i+1 {
		if v.MetaSignature == nil {
			return newErrInvalidVerificationHeader(i, errMissingMetaSig)
		}
		if err := verifyMessageSignature(m, v.MetaSignature, buf); err != nil {
			return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidMetaSig, err))
		}
		if v.OriginSignature == nil {
			return newErrInvalidVerificationHeader(i, errMissingVerifyOriginSig)
		}
		if err := verifyMessageSignature(v.Origin, v.OriginSignature, buf); err != nil {
			return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidVerifyOriginSig, err))
		}
		if v.Origin == nil {
			if v.BodySignature == nil {
				return newErrInvalidVerificationHeader(i, errMissingBodySig)
			}
			if err := verifyMessageSignature(b, v.BodySignature, buf); err != nil {
				return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidBodySig, err))
			}
			return nil
		}
		if v.BodySignature != nil {
			return newErrInvalidVerificationHeader(i, errNonOriginBodySig)
		}
	}
}

// SignResponseWithBuffer signs response parts using provided
// [neofscrypto.Signer] according to the NeoFS API protocol, and returns
// resulting verification header to attach to this response.
//
// Buffer is optional and free after the call.
func SignResponseWithBuffer[B ProtoMessage](signer Signer, r SignedResponse[B], buf []byte) (*session.ResponseVerificationHeader, error) {
	var ln int
	var err error
	vhOriginal := r.GetVerifyHeader()

	var bs []byte
	signBody := vhOriginal == nil
	if signBody { // body should be signed by the original sender only
		buf, ln = encodeMessage(r.GetBody(), buf)
		bs, err = signer.Sign(buf[:ln])
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errSignBody, err)
		}
	}

	buf, ln = encodeMessage(r.GetMetaHeader(), buf)
	ms, err := signer.Sign(buf[:ln])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSignMeta, err)
	}

	buf, ln = encodeMessage(vhOriginal, buf)
	vs, err := signer.Sign(buf[:ln])
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSignVerifyOrigin, err)
	}

	scheme := refs.SignatureScheme(signer.Scheme())
	pub := PublicKeyBytes(signer.Public())
	res := &session.ResponseVerificationHeader{
		MetaSignature:   &refs.Signature{Key: pub, Sign: ms, Scheme: scheme},
		OriginSignature: &refs.Signature{Key: pub, Sign: vs, Scheme: scheme},
		Origin:          vhOriginal,
	}
	if signBody {
		res.BodySignature = &refs.Signature{Key: pub, Sign: bs, Scheme: scheme}
	}
	return res, nil
}

// VerifyResponseWithBuffer checks whether verification header of the response
// is formed according to the NeoFS API protocol.
//
// Buffer is optional and free after the call.
func VerifyResponseWithBuffer[B ProtoMessage](r SignedResponse[B], buf []byte) error {
	v := r.GetVerifyHeader()
	if v == nil {
		return errMissingVerifyHdr
	}

	b := r.GetBody()
	m := r.GetMetaHeader()
	bs := maxEncodedSize(b, m, v)
	mo, vo := m.GetOrigin(), v.GetOrigin()
	for {
		if (mo == nil) != (vo == nil) {
			return errWrongVerifyHdrNum
		}
		if vo == nil {
			break
		}
		if s := maxEncodedSize(mo, vo); s > bs {
			bs = s
		}
		mo, vo = mo.GetOrigin(), vo.GetOrigin()
	}

	if len(buf) < bs {
		buf = make([]byte, bs)
	}

	for i := uint(0); ; m, v, i = m.Origin, v.Origin, i+1 {
		if v.MetaSignature == nil {
			return newErrInvalidVerificationHeader(i, errMissingMetaSig)
		}
		if err := verifyMessageSignature(m, v.MetaSignature, buf); err != nil {
			return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidMetaSig, err))
		}
		if v.OriginSignature == nil {
			return newErrInvalidVerificationHeader(i, errMissingVerifyOriginSig)
		}
		if err := verifyMessageSignature(v.Origin, v.OriginSignature, buf); err != nil {
			return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidVerifyOriginSig, err))
		}
		if v.Origin == nil {
			if v.BodySignature == nil {
				return newErrInvalidVerificationHeader(i, errMissingBodySig)
			}
			if err := verifyMessageSignature(b, v.BodySignature, buf); err != nil {
				return newErrInvalidVerificationHeader(i, fmt.Errorf("%w: %w", errInvalidBodySig, err))
			}
			return nil
		}
		if v.BodySignature != nil {
			return newErrInvalidVerificationHeader(i, errNonOriginBodySig)
		}
	}
}

func verifyMessageSignature(m ProtoMessage, s *refs.Signature, b []byte) error {
	if len(s.Key) == 0 {
		return errors.New("missing public key")
	}
	if s.Scheme < 0 {
		return fmt.Errorf("negative scheme %d", s.Scheme)
	}
	pubKey, err := decodePublicKey(Scheme(s.Scheme), s.Key)
	if err != nil {
		return err
	}

	var sz int
	b, sz = encodeMessage(m, b)
	if !pubKey.Verify(b[:sz], s.Sign) {
		return errors.New("signature mismatch")
	}

	return nil
}

// marshals m into buffer and returns it. Second value means buffer len occupied
// for m.
func encodeMessage(m ProtoMessage, b []byte) ([]byte, int) {
	s := m.MarshaledSize()
	if len(b) < s {
		b = make([]byte, s)
	}
	m.MarshalStable(b)
	return b, s
}

func maxEncodedSize(ms ...ProtoMessage) int {
	res := ms[0].MarshaledSize()
	for _, m := range ms[1:] {
		if s := m.MarshaledSize(); s > res {
			res = s
		}
	}
	return res
}
