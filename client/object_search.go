package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	v2object "github.com/nspcc-dev/neofs-api-go/v2/object"
	v2refs "github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/stat"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// PrmObjectSearch groups optional parameters of ObjectSearch operation.
type PrmObjectSearch struct {
	sessionContainer

	filters object.SearchFilters
}

// MarkLocal tells the server to execute the operation locally.
func (x *PrmObjectSearch) MarkLocal() {
	x.meta.SetTTL(1)
}

// WithBearerToken attaches bearer token to be used for the operation.
//
// If set, underlying eACL rules will be used in access control.
//
// Must be signed.
func (x *PrmObjectSearch) WithBearerToken(t bearer.Token) {
	var v2token acl.BearerToken
	t.WriteToV2(&v2token)
	x.meta.SetBearerToken(&v2token)
}

// WithXHeaders specifies list of extended headers (string key-value pairs)
// to be attached to the request. Must have an even length.
//
// Slice must not be mutated until the operation completes.
func (x *PrmObjectSearch) WithXHeaders(hs ...string) {
	writeXHeadersToMeta(hs, &x.meta)
}

// SetFilters sets filters by which to select objects. All container objects
// match unset/empty filters.
func (x *PrmObjectSearch) SetFilters(filters object.SearchFilters) {
	x.filters = filters
}

// ObjectListReader is designed to read list of object identifiers from NeoFS system.
//
// Must be initialized using Client.ObjectSearch, any other usage is unsafe.
type ObjectListReader struct {
	client          *Client
	cancelCtxStream context.CancelFunc
	err             error
	stream          searchObjectsResponseStream
	tail            []v2refs.ObjectID

	statisticCallback shortStatisticCallback
}

// Read reads another list of the object identifiers. Works similar to
// io.Reader.Read but copies oid.ID.
//
// Failure reason can be received via Close.
//
// Panics if buf has zero length.
func (x *ObjectListReader) Read(buf []oid.ID) (int, error) {
	if len(buf) == 0 {
		panic("empty buffer in ObjectListReader.ReadList")
	}

	read := copyIDBuffers(buf, x.tail)
	x.tail = x.tail[read:]

	if len(buf) == read {
		return read, nil
	}

	for {
		var resp v2object.SearchResponse
		x.err = x.stream.Read(&resp)
		if x.err != nil {
			return read, x.err
		}

		x.err = x.client.processResponse(&resp)
		if x.err != nil {
			return read, x.err
		}

		// read new chunk of objects
		ids := resp.GetBody().GetIDList()
		if len(ids) == 0 {
			// just skip empty lists since they are not prohibited by protocol
			continue
		}

		ln := copyIDBuffers(buf[read:], ids)
		read += ln

		if read == len(buf) {
			// save the tail
			x.tail = append(x.tail, ids[ln:]...)

			return read, nil
		}
	}
}

func copyIDBuffers(dst []oid.ID, src []v2refs.ObjectID) int {
	var i int
	for ; i < len(dst) && i < len(src); i++ {
		_ = dst[i].ReadFromV2(src[i])
	}
	return i
}

// Iterate iterates over the list of found object identifiers.
// f can return true to stop iteration earlier.
//
// Returns an error if object can't be read.
func (x *ObjectListReader) Iterate(f func(oid.ID) bool) error {
	buf := make([]oid.ID, 1)

	for {
		_, err := x.Read(buf)
		if err != nil {
			return x.Close()
		}
		if f(buf[0]) {
			return nil
		}
	}
}

// Close ends reading list of the matched objects and returns the result of the operation
// along with the final results. Must be called after using the ObjectListReader.
//
// Any client's internal or transport errors are returned as Go built-in error.
// If Client is tuned to resolve NeoFS API statuses, then NeoFS failures
// codes are returned as error.
//
// Return errors:
//   - global (see Client docs)
//   - [apistatus.ErrContainerNotFound]
//   - [apistatus.ErrObjectAccessDenied]
//   - [apistatus.ErrSessionTokenExpired]
func (x *ObjectListReader) Close() error {
	var err error
	if x.statisticCallback != nil {
		defer func() {
			x.statisticCallback(err)
		}()
	}

	defer x.cancelCtxStream()

	if x.err != nil && !errors.Is(x.err, io.EOF) {
		err = x.err
		return err
	}

	return nil
}

// ObjectSearchInit initiates object selection through a remote server using NeoFS API protocol.
//
// The call only opens the transmission channel, explicit fetching of matched objects
// is done using the ObjectListReader. Exactly one return value is non-nil.
// Resulting reader must be finally closed.
//
// Context is required and must not be nil. It is used for network communication.
//
// Signer is required and must not be nil. The operation is executed on behalf of the account corresponding to
// the specified Signer, which is taken into account, in particular, for access control.
//
// Return errors:
//   - [ErrMissingSigner]
func (c *Client) ObjectSearchInit(ctx context.Context, containerID cid.ID, signer user.Signer, prm PrmObjectSearch) (*ObjectListReader, error) {
	var err error
	defer func() {
		c.sendStatistic(stat.MethodObjectSearch, err)()
	}()

	if signer == nil {
		return nil, ErrMissingSigner
	}

	var cidV2 v2refs.ContainerID
	containerID.WriteToV2(&cidV2)

	var body v2object.SearchRequestBody
	body.SetVersion(1)
	body.SetContainerID(&cidV2)
	body.SetFilters(prm.filters.ToV2())

	// init reader
	var req v2object.SearchRequest
	req.SetBody(&body)
	c.prepareRequest(&req, &prm.meta)

	buf := c.buffers.Get().(*[]byte)
	err = signServiceMessage(signer, &req, *buf)
	c.buffers.Put(buf)
	if err != nil {
		err = fmt.Errorf("sign request: %w", err)
		return nil, err
	}

	var r ObjectListReader
	ctx, r.cancelCtxStream = context.WithCancel(ctx)

	r.stream, err = c.server.searchObjects(ctx, req)
	if err != nil {
		err = fmt.Errorf("open stream: %w", err)
		return nil, err
	}
	r.client = c
	r.statisticCallback = func(err error) {
		c.sendStatistic(stat.MethodObjectSearchStream, err)()
	}

	return &r, nil
}
