package slicer_test

import (
	"bytes"
	"context"
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
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	"github.com/nspcc-dev/neofs-sdk-go/internal/testutil"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/slicer"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	sessionv2 "github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
	"github.com/stretchr/testify/require"
)

func TestSliceDataIntoObjects(t *testing.T) {
	const limit = 1 << 10

	for _, tc := range []struct {
		name string
		ln   uint64
	}{
		{name: "no payload", ln: 0},
		{name: "limit-1B", ln: limit - 1},
		{name: "exactly limit", ln: limit},
		{name: "limitX3", ln: limit * 3},
		{name: "limitX3+1B", ln: limit*3 + 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testSlicer(t, tc.ln, limit)
		})
	}
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
	ctx := context.Background()

	in, opts := randomInput(size, sizeLimit)
	tok := sessiontest.TokenSigned(usertest.User())
	s, err := slicer.New(
		ctx,
		discardObject{opts: opts},
		in.signer,
		in.container,
		in.owner,
		&tok,
	)
	require.NoError(b, err)

	b.Run("reader", func(b *testing.B) {
		var err error
		r := bytes.NewReader(in.payload)

		b.ReportAllocs()

		for b.Loop() {
			_, err = s.Put(ctx, r, in.attributes)
			b.StopTimer()
			require.NoError(b, err)
			b.StartTimer()
		}
	})

	b.Run("writer", func(b *testing.B) {
		b.ReportAllocs()

		var err error
		var w *slicer.PayloadWriter

		for b.Loop() {
			w, err = s.InitPut(ctx, in.attributes)
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

func networkInfoFromOpts(opts slicer.Options) (netmap.NetworkInfo, error) {
	var ni netmap.NetworkInfo
	ni.SetRawNetworkParameter(string(testutil.RandByteSlice(10)), testutil.RandByteSlice(10))
	ni.SetCurrentEpoch(opts.CurrentNeoFSEpoch())
	ni.SetMaxObjectSize(opts.ObjectPayloadLimit())
	if !opts.IsHomomorphicChecksumEnabled() {
		ni.DisableHomomorphicHashing()
	}

	return ni, nil
}

type discardObject struct {
	opts slicer.Options
}

func (discardObject) ObjectPutInit(context.Context, object.Object, user.Signer, client.PrmObjectPutInit) (client.ObjectWriter, error) {
	return discardPayload{}, nil
}

func (o discardObject) NetworkInfo(_ context.Context, _ client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	return networkInfoFromOpts(o.opts)
}

type discardPayload struct{}

func (discardPayload) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (discardPayload) Close() error {
	return nil
}

func (discardPayload) GetResult() client.ResObjectPut {
	return client.ResObjectPut{}
}

type input struct {
	signer         user.Signer
	container      cid.ID
	owner          user.ID
	objectType     object.Type
	currentEpoch   uint64
	payloadLimit   uint64
	sessionToken   *session.Object
	sessionTokenV2 *sessionv2.Token
	payload        []byte
	attributes     []object.Attribute
	withHomo       bool
}

func randomInput(size, sizeLimit uint64) (input, slicer.Options) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	if err != nil {
		panic(fmt.Sprintf("generate ECDSA private key: %v", err))
	}

	attrNum := rand.Int() % 5
	attrs := make([]object.Attribute, attrNum)

	for range attrNum {
		var attr object.Attribute
		attr.SetKey(base64.StdEncoding.EncodeToString(testutil.RandByteSlice(32)))
		attr.SetValue(base64.StdEncoding.EncodeToString(testutil.RandByteSlice(32)))

		attrs = append(attrs, attr)
	}

	var in input
	in.signer = user.NewAutoIDSigner(*key)
	in.container = cidtest.ID()
	in.currentEpoch = rand.Uint64()
	in.payloadLimit = sizeLimit
	in.payload = testutil.RandByteSlice(size)
	in.attributes = attrs
	in.owner = usertest.ID()

	var opts slicer.Options
	if rand.Int()%2 == 0 {
		tok := sessiontest.TokenSigned(usertest.User())
		in.sessionTokenV2 = &tok
		opts.SetSessionV2(*in.sessionTokenV2)
	}

	in.withHomo = rand.Int()%2 == 0

	opts.SetObjectPayloadLimit(in.payloadLimit)
	opts.SetCurrentNeoFSEpoch(in.currentEpoch)
	if in.withHomo {
		opts.CalculateHomomorphicChecksum()
	}

	return in, opts
}

func testSlicer(t *testing.T, size, sizeLimit uint64) {
	testSlicerWithKnownSize(t, size, sizeLimit, true)
	testSlicerWithKnownSize(t, size, sizeLimit, false)
}

func testSlicerWithKnownSize(t *testing.T, size, sizeLimit uint64, known bool) {
	in, opts := randomInput(size, sizeLimit)

	if known {
		opts.SetPayloadSize(uint64(len(in.payload)))
	}

	checker := &slicedObjectChecker{
		opts:           opts,
		tb:             t,
		input:          in,
		chainCollector: newChainCollector(t),
	}

	for i := object.TypeRegular; i <= object.TypeLink; i++ {
		in.objectType = i

		t.Run(fmt.Sprintf("slicer with %s,known_size=%t,size=%d,limit=%d", i.String(), known, size, sizeLimit), func(t *testing.T) {
			testSlicerByHeaderType(t, checker, in, opts)
		})
	}
}

// eofOnLastChunkReader is a special reader for tests. It returns io.EOF with the last data portion.
type eofOnLastChunkReader struct {
	// this option enable returning 0, nil before fina; result with io.EOF error.
	// Only in case when we return 0, EOF result.
	isZeroNilShowed bool
	isZeroOnEOF     bool
	payload         []byte
	i               int
}

func (l *eofOnLastChunkReader) Read(p []byte) (int, error) {
	n := copy(p, l.payload[l.i:])
	l.i += n

	if l.i == len(l.payload) {
		if n == 0 {
			// nothing happened case from io.Reader docs.
			if !l.isZeroNilShowed {
				l.isZeroNilShowed = true
				return 0, nil
			}

			return 0, io.EOF
		}

		if !l.isZeroOnEOF {
			return n, io.EOF
		}
	}

	return n, nil
}

func testSlicerByHeaderType(t *testing.T, checker *slicedObjectChecker, in input, opts slicer.Options) {
	ctx := context.Background()

	t.Run("Slicer.Put", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)
		s, err := slicer.New(ctx, checker, checker.input.signer, checker.input.container, checker.input.owner, checker.input.sessionTokenV2)
		require.NoError(t, err)

		rootID, err := s.Put(ctx, bytes.NewReader(in.payload), in.attributes)
		require.NoError(t, err)
		checker.chainCollector.verify(checker.input, rootID)
	})

	t.Run("slicer.Put", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		var hdr object.Object
		hdr.SetSessionToken(opts.Session())
		hdr.SetContainerID(in.container)
		hdr.SetOwner(in.owner)
		hdr.SetAttributes(in.attributes...)

		rootID, err := slicer.Put(ctx, checker, hdr, checker.input.signer, bytes.NewReader(in.payload), opts)
		require.NoError(t, err)
		checker.chainCollector.verify(checker.input, rootID)
	})

	t.Run("Slicer.InitPut", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		// check writer with random written chunk's size
		s, err := slicer.New(ctx, checker, checker.input.signer, checker.input.container, checker.input.owner, checker.input.sessionTokenV2)
		require.NoError(t, err)

		w, err := s.InitPut(ctx, in.attributes)
		require.NoError(t, err)

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
			require.NoError(t, err)
			require.EqualValues(t, chunkSize, n)
		}

		err = w.Close()
		require.NoError(t, err)

		checker.chainCollector.verify(checker.input, w.ID())
	})

	t.Run("slicer.InitPut", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		var hdr object.Object
		hdr.SetSessionToken(opts.Session())
		hdr.SetContainerID(in.container)
		hdr.SetOwner(in.owner)
		hdr.SetAttributes(in.attributes...)

		// check writer with random written chunk's size
		w, err := slicer.InitPut(ctx, checker, hdr, checker.input.signer, opts)
		require.NoError(t, err)

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
			require.NoError(t, err)
			require.EqualValues(t, chunkSize, n)
		}

		err = w.Close()
		require.NoError(t, err)

		checker.chainCollector.verify(checker.input, w.ID())
	})

	t.Run("slicer.Put, io.EOF in last chunk", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		var hdr object.Object
		hdr.SetSessionToken(opts.Session())
		hdr.SetContainerID(in.container)
		hdr.SetOwner(in.owner)
		hdr.SetAttributes(in.attributes...)

		rootID, err := slicer.Put(ctx, checker, hdr, checker.input.signer, &eofOnLastChunkReader{payload: in.payload, isZeroNilShowed: true}, opts)
		require.NoError(t, err)
		checker.chainCollector.verify(checker.input, rootID)
	})

	t.Run("slicer.Put, zeroNil before io.EOF in last chunk", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		var hdr object.Object
		hdr.SetSessionToken(opts.Session())
		hdr.SetContainerID(in.container)
		hdr.SetOwner(in.owner)
		hdr.SetAttributes(in.attributes...)

		rootID, err := slicer.Put(ctx, checker, hdr, checker.input.signer, &eofOnLastChunkReader{payload: in.payload}, opts)
		require.NoError(t, err)
		checker.chainCollector.verify(checker.input, rootID)
	})

	t.Run("slicer.Put, io.EOF after last chunk", func(t *testing.T) {
		checker.chainCollector = newChainCollector(t)

		var hdr object.Object
		hdr.SetSessionToken(opts.Session())
		hdr.SetContainerID(in.container)
		hdr.SetOwner(in.owner)
		hdr.SetAttributes(in.attributes...)

		rootID, err := slicer.Put(ctx, checker, hdr, checker.input.signer, &eofOnLastChunkReader{payload: in.payload, isZeroOnEOF: true, isZeroNilShowed: true}, opts)
		require.NoError(t, err)
		checker.chainCollector.verify(checker.input, rootID)
	})
}

