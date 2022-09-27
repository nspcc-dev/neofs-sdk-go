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

	// Creator's private key.
	Signer ecdsa.PrivateKey

	payload io.Reader

	attributes [][2]string

	idHandler func(oid.ID)
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
// The object is stored in the parameterized container (required). Signer parameter
// is used for object creation and communication over the NeoFS protocol (required).
// There are some optional parameters, see CreateObjectPrm methods for details.
//
// Client connection MUST BE opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also CreateObject.
func CreateObjectWithClient(ctx context.Context, c *client.Client, prm CreateObjectPrm) error {
	const expirationSession = math.MaxUint64

	// send request to open the session for object writing
	// FIXME: #342 avoid session opening and create object "statically"
	var prmSession client.PrmSessionCreate
	prmSession.SetExp(expirationSession)

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
	err = tokenSession.Sign(prm.Signer)
	if err != nil {
		return fmt.Errorf("sign session token: %w", err)
	}

	// initialize object stream
	var prmPutInit client.PrmObjectPutInit
	prmPutInit.WithinSession(tokenSession)

	streamObj, err := c.ObjectPutInit(ctx, prmPutInit)
	if err != nil {
		return fmt.Errorf("init object writing on client: %w", err)
	}

	var idCreator user.ID
	user.IDFromKey(&idCreator, prm.Signer.PublicKey)

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
// objects using the same endpoint, use CreateObjectWithClient.
func CreateObject(ctx context.Context, endpoint string, prm CreateObjectPrm) error {
	var prmInit client.PrmInit
	prmInit.SetDefaultPrivateKey(prm.Signer)
	prmInit.ResolveNeoFSFailures()

	var c client.Client
	c.Init(prmInit)

	var prmDial client.PrmDial
	prmDial.SetServerURI(endpoint)

	err := c.Dial(prmDial)
	if err != nil {
		return fmt.Errorf("endpoint dial: %w", err)
	}

	return CreateObjectWithClient(ctx, &c, prm)
}
