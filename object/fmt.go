package object

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
)

// MaxHeaderLen is a maximum allowed length of binary object header to be
// created via NeoFS API protocol.
const MaxHeaderLen = 16 << 10

var (
	errCheckSumMismatch = errors.New("payload checksum mismatch")
	errCheckSumNotSet   = errors.New("payload checksum is not set")
	errIncorrectID      = errors.New("incorrect object identifier")
)

// CalculatePayloadChecksum calculates and returns checksum of
// object payload bytes.
func CalculatePayloadChecksum(payload []byte) checksum.Checksum {
	return checksum.NewSHA256(sha256.Sum256(payload))
}

// CalculateAndSetPayloadChecksum calculates checksum of current
// object payload and writes it to the object.
func (o *Object) CalculateAndSetPayloadChecksum() {
	o.SetPayloadChecksum(
		CalculatePayloadChecksum(o.Payload()),
	)
}

// VerifyPayloadChecksum checks if payload checksum in the object
// corresponds to its payload.
func (o *Object) VerifyPayloadChecksum() error {
	actual := CalculatePayloadChecksum(o.Payload())

	cs, set := o.PayloadChecksum()
	if !set {
		return errCheckSumNotSet
	}

	if !bytes.Equal(cs.Value(), actual.Value()) {
		return errCheckSumMismatch
	}

	return nil
}

// CalculateID calculates identifier for the object.
func (o *Object) CalculateID() (oid.ID, error) {
	return sha256.Sum256(o.ToV2().GetHeader().StableMarshal(nil)), nil
}

// CalculateAndSetID calculates identifier for the object
// and writes the result to it.
func (o *Object) CalculateAndSetID() error {
	id, err := o.CalculateID()
	if err != nil {
		return err
	}

	o.SetID(id)

	return nil
}

// VerifyID checks if identifier in the object corresponds to
// its structure.
func (o *Object) VerifyID() error {
	id, err := o.CalculateID()
	if err != nil {
		return err
	}

	oID := o.GetID()
	if oID.IsZero() {
		return oid.ErrZero
	}

	if id != oID {
		return errIncorrectID
	}

	return nil
}

// Sign signs object id with provided key and sets that signature to the object.
//
// See also [oid.ID.CalculateIDSignature].
func (o *Object) Sign(signer neofscrypto.Signer) error {
	oID := o.GetID()
	if oID.IsZero() {
		return oid.ErrZero
	}

	sig, err := oID.CalculateIDSignature(signer)
	if err != nil {
		return err
	}

	o.SetSignature(&sig)

	return nil
}

// SignedData returns actual payload to sign.
//
// See also [Object.Sign].
func (o *Object) SignedData() []byte {
	return o.GetID().Marshal()
}

// VerifySignature verifies object ID signature.
func (o *Object) VerifySignature() bool {
	m := (*object.Object)(o)

	sigV2 := m.GetSignature()
	if sigV2 == nil {
		return false
	}

	idV2 := m.GetObjectID()
	if idV2 == nil {
		return false
	}

	var sig neofscrypto.Signature

	return sig.ReadFromV2(*sigV2) == nil && sig.Verify(idV2.StableMarshal(nil))
}

// SetIDWithSignature sets object identifier and signature.
func (o *Object) SetIDWithSignature(signer neofscrypto.Signer) error {
	if err := o.CalculateAndSetID(); err != nil {
		return fmt.Errorf("could not set identifier: %w", err)
	}

	if err := o.Sign(signer); err != nil {
		return fmt.Errorf("could not set signature: %w", err)
	}

	return nil
}

// SetVerificationFields calculates and sets all verification fields of the object.
func (o *Object) SetVerificationFields(signer neofscrypto.Signer) error {
	o.CalculateAndSetPayloadChecksum()

	return o.SetIDWithSignature(signer)
}

// CheckVerificationFields checks all verification fields of the object.
func (o *Object) CheckVerificationFields() error {
	if err := o.CheckHeaderVerificationFields(); err != nil {
		return fmt.Errorf("invalid header structure: %w", err)
	}

	if err := o.VerifyPayloadChecksum(); err != nil {
		return fmt.Errorf("invalid payload checksum: %w", err)
	}

	return nil
}

var errInvalidSignature = errors.New("invalid signature")

// CheckHeaderVerificationFields checks all verification fields except payload.
func (o *Object) CheckHeaderVerificationFields() error {
	if !o.VerifySignature() {
		return errInvalidSignature
	}

	if err := o.VerifyID(); err != nil {
		return fmt.Errorf("invalid identifier: %w", err)
	}

	return nil
}
