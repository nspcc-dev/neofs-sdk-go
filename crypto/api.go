package neofscrypto

import (
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/api/session"
	"github.com/nspcc-dev/neofs-sdk-go/internal/proto"
)

// Request is a common interface of NeoFS API requests.
type Request interface {
	// GetMetaHeader returns meta header attached to the [Request].
	GetMetaHeader() *session.RequestMetaHeader
	// GetVerifyHeader returns verification header of the [Request].
	GetVerifyHeader() *session.RequestVerificationHeader
}

// Response is a common interface of NeoFS API responses.
type Response interface {
	// GetMetaHeader returns meta header attached to the [Response].
	GetMetaHeader() *session.ResponseMetaHeader
	// GetVerifyHeader returns verification header of the [Response].
	GetVerifyHeader() *session.ResponseVerificationHeader
}

// SignRequest signs all verified parts of the request using provided
// [neofscrypto.Signer] and returns resulting verification header. Meta header
// must be set in advance if needed.
//
// Optional buffer is used for encoding if it has sufficient size.
func SignRequest(signer Signer, req Request, body proto.Message, buf []byte) (*session.RequestVerificationHeader, error) {
	var bodySig []byte
	var err error
	originVerifyHdr := req.GetVerifyHeader()
	if originVerifyHdr == nil {
		// sign session message body
		b := encodeMessage(body, buf)
		bodySig, err = signer.Sign(b)
		if err != nil {
			return nil, fmt.Errorf("sign body: %w", err)
		}
		if len(b) > len(buf) {
			buf = b
		}
	}

	// sign meta header
	b := encodeMessage(req.GetMetaHeader(), buf)
	metaSig, err := signer.Sign(b)
	if err != nil {
		return nil, fmt.Errorf("sign meta header: %w", err)
	}
	if len(b) > len(buf) {
		buf = b
	}

	// sign verification header origin
	b = encodeMessage(originVerifyHdr, buf)
	verifyOriginSig, err := signer.Sign(b)
	if err != nil {
		return nil, fmt.Errorf("sign origin of verification header: %w", err)
	}

	scheme := refs.SignatureScheme(signer.Scheme())
	pubKey := PublicKeyBytes(signer.Public())
	res := &session.RequestVerificationHeader{
		MetaSignature:   &refs.Signature{Key: pubKey, Sign: metaSig, Scheme: scheme},
		OriginSignature: &refs.Signature{Key: pubKey, Sign: verifyOriginSig, Scheme: scheme},
		Origin:          originVerifyHdr,
	}
	if originVerifyHdr == nil {
		res.BodySignature = &refs.Signature{Key: pubKey, Sign: bodySig, Scheme: scheme}
	}
	return res, nil
}

// VerifyRequest verifies all signatures of given request and its extracted
// body.
func VerifyRequest(req Request, body proto.Message) error {
	verifyHdr := req.GetVerifyHeader()
	if verifyHdr == nil {
		return errors.New("missing verification header")
	}

	// pre-calculate max encoded size to allocate single buffer
	maxSz := body.MarshaledSize()
	metaHdr := req.GetMetaHeader()
	for {
		if metaHdr.GetOrigin() == nil != (verifyHdr.Origin == nil) { // metaHdr can be nil, verifyHdr cannot
			return errors.New("different number of meta and verification headers")
		}

		sz := verifyHdr.MarshaledSize()
		if sz > maxSz {
			maxSz = sz
		}
		sz = metaHdr.MarshaledSize()
		if sz > maxSz {
			maxSz = sz
		}

		if verifyHdr.Origin == nil {
			break
		}
		verifyHdr = verifyHdr.Origin
		metaHdr = metaHdr.Origin
	}

	var err error
	var bodySig *refs.Signature
	metaHdr = req.GetMetaHeader()
	verifyHdr = req.GetVerifyHeader()
	buf := make([]byte, maxSz)
	for {
		if verifyHdr.MetaSignature == nil {
			return errors.New("missing signature of the meta header")
		}
		if err = verifyMessageSignature(metaHdr, verifyHdr.MetaSignature, buf); err != nil {
			return fmt.Errorf("verify signature of the meta header: %w", err)
		}
		if verifyHdr.OriginSignature == nil {
			return errors.New("missing signature of the origin verification header")
		}
		if err = verifyMessageSignature(verifyHdr.Origin, verifyHdr.OriginSignature, buf); err != nil {
			return fmt.Errorf("verify signature of the origin verification header: %w", err)
		}

		if verifyHdr.Origin == nil {
			bodySig = verifyHdr.BodySignature
			break
		}

		if verifyHdr.BodySignature != nil {
			return errors.New("body signature is set for non-origin level")
		}

		verifyHdr = verifyHdr.Origin
		metaHdr = metaHdr.Origin
	}

	if bodySig == nil {
		return errors.New("missing body signature")
	}
	if err = verifyMessageSignature(body, bodySig, buf); err != nil {
		return fmt.Errorf("verify body signature: %w", err)
	}
	return nil
}

