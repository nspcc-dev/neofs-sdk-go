package object_test

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	objv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestSplitInfo(t *testing.T) {
	s := object.NewSplitInfo()
	splitID := object.NewSplitID()
	lastPart := generateID()
	link := generateID()

	s.SetSplitID(splitID)
	require.Equal(t, splitID, s.SplitID())

	s.SetLastPart(lastPart)
	lp, set := s.LastPart()
	require.True(t, set)
	require.Equal(t, lastPart, lp)

	s.SetLink(link)
	l, set := s.Link()
	require.True(t, set)
	require.Equal(t, link, l)

	t.Run("to and from v2", func(t *testing.T) {
		v2 := s.ToV2()
		newS := object.NewSplitInfoFromV2(v2)

		require.Equal(t, s, newS)
	})

	t.Run("marshal and unmarshal", func(t *testing.T) {
		data, err := s.Marshal()
		require.NoError(t, err)

		newS := object.NewSplitInfo()

		err = newS.Unmarshal(data)
		require.NoError(t, err)
		require.Equal(t, s, newS)
	})
}

func generateID() oid.ID {
	var buf [32]byte
	_, _ = rand.Read(buf[:])

	var id oid.ID
	id.SetSHA256(buf)

	return id
}

func TestNewSplitInfoFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *objv2.SplitInfo

		require.Nil(t, object.NewSplitInfoFromV2(x))
	})
}

func TestSplitInfo_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *object.SplitInfo

		require.Nil(t, x.ToV2())
	})
}

func TestNewSplitInfo(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		si := object.NewSplitInfo()

		// check initial values
		require.Nil(t, si.SplitID())
		_, set := si.LastPart()
		require.False(t, set)
		_, set = si.Link()
		require.False(t, set)

		// convert to v2 message
		siV2 := si.ToV2()

		require.Nil(t, siV2.GetSplitID())
		require.Nil(t, siV2.GetLastPart())
		require.Nil(t, siV2.GetLink())
	})
}

func TestSplitInfoMarshalJSON(t *testing.T) {
	s := object.NewSplitInfo()
	s.SetSplitID(object.NewSplitID())
	s.SetLastPart(generateID())
	s.SetLink(generateID())

	data, err := s.MarshalJSON()
	require.NoError(t, err)

	actual := object.NewSplitInfo()
	require.NoError(t, json.Unmarshal(data, actual))
	require.Equal(t, s, actual)
}
