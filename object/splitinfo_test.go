package object_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
	"github.com/stretchr/testify/require"
)

var validSplitInfo object.SplitInfo // set by init.

// with required fields only.
var validMinSplitInfos []object.SplitInfo // set by init.

func init() {
	validSplitInfo.SetSplitID(anyValidSplitID)
	validSplitInfo.SetLastPart(anyValidIDs[0])
	validSplitInfo.SetLink(anyValidIDs[1])
	validSplitInfo.SetFirstPart(anyValidIDs[2])

	validMinSplitInfos = make([]object.SplitInfo, 2)
	validMinSplitInfos[0].SetLastPart(anyValidIDs[0])
	validMinSplitInfos[1].SetLink(anyValidIDs[1])
}

// corresponds to validSplitInfo.
var validBinSplitInfo = []byte{
	10, 16, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41, 18, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110,
	125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 26, 34, 10,
	32, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67, 213, 243,
	57, 253, 127, 179, 235, 34, 34, 10, 32, 206, 228, 247, 217, 41, 247, 159, 215, 79, 226, 53, 153, 133, 16, 102, 104, 2, 234,
	35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197, 242,
}

// corresponds to validMinSplitInfos.
var validBinMinSplitInfos = [][]byte{{
	18, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70,
	246, 8, 139, 247, 174, 53, 60,
}, {
	26, 34, 10, 32, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46, 135, 32, 116, 172, 67,
	213, 243, 57, 253, 127, 179, 235,
}}

// corresponds to validSplitInfo.
var validJSONSplitInfo = `
{
 "splitId": "4IQDUCAsRbi5IOLJzsSTKQ==",
 "lastPart": {
  "value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="
 },
 "link": {
  "value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+s="
 },
 "firstPart": {
  "value": "zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfI="
 }
}
`

// corresponds to validMinSplitInfos.
var validJSONMinSplitInfos = []string{`
{
 "splitId": "",
 "lastPart": {
  "value": "sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="
 },
 "link": null,
 "firstPart": null
}
`, `
{
 "splitId": "",
 "lastPart": null,
 "link": {
  "value": "5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+s="
 },
 "firstPart": null
}
`}

func TestSplitInfo_SetSplitID(t *testing.T) {
	var s object.SplitInfo
	require.Zero(t, s.SplitID())

	s.SetSplitID(anyValidSplitID)
	require.Equal(t, anyValidSplitID, s.SplitID())

	b := bytes.Clone(anyValidSplitIDBytes)
	b[0]++
	otherSplitID := object.NewSplitIDFromV2(b)
	s.SetSplitID(otherSplitID)
	require.Equal(t, otherSplitID, s.SplitID())
}

func testSplitInfoIDField(
	t testing.TB,
	get func(info object.SplitInfo) oid.ID,
	getFlag func(info object.SplitInfo) (oid.ID, bool),
	set func(*object.SplitInfo, oid.ID),
) {
	var s object.SplitInfo
	require.True(t, get(s).IsZero())
	_, ok := getFlag(s)
	require.False(t, ok)

	set(&s, anyValidIDs[0])
	require.Equal(t, anyValidIDs[0], get(s))
	res, ok := getFlag(s)
	require.True(t, ok)
	require.Equal(t, anyValidIDs[0], res)

	set(&s, anyValidIDs[1])
	require.Equal(t, anyValidIDs[1], get(s))
	res, ok = getFlag(s)
	require.True(t, ok)
	require.Equal(t, anyValidIDs[1], res)
}

func TestSplitInfo_SetLastPart(t *testing.T) {
	testSplitInfoIDField(t, object.SplitInfo.GetLastPart, object.SplitInfo.LastPart, (*object.SplitInfo).SetLastPart)
}

func TestSplitInfo_SetLink(t *testing.T) {
	testSplitInfoIDField(t, object.SplitInfo.GetLink, object.SplitInfo.Link, (*object.SplitInfo).SetLink)
}