// SignResponse signs all verified parts of the response using provided
// [neofscrypto.Signer] and returns resulting verification header. Meta header
// must be set in advance if needed.
//
// Optional buffer is used for encoding if it has sufficient size.
func SignResponse(signer Signer, resp Response, body proto.Message, buf []byte) (*session.ResponseVerificationHeader, error) {
	var bodySig []byte
	var err error
	originVerifyHdr := resp.GetVerifyHeader()
	if originVerifyHdr == nil {
		// sign session message body
		b := encodeMessage(body, buf)
		bodySig, err = signer.Sign(b)
		if err != nil {
			return nil, fmt.Errorf("sign body: %w", err)
		}
		if len(b) > len(buf) {
			buf = b
		}
	}

	// sign meta header
	b := encodeMessage(resp.GetMetaHeader(), buf)
	metaSig, err := signer.Sign(b)
	if err != nil {
		return nil, fmt.Errorf("sign meta header: %w", err)
	}
	if len(b) > len(buf) {
		buf = b
	}

	// sign verification header origin
	b = encodeMessage(originVerifyHdr, buf)
	verifyOriginSig, err := signer.Sign(b)
	if err != nil {
		return nil, fmt.Errorf("sign origin of verification header: %w", err)
	}

	scheme := refs.SignatureScheme(signer.Scheme())
	pubKey := PublicKeyBytes(signer.Public())
	res := &session.ResponseVerificationHeader{
		MetaSignature:   &refs.Signature{Key: pubKey, Sign: metaSig, Scheme: scheme},
		OriginSignature: &refs.Signature{Key: pubKey, Sign: verifyOriginSig, Scheme: scheme},
		Origin:          originVerifyHdr,
	}
	if originVerifyHdr == nil {
		res.BodySignature = &refs.Signature{Key: pubKey, Sign: bodySig, Scheme: scheme}
	}
	return res, nil
}

// VerifyResponse verifies all signatures of given response and its extracted
// body.
func VerifyResponse(resp Response, body proto.Message) error {
	verifyHdr := resp.GetVerifyHeader()
	if verifyHdr == nil {
		return errors.New("missing verification header")
	}

	// pre-calculate max encoded size to allocate single buffer
	maxSz := body.MarshaledSize()
	metaHdr := resp.GetMetaHeader()
	for {
		if metaHdr.GetOrigin() == nil != (verifyHdr.Origin == nil) { // metaHdr can be nil, verifyHdr cannot
			return errors.New("different number of meta and verification headers")
		}

		sz := verifyHdr.MarshaledSize()
		if sz > maxSz {
			maxSz = sz
		}
		sz = metaHdr.MarshaledSize()
		if sz > maxSz {
			maxSz = sz
		}

		if verifyHdr.Origin == nil {
			break
		}
		verifyHdr = verifyHdr.Origin
		metaHdr = metaHdr.Origin
	}

	var err error
	var bodySig *refs.Signature
	metaHdr = resp.GetMetaHeader()
	verifyHdr = resp.GetVerifyHeader()
	buf := make([]byte, maxSz)
	for {
		if verifyHdr.MetaSignature == nil {
			return errors.New("missing signature of the meta header")
		}
		if err = verifyMessageSignature(metaHdr, verifyHdr.MetaSignature, buf); err != nil {
			return fmt.Errorf("verify signature of the meta header: %w", err)
		}
		if verifyHdr.OriginSignature == nil {
			return errors.New("missing signature of the origin verification header")
		}
		if err = verifyMessageSignature(verifyHdr.Origin, verifyHdr.OriginSignature, buf); err != nil {
			return fmt.Errorf("verify signature of the origin verification header: %w", err)
		}

		if verifyHdr.Origin == nil {
			bodySig = verifyHdr.BodySignature
			break
		}

		if verifyHdr.BodySignature != nil {
			return errors.New("body signature is set for non-origin level")
		}

		verifyHdr = verifyHdr.Origin
		metaHdr = metaHdr.Origin
	}

	if bodySig == nil {
		return errors.New("missing body signature")
	}
	if err = verifyMessageSignature(body, bodySig, buf); err != nil {
		return fmt.Errorf("verify body signature: %w", err)
	}
	return nil
}

func verifyMessageSignature(msg proto.Message, sig *refs.Signature, buf []byte) error {
	if len(sig.Key) == 0 {
		return errors.New("missing public key")
	} else if sig.Scheme < 0 {
		return fmt.Errorf("invalid scheme %d", sig.Scheme)
	}

	pubKey, err := decodePublicKey(Scheme(sig.Scheme), sig.Key)
	if err != nil {
		return err
	}

	if !pubKey.Verify(encodeMessage(msg, buf), sig.Sign) {
		return errors.New("signature mismatch")
	}
	return nil
}

func encodeMessage(m proto.Message, buf []byte) []byte {
	sz := m.MarshaledSize()
	var b []byte
	if len(buf) >= sz {
		b = buf[:sz]
	} else {
		b = make([]byte, sz)
	}
	m.MarshalStable(b)
	return b
}
