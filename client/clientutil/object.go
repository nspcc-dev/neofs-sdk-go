package clientutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
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

	signer neofscrypto.Signer

	payload io.Reader

	attributes [][2]string

	idHandler func(oid.ID)
}

func NewCreateObjectPrm(container cid.ID, payload io.Reader) *CreateObjectPrm {
	return &CreateObjectPrm{
		Container: container,
		payload:   payload,
	}
}

// SetSigner specifies optional signing component. Signer MUST be correctly
// initialized.
func (x *CreateObjectPrm) SetSigner(signer neofscrypto.Signer) *CreateObjectPrm {
	x.signer = signer
	return x
}

// AddAttribute adds optional key-value attribute to be assigned to an object.
// Can be called multiple times.
//
// Both key and value MUST NOT be empty.
func (x *CreateObjectPrm) AddAttribute(key, val string) *CreateObjectPrm {
	x.attributes = append(x.attributes, [2]string{key, val})
	return x
}

// SetPayload sets optional object payload encapsulated in io.Reader provider.
//
// Reader SHOULD NOT be nil.
func (x *CreateObjectPrm) SetPayload(r io.Reader) *CreateObjectPrm {
	x.payload = r
	return x
}

// SetIDHandler sets optional function to pass the identifier of the stored object.
//
// Handler SHOULD NOT be nil.
func (x *CreateObjectPrm) SetIDHandler(f func(oid.ID)) *CreateObjectPrm {
	x.idHandler = f
	return x
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
// By default, object is created without payload. Use CreateObjectPrm.SetPayload
// to specify the data source.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also CreateObject.
func CreateObjectWithClient(ctx context.Context, c *client.Client, prm CreateObjectPrm) error {
	const expirationSession = math.MaxUint64

	signer := signerDefault
	if prm.signer != nil {
		signer = prm.signer
	}

	// send request to open the session for object writing
	// FIXME: #342 avoid session opening and create object "statically"
	var prmSession client.PrmSessionCreate
	prmSession.SetExp(expirationSession)
	prmSession.UseSigner(signer)

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
	prmPutInit.UseSigner(signer)

	streamObj, err := c.ObjectPutInit(ctx, prmPutInit)
	if err != nil {
		return fmt.Errorf("init object writing on client: %w", err)
	}

	var idCreator user.ID
	if err = user.IDFromSigner(&idCreator, signer); err != nil {
		return fmt.Errorf("IDFromSigner: %w", err)
	}

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
	prmGet.UseSigner(signerDefault)

	streamObj, err := c.ObjectGetInit(ctx, prm.Container, prm.Object, prmGet)
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

	if err = streamObj.Close(); err != nil {
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
	prmDel.UseSigner(signerDefault)

	_, err := c.ObjectDelete(ctx, prm.Container, prm.Object, prmDel)
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

type searchQuery = func(*object.SearchFilters)

func queryFileName(name string) searchQuery {
	return func(fs *object.SearchFilters) {
		fs.AddFilter(object.AttributeFileName, name, object.MatchStringEqual)
	}
}

func selectObjectsWithClient(ctx context.Context, c *client.Client, cnr cid.ID, query searchQuery, handler func(oid.ID) bool) error {
	var prm client.PrmObjectSearch
	prm.UseSigner(signerDefault)

	if query != nil {
		var filters object.SearchFilters
		query(&filters)

		prm.SetFilters(filters)
	}

	stream, err := c.ObjectSearchInit(ctx, cnr, prm)
	if err != nil {
		return fmt.Errorf("search objects via client: %w", err)
	}

	err = stream.Iterate(func(id oid.ID) bool {
		if handler != nil {
			handler(id)
		}

		return false
	})

	return err
}

func selectAllObjectsWithClient(ctx context.Context, c *client.Client, cnr cid.ID, handler func(oid.ID)) error {
	return selectObjectsWithClient(ctx, c, cnr, nil, func(id oid.ID) bool {
		if handler != nil {
			handler(id)
		}

		return false
	})
}

func selectObjects(ctx context.Context, endpoint string, cnr cid.ID, query searchQuery, handler func(oid.ID) bool) error {
	c, err := createClient(endpoint)
	if err != nil {
		return err
	}

	return selectObjectsWithClient(ctx, c, cnr, query, handler)
}

func selectAllObjects(ctx context.Context, endpoint string, cnr cid.ID, handler func(id oid.ID)) error {
	return selectObjects(ctx, endpoint, cnr, nil, func(id oid.ID) bool {
		if handler != nil {
			handler(id)
		}

		return false
	})
}

// ListObjectsPrm groups parameters of ListObjects operation.
type ListObjectsPrm struct {
	// Target NeoFS container.
	Container cid.ID

	handler func(oid.ID)
}

// SetHandler sets optional handler to pass the list elements. Handler
// SHOULD NOT be nil.
func (x *ListObjectsPrm) SetHandler(f func(id oid.ID)) {
	x.handler = f
}

// ListObjectsWithClient reads set of all container objects from the NeoFS
// network using the given client.
//
// Objects are listed in the container referenced by ListObjectsPrm.Container
// which MUST be explicitly set.
//
// ListObjectsWithClient does not return error if no objects are found in the
// container. By default, ListObjectsWithClient discards all found objects.
// This can be useful to check container searching operability. To process
// listed objects, use ListObjectsPrm.SetHandler method.
//
// Container SHOULD be public-read. ListObjectsWithClient uses random private
// key for communication over the NeoFS protocol. This is also suitable in the
// absence of a specific key.
//
// Client connection MUST be opened in advance, see Dial method for details.
// Network communication is carried out within a given context, so it MUST NOT
// be nil.
//
// See also ListObjects.
func ListObjectsWithClient(ctx context.Context, c *client.Client, prm ListObjectsPrm) error {
	return selectAllObjectsWithClient(ctx, c, prm.Container, prm.handler)
}

// ListObjects reads set of all container objects from the NeoFS network
// through the given endpoint.
//
// RemoveObject is well suited for one-time data removal. To delete multiple
// objects using the same endpoint, use ListObjectsWithClient. ListObjects
// inherits behavior of ListObjectsWithClient.
func ListObjects(ctx context.Context, endpoint string, prm ListObjectsPrm) error {
	return selectAllObjects(ctx, endpoint, prm.Container, prm.handler)
}

// UploadFilePrm groups parameters of UploadFile operation.
type UploadFilePrm struct {
	// Target NeoFS container.
	Container cid.ID

	// Associated file name.
	Name string

	createPrm CreateObjectPrm
}

// SetFileData specifies optional file data source. Reader SHOULD NOT be nil.
func (x *UploadFilePrm) SetFileData(r io.Reader) {
	x.createPrm.SetPayload(r)
}

// UploadFile uploads file into the NeoFS network through the given endpoint.
//
// New object is created in the container referenced by UploadFilePrm.Container
// which MUST be explicitly set. Container MUST be public-write. The new object
// is associated with the file by file name which MUST NOT be empty.
//
// By default, objects corresponds to empty file. Use UploadFilePrm.SetFileData
// to specify the file data source.
//
// See also DownloadFile.
func UploadFile(ctx context.Context, endpoint string, prm UploadFilePrm) error {
	if prm.Name == "" {
		panic("empty file name")
	}

	prm.createPrm.Container = prm.Container
	prm.createPrm.AddAttribute(object.AttributeFileName, prm.Name)

	return CreateObject(ctx, endpoint, prm.createPrm)
}

// UploadOpenedFile is a helping wrapper over UploadFile which processes os.File.
// File MUST be correctly opened.
func UploadOpenedFile(ctx context.Context, endpoint string, cnr cid.ID, f *os.File) error {
	prm := UploadFilePrm{
		Container: cnr,
		Name:      f.Name(),
	}

	prm.SetFileData(f)

	return UploadFile(ctx, endpoint, prm)
}

// UploadFileByPath is a helping wrapper over UploadOpenedFile which preliminary
// opens a file.
func UploadFileByPath(ctx context.Context, endpoint string, cnr cid.ID, fPath string) error {
	f, err := os.Open(fPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	return UploadOpenedFile(ctx, endpoint, cnr, f)
}

// DownloadFilePrm groups parameters of DownloadFile operation.
type DownloadFilePrm struct {
	// Target NeoFS container.
	Container cid.ID

	// Associated file name.
	Name string

	readPrm ReadObjectPrm
}

// WriteFileTo specifies optional destination of the file data. Writer SHOULD
// NOT be nil.
func (x *DownloadFilePrm) WriteFileTo(w io.Writer) {
	x.readPrm.WritePayloadTo(w)
}

// DownloadFile downloads file from the NeoFS network through the given endpoint.
//
// The associated object is read from the container referenced by
// DownloadFilePrm.Container which MUST be explicitly set. Container MUST be
// public-read. The exact objects is selected by associated file name specified
// in DownloadFilePrm.Name which MUST NOT be empty. If there is no object
// associated with the file name, fs.ErrNotExist returns. If there are multiple
// objects associated with the file name, DownloadFile fails.
//
// By default, DownloadFile discards file data. Use DownloadFilePrm.WriteFileTo
// to specify the file data destination.
//
// See also UploadFile.
func DownloadFile(ctx context.Context, endpoint string, prm DownloadFilePrm) error {
	if prm.Name == "" {
		panic("empty file name")
	}

	count := 0

	err := selectObjects(ctx, endpoint, prm.Container, queryFileName(prm.Name), func(id oid.ID) bool {
		prm.readPrm.Object = id
		count++
		return count > 1
	})
	if err != nil {
		return fmt.Errorf("select object by file name: %w", err)
	} else if count == 0 {
		return fs.ErrNotExist
	} else if count > 1 {
		return errors.New("multiple match")
	}

	prm.readPrm.Container = prm.Container

	return ReadObject(ctx, endpoint, prm.readPrm)
}

// RestoreFile is a helping wrapper over DownloadFile which processes os.File.
// File MUST be correctly opened.
func RestoreFile(ctx context.Context, endpoint string, cnr cid.ID, f *os.File) error {
	prm := DownloadFilePrm{
		Container: cnr,
		Name:      f.Name(),
	}

	prm.WriteFileTo(f)

	return DownloadFile(ctx, endpoint, prm)
}

// RestoreFileByPath is a helping wrapper over RestoreFile which preliminary
// creates or truncates the file. RestoreFileByPath does not remove the created
// files if download fails.
func RestoreFileByPath(ctx context.Context, endpoint string, cnr cid.ID, fPath string) error {
	f, err := os.Create(fPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	return RestoreFile(ctx, endpoint, cnr, f)
}
