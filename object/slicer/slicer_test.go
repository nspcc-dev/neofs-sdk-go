package slicer_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsecdsa "github.com/nspcc-dev/neofs-sdk-go/crypto/ecdsa"
	"github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/slicer"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

const defaultLimit = 1 << 20

func TestSliceDataIntoObjects(t *testing.T) {
	const size = 1 << 10

	t.Run("known limit", func(t *testing.T) {
		t.Run("under limit", func(t *testing.T) {
			testSlicer(t, size, size)
			testSlicer(t, size, size+1)
		})

		t.Run("multiple size", func(t *testing.T) {
			testSlicer(t, size, 3*size)
			testSlicer(t, size, 3*size+1)
		})
	})

	t.Run("unknown limit", func(t *testing.T) {
		t.Run("under limit", func(t *testing.T) {
			testSlicer(t, defaultLimit-1, 0)
			testSlicer(t, defaultLimit, 0)
		})

		t.Run("multiple size", func(t *testing.T) {
			testSlicer(t, 3*defaultLimit, 0)
			testSlicer(t, 3*defaultLimit+1, 0)
		})
	})

	t.Run("no payload", func(t *testing.T) {
		testSlicer(t, 0, 0)
		testSlicer(t, 0, 1024)
	})
}

func BenchmarkSliceDataIntoObjects(b *testing.B) {
	const limit = 1 << 7
	const stepFactor = 4
	for size := uint64(1); size <= 1<<20; size *= stepFactor {
		b.Run(fmt.Sprintf("slice_%d-%d", size, limit), func(b *testing.B) {
			benchmarkSliceDataIntoObjects(b, size, limit)
		})
	}
}

func benchmarkSliceDataIntoObjects(b *testing.B, size, sizeLimit uint64) {
	in, opts := randomInput(b, size, sizeLimit)
	s := slicer.NewSession(in.signer, in.container, *sessiontest.ObjectSigned(test.RandomSigner(b)), discardObject{}, opts)

	b.Run("reader", func(b *testing.B) {
		var err error
		r := bytes.NewReader(in.payload)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err = s.Slice(r, in.attributes...)
			b.StopTimer()
			require.NoError(b, err)
			b.StartTimer()
		}
	})

	b.Run("writer", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		var err error
		var w *slicer.PayloadWriter

		for i := 0; i < b.N; i++ {
			w, err = s.InitPayloadStream(in.attributes...)
			b.StopTimer()
			require.NoError(b, err)
			b.StartTimer()

			_, err = w.Write(in.payload)
			if err == nil {
				err = w.Close()
			}

			b.StopTimer()
			require.NoError(b, err)
			b.StartTimer()
		}
	})
}

type discardObject struct{}

func (discardObject) InitDataStream(object.Object, neofscrypto.Signer) (io.Writer, error) {
	return discardPayload{}, nil
}

type discardPayload struct{}

func (discardPayload) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type input struct {
	signer       neofscrypto.Signer
	container    cid.ID
	owner        user.ID
	currentEpoch uint64
	payloadLimit uint64
	sessionToken *session.Object
	payload      []byte
	attributes   []string
	withHomo     bool
}

func randomData(size uint64) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}

func randomInput(tb testing.TB, size, sizeLimit uint64) (input, slicer.Options) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		panic(fmt.Sprintf("generate ECDSA private key: %v", err))
	}

	attrNum := rand.Int() % 5
	attrs := make([]string, 2*attrNum)

	for i := 0; i < len(attrs); i += 2 {
		attrs[i] = base64.StdEncoding.EncodeToString(randomData(32))
		attrs[i+1] = base64.StdEncoding.EncodeToString(randomData(32))
	}

	var in input
	in.signer = neofsecdsa.Signer(*key)
	in.container = cidtest.ID()
	in.currentEpoch = rand.Uint64()
	if sizeLimit > 0 {
		in.payloadLimit = sizeLimit
	} else {
		in.payloadLimit = defaultLimit
	}
	in.payload = randomData(size)
	in.attributes = attrs

	if rand.Int()%2 == 0 {
		in.sessionToken = sessiontest.ObjectSigned(test.RandomSigner(tb))
	} else {
		in.owner = *usertest.ID(tb)
	}

	in.withHomo = rand.Int()%2 == 0

	var opts slicer.Options
	opts.SetObjectPayloadLimit(in.payloadLimit)
	opts.SetCurrentNeoFSEpoch(in.currentEpoch)
	if in.withHomo {
		opts.CalculateHomomorphicChecksum()
	}

	return in, opts
}