type slicedObjectChecker struct {
	opts slicer.Options

	tb testing.TB

	input input

	chainCollector *chainCollector
}

func (x *slicedObjectChecker) NetworkInfo(_ context.Context, _ client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	return networkInfoFromOpts(x.opts)
}

func (x *slicedObjectChecker) ObjectPutInit(_ context.Context, hdr object.Object, _ user.Signer, _ client.PrmObjectPutInit) (client.ObjectWriter, error) {
	checkStaticMetadata(x.tb, hdr, x.input)

	buf := bytes.NewBuffer(nil)

	x.chainCollector.handleOutgoingObject(hdr, buf)

	return newSizeChecker(x.tb, hdr, buf, x.input.payloadLimit), nil
}

type writeSizeChecker struct {
	tb        testing.TB
	hdr       object.Object
	limit     uint64
	processed uint64
	base      *bytes.Buffer
}

func newSizeChecker(tb testing.TB, hdr object.Object, base *bytes.Buffer, sizeLimit uint64) *writeSizeChecker {
	return &writeSizeChecker{
		tb:    tb,
		hdr:   hdr,
		limit: sizeLimit,
		base:  base,
	}
}

func (x *writeSizeChecker) Write(p []byte) (int, error) {
	require.NotZero(x.tb, len(p), "non of the split object should be empty")

	n, err := x.base.Write(p)
	x.processed += uint64(n)
	return n, err
}

