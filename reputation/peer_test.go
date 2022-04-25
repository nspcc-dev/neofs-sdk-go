package reputation_test

import (
	"testing"

	reputationV2 "github.com/nspcc-dev/neofs-api-go/v2/reputation"
	"github.com/nspcc-dev/neofs-sdk-go/reputation"
	reputationtest "github.com/nspcc-dev/neofs-sdk-go/reputation/test"
	"github.com/stretchr/testify/require"
)

func TestPeerID_ToV2(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var x *reputation.PeerID

		require.Nil(t, x.ToV2())
	})

	t.Run("nil", func(t *testing.T) {
		peerID := reputationtest.PeerID()

		require.Equal(t, peerID, reputation.PeerIDFromV2(peerID.ToV2()))
	})
}

func TestPeerID_String(t *testing.T) {
	t.Run("Parse/String", func(t *testing.T) {
		id := reputationtest.PeerID()

		strID := id.String()

		id2 := reputation.NewPeerID()

		err := id2.Parse(strID)
		require.NoError(t, err)

		require.Equal(t, id, id2)
	})

	t.Run("nil", func(t *testing.T) {
		id := reputation.NewPeerID()

		require.Empty(t, id.String())
	})
}

func TestPeerIDEncoding(t *testing.T) {
	id := reputationtest.PeerID()

	t.Run("binary", func(t *testing.T) {
		data := id.Marshal()

		id2 := reputation.NewPeerID()
		require.NoError(t, id2.Unmarshal(data))

		require.Equal(t, id, id2)
	})

	t.Run("json", func(t *testing.T) {
		data, err := id.MarshalJSON()
		require.NoError(t, err)

		id2 := reputation.NewPeerID()
		require.NoError(t, id2.UnmarshalJSON(data))

		require.Equal(t, id, id2)
	})
}

func TestPeerIDFromV2(t *testing.T) {
	t.Run("from nil", func(t *testing.T) {
		var x *reputationV2.PeerID

		require.Nil(t, reputation.PeerIDFromV2(x))
	})
}

func TestNewPeerID(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		id := reputation.NewPeerID()

		// convert to v2 message
		idV2 := id.ToV2()

		require.Nil(t, idV2.GetPublicKey())
	})
}