func testSlicer(tb testing.TB, size, sizeLimit uint64) {
	in, opts := randomInput(tb, size, sizeLimit)

	checker := &slicedObjectChecker{
		tb:             tb,
		input:          in,
		chainCollector: newChainCollector(tb),
	}

	var s *slicer.Slicer
	if checker.input.sessionToken != nil {
		s = slicer.NewSession(in.signer, checker.input.container, *checker.input.sessionToken, checker, opts)
	} else {
		s = slicer.New(in.signer, checker.input.container, checker.input.owner, checker, opts)
	}

	// check reader
	rootID, err := s.Slice(bytes.NewReader(in.payload), in.attributes...)
	require.NoError(tb, err)
	checker.chainCollector.verify(checker.input, rootID)

	// check writer with random written chunk's size
	checker.chainCollector = newChainCollector(tb)

	w, err := s.InitPayloadStream(in.attributes...)
	require.NoError(tb, err)

	var chunkSize int
	if len(in.payload) > 0 {
		chunkSize = rand.Int() % len(in.payload)
		if chunkSize == 0 {
			chunkSize = 1
		}
	}

	for payload := in.payload; len(payload) > 0; payload = payload[chunkSize:] {
		if chunkSize > len(payload) {
			chunkSize = len(payload)
		}
		n, err := w.Write(payload[:chunkSize])
		require.NoError(tb, err)
		require.EqualValues(tb, chunkSize, n)
	}

	err = w.Close()
	require.NoError(tb, err)

	checker.chainCollector.verify(checker.input, w.ID())
}

type slicedObjectChecker struct {
	tb testing.TB

	input input

	chainCollector *chainCollector
}

func (x *slicedObjectChecker) InitDataStream(hdr object.Object, _ neofscrypto.Signer) (io.Writer, error) {
	checkStaticMetadata(x.tb, hdr, x.input)

	buf := bytes.NewBuffer(nil)

	x.chainCollector.handleOutgoingObject(hdr, buf)

	return newSizeChecker(x.tb, hdr, buf, x.input.payloadLimit), nil
}

type writeSizeChecker struct {
	tb          testing.TB
	hdr         object.Object
	limit       uint64
	processed   uint64
	base        io.Writer
	payloadSeen bool
}

func newSizeChecker(tb testing.TB, hdr object.Object, base io.Writer, sizeLimit uint64) io.Writer {
	return &writeSizeChecker{
		tb:    tb,
		hdr:   hdr,
		limit: sizeLimit,
		base:  base,
	}
}

func (x *writeSizeChecker) Write(p []byte) (int, error) {
	if !x.payloadSeen && len(p) > 0 {
		x.payloadSeen = true
	}

	if x.payloadSeen {
		if len(x.hdr.Children()) == 0 {
			// only linking objects should be streamed with
			// empty payload
			require.NotZero(x.tb, len(p))
		} else {
			// linking object should have empty payload
			require.Zero(x.tb, x.hdr.PayloadSize())
		}
	}

	n, err := x.base.Write(p)
	x.processed += uint64(n)
	return n, err
}

func (x *writeSizeChecker) Close() error {
	require.LessOrEqual(x.tb, x.processed, x.limit, "object payload must not overflow the limit")
	return nil
}

type payloadWithChecksum struct {
	r  io.Reader
	cs []checksum.Checksum
	hs []hash.Hash
}

type chainCollector struct {
	tb testing.TB

	mProcessed map[oid.ID]struct{}

	parentHeaderSet bool
	parentHeader    object.Object

	splitID *object.SplitID

	firstSet    bool
	first       oid.ID
	firstHeader object.Object

	mNext map[oid.ID]oid.ID

	mPayloads map[oid.ID]payloadWithChecksum

	children []oid.ID
}

func newChainCollector(tb testing.TB) *chainCollector {
	return &chainCollector{
		tb:         tb,
		mProcessed: make(map[oid.ID]struct{}),
		mNext:      make(map[oid.ID]oid.ID),
		mPayloads:  make(map[oid.ID]payloadWithChecksum),
	}
}

