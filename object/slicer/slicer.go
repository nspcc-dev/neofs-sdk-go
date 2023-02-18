package slicer

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

// ObjectWriter represents a virtual object recorder.
type ObjectWriter interface {
	// InitDataStream initializes and returns a stream of writable data associated
	// with the object according to its header. Provided header includes at least
	// container, owner and object ID fields.
	InitDataStream(header object.Object) (dataStream io.Writer, err error)
}

// Slicer converts input raw data streams into NeoFS objects. Working Slicer
// must be constructed via New.
type Slicer struct {
	signer neofscrypto.Signer

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
func New(signer neofscrypto.Signer, cnr cid.ID, owner user.ID, w ObjectWriter, opts Options) *Slicer {
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
func NewSession(signer neofscrypto.Signer, cnr cid.ID, token session.Object, w ObjectWriter, opts Options) *Slicer {
	return &Slicer{
		signer:       signer,
		cnr:          cnr,
		owner:        token.Issuer(),
		w:            w,
		opts:         opts,
		sessionToken: &token,
	}
}

// fillCommonMetadata writes to the object metadata common to all objects of the
// same stream.
func (x *Slicer) fillCommonMetadata(obj *object.Object) {
	currentVersion := version.Current()
	obj.SetVersion(&currentVersion)
	obj.SetContainerID(x.cnr)
	obj.SetCreationEpoch(x.opts.currentNeoFSEpoch)
	obj.SetType(object.TypeRegular)
	obj.SetOwnerID(&x.owner)
	obj.SetSessionToken(x.sessionToken)
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
func (x *Slicer) Slice(data io.Reader, attributes ...string) (oid.ID, error) {
	if len(attributes)%2 != 0 {
		panic("attributes must be even number of strings")
	}

	if x.opts.objectPayloadLimit == 0 {
		x.opts.objectPayloadLimit = 1 << 20
	}

	var rootID oid.ID
	var rootHeader object.Object
	var rootMeta dynamicObjectMetadata
	var offset uint64
	var isSplit bool
	var childMeta dynamicObjectMetadata
	var writtenChildren []oid.ID
	var childHeader object.Object
	bChunk := make([]byte, x.opts.objectPayloadLimit+1)

	x.fillCommonMetadata(&rootHeader)
	rootMeta.reset()

	for {
		n, err := data.Read(bChunk[offset:])
		if err == nil {
			if last := offset + uint64(n); last <= x.opts.objectPayloadLimit {
				rootMeta.accumulateNextPayloadChunk(bChunk[offset:last])
				if isSplit {
					childMeta.accumulateNextPayloadChunk(bChunk[offset:last])
				}
				offset = last
				// data is not over, and we expect more bytes to form next object
				continue
			}
		} else {
			if !errors.Is(err, io.EOF) {
				return rootID, fmt.Errorf("read payload chunk: %w", err)
			}

			// there will be no more data

			toSend := offset + uint64(n)
			if toSend <= x.opts.objectPayloadLimit {
				// we can finalize the root object and send last part

				if len(attributes) > 0 {
					attrs := make([]object.Attribute, len(attributes)/2)

					for i := 0; i < len(attrs); i++ {
						attrs[i].SetKey(attributes[2*i])
						attrs[i].SetValue(attributes[2*i+1])
					}

					rootHeader.SetAttributes(attrs...)
				}

				rootID, err = flushObjectMetadata(x.signer, rootMeta, &rootHeader)
				if err != nil {
					return rootID, fmt.Errorf("form root object: %w", err)
				}

				if isSplit {
					// when splitting, root object's header is written into its last child
					childHeader.SetParent(&rootHeader)
					childHeader.SetPreviousID(writtenChildren[len(writtenChildren)-1])

					childID, err := writeInMemObject(x.signer, x.w, childHeader, bChunk[:toSend], childMeta)
					if err != nil {
						return rootID, fmt.Errorf("write child object: %w", err)
					}

					writtenChildren = append(writtenChildren, childID)
				} else {
					// root object is single (full < limit), so send it directly
					rootID, err = writeInMemObject(x.signer, x.w, rootHeader, bChunk[:toSend], rootMeta)
					if err != nil {
						return rootID, fmt.Errorf("write single root object: %w", err)
					}

					return rootID, nil
				}

				break
			}

			// otherwise, form penultimate object, then do one more iteration for
			// simplicity: according to io.Reader, we'll get io.EOF again, but the overflow
			// will no longer occur, so we'll finish the loop
		}

		// according to buffer size, here we can overflow the object payload limit, e.g.
		//  1. full=11B,limit=10B,read=11B (no objects created yet)
		//  2. full=21B,limit=10B,read=11B (one object has been already sent with size=10B)

		toSend := offset + uint64(n)
		overflow := toSend > x.opts.objectPayloadLimit
		if overflow {
			toSend = x.opts.objectPayloadLimit
		}

		// we could read some data even in case of io.EOF, so don't forget pick up the tail
		if n > 0 {
			rootMeta.accumulateNextPayloadChunk(bChunk[offset:toSend])
			if isSplit {
				childMeta.accumulateNextPayloadChunk(bChunk[offset:toSend])
			}
		}

		if overflow {
			isSplitCp := isSplit // we modify it in next condition below but need after it
			if !isSplit {
				// we send only child object below, but we can get here at the beginning (see
				// option 1 described above), so we need to pre-init child resources
				isSplit = true
				x.fillCommonMetadata(&childHeader)
				childHeader.SetSplitID(object.NewSplitID())
				childMeta = rootMeta
				// we do shallow copy of rootMeta because below we take this into account and do
				// not corrupt it
			} else {
				childHeader.SetPreviousID(writtenChildren[len(writtenChildren)-1])
			}

			childID, err := writeInMemObject(x.signer, x.w, childHeader, bChunk[:toSend], childMeta)
			if err != nil {
				return rootID, fmt.Errorf("write child object: %w", err)
			}

			writtenChildren = append(writtenChildren, childID)

			// shift overflow bytes to the beginning
			if !isSplitCp {
				childMeta = dynamicObjectMetadata{} // to avoid rootMeta corruption
			}
			childMeta.reset()
			childMeta.accumulateNextPayloadChunk(bChunk[toSend:])
			rootMeta.accumulateNextPayloadChunk(bChunk[toSend:])
			offset = uint64(copy(bChunk, bChunk[toSend:]))
		}
	}

	// linking object
	childMeta.reset()
	childHeader.ResetPreviousID()
	childHeader.SetChildren(writtenChildren...)

	_, err := writeInMemObject(x.signer, x.w, childHeader, nil, childMeta)
	if err != nil {
		return rootID, fmt.Errorf("write linking object: %w", err)
	}

	return rootID, nil
}

func flushObjectMetadata(signer neofscrypto.Signer, meta dynamicObjectMetadata, header *object.Object) (oid.ID, error) {
	var cs checksum.Checksum

	var csBytes [sha256.Size]byte
	copy(csBytes[:], meta.checksum.Sum(nil))

	cs.SetSHA256(csBytes)
	header.SetPayloadChecksum(cs)

	var csHomoBytes [tz.Size]byte
	copy(csHomoBytes[:], meta.homomorphicChecksum.Sum(nil))

	cs.SetTillichZemor(csHomoBytes)
	header.SetPayloadHomomorphicHash(cs)

	header.SetPayloadSize(meta.length)

	id, err := object.CalculateID(header)
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

func writeInMemObject(signer neofscrypto.Signer, w ObjectWriter, header object.Object, payload []byte, meta dynamicObjectMetadata) (oid.ID, error) {
	id, err := flushObjectMetadata(signer, meta, &header)
	if err != nil {
		return id, err
	}

	stream, err := w.InitDataStream(header)
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

// accumulateNextPayloadChunk handles the next payload chunk and updates the
// accumulated metadata.
func (x *dynamicObjectMetadata) accumulateNextPayloadChunk(chunk []byte) {
	x.length += uint64(len(chunk))
	x.checksum.Write(chunk)
	x.homomorphicChecksum.Write(chunk)
}

// reset resets all accumulated metadata.
func (x *dynamicObjectMetadata) reset() {
	x.length = 0

	if x.checksum != nil {
		x.checksum.Reset()
	} else {
		x.checksum = sha256.New()
	}

	if x.homomorphicChecksum != nil {
		x.homomorphicChecksum.Reset()
	} else {
		x.homomorphicChecksum = tz.New()
	}
}