func (x *writeSizeChecker) Close() error {
	if x.hdr.Type() == object.TypeLink {
		payload := x.base.Bytes()

		var testLink object.Link
		require.NoError(x.tb, testLink.Unmarshal(payload), "link object's payload must be structured")
	}

	require.LessOrEqual(x.tb, x.processed, x.limit, "object payload must not overflow the limit")

	require.Equal(x.tb, x.processed, x.hdr.PayloadSize())

	// deprecated things
	require.Nil(x.tb, x.hdr.SplitID(), "no split ID should be presented")
	require.Empty(x.tb, x.hdr.Children(), "no child should be stored in the headers")

	return nil
}

func (x *writeSizeChecker) GetResult() client.ResObjectPut {
	return client.ResObjectPut{}
}

type payloadWithChecksum struct {
	r  io.Reader
	cs []checksum.Checksum
	hs []hash.Hash
}

type chainCollector struct {
	tb testing.TB

	mProcessed map[oid.ID]struct{}

	shortParentHeader *object.Object

	parentHeaderSet bool
	parentHeader    object.Object

	firstSet    bool
	first       oid.ID
	firstHeader object.Object

	mNext map[oid.ID]oid.ID

	mPayloads map[oid.ID]payloadWithChecksum

	link oid.ID
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
	cnr := header.GetContainerID()
	require.False(tb, cnr.IsZero(), "all objects must be bound to some container")
	require.True(tb, cnr == in.container, "the container must be set to the configured one")

	owner := header.Owner()
	if in.sessionTokenV2 != nil {
		require.True(tb, in.sessionTokenV2.Issuer() == owner, "owner must be set to the session issuer")
	} else if in.sessionToken != nil {
		require.True(tb, in.sessionToken.Issuer() == owner, "owner must be set to the session issuer")
	} else {
		require.True(tb, owner == in.owner, "owner must be set to the particular user")
	}

	ver := header.Version()
	require.NotNil(tb, ver, "version must be set in all objects")
	require.Equal(tb, version.Current(), *ver, "the version must be set to current SDK one")

	typ := header.Type()
	require.True(tb, typ == object.TypeRegular || typ == object.TypeLink, "only regular and link objects must be produced")

	require.EqualValues(tb, in.currentEpoch, header.CreationEpoch(), "configured current epoch must be set as creation epoch")
	require.Equal(tb, in.sessionToken, header.SessionToken(), "configured session token must be written into objects")

	require.NoError(tb, header.CheckHeaderVerificationFields(), "verification fields must be correctly set in header")

	_, ok := header.PayloadHomomorphicHash()
	require.Equal(tb, in.withHomo, ok)
}

