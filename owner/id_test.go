package owner_test

import (
	"crypto/ecdsa"
	"testing"

	"github.com/mr-tron/base58"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	. "github.com/nspcc-dev/neofs-sdk-go/owner"
	ownertest "github.com/nspcc-dev/neofs-sdk-go/owner/test"
	"github.com/stretchr/testify/require"
)

func TestIDV2(t *testing.T) {
	id := ownertest.GenerateID()

	idV2 := id.ToV2()

	require.Equal(t, id, NewIDFromV2(idV2))
}

func TestNewIDFromNeo3Wallet(t *testing.T) {
	p, err := keys.NewPrivateKey()
	require.NoError(t, err)

	wallet, err := NEO3WalletFromPublicKey((*ecdsa.PublicKey)(p.PublicKey()))
	require.NoError(t, err)

	id := NewIDFromNeo3Wallet(wallet)
	require.Equal(t, id.ToV2().GetValue(), wallet.Bytes())
}

func TestID_Parse(t *testing.T) {
	t.Run("should parse successful", func(t *testing.T) {
		p, err := keys.NewPrivateKey()
		require.NoError(t, err)

		wallet, err := NEO3WalletFromPublicKey((*ecdsa.PublicKey)(p.PublicKey()))
		require.NoError(t, err)

		eid := NewIDFromNeo3Wallet(wallet)
		aid := NewID()

		require.NoError(t, aid.Parse(eid.String()))
		require.Equal(t, eid, aid)
	})

	t.Run("should failure on parse", func(t *testing.T) {
		cs := []byte{1, 2, 3, 4, 5, 6}
		str := base58.Encode(cs)
		cid := NewID()

		require.Error(t, cid.Parse(str))
	})
}

func TestIDEncoding(t *testing.T) {
	id := ownertest.GenerateID()

	t.Run("binary", func(t *testing.T) {
		data, err := id.Marshal()
		require.NoError(t, err)

		id2 := NewID()
		require.NoError(t, id2.Unmarshal(data))

		require.Equal(t, id, id2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := id.MarshalJSON()
		require.NoError(t, err)

		a2 := NewID()
		require.NoError(t, a2.UnmarshalJSON(data))

		require.Equal(t, id, a2)
	})
}

func TestID_Equal(t *testing.T) {
	var (
		data1 = []byte{1, 2, 3}
		data2 = data1
		data3 = append(data1, 255)
	)

	id1 := ownertest.GenerateIDFromBytes(data1)

	require.True(t, id1.Equal(
		ownertest.GenerateIDFromBytes(data2),
	))

	require.False(t, id1.Equal(
		ownertest.GenerateIDFromBytes(data3),
	))
}

func TestNewIDFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *refs.OwnerID

		require.Nil(t, NewIDFromV2(x))
	})
}

func TestID_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *ID

		require.Nil(t, x.ToV2())
	})
}

func TestID_String(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		id := NewID()

		require.Empty(t, id.String())
	})
}

func TestNewID(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		id := NewID()

		// convert to v2 message
		idV2 := id.ToV2()

		require.Nil(t, idV2.GetValue())
	})
}
