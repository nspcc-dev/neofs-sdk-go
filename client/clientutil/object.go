package clientutil

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// CreateObjectPrm groups parameters of CreateObject operation.
type CreateObjectPrm struct {
	// Target NeoFS container.
	Container cid.ID

	signerSet bool
	signer    ecdsa.PrivateKey

	payload io.Reader

	attributes [][2]string

	idHandler func(oid.ID)
}

// SetSigner specifies optional signing component. Signer MUST be correctly
// initialized.
func (x *CreateObjectPrm) SetSigner(signer ecdsa.PrivateKey) {
	x.signer = signer
	x.signerSet = true
}

// AddAttribute adds optional key-value attribute to be assigned to an object.
// Can be called multiple times.
//
// Both key and value MUST NOT be empty.
func (x *CreateObjectPrm) AddAttribute(key, val string) {
	x.attributes = append(x.attributes, [2]string{key, val})
}

// SetPayload sets optional object payload encapsulated in io.Reader provider.
//
// Reader SHOULD NOT be nil.
func (x *CreateObjectPrm) SetPayload(r io.Reader) {
	x.payload = r
}

// SetIDHandler sets optional function to pass the identifier of the stored object.
//
// Handler SHOULD NOT be nil.
func (x *CreateObjectPrm) SetIDHandler(f func(oid.ID)) {
	x.idHandler = f
}

// CreateObjectWithClient creates new NeoFS object and stores it into the
// NeoFS network using the given client.
//
// The object is stored in the container referenced by CreateObjectPrm.Container
// which MUST be explicitly set.
//
// Container SHOULD be public-write or sender SHOULD have corresponding rights.
// CreateObjectWithClient uses random private key for object creation and
// communication over the NeoFS protocol. This is suitable for working with
// public-write containers or in the absence of a specific key. To explicitly
// specify the signer, use CreateObjectPrm.SetSigner method.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also CreateObject.
func CreateObjectWithClient(ctx context.Context, c *client.Client, prm CreateObjectPrm) error {
	const expirationSession = math.MaxUint64

	var signer ecdsa.PrivateKey
	if prm.signerSet {
		signer = prm.signer
	} else {
		signer = signerDefault
	}

	// send request to open the session for object writing
	// FIXME: #342 avoid session opening and create object "statically"
	var prmSession client.PrmSessionCreate
	prmSession.SetExp(expirationSession)
	prmSession.UseKey(signer)

	resSession, err := c.SessionCreate(ctx, prmSession)
	if err != nil {
		return fmt.Errorf("open session with the remote node: %w", err)
	}

	// decode session ID
	var idSession uuid.UUID

	err = idSession.UnmarshalBinary(resSession.ID())
	if err != nil {
		return fmt.Errorf("invalid session ID in session response: %w", err)
	}

	// decode session public key
	var keySession neofsecdsa.PublicKey

	err = keySession.Decode(resSession.PublicKey())
	if err != nil {
		return fmt.Errorf("invalid session public key in session response: %w", err)
	}

	// form token of the object session
	var tokenSession session.Object
	tokenSession.SetID(idSession)
	tokenSession.SetExp(expirationSession)
	tokenSession.BindContainer(prm.Container)
	tokenSession.ForVerb(session.VerbObjectPut)
	tokenSession.SetAuthKey(&keySession)

	// sign the session token
	err = tokenSession.Sign(signer)
	if err != nil {
		return fmt.Errorf("sign session token: %w", err)
	}

	// initialize object stream
	var prmPutInit client.PrmObjectPutInit
	prmPutInit.WithinSession(tokenSession)
	prmPutInit.UseKey(signer)

	streamObj, err := c.ObjectPutInit(ctx, prmPutInit)
	if err != nil {
		return fmt.Errorf("init object writing on client: %w", err)
	}

	var idCreator user.ID
	user.IDFromKey(&idCreator, signer.PublicKey)

	// form the minimum required object structure
	var obj object.Object
	obj.SetContainerID(prm.Container)
	obj.SetOwnerID(&idCreator)

	// add attributes
	if prm.attributes != nil {
		attributes := make([]object.Attribute, len(prm.attributes))

		for i := range prm.attributes {
			attributes[i].SetKey(prm.attributes[i][0])
			attributes[i].SetValue(prm.attributes[i][1])
		}

		obj.SetAttributes(attributes...)
	}

	// write header
	if streamObj.WriteHeader(obj) && prm.payload != nil { // do not swap the conditions!
		var n int
		buf := make([]byte, 100<<10)

		// write payload
		for {
			n, err = prm.payload.Read(buf)
			if n > 0 {
				if !streamObj.WritePayloadChunk(buf[:n]) {
					break
				}

				continue
			}

			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read payload: %w", err)
		}
	}

	res, err := streamObj.Close()
	if err != nil {
		return fmt.Errorf("write object: %w", err)
	}

	if prm.idHandler != nil {
		prm.idHandler(res.StoredObjectID())
	}

	return nil
}

