package slicer

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

var (
	// ErrIncompleteHeader indicates some fields are missing in header.
	ErrIncompleteHeader = errors.New("incomplete header")
)

// ObjectWriter represents a virtual object recorder.
type ObjectWriter interface {
	// ObjectPutInit initializes and returns a stream of writable data associated
	// with the object according to its header. Provided header includes at least
	// container, owner and object ID fields. Header length is limited to
	// [object.MaxHeaderLen].
	//
	// Signer is required and must not be nil. The operation is executed on behalf of
	// the account corresponding to the specified Signer, which is taken into account, in particular, for access control.
	ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm client.PrmObjectPutInit) (client.ObjectWriter, error)
}

// NetworkedClient represents a virtual object recorder with possibility to get actual [netmap.NetworkInfo] data.
type NetworkedClient interface {
	ObjectWriter

	NetworkInfo(ctx context.Context, prm client.PrmNetworkInfo) (netmap.NetworkInfo, error)
}

// Slicer converts input raw data streams into NeoFS objects. Working Slicer
// must be constructed via New.
type Slicer struct {
	signer user.Signer

	w ObjectWriter

	opts Options

	hdr object.Object
}

// New constructs Slicer which writes sliced ready-to-go objects owned by
// particular user into the specified container using provided ObjectWriter.
// All objects are signed using provided neofscrypto.Signer.
//
// If ObjectWriter returns data streams which provide io.Closer, they are closed
// in Slicer.Slice after the payload of any object has been written. In this
// case, Slicer.Slice fails immediately on Close error.
//
// NetworkedClient parameter allows to extract all required network-depended information for default Slicer behavior
// tuning.
//
// If payload size limit is specified via Options.SetObjectPayloadLimit,
// outgoing objects has payload not bigger than the limit. NeoFS stores the
// corresponding value in the network configuration. Ignore this option if you
// don't (want to) have access to it. By default, single object is limited by
// 1MB. Slicer uses this value to enforce the maximum object payload size limit
// described in the NeoFS Specification. If the total amount of data exceeds the
// specified limit, Slicer applies the slicing algorithm described within the
// same specification. The outcome will be a group of "small" objects containing
// a chunk of data, as well as an auxiliary linking object. All derived objects
// are written to the parameterized ObjectWriter. If the amount of data is
// within the limit, one object is produced. Note that Slicer can write multiple
// objects, but returns the root object ID only.
//
// Parameter sessionToken may be nil, if no session is used.
func New(ctx context.Context, nw NetworkedClient, signer user.Signer, cnr cid.ID, owner user.ID, sessionToken *session.Object) (*Slicer, error) {
	ni, err := nw.NetworkInfo(ctx, client.PrmNetworkInfo{})
	if err != nil {
		return nil, fmt.Errorf("network info: %w", err)
	}

	opts := Options{
		objectPayloadLimit: ni.MaxObjectSize(),
		currentNeoFSEpoch:  ni.CurrentEpoch(),
		sessionToken:       sessionToken,
	}

	if !ni.HomomorphicHashingDisabled() {
		opts.CalculateHomomorphicChecksum()
	}

	var hdr object.Object
	hdr.SetContainerID(cnr)
	hdr.SetType(object.TypeRegular)
	hdr.SetOwnerID(&owner)
	hdr.SetCreationEpoch(ni.CurrentEpoch())
	hdr.SetSessionToken(sessionToken)
	return &Slicer{
		opts:   opts,
		w:      nw,
		signer: signer,
		hdr:    hdr,
	}, nil
}

// Put creates new NeoFS object from the input data stream, associates the
// object with the configured container and writes the object via underlying
// [ObjectWriter]. After a successful write, Put returns an [oid.ID] which is a
// unique reference to the object in the container. Put sets all required
// calculated fields like payload length, checksum, etc.
//
// Put allows you to specify [object.Attribute] parameters to be written to the
// resulting object's metadata. Keys SHOULD NOT start with system-reserved
// '__NEOFS__' prefix.
//
// See [New] for details.
func (x *Slicer) Put(ctx context.Context, data io.Reader, attrs []object.Attribute) (oid.ID, error) {
	x.hdr.SetAttributes(attrs...)
	return slice(ctx, x.w, x.hdr, data, x.signer, x.opts)
}

