package object

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	signatureV2 "github.com/nspcc-dev/neofs-api-go/v2/signature"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	sigutil "github.com/nspcc-dev/neofs-sdk-go/util/signature"
)

var (
	errCheckSumMismatch = errors.New("payload checksum mismatch")
	errCheckSumNotSet   = errors.New("payload checksum is not set")
	errIncorrectID      = errors.New("incorrect object identifier")
)

// CalculatePayloadChecksum calculates and returns checksum of
// object payload bytes.
func CalculatePayloadChecksum(payload []byte) checksum.Checksum {
	var res checksum.Checksum
	checksum.Calculate(&res, checksum.SHA256, payload)

	return res
}

// CalculateAndSetPayloadChecksum calculates checksum of current
// object payload and writes it to the object.
func CalculateAndSetPayloadChecksum(obj *Object) {
	obj.SetPayloadChecksum(
		CalculatePayloadChecksum(obj.Payload()),
	)
}

// VerifyPayloadChecksum checks if payload checksum in the object
// corresponds to its payload.
func VerifyPayloadChecksum(obj *Object) error {
	actual := CalculatePayloadChecksum(obj.Payload())

	cs, set := obj.PayloadChecksum()
	if !set {
		return errCheckSumNotSet
	}

	if !bytes.Equal(cs.Value(), actual.Value()) {
		return errCheckSumMismatch
	}

	return nil
}

// CalculateID calculates identifier for the object.
func CalculateID(obj *Object) (oid.ID, error) {
	data, err := obj.ToV2().GetHeader().StableMarshal(nil)
	if err != nil {
		return oid.ID{}, err
	}

	var id oid.ID
	id.SetSHA256(sha256.Sum256(data))

	return id, nil
}

// CalculateAndSetID calculates identifier for the object
// and writes the result to it.
func CalculateAndSetID(obj *Object) error {
	id, err := CalculateID(obj)
	if err != nil {
		return err
	}

	obj.SetID(id)

	return nil
}

// VerifyID checks if identifier in the object corresponds to
// its structure.
func VerifyID(obj *Object) error {
	id, err := CalculateID(obj)
	if err != nil {
		return err
	}

	oID, set := obj.ID()
	if !set {
		return errOIDNotSet
	}

	if !id.Equals(oID) {
		return errIncorrectID
	}

	return nil
}

// CalculateAndSetSignature signs id with provided key and sets that signature to
// the object.
func CalculateAndSetSignature(key ecdsa.PrivateKey, obj *Object) error {
	oID, set := obj.ID()
	if !set {
		return errOIDNotSet
	}

	sig, err := oID.CalculateIDSignature(key)
	if err != nil {
		return err
	}

	obj.SetSignature(&sig)

	return nil
}

// VerifyIDSignature verifies object ID signature.
func (o *Object) VerifyIDSignature() bool {
	oID, set := o.ID()
	if !set {
		return false
	}

	var idV2 refs.ObjectID
	oID.WriteToV2(&idV2)

	sig := o.Signature()

	err := sigutil.VerifyData(
		signatureV2.StableMarshalerWrapper{
			SM: &idV2,
		},
		sig,
	)

	return err == nil
}

// SetIDWithSignature sets object identifier and signature.
func SetIDWithSignature(key ecdsa.PrivateKey, obj *Object) error {
	if err := CalculateAndSetID(obj); err != nil {
		return fmt.Errorf("could not set identifier: %w", err)
	}

	if err := CalculateAndSetSignature(key, obj); err != nil {
		return fmt.Errorf("could not set signature: %w", err)
	}

	return nil
}

// SetVerificationFields calculates and sets all verification fields of the object.
func SetVerificationFields(key ecdsa.PrivateKey, obj *Object) error {
	CalculateAndSetPayloadChecksum(obj)

	return SetIDWithSignature(key, obj)
}

// CheckVerificationFields checks all verification fields of the object.
func CheckVerificationFields(obj *Object) error {
	if err := CheckHeaderVerificationFields(obj); err != nil {
		return fmt.Errorf("invalid header structure: %w", err)
	}

	if err := VerifyPayloadChecksum(obj); err != nil {
		return fmt.Errorf("invalid payload checksum: %w", err)
	}

	return nil
}

var errInvalidSignature = errors.New("invalid signature")

// CheckHeaderVerificationFields checks all verification fields except payload.
func CheckHeaderVerificationFields(obj *Object) error {
	if !obj.VerifyIDSignature() {
		return errInvalidSignature
	}

	if err := VerifyID(obj); err != nil {
		return fmt.Errorf("invalid identifier: %w", err)
	}

	return nil
}