func (x *chainCollector) handleOutgoingObject(headerOriginal object.Object, payload io.Reader) {
	// copy the header cause some slicer code is written considering
	// that sent object is a safe-to-change object, while tests store
	// and share the "sent" objects
	var header object.Object
	headerOriginal.CopyTo(&header)

	require.Empty(x.tb, header.Payload(), "payload must be unset in header")

	id := header.GetID()
	require.False(x.tb, id.IsZero(), "all objects must have an ID")

	idCalc, err := header.CalculateID()
	require.NoError(x.tb, err)

	require.True(x.tb, idCalc == id)

	_, ok := x.mProcessed[id]
	require.False(x.tb, ok, "object must be written exactly once")

	x.mProcessed[id] = struct{}{}

	require.Nil(x.tb, header.SplitID(), "split ID is deprecated and must be nil")

	parent := header.Parent()
	if parent != nil {
		require.Nil(x.tb, parent.Parent(), "multi-level genealogy is not supported")

		if x.shortParentHeader == nil {
			// parent in the first part

			require.True(x.tb, parent.GetID().IsZero(), "first object's parent cannot have ID")

			require.Nil(x.tb, parent.Signature(), "first object's parent cannot have signature")

			x.shortParentHeader = parent
		} else if !x.parentHeaderSet {
			x.parentHeaderSet = true
			x.parentHeader = *parent
		} else {
			require.Equal(x.tb, x.parentHeader, *parent, "root header must the same")

			var parentNoPayloadInfo object.Object

			cID := x.parentHeader.GetContainerID()
			parentNoPayloadInfo.SetVersion(x.parentHeader.Version())
			parentNoPayloadInfo.SetContainerID(cID)
			parentNoPayloadInfo.SetCreationEpoch(x.parentHeader.CreationEpoch())
			parentNoPayloadInfo.SetType(x.parentHeader.Type())
			parentNoPayloadInfo.SetOwner(x.parentHeader.Owner())
			parentNoPayloadInfo.SetSessionToken(x.parentHeader.SessionToken())
			parentNoPayloadInfo.SetSessionTokenV2(x.parentHeader.SessionTokenV2())
			parentNoPayloadInfo.SetAttributes(x.parentHeader.Attributes()...)

			require.Equal(x.tb, *x.shortParentHeader, parentNoPayloadInfo, "first object's parent should be equal to the resulting (without payload info)")
		}
	}

	prev := header.GetPreviousID()
	if !prev.IsZero() {
		_, ok := x.mNext[prev]
		require.False(x.tb, ok, "split-chain must not be forked")

		for k := range x.mNext {
			require.False(x.tb, k == prev, "split-chain must not be cycled")
		}

		x.mNext[prev] = id
	} else if header.HasParent() {
		if !x.firstSet {
			// 1st split-chain

			require.Equal(x.tb, object.TypeRegular, header.Type())
			require.NotNil(x.tb, header.Parent())
		} else {
			// linking object

			require.Equal(x.tb, object.TypeLink, header.Type())

			var testLink object.Link
			require.NoError(x.tb, header.ReadLink(&testLink))

			x.link = id
		}
	}

	if !x.firstSet {
		x.firstSet = true
		x.first = id
		x.firstHeader = header
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

type payloadCounter struct {
	res int
}

func (p *payloadCounter) Write(payload []byte) (n int, err error) {
	read := len(payload)
	p.res += read

	return read, nil
}

func (x *chainCollector) verify(in input, rootID oid.ID) {
	require.True(x.tb, x.firstSet, "initial split-chain element must be set")

	rootObj := x.parentHeader
	if !x.parentHeaderSet {
		rootObj = x.firstHeader
	}

	var firstObject object.MeasuredObject
	firstObject.SetObjectID(x.first)
	firstObject.SetObjectSize(uint32(x.firstHeader.PayloadSize()))

	restoredChain := []object.MeasuredObject{firstObject}
	restoredPayload := bytes.NewBuffer(make([]byte, 0, rootObj.PayloadSize()))

	require.Equal(x.tb, in.objectType, rootObj.Type())

	for {
		v, ok := x.mPayloads[restoredChain[len(restoredChain)-1].ObjectID()]
		require.True(x.tb, ok)

		var counter payloadCounter

		ws := []io.Writer{restoredPayload, &counter}
		for i := range v.hs {
			ws = append(ws, v.hs[i])
		}

		_, err := io.Copy(io.MultiWriter(ws...), v.r)
		require.NoError(x.tb, err)

		restoredChain[len(restoredChain)-1].SetObjectSize(uint32(counter.res))

		for i := range v.cs {
			require.True(x.tb, bytes.Equal(v.cs[i].Value(), v.hs[i].Sum(nil)))
		}

		nextObjectID, ok := x.mNext[restoredChain[len(restoredChain)-1].ObjectID()]
		if !ok {
			break
		}

		var nextObject object.MeasuredObject
		nextObject.SetObjectID(nextObjectID)

		restoredChain = append(restoredChain, nextObject)
	}

	rootObj.SetPayload(restoredPayload.Bytes())

	if uint64(len(in.payload)) <= in.payloadLimit {
		require.Zero(x.tb, x.link)
	} else {
		require.NotZero(x.tb, x.link)
		p, ok := x.mPayloads[x.link]
		require.True(x.tb, ok)
		payload, err := io.ReadAll(p.r)
		require.NoError(x.tb, err)

		var l object.Link
		require.NoError(x.tb, l.Unmarshal(payload))
		require.Equal(x.tb, l.Objects(), restoredChain)
	}

	id := rootObj.GetID()
	require.False(x.tb, id.IsZero(), "root object must have an ID")
	require.True(x.tb, id == rootID, "root ID in root object must be returned in the result")

	checkStaticMetadata(x.tb, rootObj, in)

	attrs := rootObj.Attributes()
	require.Len(x.tb, attrs, len(in.attributes))
	for i := range attrs {
		require.Equal(x.tb, in.attributes[i].Key(), attrs[i].Key())
		require.Equal(x.tb, in.attributes[i].Value(), attrs[i].Value())
	}

	require.Equal(x.tb, in.payload, rootObj.Payload())
	require.NoError(x.tb, rootObj.VerifyPayloadChecksum(), "payload checksum must be correctly set")
}

type memoryWriter struct {
	opts        slicer.Options
	headers     []object.Object
	firstObject *oid.ID
}

func (w *memoryWriter) ObjectPutInit(_ context.Context, hdr object.Object, _ user.Signer, _ client.PrmObjectPutInit) (client.ObjectWriter, error) {
	var objectCopy object.Object
	hdr.CopyTo(&objectCopy)
	w.headers = append(w.headers, objectCopy)

	if w.firstObject == nil {
		first := hdr.GetFirstID()
		if !first.IsZero() {
			w.firstObject = &first
		}
	}

	return &memoryPayload{}, nil
}

func (w *memoryWriter) NetworkInfo(_ context.Context, _ client.PrmNetworkInfo) (netmap.NetworkInfo, error) {
	return networkInfoFromOpts(w.opts)
}

type memoryPayload struct {
}

func (p *memoryPayload) Write(data []byte) (int, error) {
	return len(data), nil
}

func (p *memoryPayload) Close() error {
	return nil
}

func (p *memoryPayload) GetResult() client.ResObjectPut {
	return client.ResObjectPut{}
}

func TestSlicedObjectsHaveSplitID(t *testing.T) {
	maxObjectSize := uint64(200) // some reasonable value that a link object can fit
	overheadAmount := uint64(3)
	ctx := context.Background()

	containerID := cidtest.ID()

	usr := usertest.User()
	usrID := usr.UserID()

	opts := slicer.Options{}
	opts.SetObjectPayloadLimit(maxObjectSize)
	opts.SetCurrentNeoFSEpoch(10)

	checkParentWithoutSplitInfo := func(hdr object.Object) {
		for o := hdr.Parent(); o != nil; o = o.Parent() {
			require.Nil(t, o.ProtoMessage().GetHeader().GetSplit())
		}
	}

	t.Run("slice", func(t *testing.T) {
		writer := &memoryWriter{
			opts: opts,
		}
		sl, err := slicer.New(ctx, writer, usr, containerID, usrID, nil)
		require.NoError(t, err)

		payload := testutil.RandByteSlice(maxObjectSize * overheadAmount)

		_, err = sl.Put(ctx, bytes.NewBuffer(payload), nil)
		require.NoError(t, err)

		require.Equal(t, overheadAmount+1, uint64(len(writer.headers)))

		for i, h := range writer.headers {
			first := h.GetFirstID()

			if i == 0 {
				require.True(t, first.IsZero())
			} else {
				require.False(t, first.IsZero())
				require.Equal(t, *writer.firstObject, first)
			}

			require.Nil(t, h.SplitID())

			checkParentWithoutSplitInfo(h)
		}
	})

	t.Run("InitPayloadStream", func(t *testing.T) {
		writer := &memoryWriter{
			opts: opts,
		}
		sl, err := slicer.New(ctx, writer, usr, containerID, usrID, nil)
		require.NoError(t, err)

		payloadWriter, err := sl.InitPut(ctx, nil)
		require.NoError(t, err)

		for range overheadAmount {
			payload := testutil.RandByteSlice(maxObjectSize)

			_, err := payloadWriter.Write(payload)
			require.NoError(t, err)
		}

		require.NoError(t, payloadWriter.Close())
		require.Equal(t, overheadAmount+1, uint64(len(writer.headers)))

		for i, h := range writer.headers {
			first := h.GetFirstID()

			if i == 0 {
				require.True(t, first.IsZero())
			} else {
				require.False(t, first.IsZero())
				require.Equal(t, *writer.firstObject, first)
			}

			require.Nil(t, h.SplitID())

			checkParentWithoutSplitInfo(h)
		}
	})

	t.Run("no split info if no overflow", func(t *testing.T) {
		writer := &memoryWriter{
			opts: opts,
		}
		sl, err := slicer.New(ctx, writer, usr, containerID, usrID, nil)
		require.NoError(t, err)

		payload := testutil.RandByteSlice(maxObjectSize - 1)

		_, err = sl.Put(ctx, bytes.NewBuffer(payload), nil)
		require.NoError(t, err)

		require.Equal(t, uint64(1), uint64(len(writer.headers)))

		for _, h := range writer.headers {
			splitID := h.SplitID()
			require.Nil(t, splitID)
			checkParentWithoutSplitInfo(h)
		}
	})
}

func BenchmarkWritePayloadBuffer(b *testing.B) {
	for _, tc := range []struct {
		sizeLimit uint64
		size      uint64
	}{
		{sizeLimit: 1 << 10, size: 1},
		{sizeLimit: 1 << 10, size: 1 << 10},
		{sizeLimit: 1 << 10, size: 10 << 10},
		{sizeLimit: 1 << 10, size: 200 << 10},
		{sizeLimit: 1 << 26, size: 1 << 10},
	} {
		b.Run(fmt.Sprintf("limit=%d,size=%d", tc.sizeLimit, tc.size), func(b *testing.B) {
			ctx := context.Background()
			in, opts := randomInput(tc.size, tc.sizeLimit)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			b.Run("with payload buffer", func(b *testing.B) {
				opts := opts
				opts.SetPayloadBuffer(make([]byte, tc.sizeLimit+1))

				b.ReportAllocs()

				for b.Loop() {
					w, err := slicer.InitPut(ctx, discardObject{opts: opts}, hdr, in.signer, opts)
					require.NoError(b, err)

					_, err = w.Write(in.payload)
					if err == nil {
						err = w.Close()
					}
					require.NoError(b, err)
				}
			})

			b.Run("without payload buffer", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					w, err := slicer.InitPut(ctx, discardObject{opts: opts}, hdr, in.signer, opts)
					require.NoError(b, err)

					_, err = w.Write(in.payload)
					if err == nil {
						err = w.Close()
					}
					require.NoError(b, err)
				}
			})
		})
	}
}