// CreateObject creates new NeoFS object and stores it into the NeoFS network
// through the given endpoint.
//
// CreateObject is well suited for one-time data storage. To create multiple
// objects using the same endpoint, use CreateObjectWithClient. CreateObject
// inherits behavior of CreateObjectWithClient.
func CreateObject(ctx context.Context, endpoint string, prm CreateObjectPrm) error {
	c, err := createClient(endpoint)
	if err != nil {
		return err
	}

	return CreateObjectWithClient(ctx, c, prm)
}

// ReadObjectPrm groups parameters of ReadObject operation.
type ReadObjectPrm struct {
	// Target NeoFS container.
	Container cid.ID

	// Object's reference to read from the Container.
	Object oid.ID

	payloadWriter io.Writer
}

// WritePayloadTo sets optional io.Writer to write payload of the request
// object to.
func (x *ReadObjectPrm) WritePayloadTo(w io.Writer) {
	x.payloadWriter = w
}

// ReadObjectWithClient reads object from the NeoFS network using the given
// client.
//
// The object is read from the container referenced by ReadObjectPrm.Container
// which MUST be explicitly set. Exact object is referenced by
// ReadObjectPrm.Object which MUST be explicitly set.
//
// Container SHOULD be public-read. ReadObjectWithClient uses random private key
// for communication over the NeoFS protocol. This is also suitable in the absence
// of a specific key.
//
// By default, ReadObjectWithClient reads but discards the object data. It can
// be used to check object availability. To explicitly specify payload target,
// use ReadObjectPrm.WritePayloadTo method.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also ReadObject.
func ReadObjectWithClient(ctx context.Context, c *client.Client, prm ReadObjectPrm) error {
	// initialize object stream
	var prmGet client.PrmObjectGet
	prmGet.FromContainer(prm.Container)
	prmGet.ByID(prm.Object)
	prmGet.UseKey(signerDefault)

	streamObj, err := c.ObjectGetInit(ctx, prmGet)
	if err != nil {
		return fmt.Errorf("init object writing on client: %w", err)
	}

	// read and discard header
	if streamObj.ReadHeader(new(object.Object)) {
		if prm.payloadWriter == nil {
			// discard payload by default
			prm.payloadWriter = io.Discard
		}

		// copy payload to the destination
		_, err = io.Copy(prm.payloadWriter, streamObj)
		if err != nil {
			return err
		}
	}

	_, err = streamObj.Close()
	if err != nil {
		return fmt.Errorf("read object: %w", err)
	}

	return nil
}

// ReadObject reads object from the NeoFS network through the given endpoint.
//
// ReadObject is well suited for one-time data reading. To read multiple
// objects using the same endpoint, use ReadObjectWithClient. ReadObject
// inherits behavior of ReadObjectWithClient.
func ReadObject(ctx context.Context, endpoint string, prm ReadObjectPrm) error {
	c, err := createClient(endpoint)
	if err != nil {
		return err
	}

	return ReadObjectWithClient(ctx, c, prm)
}

// RemoveObjectPrm groups parameters of RemoveObject operation.
type RemoveObjectPrm struct {
	// Target NeoFS container.
	Container cid.ID

	// Reference to the object to be removed from the Container.
	Object oid.ID
}

// RemoveObjectWithClient removes object from the NeoFS network using the given
// client. Successful RemoveObjectWithClient does not guarantee synchronous
// physical removal: object becomes unavailable, but can be purged later.
//
// The object is removed from the container referenced by RemoveObjectPrm.Container
// which MUST be explicitly set. Exact object is referenced by RemoveObjectPrm.Object
// which MUST be explicitly set.
//
// Container SHOULD be public-write. RemoveObjectWithClient uses random private
// key for removal witness and communication over the NeoFS protocol. This is
// also suitable in the absence of a specific key.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also RemoveObject.
func RemoveObjectWithClient(ctx context.Context, c *client.Client, prm RemoveObjectPrm) error {
	var prmDel client.PrmObjectDelete
	prmDel.FromContainer(prm.Container)
	prmDel.ByID(prm.Object)
	prmDel.UseKey(signerDefault)

	_, err := c.ObjectDelete(ctx, prmDel)
	if err != nil {
		return fmt.Errorf("remove object via client: %w", err)
	}

	return nil
}

// RemoveObject removes object from the NeoFS network through the given endpoint.
// Successful RemoveObject does not guarantee synchronous physical removal:
// object becomes unavailable, but can be purged later.
//
// RemoveObject is well suited for one-time data removal. To delete multiple
// objects using the same endpoint, use RemoveObjectWithClient. RemoveObject
// inherits behavior of RemoveObjectWithClient.
func RemoveObject(ctx context.Context, endpoint string, prm RemoveObjectPrm) error {
	c, err := createClient(endpoint)
	if err != nil {
		return err
	}

	return RemoveObjectWithClient(ctx, c, prm)
}