func TestSplitInfo_SetFirstPart(t *testing.T) {
	testSplitInfoIDField(t, object.SplitInfo.GetFirstPart, object.SplitInfo.FirstPart, (*object.SplitInfo).SetFirstPart)
}

func TestSplitInfo(t *testing.T) {
	s := object.NewSplitInfo()
	splitID := object.NewSplitID()
	lastPart := oidtest.ID()
	link := oidtest.ID()
	firstPart := oidtest.ID()

	s.SetSplitID(splitID)
	require.Equal(t, splitID, s.SplitID())

	s.SetLastPart(lastPart)
	require.Equal(t, lastPart, s.GetLastPart())

	s.SetLink(link)
	require.Equal(t, link, s.GetLink())

	s.SetFirstPart(firstPart)
	require.Equal(t, firstPart, s.GetFirstPart())
}

func TestSplitInfoMarshal(t *testing.T) {
	testMessage := func(t *testing.T, s object.SplitInfo) {
		var newS object.SplitInfo
		require.NoError(t, newS.FromProtoMessage(s.ProtoMessage()))

		require.Equal(t, s, newS)
	}
	testMarshal := func(t *testing.T, s object.SplitInfo) {
		var newS object.SplitInfo

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

		testMessage(t, *s)
		testMarshal(t, *s)
	})
	t.Run("good, only link is set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLink(oidtest.ID())

		testMessage(t, *s)
		testMarshal(t, *s)
	})
	t.Run("good, only last part is set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())
		s.SetLastPart(oidtest.ID())

		testMessage(t, *s)
		testMarshal(t, *s)
	})
	t.Run("bad, no fields are set", func(t *testing.T) {
		s := object.NewSplitInfo()
		s.SetSplitID(object.NewSplitID())

		require.Error(t, object.NewSplitInfo().Unmarshal(s.Marshal()))
	})
}