func BenchmarkReadPayloadBuffer(b *testing.B) {
	for _, tc := range []struct {
		sizeLimit uint64
		size      uint64
	}{
		{sizeLimit: 1 << 10, size: 1},
		{sizeLimit: 1 << 10, size: 1 << 10},
		{sizeLimit: 1 << 10, size: 10 << 10},
		{sizeLimit: 1 << 10, size: 200 << 10},
		{sizeLimit: 1 << 26, size: 1 << 10},
	} {
		b.Run(fmt.Sprintf("limit=%d,size=%d", tc.sizeLimit, tc.size), func(b *testing.B) {
			ctx := context.Background()
			in, opts := randomInput(tc.size, tc.sizeLimit)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			b.Run("with payload buffer", func(b *testing.B) {
				opts := opts
				opts.SetPayloadBuffer(make([]byte, tc.sizeLimit+1))

				b.ReportAllocs()

				for b.Loop() {
					_, err := slicer.Put(ctx, discardObject{opts: opts}, hdr, in.signer, bytes.NewReader(in.payload), opts)
					require.NoError(b, err)
				}
			})

			b.Run("without payload buffer", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_, err := slicer.Put(ctx, discardObject{opts: opts}, hdr, in.signer, bytes.NewReader(in.payload), opts)
					require.NoError(b, err)
				}
			})
		})
	}
}