// InitPut works similar to [Slicer.Put] but provides [PayloadWriter] allowing
// the caller to write data himself.
func (x *Slicer) InitPut(ctx context.Context, attrs []object.Attribute) (*PayloadWriter, error) {
	x.hdr.SetAttributes(attrs...)
	return initPayloadStream(ctx, x.w, x.hdr, x.signer, x.opts)
}

// Put works similar to [Slicer.Put], but allows flexible configuration of object header.
// The method accepts [Options] for adjusting max object size, epoch, session token, etc.
func Put(ctx context.Context, ow ObjectWriter, header object.Object, signer user.Signer, data io.Reader, opts Options) (oid.ID, error) {
	return slice(ctx, ow, header, data, signer, opts)
}

// InitPut works similar to [slicer.Put], but provides [ObjectWriter] allowing
// the caller to write data himself.
func InitPut(ctx context.Context, ow ObjectWriter, header object.Object, signer user.Signer, opts Options) (*PayloadWriter, error) {
	return initPayloadStream(ctx, ow, header, signer, opts)
}

const defaultPayloadSizeLimit = 1 << 20

// childPayloadSizeLimit returns configured size limit of the child object's
// payload which defaults to 1MB.
func childPayloadSizeLimit(opts Options) uint64 {
	if opts.objectPayloadLimit > 0 {
		return opts.objectPayloadLimit
	}
	return defaultPayloadSizeLimit
}

func slice(ctx context.Context, ow ObjectWriter, header object.Object, data io.Reader, signer user.Signer, opts Options) (oid.ID, error) {
	var rootID oid.ID

	objectPayloadLimit := childPayloadSizeLimit(opts)

	var n int

	writer, err := initPayloadStream(ctx, ow, header, signer, opts)
	if err != nil {
		return rootID, fmt.Errorf("init writter: %w", err)
	}

	var buf []byte
	var buffered uint64

	for {
		buffered = writer.rootMeta.length
		if writer.withSplit {
			buffered = writer.childMeta.length
		}

		if buffered == objectPayloadLimit && uint64(len(writer.payloadBuffer)) <= objectPayloadLimit {
			// in this case, the read buffers are exhausted, and it is unclear whether there
			// will be more data. We need to know this right away, because the format of the
			// final objects depends on it. So read to temp buffer
			buf = []byte{0}
		} else {
			payloadBufferLen := uint64(len(writer.payloadBuffer))
			if payloadBufferLen > buffered {
				buf = writer.payloadBuffer[buffered:]
			} else {
				if writer.extraPayloadBuffer == nil {
					// TODO(#544): support external buffer pools
					writer.extraPayloadBuffer = make([]byte, writer.payloadSizeLimit-payloadBufferLen)
					buf = writer.extraPayloadBuffer
				} else {
					buf = writer.extraPayloadBuffer[buffered-payloadBufferLen:]
				}
			}
		}

		n, err = data.Read(buf)
		if n > 0 {
			if _, err = writer.Write(buf[:n]); err != nil {
				return oid.ID{}, err
			}
		}

		if err == nil {
			continue
		}

		if !errors.Is(err, io.EOF) {
			return rootID, fmt.Errorf("read payload chunk: %w", err)
		}

		if writer.payloadSizeFixed && writer.rootMeta.length < writer.payloadSize {
			return oid.ID{}, io.ErrUnexpectedEOF
		}

		if err = writer.Close(); err != nil {
			return rootID, fmt.Errorf("writer close: %w", err)
		}

		rootID = writer.ID()
		break
	}

	return rootID, nil
}

