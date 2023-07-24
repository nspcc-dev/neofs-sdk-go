package slicer

import (
	"bytes"
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
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

var (
	// ErrInvalidAttributeAmount indicates wrong number of arguments. Amount of arguments MUST be even number.
	ErrInvalidAttributeAmount = errors.New("attributes must be even number of strings")
)

// ObjectWriter represents a virtual object recorder.
type ObjectWriter interface {
	// ObjectPutInit initializes and returns a stream of writable data associated
	// with the object according to its header. Provided header includes at least
	// container, owner and object ID fields.
	//
	// Signer is required and must not be nil. The operation is executed on behalf of
	// the account corresponding to the specified Signer, which is taken into account, in particular, for access control.
	ObjectPutInit(ctx context.Context, hdr object.Object, signer user.Signer, prm client.PrmObjectPutInit) (client.ObjectWriter, error)
}

// Slicer converts input raw data streams into NeoFS objects. Working Slicer
// must be constructed via New.
type Slicer struct {
	signer user.Signer

	cnr cid.ID

	owner user.ID

	w ObjectWriter

	opts Options

	sessionToken *session.Object
}

// New constructs Slicer which writes sliced ready-to-go objects owned by
// particular user into the specified container using provided ObjectWriter.
// All objects are signed using provided neofscrypto.Signer.
//
// If ObjectWriter returns data streams which provide io.Closer, they are closed
// in Slicer.Slice after the payload of any object has been written. In this
// case, Slicer.Slice fails immediately on Close error.
//
// Options parameter allows you to provide optional parameters which tune
// the default Slicer behavior. They are detailed below.
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
// If current NeoFS epoch is specified via Options.SetCurrentNeoFSEpoch, it is
// written to the metadata of all resulting objects as a creation epoch.
//
// See also NewSession.
func New(signer user.Signer, cnr cid.ID, owner user.ID, w ObjectWriter, opts Options) *Slicer {
	return &Slicer{
		signer: signer,
		cnr:    cnr,
		owner:  owner,
		w:      w,
		opts:   opts,
	}
}

// NewSession creates Slicer which generates objects within provided session.
// NewSession work similar to New with the detail that the session issuer owns
// the produced objects. Specified session token is written to the metadata of
// all resulting objects. In this case, the object is considered to be created
// by a proxy on behalf of the session issuer.
func NewSession(signer user.Signer, cnr cid.ID, token session.Object, w ObjectWriter, opts Options) *Slicer {
	return &Slicer{
		signer:       signer,
		cnr:          cnr,
		owner:        token.Issuer(),
		w:            w,
		opts:         opts,
		sessionToken: &token,
	}
}

const defaultPayloadSizeLimit = 1 << 20

// childPayloadSizeLimit returns configured size limit of the child object's
// payload which defaults to 1MB.
func (x *Slicer) childPayloadSizeLimit() uint64 {
	if x.opts.objectPayloadLimit > 0 {
		return x.opts.objectPayloadLimit
	}
	return defaultPayloadSizeLimit
}

// Slice creates new NeoFS object from the input data stream, associates the
// object with the configured container and writes the object via underlying
// ObjectWriter. After a successful write, Slice returns an oid.ID which is a
// unique reference to the object in the container. Slice sets all required
// calculated fields like payload length, checksum, etc.
//
// Slice allows you to specify string key-value pairs to be written to the
// resulting object's metadata as object attributes. Corresponding argument MUST
// NOT be empty or have odd length. Keys SHOULD NOT start with system-reserved
// '__NEOFS__' prefix.
//
// See New for details.
func (x *Slicer) Slice(data io.Reader, attributes ...object.Attribute) (oid.ID, error) {
	var rootID oid.ID

	writer, err := x.InitPayloadStream(attributes...)
	if err != nil {
		return rootID, fmt.Errorf("init writer: %w", err)
	}

	var n int
	bChunk := make([]byte, x.opts.objectPayloadLimit)

	for {
		n, err = data.Read(bChunk)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return rootID, fmt.Errorf("read chunk: %w", err)
			}

			// no more data to read

			if err = writer.Close(); err != nil {
				return rootID, fmt.Errorf("close writer: %w", err)
			}

			rootID = writer.ID()
			break
		}

		if _, err = writer.Write(bChunk[:n]); err != nil {
			return oid.ID{}, err
		}
	}

	return rootID, nil
}