func TestOptions_SetPayloadBuffer(t *testing.T) {
	for _, tc := range []struct {
		dataSize     uint64
		payloadLimit uint64
		bufSize      int
	}{
		// buffer smaller than limit
		{dataSize: 0, payloadLimit: 200, bufSize: 50},
		{dataSize: 10, payloadLimit: 200, bufSize: 50},
		{dataSize: 50, payloadLimit: 200, bufSize: 50},
		{dataSize: 60, payloadLimit: 200, bufSize: 50},
		{dataSize: 100, payloadLimit: 200, bufSize: 50},
		{dataSize: 120, payloadLimit: 200, bufSize: 50},
		{dataSize: 150, payloadLimit: 200, bufSize: 50},
		{dataSize: 200, payloadLimit: 200, bufSize: 50},
		{dataSize: 210, payloadLimit: 200, bufSize: 50},
		// buffer of limit size
		{dataSize: 0, payloadLimit: 200, bufSize: 200},
		{dataSize: 10, payloadLimit: 200, bufSize: 200},
		{dataSize: 50, payloadLimit: 200, bufSize: 200},
		{dataSize: 60, payloadLimit: 200, bufSize: 200},
		{dataSize: 100, payloadLimit: 200, bufSize: 200},
		{dataSize: 120, payloadLimit: 200, bufSize: 200},
		{dataSize: 150, payloadLimit: 200, bufSize: 200},
		{dataSize: 200, payloadLimit: 200, bufSize: 200},
		{dataSize: 210, payloadLimit: 200, bufSize: 200},
		// buffer bigger than limit
		{dataSize: 0, payloadLimit: 200, bufSize: 210},
		{dataSize: 10, payloadLimit: 200, bufSize: 210},
		{dataSize: 50, payloadLimit: 200, bufSize: 210},
		{dataSize: 60, payloadLimit: 200, bufSize: 210},
		{dataSize: 100, payloadLimit: 200, bufSize: 210},
		{dataSize: 120, payloadLimit: 200, bufSize: 210},
		{dataSize: 150, payloadLimit: 200, bufSize: 210},
		{dataSize: 200, payloadLimit: 200, bufSize: 210},
		{dataSize: 210, payloadLimit: 200, bufSize: 210},
	} {
		t.Run(fmt.Sprintf("with_buffer=%d_data=%d_limit=%d", tc.bufSize, tc.dataSize, tc.payloadLimit), func(t *testing.T) {
			in, opts := randomInput(tc.dataSize, tc.payloadLimit)
			if tc.bufSize > 0 {
				opts.SetPayloadBuffer(make([]byte, tc.bufSize))
			}

			checker := &slicedObjectChecker{
				opts:           opts,
				tb:             t,
				input:          in,
				chainCollector: newChainCollector(t),
			}

			testSlicerByHeaderType(t, checker, in, opts)
		})
	}
}