// headerData extract required fields from header, otherwise throw the error.
func headerData(header object.Object) (cid.ID, user.ID, error) {
	containerID, isSet := header.ContainerID()
	if !isSet {
		return cid.ID{}, user.ID{}, fmt.Errorf("container-id: %w", ErrIncompleteHeader)
	}

	owner := header.OwnerID()
	if owner == nil {
		return cid.ID{}, user.ID{}, fmt.Errorf("owner: %w", ErrIncompleteHeader)
	}

	return containerID, *owner, nil
}

func initPayloadStream(ctx context.Context, ow ObjectWriter, header object.Object, signer user.Signer, opts Options) (*PayloadWriter, error) {
	containerID, owner, err := headerData(header)
	if err != nil {
		return nil, err
	}

	var prm client.PrmObjectPutInit
	prm.SetCopiesNumber(opts.copiesNumber)

	if opts.sessionToken != nil {
		prm.WithinSession(*opts.sessionToken)
		header.SetSessionToken(opts.sessionToken)
		// session issuer is a container owner.
		issuer := opts.sessionToken.Issuer()
		owner = issuer
		header.SetOwnerID(&owner)
	} else if opts.bearerToken != nil {
		prm.WithBearerToken(*opts.bearerToken)
		// token issuer is a container owner.
		issuer := opts.bearerToken.Issuer()
		owner = issuer
		header.SetOwnerID(&owner)
	}

	header.SetCreationEpoch(opts.currentNeoFSEpoch)
	currentVersion := version.Current()
	header.SetVersion(&currentVersion)

	var stubObject object.Object
	stubObject.SetVersion(&currentVersion)
	stubObject.SetContainerID(containerID)
	stubObject.SetCreationEpoch(opts.currentNeoFSEpoch)
	stubObject.SetType(object.TypeRegular)
	stubObject.SetOwnerID(&owner)
	stubObject.SetSessionToken(opts.sessionToken)

	res := &PayloadWriter{
		ctx:              ctx,
		headerObject:     header,
		stream:           ow,
		signer:           signer,
		container:        containerID,
		owner:            owner,
		currentEpoch:     opts.currentNeoFSEpoch,
		sessionToken:     opts.sessionToken,
		rootMeta:         newDynamicObjectMetadata(opts.withHomoChecksum),
		childMeta:        newDynamicObjectMetadata(opts.withHomoChecksum),
		payloadSizeLimit: childPayloadSizeLimit(opts),
		payloadSizeFixed: opts.payloadSizeFixed,
		payloadSize:      opts.payloadSize,
		prmObjectPutInit: prm,
		stubObject:       &stubObject,
	}

	if res.payloadSizeFixed && res.payloadSize < res.payloadSizeLimit {
		res.payloadSizeLimit = res.payloadSize
	}

	res.payloadBuffer = opts.payloadBuffer
	res.rootMeta.reset()
	res.metaWriter = &res.rootMeta

	return res, nil
}

// PayloadWriter is a single-object payload stream provided by Slicer.
type PayloadWriter struct {
	ctx    context.Context
	stream ObjectWriter

	rootID       oid.ID
	headerObject object.Object

	signer       user.Signer
	container    cid.ID
	owner        user.ID
	currentEpoch uint64
	sessionToken *session.Object

	payloadBuffer      []byte
	extraPayloadBuffer []byte

	rootMeta  dynamicObjectMetadata
	childMeta dynamicObjectMetadata

	// max payload size of produced objects in bytes
	payloadSizeLimit uint64
	payloadSizeFixed bool
	payloadSize      uint64

	metaWriter io.Writer

	withSplit   bool
	firstObject *oid.ID

	writtenChildren  []object.MeasuredObject
	prmObjectPutInit client.PrmObjectPutInit
	stubObject       *object.Object
}

var errPayloadSizeExceeded = errors.New("payload size exceeded")

