package object

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type unsupportedOneOf struct{}

func (unsupportedOneOf) isHeadResponse_Body_Head()          {} //nolint:revive
func (unsupportedOneOf) isGetResponse_Body_ObjectPart()     {} //nolint:revive
func (unsupportedOneOf) isGetRangeResponse_Body_RangePart() {} //nolint:revive
func (unsupportedOneOf) isPutRequest_Body_ObjectPart()      {} //nolint:revive

func TestGetResponse_Body(t *testing.T) {
	var v GetResponse_Body
	v.ObjectPart = unsupportedOneOf{}
	require.PanicsWithValue(t, "unexpected object part object.unsupportedOneOf", func() {
		v.MarshaledSize()
	})
	require.PanicsWithValue(t, "unexpected object part object.unsupportedOneOf", func() {
		v.MarshalStable(make([]byte, 100))
	})
}

func TestHeadResponse_Body(t *testing.T) {
	var v HeadResponse_Body
	v.Head = unsupportedOneOf{}
	require.PanicsWithValue(t, "unexpected head part object.unsupportedOneOf", func() {
		v.MarshaledSize()
	})
	require.PanicsWithValue(t, "unexpected head part object.unsupportedOneOf", func() {
		v.MarshalStable(make([]byte, 100))
	})
}

func TestGetRangeResponse_Body(t *testing.T) {
	var v GetRangeResponse_Body
	v.RangePart = unsupportedOneOf{}
	require.PanicsWithValue(t, "unexpected range part object.unsupportedOneOf", func() {
		v.MarshaledSize()
	})
	require.PanicsWithValue(t, "unexpected range part object.unsupportedOneOf", func() {
		v.MarshalStable(make([]byte, 100))
	})
}

func TestPutRequest_Body(t *testing.T) {
	var v PutRequest_Body
	v.ObjectPart = unsupportedOneOf{}
	require.PanicsWithValue(t, "unexpected object part object.unsupportedOneOf", func() {
		v.MarshaledSize()
	})
	require.PanicsWithValue(t, "unexpected object part object.unsupportedOneOf", func() {
		v.MarshalStable(make([]byte, 100))
	})
}