func TestKnownPayloadSize(t *testing.T) {
	ctx := context.Background()
	t.Run("overflow", func(t *testing.T) {
		t.Run("read", func(t *testing.T) {
			in, opts := randomInput(1, 1)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			opts.SetPayloadSize(20)
			r := bytes.NewReader(make([]byte, 21))

			_, err := slicer.Put(ctx, discardObject{opts: opts}, hdr, in.signer, r, opts)
			require.ErrorContains(t, err, "payload size exceeded")
		})

		t.Run("write", func(t *testing.T) {
			in, opts := randomInput(1, 1)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			opts.SetPayloadSize(20)

			w, err := slicer.InitPut(ctx, discardObject{opts: opts}, hdr, in.signer, opts)
			require.NoError(t, err)

			for i := range byte(21) {
				_, err = w.Write([]byte{1})
				if i < 20 {
					require.NoError(t, err)
				} else {
					require.ErrorContains(t, err, "payload size exceeded")
				}
			}
		})
	})

	t.Run("flaw", func(t *testing.T) {
		t.Run("read", func(t *testing.T) {
			in, opts := randomInput(1, 1)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			opts.SetPayloadSize(20)
			r := bytes.NewReader(make([]byte, 19))

			_, err := slicer.Put(ctx, discardObject{opts: opts}, hdr, in.signer, r, opts)
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})

		t.Run("write", func(t *testing.T) {
			in, opts := randomInput(1, 1)
			obj := objecttest.Object()
			hdr := *obj.CutPayload()

			opts.SetPayloadSize(20)

			w, err := slicer.InitPut(ctx, discardObject{opts: opts}, hdr, in.signer, opts)
			require.NoError(t, err)

			_, err = w.Write(make([]byte, 19))
			require.NoError(t, err)

			err = w.Close()
			require.ErrorIs(t, err, io.ErrUnexpectedEOF)
		})
	})
}