// Write writes next chunk of the object data. Concatenation of all chunks forms
// the payload of the final object. When the data is over, the PayloadWriter
// should be closed.
func (x *PayloadWriter) Write(chunk []byte) (int, error) {
	if len(chunk) == 0 {
		// not explicitly prohibited in the io.Writer documentation
		return 0, nil
	}

	if x.payloadSizeFixed && x.rootMeta.length+uint64(len(chunk)) > x.payloadSize {
		return 0, errPayloadSizeExceeded
	}

	buffered := x.rootMeta.length
	if x.withSplit {
		buffered = x.childMeta.length
	}

	if buffered+uint64(len(chunk)) <= x.payloadSizeLimit {
		// buffer data to produce as few objects as possible for better storage efficiency
		_, err := x.metaWriter.Write(chunk)
		if err != nil {
			return 0, err
		}

		var n int
		payloadBufferLen := uint64(len(x.payloadBuffer))
		if payloadBufferLen > buffered {
			n = copy(x.payloadBuffer[buffered:], chunk)
			if n == len(chunk) {
				return n, nil
			}

			chunk = chunk[n:]
			buffered += uint64(n)
		}

		if x.extraPayloadBuffer == nil {
			if x.payloadSizeFixed {
				// in this case x.payloadSizeLimit >= x.payloadSize
				x.extraPayloadBuffer = make([]byte, x.payloadSizeLimit)
			} else {
				// if here for the first time, then allocate the minimum buffer sufficient for
				// writing: user may do one Write followed by Close. In such cases there is no
				// point in allocating a buffer of payloadSizeLimit size.
				// TODO(#544): support external buffer pools
				x.extraPayloadBuffer = make([]byte, len(chunk))
			}
		} else if payloadBufferLen+uint64(len(x.extraPayloadBuffer)) < x.payloadSizeLimit {
			b := make([]byte, uint64(len(x.extraPayloadBuffer))+x.payloadSizeLimit-buffered)
			copy(b, x.extraPayloadBuffer)
			x.extraPayloadBuffer = b
		}

		return n + copy(x.extraPayloadBuffer[buffered-payloadBufferLen:], chunk), nil
	}

	// at this point there is enough data to flush the buffer by sending the next
	n := int(x.payloadSizeLimit - buffered)
	_, err := x.metaWriter.Write(chunk[:n])
	if err != nil {
		return 0, err
	}

	payloadBuffers := make([][]byte, 0, 3)
	if buffered > 0 {
		if len(x.payloadBuffer) > 0 {
			if uint64(len(x.payloadBuffer)) >= buffered {
				payloadBuffers = append(payloadBuffers, x.payloadBuffer[:buffered])
			} else {
				payloadBuffers = append(payloadBuffers, x.payloadBuffer, x.extraPayloadBuffer[:buffered-uint64(len(x.payloadBuffer))])
			}
		} else {
			payloadBuffers = append(payloadBuffers, x.extraPayloadBuffer[:buffered])
		}
	}

	if n > 0 {
		payloadBuffers = append(payloadBuffers, chunk[:n])
	}

	if !x.withSplit {
		// note: don't move next row, the value of this flag will be used inside writeIntermediateChild
		// to fill splitInfo in all child objects.
		x.withSplit = true

		err := x.writeIntermediateChild(x.ctx, x.rootMeta, payloadBuffers)
		if err != nil {
			return n, fmt.Errorf("write 1st child: %w", err)
		}

		x.metaWriter = io.MultiWriter(&x.rootMeta, &x.childMeta)
	} else {
		err := x.writeIntermediateChild(x.ctx, x.childMeta, payloadBuffers)
		if err != nil {
			return n, fmt.Errorf("write next child: %w", err)
		}
	}

	x.childMeta.reset()

	n2, err := x.Write(chunk[n:]) // here n > 0 so infinite recursion shouldn't occur

	return n + n2, err
}

