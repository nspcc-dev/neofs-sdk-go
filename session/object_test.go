package session_test

import (
	"testing"

	v2session "github.com/nspcc-dev/neofs-api-go/v2/session"
	objecttest "github.com/nspcc-dev/neofs-sdk-go/object/address/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	sessiontest "github.com/nspcc-dev/neofs-sdk-go/session/test"
	"github.com/stretchr/testify/require"
)

func TestObjectContextVerbs(t *testing.T) {
	c := session.NewObjectContext()

	assert := func(setter func(), getter func() bool, verb v2session.ObjectSessionVerb) {
		setter()

		require.True(t, getter())

		require.Equal(t, verb, c.ToV2().GetVerb())
	}

	t.Run("PUT", func(t *testing.T) {
		assert(c.ForPut, c.IsForPut, v2session.ObjectVerbPut)
	})

	t.Run("DELETE", func(t *testing.T) {
		assert(c.ForDelete, c.IsForDelete, v2session.ObjectVerbDelete)
	})

	t.Run("GET", func(t *testing.T) {
		assert(c.ForGet, c.IsForGet, v2session.ObjectVerbGet)
	})

	t.Run("SEARCH", func(t *testing.T) {
		assert(c.ForSearch, c.IsForSearch, v2session.ObjectVerbSearch)
	})

	t.Run("RANGE", func(t *testing.T) {
		assert(c.ForRange, c.IsForRange, v2session.ObjectVerbRange)
	})

	t.Run("RANGEHASH", func(t *testing.T) {
		assert(c.ForRangeHash, c.IsForRangeHash, v2session.ObjectVerbRangeHash)
	})

	t.Run("HEAD", func(t *testing.T) {
		assert(c.ForHead, c.IsForHead, v2session.ObjectVerbHead)
	})
}

func TestObjectContext_ApplyTo(t *testing.T) {
	c := session.NewObjectContext()
	id := objecttest.Address()

	t.Run("method", func(t *testing.T) {
		c.ApplyTo(id)

		require.Equal(t, id, c.Address())

		c.ApplyTo(nil)

		require.Nil(t, c.Address())
	})
}

func TestObjectFilter_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *session.ObjectContext

		require.Nil(t, x.ToV2())
	})

	t.Run("default values", func(t *testing.T) {
		c := session.NewObjectContext()

		// check initial values
		require.Nil(t, c.Address())

		for _, op := range []func() bool{
			c.IsForPut,
			c.IsForDelete,
			c.IsForGet,
			c.IsForHead,
			c.IsForRange,
			c.IsForRangeHash,
			c.IsForSearch,
		} {
			require.False(t, op())
		}

		// convert to v2 message
		cV2 := c.ToV2()

		require.Equal(t, v2session.ObjectVerbUnknown, cV2.GetVerb())
		require.Nil(t, cV2.GetAddress())
	})
}

func TestObjectContextEncoding(t *testing.T) {
	c := sessiontest.ObjectContext()

	t.Run("binary", func(t *testing.T) {
		data, err := c.Marshal()
		require.NoError(t, err)

		c2 := session.NewObjectContext()
		require.NoError(t, c2.Unmarshal(data))

		require.Equal(t, c, c2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := c.MarshalJSON()
		require.NoError(t, err)

		c2 := session.NewObjectContext()
		require.NoError(t, c2.UnmarshalJSON(data))

		require.Equal(t, c, c2)
	})
}