func BenchmarkKnownPayloadSize(b *testing.B) {
	ctx := context.Background()
	for _, tc := range []struct {
		sizeLimit uint64
		size      uint64
	}{
		{sizeLimit: 1 << 10, size: 1},
		{sizeLimit: 1 << 10, size: 1 << 10},
		{sizeLimit: 1 << 10, size: 10 << 10},
	} {
		b.Run(fmt.Sprintf("limit=%d,size=%d", tc.sizeLimit, tc.size), func(b *testing.B) {
			b.Run("read", func(b *testing.B) {
				obj := objecttest.Object()
				hdr := *obj.CutPayload()
				signer := user.NewSigner(neofscryptotest.Signer(), usertest.ID())
				payload := testutil.RandByteSlice(tc.size)

				var opts slicer.Options
				opts.SetObjectPayloadLimit(tc.sizeLimit)
				opts.SetPayloadSize(tc.size)

				b.ReportAllocs()

				for b.Loop() {
					_, err := slicer.Put(ctx, discardObject{opts: opts}, hdr, signer, bytes.NewReader(payload), opts)
					require.NoError(b, err)
				}
			})

			b.Run("write", func(b *testing.B) {
				obj := objecttest.Object()
				hdr := *obj.CutPayload()
				signer := user.NewSigner(neofscryptotest.Signer(), usertest.ID())
				payload := testutil.RandByteSlice(tc.size)

				var opts slicer.Options
				opts.SetObjectPayloadLimit(tc.sizeLimit)
				opts.SetPayloadSize(tc.size)

				b.ReportAllocs()

				for b.Loop() {
					w, err := slicer.InitPut(ctx, discardObject{opts: opts}, hdr, signer, opts)
					require.NoError(b, err)

					_, err = w.Write(payload)
					if err == nil {
						err = w.Close()
					}
					require.NoError(b, err)
				}
			})
		})
	}
}