// Close finalizes object with written payload data, saves the object and closes
// the stream. Reference to the stored object can be obtained by ID method.
func (x *PayloadWriter) Close() error {
	if x.payloadSizeFixed && x.rootMeta.length < x.payloadSize {
		return io.ErrUnexpectedEOF
	}

	buffered := x.rootMeta.length
	if x.withSplit {
		buffered = x.childMeta.length
	}

	var payloadBuffers [][]byte
	if buffered > 0 {
		if len(x.payloadBuffer) > 0 {
			if uint64(len(x.payloadBuffer)) >= buffered {
				payloadBuffers = [][]byte{x.payloadBuffer[:buffered]}
			} else {
				payloadBuffers = [][]byte{x.payloadBuffer, x.extraPayloadBuffer[:buffered-uint64(len(x.payloadBuffer))]}
			}
		} else {
			payloadBuffers = [][]byte{x.extraPayloadBuffer[:buffered]}
		}
	}

	if x.withSplit {
		return x.writeLastChild(x.ctx, x.childMeta, payloadBuffers, x.setID)
	}
	return x.writeLastChild(x.ctx, x.rootMeta, payloadBuffers, x.setID)
}

func (x *PayloadWriter) setID(id oid.ID) {
	x.rootID = id
}

// ID returns unique identifier of the stored object representing its reference
// in the configured container.
//
// ID MUST NOT be called before successful Close (undefined behavior otherwise).
func (x *PayloadWriter) ID() oid.ID {
	return x.rootID
}

// writeIntermediateChild writes intermediate split-chain element with specified
// dynamicObjectMetadata to the configured ObjectWriter.
func (x *PayloadWriter) writeIntermediateChild(ctx context.Context, meta dynamicObjectMetadata, payloadBuffers [][]byte) error {
	return x._writeChild(ctx, meta, payloadBuffers, false, nil)
}

// writeIntermediateChild writes last split-chain element with specified
// dynamicObjectMetadata to the configured ObjectWriter. If rootIDHandler is
// specified, ID of the resulting root object is passed into it.
func (x *PayloadWriter) writeLastChild(ctx context.Context, meta dynamicObjectMetadata, payloadBuffers [][]byte, rootIDHandler func(id oid.ID)) error {
	return x._writeChild(ctx, meta, payloadBuffers, true, rootIDHandler)
}

func (x *PayloadWriter) _writeChild(ctx context.Context, meta dynamicObjectMetadata, payloadBuffers [][]byte, last bool, rootIDHandler func(id oid.ID)) error {
	obj := *x.stubObject
	obj.SetSplitID(nil)
	obj.ResetPreviousID()
	obj.SetParent(nil)
	obj.ResetParentID()
	obj.SetSignature(nil)
	obj.ResetID()

	if x.withSplit {
		if x.firstObject == nil {
			// first child object, has parent header,
			// does not have split chain ID
			obj.SetParent(&x.headerObject)
		} else {
			// any non-first split object, has
			// the first child object ID as a
			// split chain identifier
			obj.SetFirstID(*x.firstObject)
		}
	}
	if len(x.writtenChildren) > 0 {
		obj.SetPreviousID(x.writtenChildren[len(x.writtenChildren)-1].ObjectID())
	}
	if last {
		rootID, err := flushObjectMetadata(x.signer, x.rootMeta, &x.headerObject)
		if err != nil {
			return fmt.Errorf("form root object: %w", err)
		}

		if rootIDHandler != nil {
			rootIDHandler(rootID)
		}

		if x.withSplit {
			obj.SetParentID(rootID)
			obj.SetParent(&x.headerObject)
		} else {
			obj = x.headerObject
		}
	}

	var id oid.ID
	var err error

	id, err = writeInMemObject(ctx, x.signer, x.stream, obj, payloadBuffers, meta, x.prmObjectPutInit)
	if err != nil {
		return fmt.Errorf("write formed object: %w", err)
	}

	if x.withSplit && x.firstObject == nil {
		x.firstObject = &id
	}

	var measuredObject object.MeasuredObject
	measuredObject.SetObjectSize(uint32(meta.length))
	measuredObject.SetObjectID(id)

	x.writtenChildren = append(x.writtenChildren, measuredObject)

	if x.withSplit && last {
		var linkObj object.Link
		linkObj.SetObjects(x.writtenChildren)
		obj.WriteLink(linkObj)

		obj.ResetPreviousID()
		// we reuse already written object, we should reset these fields, to eval them one more time in writeInMemObject.
		obj.ResetID()
		obj.SetSignature(nil)

		payload := obj.Payload()
		payloadAsBuffers := [][]byte{obj.Payload()}

		if uint64(len(payload)) > x.payloadSizeLimit {
			return fmt.Errorf("link's payload exceeds max available size: %d > %d", uint64(len(payload)), x.payloadSizeLimit)
		}

		meta.reset()
		meta.accumulateNextPayloadChunk(payload)

		_, err = writeInMemObject(ctx, x.signer, x.stream, obj, payloadAsBuffers, meta, x.prmObjectPutInit)
		if err != nil {
			return fmt.Errorf("write linking object: %w", err)
		}
	}

	return nil
}

