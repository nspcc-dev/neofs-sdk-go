package user_test

import (
	"crypto/elliptic"
	"encoding/hex"
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/stretchr/testify/require"
)

func TestIDFromKey(t *testing.T) {
	// examples are taken from https://docs.neo.org/docs/en-us/basic/concept/wallets.html
	rawPub, _ := hex.DecodeString("03cdb067d930fd5adaa6c68545016044aaddec64ba39e548250eaea551172e535c")
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), rawPub)
	require.True(t, x != nil && y != nil)

	var id user.ID

	require.NoError(t, user.IDFromKey(&id, rawPub))

	require.Equal(t, "NNLi44dJNXtDNSBkofB48aTVYtb1zZrNEs", id.EncodeToString())
}