func TestSplitInfo_FromProtoMessage(t *testing.T) {
	m := &protoobject.SplitInfo{
		SplitId:   anyValidSplitIDBytes,
		LastPart:  protoIDFromBytes(anyValidIDs[0][:]),
		Link:      protoIDFromBytes(anyValidIDs[1][:]),
		FirstPart: protoIDFromBytes(anyValidIDs[2][:]),
	}

	var s object.SplitInfo
	require.NoError(t, s.FromProtoMessage(m))
	require.Equal(t, anyValidSplitID, s.SplitID())
	require.Equal(t, anyValidIDs[0], s.GetLastPart())
	id, ok := s.LastPart()
	require.True(t, ok)
	require.Equal(t, anyValidIDs[0], id)
	require.Equal(t, anyValidIDs[1], s.GetLink())
	id, ok = s.Link()
	require.True(t, ok)
	require.Equal(t, anyValidIDs[1], id)
	require.Equal(t, anyValidIDs[2], s.GetFirstPart())
	id, ok = s.FirstPart()
	require.True(t, ok)
	require.Equal(t, anyValidIDs[2], id)

	// reset optional fields
	m.SplitId = nil
	m.FirstPart = nil
	m.Link = nil
	s2 := s
	require.NoError(t, s2.FromProtoMessage(m))

	require.Zero(t, s2.SplitID())
	require.True(t, s2.GetFirstPart().IsZero())
	_, ok = s2.FirstPart()
	require.False(t, ok)
	require.Equal(t, anyValidIDs[0], s2.GetLastPart())
	require.True(t, s2.GetFirstPart().IsZero())
	_, ok = s2.Link()
	require.True(t, s2.GetLink().IsZero())
	require.False(t, ok)
	_, ok = s2.Link()
	require.False(t, ok)

	// either linking or last part must be set, so lets swap
	m.Link = protoIDFromBytes(anyValidIDs[1][:])
	m.LastPart = nil
	require.NoError(t, s2.FromProtoMessage(m))

	require.Equal(t, anyValidIDs[1], s2.GetLink())
	require.True(t, s2.GetLastPart().IsZero())
	_, ok = s2.LastPart()
	require.False(t, ok)

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name, err string
			corrupt   func(*protoobject.SplitInfo)
		}{
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				corrupt: func(m *protoobject.SplitInfo) { m.SplitId = anyValidSplitIDBytes[:15] }},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				corrupt: func(m *protoobject.SplitInfo) { m.SplitId = append(anyValidSplitIDBytes[:], 1) }},
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
				corrupt: func(m *protoobject.SplitInfo) {
					m.SplitId = bytes.Clone(anyValidSplitIDBytes[:])
					m.SplitId[6] = 3 << 4
				}},
			{name: "last part/nil value", err: "could not convert last part object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.LastPart = protoIDFromBytes(nil) }},
			{name: "last part/empty value", err: "could not convert last part object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.LastPart = protoIDFromBytes([]byte{}) }},
			{name: "last part/undersize", err: "could not convert last part object ID: invalid length 31",
				corrupt: func(m *protoobject.SplitInfo) { m.LastPart = protoIDFromBytes(make([]byte, 31)) }},
			{name: "last part/oversize", err: "could not convert last part object ID: invalid length 33",
				corrupt: func(m *protoobject.SplitInfo) { m.LastPart = protoIDFromBytes(make([]byte, 33)) }},
			{name: "link/nil value", err: "could not convert link object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.Link = protoIDFromBytes(nil) }},
			{name: "link/empty value", err: "could not convert link object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.Link = protoIDFromBytes([]byte{}) }},
			{name: "link/undersize", err: "could not convert link object ID: invalid length 31",
				corrupt: func(m *protoobject.SplitInfo) { m.Link = protoIDFromBytes(make([]byte, 31)) }},
			{name: "link/oversize", err: "could not convert link object ID: invalid length 33",
				corrupt: func(m *protoobject.SplitInfo) { m.Link = protoIDFromBytes(make([]byte, 33)) }},
			{name: "first part/nil value", err: "could not convert first part object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.FirstPart = protoIDFromBytes(nil) }},
			{name: "first part/empty value", err: "could not convert first part object ID: invalid length 0",
				corrupt: func(m *protoobject.SplitInfo) { m.FirstPart = protoIDFromBytes([]byte{}) }},
			{name: "first part/undersize", err: "could not convert first part object ID: invalid length 31",
				corrupt: func(m *protoobject.SplitInfo) { m.FirstPart = protoIDFromBytes(make([]byte, 31)) }},
			{name: "first part/oversize", err: "could not convert first part object ID: invalid length 33",
				corrupt: func(m *protoobject.SplitInfo) { m.FirstPart = protoIDFromBytes(make([]byte, 33)) }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				s2 := s
				m := s2.ProtoMessage()
				tc.corrupt(m)
				require.EqualError(t, new(object.SplitInfo).FromProtoMessage(m), tc.err)
			})
		}
	})
}

func TestSplitInfo_ProtoMessage(t *testing.T) {
	var s object.SplitInfo

	// zero
	m := s.ProtoMessage()
	require.Zero(t, m.GetSplitId())
	require.Zero(t, m.GetFirstPart())
	require.Zero(t, m.GetLink())
	require.Zero(t, m.GetFirstPart())

	// filled
	m = validSplitInfo.ProtoMessage()
	require.EqualValues(t, anyValidSplitIDBytes, m.GetSplitId())
	require.Equal(t, anyValidIDs[0][:], m.GetLastPart().GetValue())
	require.Equal(t, anyValidIDs[1][:], m.GetLink().GetValue())
	require.Equal(t, anyValidIDs[2][:], m.GetFirstPart().GetValue())
}

func TestSplitInfo_Marshal(t *testing.T) {
	require.Equal(t, validBinSplitInfo, validSplitInfo.Marshal())
	for i := range validMinSplitInfos {
		require.Equal(t, validBinMinSplitInfos[i], validMinSplitInfos[i].Marshal())
	}
}