func flushObjectMetadata(signer neofscrypto.Signer, meta dynamicObjectMetadata, header *object.Object) (oid.ID, error) {
	var cs checksum.Checksum

	cs.SetSHA256([sha256.Size]byte(meta.checksum.Sum(nil)))
	header.SetPayloadChecksum(cs)

	if meta.homomorphicChecksum != nil {
		cs.SetTillichZemor([tz.Size]byte(meta.homomorphicChecksum.Sum(nil)))
		header.SetPayloadHomomorphicHash(cs)
	}

	header.SetPayloadSize(meta.length)

	id, err := header.CalculateID()
	if err != nil {
		return id, fmt.Errorf("calculate ID: %w", err)
	}

	header.SetID(id)

	bID, err := id.Marshal()
	if err != nil {
		return id, fmt.Errorf("marshal object ID: %w", err)
	}

	var sig neofscrypto.Signature

	err = sig.Calculate(signer, bID)
	if err != nil {
		return id, fmt.Errorf("sign object ID: %w", err)
	}

	header.SetSignature(&sig)

	return id, nil
}

func writeInMemObject(ctx context.Context, signer user.Signer, w ObjectWriter, header object.Object, payloadBuffers [][]byte, meta dynamicObjectMetadata, prm client.PrmObjectPutInit) (oid.ID, error) {
	var (
		id    oid.ID
		err   error
		isSet bool
	)

	id, isSet = header.ID()
	if !isSet || header.Signature() == nil {
		id, err = flushObjectMetadata(signer, meta, &header)

		if err != nil {
			return id, err
		}
	}

	stream, err := w.ObjectPutInit(ctx, header, signer, prm)
	if err != nil {
		return id, fmt.Errorf("init data stream for next object: %w", err)
	}

	for i := range payloadBuffers {
		_, err = stream.Write(payloadBuffers[i])
		if err != nil {
			return id, fmt.Errorf("write object payload: %w", err)
		}
	}

	if c, ok := stream.(io.Closer); ok {
		err = c.Close()
		if err != nil {
			return id, fmt.Errorf("finish object stream: %w", err)
		}
	}

	return id, nil
}

// dynamicObjectMetadata groups accumulated object metadata which depends on
// payload.
type dynamicObjectMetadata struct {
	length              uint64
	checksum            hash.Hash
	homomorphicChecksum hash.Hash
}

func newDynamicObjectMetadata(withHomoChecksum bool) dynamicObjectMetadata {
	m := dynamicObjectMetadata{
		checksum: sha256.New(),
	}

	if withHomoChecksum {
		m.homomorphicChecksum = tz.New()
	}

	return m
}

func (x *dynamicObjectMetadata) Write(chunk []byte) (int, error) {
	x.accumulateNextPayloadChunk(chunk)
	return len(chunk), nil
}

// accumulateNextPayloadChunk handles the next payload chunk and updates the
// accumulated metadata.
func (x *dynamicObjectMetadata) accumulateNextPayloadChunk(chunk []byte) {
	x.length += uint64(len(chunk))
	x.checksum.Write(chunk)
	if x.homomorphicChecksum != nil {
		x.homomorphicChecksum.Write(chunk)
	}
}

// reset resets all accumulated metadata.
func (x *dynamicObjectMetadata) reset() {
	x.length = 0
	x.checksum.Reset()
	if x.homomorphicChecksum != nil {
		x.homomorphicChecksum.Reset()
	}
}