// InitPayloadStream works similar to Slice but provides PayloadWriter allowing
// the caller to write data himself.
func (x *Slicer) InitPayloadStream(attributes ...object.Attribute) (*PayloadWriter, error) {
	res := &PayloadWriter{
		stream:       x.w,
		signer:       x.signer,
		container:    x.cnr,
		owner:        x.owner,
		currentEpoch: x.opts.currentNeoFSEpoch,
		sessionToken: x.sessionToken,
		attributes:   attributes,
		rootMeta:     newDynamicObjectMetadata(x.opts.withHomoChecksum),
		childMeta:    newDynamicObjectMetadata(x.opts.withHomoChecksum),
	}

	res.buf.Grow(int(x.childPayloadSizeLimit()))
	res.rootMeta.reset()
	res.currentWriter = newLimitedWriter(io.MultiWriter(&res.buf, &res.rootMeta), x.childPayloadSizeLimit())

	return res, nil
}

// PayloadWriter is a single-object payload stream provided by Slicer.
type PayloadWriter struct {
	stream ObjectWriter

	rootID oid.ID

	signer       user.Signer
	container    cid.ID
	owner        user.ID
	currentEpoch uint64
	sessionToken *session.Object
	attributes   []object.Attribute

	buf bytes.Buffer

	rootMeta  dynamicObjectMetadata
	childMeta dynamicObjectMetadata

	currentWriter limitedWriter

	withSplit bool
	splitID   *object.SplitID

	writtenChildren []oid.ID
}

// Write writes next chunk of the object data. Concatenation of all chunks forms
// the payload of the final object. When the data is over, the PayloadWriter
// should be closed.
func (x *PayloadWriter) Write(chunk []byte) (int, error) {
	if len(chunk) == 0 {
		// not explicitly prohibited in the io.Writer documentation
		return 0, nil
	}

	n, err := x.currentWriter.Write(chunk)
	if err == nil || !errors.Is(err, errOverflow) {
		return n, err
	}

	if !x.withSplit {
		x.splitID = object.NewSplitID()
		// note: don't move next row, the value of this flag will be used inside writeIntermediateChild
		// to fill splitInfo in all child objects.
		x.withSplit = true

		err = x.writeIntermediateChild(x.rootMeta)
		if err != nil {
			return n, fmt.Errorf("write 1st child: %w", err)
		}

		x.currentWriter.reset(io.MultiWriter(&x.buf, &x.rootMeta, &x.childMeta))
	} else {
		err = x.writeIntermediateChild(x.childMeta)
		if err != nil {
			return n, fmt.Errorf("write next child: %w", err)
		}

		x.currentWriter.resetProgress()
	}

	x.buf.Reset()
	x.childMeta.reset()

	n2, err := x.Write(chunk[n:]) // here n > 0 so infinite recursion shouldn't occur

	return n + n2, err
}