func TestSplitInfo_Unmarshal(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("protobuf", func(t *testing.T) {
			err := new(object.SplitInfo).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range []struct {
			name, err string
			b         []byte
		}{
			{name: "empty", err: "neither link object ID nor last part object ID is set",
				b: []byte{}},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				b: []byte{10, 15, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 18, 34, 10, 32, 178, 74,
					58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70, 246,
					8, 139, 247, 174, 53, 60}},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				b: []byte{10, 17, 224, 132, 3, 80, 32, 44, 69, 184, 185, 32, 226, 201, 206, 196, 147, 41, 1, 18, 34, 10, 32,
					178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173,
					70, 246, 8, 139, 247, 174, 53, 60}},
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
				b: []byte{10, 16, 224, 132, 3, 80, 32, 44, 48, 184, 185, 32, 226, 201, 206, 196, 147, 41, 18, 34, 10, 32, 178,
					74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44, 18, 56, 117, 173, 70,
					246, 8, 139, 247, 174, 53, 60}},
			{name: "last part/empty value", err: "could not convert last part object ID: invalid length 0",
				b: []byte{18, 0}},
			{name: "last part/undersize", err: "could not convert last part object ID: invalid length 31",
				b: []byte{18, 33, 10, 31, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53}},
			{name: "last part/oversize", err: "could not convert last part object ID: invalid length 33",
				b: []byte{18, 35, 10, 33, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 1}},
			{name: "link/empty value", err: "could not convert link object ID: invalid length 0",
				b: []byte{26, 0}},
			{name: "link/undersize", err: "could not convert link object ID: invalid length 31",
				b: []byte{26, 33, 10, 31, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46,
					135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179}},
			{name: "link/oversize", err: "could not convert link object ID: invalid length 33",
				b: []byte{26, 35, 10, 33, 229, 77, 63, 235, 2, 9, 165, 123, 116, 123, 47, 65, 22, 34, 214, 76, 45, 225, 21, 46,
					135, 32, 116, 172, 67, 213, 243, 57, 253, 127, 179, 235, 1}},
			{name: "first part/empty value", err: "could not convert first part object ID: invalid length 0",
				b: []byte{18, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 34, 0}},
			{name: "first part/undersize", err: "could not convert first part object ID: invalid length 31",
				b: []byte{18, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 34, 33, 10, 31, 206, 228, 247, 217, 41, 247, 159, 215, 79,
					226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197}},
			{name: "first part/oversize", err: "could not convert first part object ID: invalid length 33",
				b: []byte{18, 34, 10, 32, 178, 74, 58, 219, 46, 3, 110, 125, 220, 81, 238, 35, 27, 6, 228, 193, 190, 224, 77, 44,
					18, 56, 117, 173, 70, 246, 8, 139, 247, 174, 53, 60, 34, 35, 10, 33, 206, 228, 247, 217, 41, 247, 159, 215,
					79, 226, 53, 153, 133, 16, 102, 104, 2, 234, 35, 220, 236, 112, 101, 24, 235, 126, 173, 229, 161, 202, 197,
					242, 1}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.SplitInfo).Unmarshal(tc.b), tc.err)
			})
		}
	})

	var si object.SplitInfo
	// min
	require.NoError(t, si.Unmarshal(validBinMinSplitInfos[0]))
	require.Zero(t, si.SplitID())
	require.True(t, si.GetLink().IsZero())
	require.True(t, si.GetFirstPart().IsZero())
	require.NoError(t, si.Unmarshal(validBinMinSplitInfos[1]))
	require.Zero(t, si.SplitID())
	require.True(t, si.GetLastPart().IsZero())
	require.True(t, si.GetFirstPart().IsZero())

	// filled
	require.NoError(t, si.Unmarshal(validBinSplitInfo))
	require.Equal(t, validSplitInfo, si)
}

