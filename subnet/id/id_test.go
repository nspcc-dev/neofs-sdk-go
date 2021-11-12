package subnetid_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	subnetid "github.com/nspcc-dev/neofs-sdk-go/subnet/id"
	subnetidtest "github.com/nspcc-dev/neofs-sdk-go/subnet/id/test"
	"github.com/stretchr/testify/require"
)

func TestIsZero(t *testing.T) {
	var id subnetid.ID

	require.True(t, subnetid.IsZero(id))

	id.SetNumber(13)
	require.False(t, subnetid.IsZero(id))

	id.SetNumber(0)
	require.True(t, subnetid.IsZero(id))
}

func TestID_FromV2(t *testing.T) {
	const num = 13

	var id1 subnetid.ID
	id1.SetNumber(num)

	var idv2 refs.SubnetID
	idv2.SetValue(num)

	var id2 subnetid.ID
	id2.FromV2(idv2)

	require.True(t, id1.Equals(&id2))
}

func TestID_WriteToV2(t *testing.T) {
	const num = 13

	var (
		id   subnetid.ID
		idv2 refs.SubnetID
	)

	id.WriteToV2(&idv2)
	require.Zero(t, idv2.GetValue())

	id.SetNumber(num)

	id.WriteToV2(&idv2)
	require.EqualValues(t, num, idv2.GetValue())
}

func TestID_Equals(t *testing.T) {
	const num = 13

	var id1, id2, idOther, id0 subnetid.ID

	id0.Equals(nil)

	id1.SetNumber(num)
	id2.SetNumber(num)
	idOther.SetNumber(num + 1)

	require.True(t, id1.Equals(&id2))
	require.False(t, id1.Equals(&idOther))
	require.False(t, id2.Equals(&idOther))
}

func TestSubnetIDEncoding(t *testing.T) {
	id := subnetidtest.ID()

	t.Run("binary", func(t *testing.T) {
		data, err := id.Marshal()
		require.NoError(t, err)

		var id2 subnetid.ID
		require.NoError(t, id2.Unmarshal(data))

		require.True(t, id2.Equals(id))
	})

	t.Run("text", func(t *testing.T) {
		data, err := id.MarshalText()
		require.NoError(t, err)

		var id2 subnetid.ID
		require.NoError(t, id2.UnmarshalText(data))

		require.True(t, id2.Equals(id))
	})
}

func TestMakeZero(t *testing.T) {
	var id subnetid.ID

	id.SetNumber(13)

	require.False(t, subnetid.IsZero(id))

	subnetid.MakeZero(&id)

	require.True(t, subnetid.IsZero(id))
}