// Close finalizes object with written payload data, saves the object and closes
// the stream. Reference to the stored object can be obtained by ID method.
func (x *PayloadWriter) Close() error {
	if x.withSplit {
		return x.writeLastChild(x.childMeta, x.setID)
	}
	return x.writeLastChild(x.rootMeta, x.setID)
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
func (x *PayloadWriter) writeIntermediateChild(meta dynamicObjectMetadata) error {
	return x._writeChild(meta, false, nil)
}

// writeIntermediateChild writes last split-chain element with specified
// dynamicObjectMetadata to the configured ObjectWriter. If rootIDHandler is
// specified, ID of the resulting root object is passed into it.
func (x *PayloadWriter) writeLastChild(meta dynamicObjectMetadata, rootIDHandler func(id oid.ID)) error {
	return x._writeChild(meta, true, rootIDHandler)
}

func (x *PayloadWriter) _writeChild(meta dynamicObjectMetadata, last bool, rootIDHandler func(id oid.ID)) error {
	currentVersion := version.Current()

	fCommon := func(obj *object.Object) {
		obj.SetVersion(&currentVersion)
		obj.SetContainerID(x.container)
		obj.SetCreationEpoch(x.currentEpoch)
		obj.SetType(object.TypeRegular)
		obj.SetOwnerID(&x.owner)
		obj.SetSessionToken(x.sessionToken)
	}

	var obj object.Object
	fCommon(&obj)
	if x.withSplit {
		obj.SetSplitID(x.splitID)
	}
	if len(x.writtenChildren) > 0 {
		obj.SetPreviousID(x.writtenChildren[len(x.writtenChildren)-1])
	}
	if last {
		var rootObj *object.Object
		if x.withSplit {
			rootObj = new(object.Object)
		} else {
			rootObj = &obj
		}

		fCommon(rootObj)
		rootObj.SetAttributes(x.attributes...)

		rootID, err := flushObjectMetadata(x.signer, x.rootMeta, rootObj)
		if err != nil {
			return fmt.Errorf("form root object: %w", err)
		}

		if rootIDHandler != nil {
			rootIDHandler(rootID)
		}

		if x.withSplit {
			obj.SetParentID(rootID)
			obj.SetParent(rootObj)
		}
	}

	id, err := writeInMemObject(x.signer, x.stream, obj, x.buf.Bytes(), meta, x.sessionToken)
	if err != nil {
		return fmt.Errorf("write formed object: %w", err)
	}

	x.writtenChildren = append(x.writtenChildren, id)

	if x.withSplit && last {
		meta.reset()
		obj.ResetPreviousID()
		obj.SetChildren(x.writtenChildren...)

		_, err = writeInMemObject(x.signer, x.stream, obj, nil, meta, x.sessionToken)
		if err != nil {
			return fmt.Errorf("write linking object: %w", err)
		}
	}

	return nil
}

func flushObjectMetadata(signer neofscrypto.Signer, meta dynamicObjectMetadata, header *object.Object) (oid.ID, error) {
	var cs checksum.Checksum

	var csBytes [sha256.Size]byte
	copy(csBytes[:], meta.checksum.Sum(nil))

	cs.SetSHA256(csBytes)
	header.SetPayloadChecksum(cs)

	if meta.homomorphicChecksum != nil {
		var csHomoBytes [tz.Size]byte
		copy(csHomoBytes[:], meta.homomorphicChecksum.Sum(nil))

		cs.SetTillichZemor(csHomoBytes)
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

func writeInMemObject(signer user.Signer, w ObjectWriter, header object.Object, payload []byte, meta dynamicObjectMetadata, session *session.Object) (oid.ID, error) {
	id, err := flushObjectMetadata(signer, meta, &header)
	if err != nil {
		return id, err
	}

	var prm client.PrmObjectPutInit
	if session != nil {
		prm.WithinSession(*session)
	}

	stream, err := w.ObjectPutInit(context.TODO(), header, signer, prm)
	if err != nil {
		return id, fmt.Errorf("init data stream for next object: %w", err)
	}

	_, err = stream.Write(payload)
	if err != nil {
		return id, fmt.Errorf("write object payload: %w", err)
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

var errOverflow = errors.New("overflow")

// limitedWriter provides io.Writer limiting data volume.
type limitedWriter struct {
	base io.Writer

	limit, written uint64
}

// newLimitedWriter initializes limiterWriter which writes data to the base
// writer before the specified limit.
func newLimitedWriter(base io.Writer, limit uint64) limitedWriter {
	return limitedWriter{
		base:  base,
		limit: limit,
	}
}

// reset resets progress to zero and sets the base target for writing subsequent
// data.
func (x *limitedWriter) reset(base io.Writer) {
	x.base = base
	x.resetProgress()
}

// resetProgress resets progress to zero.
func (x *limitedWriter) resetProgress() {
	x.written = 0
}

// Write writes next chunk of the data to the base writer. If chunk along with
// already written data overflows configured limit, Write returns errOverflow.
func (x *limitedWriter) Write(p []byte) (n int, err error) {
	overflow := uint64(len(p)) > x.limit-x.written

	if overflow {
		n, err = x.base.Write(p[:x.limit-x.written])
	} else {
		n, err = x.base.Write(p)
	}

	x.written += uint64(n)

	if overflow && err == nil {
		return n, errOverflow
	}

	return n, err
}