func checkStaticMetadata(tb testing.TB, header object.Object, in input) {
	cnr, ok := header.ContainerID()
	require.True(tb, ok, "all objects must be bound to some container")
	require.True(tb, cnr.Equals(in.container), "the container must be set to the configured one")

	owner := header.OwnerID()
	require.NotNil(tb, owner, "any object must be owned by somebody")
	if in.sessionToken != nil {
		require.True(tb, in.sessionToken.Issuer().Equals(*owner), "owner must be set to the session issuer")
	} else {
		require.True(tb, owner.Equals(in.owner), "owner must be set to the particular user")
	}

	ver := header.Version()
	require.NotNil(tb, ver, "version must be set in all objects")
	require.Equal(tb, version.Current(), *ver, "the version must be set to current SDK one")

	require.Equal(tb, object.TypeRegular, header.Type(), "only regular objects must be produced")
	require.EqualValues(tb, in.currentEpoch, header.CreationEpoch(), "configured current epoch must be set as creation epoch")
	require.Equal(tb, in.sessionToken, header.SessionToken(), "configured session token must be written into objects")

	require.NoError(tb, header.CheckHeaderVerificationFields(), "verification fields must be correctly set in header")

	_, ok = header.PayloadHomomorphicHash()
	require.Equal(tb, in.withHomo, ok)
}

func (x *chainCollector) handleOutgoingObject(header object.Object, payload io.Reader) {
	id, ok := header.ID()
	require.True(x.tb, ok, "all objects must have an ID")

	idCalc, err := header.CalculateID()
	require.NoError(x.tb, err)

	require.True(x.tb, idCalc.Equals(id))

	_, ok = x.mProcessed[id]
	require.False(x.tb, ok, "object must be written exactly once")

	x.mProcessed[id] = struct{}{}

	splitID := header.SplitID()
	if x.splitID == nil && splitID != nil {
		x.splitID = splitID
	} else {
		require.Equal(x.tb, x.splitID, splitID, "split ID must the same in all objects")
	}

	parent := header.Parent()
	if parent != nil {
		require.Nil(x.tb, parent.Parent(), "multi-level genealogy is not supported")

		if x.parentHeaderSet {
			require.Equal(x.tb, x.parentHeader, *parent, "root header must the same")
		} else {
			x.parentHeaderSet = true
			x.parentHeader = *parent
		}
	}

	prev, ok := header.PreviousID()
	if ok {
		_, ok := x.mNext[prev]
		require.False(x.tb, ok, "split-chain must not be forked")

		for k := range x.mNext {
			require.False(x.tb, k.Equals(prev), "split-chain must not be cycled")
		}

		x.mNext[prev] = id
	} else if len(header.Children()) == 0 { // 1st split-chain or linking object
		require.False(x.tb, x.firstSet, "there must not be multiple split-chains")
		x.firstSet = true
		x.first = id
		x.firstHeader = header
	}

	children := header.Children()
	if len(children) > 0 {
		if len(x.children) > 0 {
			require.Equal(x.tb, x.children, children, "children list must be the same")
		} else {
			x.children = children
		}
	}

	cs, ok := header.PayloadChecksum()
	require.True(x.tb, ok)

	pcs := payloadWithChecksum{
		r:  payload,
		cs: []checksum.Checksum{cs},
		hs: []hash.Hash{sha256.New()},
	}

	csHomo, ok := header.PayloadHomomorphicHash()
	if ok {
		pcs.cs = append(pcs.cs, csHomo)
		pcs.hs = append(pcs.hs, tz.New())
	}

	x.mPayloads[id] = pcs
}

