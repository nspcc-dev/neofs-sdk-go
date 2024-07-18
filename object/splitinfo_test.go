package object_test

import (
	"encoding/json"
	"testing"

	objv2 "github.com/nspcc-dev/neofs-api-go/v2/object"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/stretchr/testify/require"
)

func TestSplitInfo(t *testing.T) {
	s := object.NewSplitInfo()
	splitID := object.NewSplitID()
	lastPart := oidtest.ID()
	link := oidtest.ID()
	firstPart := oidtest.ID()

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

	s.SetFirstPart(firstPart)
	ip, set := s.FirstPart()
	require.True(t, set)
	require.Equal(t, firstPart, ip)
}

func TestSplitInfoMarshal(t *testing.T) {
	testToV2 := func(t *testing.T, s *object.SplitInfo) {
		v2 := s.ToV2()
		newS := object.NewSplitInfoFromV2(v2)

		require.Equal(t, s, newS)
	}
	testMarshal := func(t *testing.T, s *object.SplitInfo) {
		newS := object.NewSplitInfo()

		err := newS.Unmarshal(s.Marshal())
		require.NoError(t, err)
		require.Equal(t, s, newS)
	}

	t.Run("good, all fields are set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLink(oidtest.ID())
		s.SetLastPart(oidtest.ID())
		s.SetFirstPart(oidtest.ID())

		testToV2(t, s)
		testMarshal(t, s)
	})
	t.Run("good, only link is set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLink(oidtest.ID())

		testToV2(t, s)
		testMarshal(t, s)
	})
	t.Run("good, only last part is set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLastPart(oidtest.ID())

		testToV2(t, s)
		testMarshal(t, s)
	})
	t.Run("bad, no fields are set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())

		require.Error(t, object.NewSplitInfo().Unmarshal(s.Marshal()))
	})
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
		_, set = si.FirstPart()
		require.False(t, set)

		// convert to v2 message
		siV2 := si.ToV2()

		require.Nil(t, siV2.GetSplitID())
		require.Nil(t, siV2.GetLastPart())
		require.Nil(t, siV2.GetLink())
		require.Nil(t, siV2.GetFirstPart())
	})
}

func TestSplitInfoMarshalJSON(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLastPart(oidtest.ID())
		s.SetLink(oidtest.ID())
		s.SetFirstPart(oidtest.ID())

		data, err := s.MarshalJSON()
		require.NoError(t, err)

		actual := object.NewSplitInfo()
		require.NoError(t, json.Unmarshal(data, actual))
		require.Equal(t, s, actual)
	})
	t.Run("bad link", func(t *testing.T) {
		data := `{"splitId":"Sn707289RrqDyJOrZMbMoQ==","lastPart":{"value":"Y7baWE0UdUOBr1ELKX3Q5v1LKRubQUbI81Q5UxCVeow="},"link":{"value":"bad"}}`
		require.Error(t, json.Unmarshal([]byte(data), object.NewSplitInfo()))
	})
	t.Run("bad last part", func(t *testing.T) {
		data := `{"splitId":"Sn707289RrqDyJOrZMbMoQ==","lastPart":{"value":"bad"},"link":{"value":"eRyPNCNNxHfxPcjijlv05HEcdoep/b7eHNLRSmDlnts="}}`
		require.Error(t, json.Unmarshal([]byte(data), object.NewSplitInfo()))
	})
	t.Run("bad first part", func(t *testing.T) {
		data := `{"splitId":"Sn707289RrqDyJOrZMbMoQ==","firstPart":{"value":"bad"},"link":{"value":"eRyPNCNNxHfxPcjijlv05HEcdoep/b7eHNLRSmDlnts="}}`
		require.Error(t, json.Unmarshal([]byte(data), object.NewSplitInfo()))
	})
}
