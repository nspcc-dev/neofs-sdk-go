package object_test

import (
	"crypto/rand"
	"encoding/json"
	"testing"

	objectv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/stretchr/testify/require"
)

func TestSplitInfo(t *testing.T) {
	var s object.SplitInfo
	splitID := object.NewSplitID()
	lastPart := generateID()
	link := generateID()

	s.SetSplitID(splitID)
	require.Equal(t, splitID, s.SplitID())

	s.SetLastPart(lastPart)
	require.Equal(t, lastPart, s.LastPart())

	s.SetLink(link)
	require.Equal(t, link, s.Link())

	t.Run("to and from v2", func(t *testing.T) {
		var v2 objectv2.SplitInfo
		s.WriteToV2(&v2)

		var newS object.SplitInfo
		newS.ReadFromV2(v2)

		require.Equal(t, s, newS)
	})

	t.Run("marshal and unmarshal", func(t *testing.T) {
		data, err := s.Marshal()
		require.NoError(t, err)

		var newS object.SplitInfo

		err = newS.Unmarshal(data)
		require.NoError(t, err)
		require.Equal(t, s, newS)
	})
}

func generateID() *oid.ID {
	var buf [32]byte
	_, _ = rand.Read(buf[:])

	var id oid.ID
	id.SetSHA256(buf)

	return &id
}

func TestNewSplitInfoFromV2(t *testing.T) {
	t.Run("from zero V2", func(t *testing.T) {
		var (
			v2 objectv2.SplitInfo
			x  object.SplitInfo
		)

		x.ReadFromV2(v2)

		require.Nil(t, x.SplitID())
		require.Nil(t, x.Link())
		require.Nil(t, x.LastPart())
	})
}

func TestSplitInfo_ToV2(t *testing.T) {
	t.Run("zero to V2", func(t *testing.T) {
		var (
			x  object.SplitInfo
			v2 objectv2.SplitInfo
		)

		x.WriteToV2(&v2)

		require.Nil(t, v2.GetSplitID())
		require.Nil(t, v2.GetLink())
		require.Nil(t, v2.GetLastPart())
	})
}

func TestNewSplitInfo(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		var si object.SplitInfo

		// check initial values
		require.Nil(t, si.SplitID())
		require.Nil(t, si.LastPart())
		require.Nil(t, si.Link())
	})
}

func TestSplitInfoMarshalJSON(t *testing.T) {
	var s object.SplitInfo
	s.SetSplitID(object.NewSplitID())
	s.SetLastPart(generateID())
	s.SetLink(generateID())

	data, err := s.MarshalJSON()
	require.NoError(t, err)

	var actual object.SplitInfo
	require.NoError(t, json.Unmarshal(data, &actual))
	require.Equal(t, s, actual)
}