func TestNewSplitInfo(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		si := object.NewSplitInfo()

		// check initial values
		require.Nil(t, si.SplitID())
		require.True(t, si.GetLastPart().IsZero())
		require.True(t, si.GetLink().IsZero())
		require.True(t, si.GetFirstPart().IsZero())

		// convert to v2 message
		m := si.ProtoMessage()

		require.Nil(t, m.GetSplitId())
		require.Nil(t, m.GetLastPart())
		require.Nil(t, m.GetLink())
		require.Nil(t, m.GetFirstPart())
	})
}

func TestSplitInfo_MarshalJSON(t *testing.T) {
	b, err := json.MarshalIndent(validSplitInfo, "", " ")
	require.NoError(t, err)
	require.JSONEq(t, validJSONSplitInfo, string(b))

	for i := range validMinSplitInfos {
		b, err = json.MarshalIndent(validMinSplitInfos[i], "", " ")
		require.NoError(t, err)
		require.JSONEq(t, validJSONMinSplitInfos[i], string(b))
	}
}

func TestSplitInfo_UnmarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		t.Run("JSON", func(t *testing.T) {
			err := new(object.SplitInfo).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range []struct{ name, err, j string }{
			{name: "empty", err: "neither link object ID nor last part object ID is set",
				j: `{}`},
			{name: "split ID/undersize", err: "invalid split ID: invalid UUID (got 15 bytes)",
				j: `{"splitId":"4IQDUCAsRbi5IOLJzsST"}`},
			{name: "split ID/oversize", err: "invalid split ID: invalid UUID (got 17 bytes)",
				j: `{"splitId":"4IQDUCAsRbi5IOLJzsSTKQE="}`},
			{name: "split ID/wrong version", err: "invalid split ID: wrong UUID version 3, expected 4",
				j: `{"splitId":"4IQDUCAsMLi5IOLJzsSTKQ=="}`},
			{name: "last part/empty value", err: "could not convert last part object ID: invalid length 0",
				j: `{"lastPart":{"value":""}}`},
			{name: "last part/undersize", err: "could not convert last part object ID: invalid length 31",
				j: `{"lastPart":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNQ=="}}`},
			{name: "last part/oversize", err: "could not convert last part object ID: invalid length 33",
				j: `{"lastPart":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTwB"}}`},
			{name: "link/empty value", err: "could not convert link object ID: invalid length 0",
				j: `{"link":{"value":""}}`},
			{name: "link/undersize", err: "could not convert link object ID: invalid length 31",
				j: `{"link":{"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/sw=="}}`},
			{name: "link/oversize", err: "could not convert link object ID: invalid length 33",
				j: `{"link":{"value":"5U0/6wIJpXt0ey9BFiLWTC3hFS6HIHSsQ9XzOf1/s+sB"}}`},
			{name: "first part/empty value", err: "could not convert first part object ID: invalid length 0",
				j: `{"lastPart":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="}, "firstPart":{"value":""}}`},
			{name: "first part/undersize", err: "could not convert first part object ID: invalid length 31",
				j: `{"lastPart":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},"firstPart":{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxQ=="}}`},
			{name: "first part/oversize", err: "could not convert first part object ID: invalid length 33",
				j: `{"lastPart":{"value":"sko62y4Dbn3cUe4jGwbkwb7gTSwSOHWtRvYIi/euNTw="},"firstPart":{"value":"zuT32Sn3n9dP4jWZhRBmaALqI9zscGUY636t5aHKxfIB"}}`},
		} {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(object.SplitInfo).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})

	var si object.SplitInfo
	// min
	require.NoError(t, si.UnmarshalJSON([]byte(validJSONMinSplitInfos[0])))
	require.Zero(t, si.SplitID())
	require.True(t, si.GetLink().IsZero())
	require.True(t, si.GetFirstPart().IsZero())
	require.NoError(t, si.UnmarshalJSON([]byte(validJSONMinSplitInfos[1])))
	require.Zero(t, si.SplitID())
	require.True(t, si.GetLastPart().IsZero())
	require.True(t, si.GetFirstPart().IsZero())

	// filled
	require.NoError(t, si.UnmarshalJSON([]byte(validJSONSplitInfo)))
	require.Equal(t, validSplitInfo, si)
}