func (x *chainCollector) verify(in input, rootID oid.ID) {
	require.True(x.tb, x.firstSet, "initial split-chain element must be set")

	rootObj := x.parentHeader
	if !x.parentHeaderSet {
		rootObj = x.firstHeader
	}

	restoredChain := []oid.ID{x.first}
	restoredPayload := bytes.NewBuffer(make([]byte, 0, rootObj.PayloadSize()))

	for {
		v, ok := x.mPayloads[restoredChain[len(restoredChain)-1]]
		require.True(x.tb, ok)

		ws := []io.Writer{restoredPayload}
		for i := range v.hs {
			ws = append(ws, v.hs[i])
		}

		_, err := io.Copy(io.MultiWriter(ws...), v.r)
		require.NoError(x.tb, err)

		for i := range v.cs {
			require.True(x.tb, bytes.Equal(v.cs[i].Value(), v.hs[i].Sum(nil)))
		}

		next, ok := x.mNext[restoredChain[len(restoredChain)-1]]
		if !ok {
			break
		}

		restoredChain = append(restoredChain, next)
	}

	rootObj.SetPayload(restoredPayload.Bytes())

	if uint64(len(in.payload)) <= in.payloadLimit {
		require.Empty(x.tb, x.children)
	} else {
		require.Equal(x.tb, x.children, restoredChain)
	}

	id, ok := rootObj.ID()
	require.True(x.tb, ok, "root object must have an ID")
	require.True(x.tb, id.Equals(rootID), "root ID in root object must be returned in the result")

	checkStaticMetadata(x.tb, rootObj, in)

	attrs := rootObj.Attributes()
	require.Len(x.tb, attrs, len(in.attributes)/2)
	for i := range attrs {
		require.Equal(x.tb, in.attributes[2*i], attrs[i].Key())
		require.Equal(x.tb, in.attributes[2*i+1], attrs[i].Value())
	}

	require.Equal(x.tb, in.payload, rootObj.Payload())
	require.NoError(x.tb, rootObj.VerifyPayloadChecksum(), "payload checksum must be correctly set")
}

type memoryWriter struct {
	headers []object.Object
	splitID *object.SplitID
}

func (w *memoryWriter) InitDataStream(hdr object.Object) (io.Writer, error) {
	w.headers = append(w.headers, hdr)
	if w.splitID == nil && hdr.SplitID() != nil {
		w.splitID = hdr.SplitID()
	}

	return &memoryPayload{}, nil
}

type memoryPayload struct {
}

func (p *memoryPayload) Write(data []byte) (int, error) {
	return len(data), nil
}

func TestSlicedObjectsHaveSplitID(t *testing.T) {
	maxObjectSize := uint64(10)
	overheadAmount := uint64(3)

	var containerID cid.ID
	id := make([]byte, sha256.Size)
	_, err := rand.Read(id)
	require.NoError(t, err)
	containerID.Encode(id)

	var ownerID user.ID
	signer := test.RandomSignerRFC6979(t)
	require.NoError(t, user.IDFromSigner(&ownerID, signer))

	opts := slicer.Options{}
	opts.SetObjectPayloadLimit(maxObjectSize)
	opts.SetCurrentNeoFSEpoch(10)

	t.Run("slice", func(t *testing.T) {
		writer := &memoryWriter{}
		sl := slicer.New(signer, containerID, ownerID, writer, opts)

		payload := make([]byte, maxObjectSize*overheadAmount)
		_, err = rand.Read(payload)
		require.NoError(t, err)

		_, err = sl.Slice(bytes.NewBuffer(payload))
		require.NoError(t, err)

		require.Equal(t, overheadAmount+1, uint64(len(writer.headers)))

		for _, h := range writer.headers {
			splitID := h.SplitID()
			require.NotNil(t, splitID)
			require.Equal(t, writer.splitID.ToV2(), splitID.ToV2())
		}
	})

	t.Run("InitPayloadStream", func(t *testing.T) {
		writer := &memoryWriter{}
		sl := slicer.New(signer, containerID, ownerID, writer, opts)

		payloadWriter, err := sl.InitPayloadStream()
		require.NoError(t, err)

		for i := uint64(0); i < overheadAmount; i++ {
			payload := make([]byte, maxObjectSize)
			_, err = rand.Read(payload)
			require.NoError(t, err)

			_, err := payloadWriter.Write(payload)
			require.NoError(t, err)
		}

		require.NoError(t, payloadWriter.Close())
		require.Equal(t, overheadAmount+1, uint64(len(writer.headers)))

		for _, h := range writer.headers {
			splitID := h.SplitID()
			require.NotNil(t, splitID)
			require.Equal(t, writer.splitID.ToV2(), splitID.ToV2())
		}
	})

	t.Run("no split info if no overflow", func(t *testing.T) {
		writer := &memoryWriter{}
		sl := slicer.New(signer, containerID, ownerID, writer, opts)

		payload := make([]byte, maxObjectSize-1)
		_, err = rand.Read(payload)
		require.NoError(t, err)

		_, err = sl.Slice(bytes.NewBuffer(payload))
		require.NoError(t, err)

		require.Equal(t, uint64(1), uint64(len(writer.headers)))

		for _, h := range writer.headers {
			splitID := h.SplitID()
			require.Nil(t, splitID)
		}
	})
}
