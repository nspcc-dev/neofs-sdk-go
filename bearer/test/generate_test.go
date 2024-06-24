package bearertest_test

import (
	"testing"

	"github.com/nspcc-dev/neofs-sdk-go/api/acl"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	bearertest "github.com/nspcc-dev/neofs-sdk-go/bearer/test"
	"github.com/stretchr/testify/require"
)

func TestUnsignedToken(t *testing.T) {
	tok := bearertest.UnsignedToken()
	require.False(t, tok.VerifySignature())
	require.NotEqual(t, tok, bearertest.UnsignedToken())

	var tok2 bearer.Token
	require.NoError(t, tok2.Unmarshal(tok.Marshal()))
	require.Equal(t, tok, tok2)

	var m acl.BearerToken
	tok.WriteToV2(&m)
	var tok3 bearer.Token
	require.NoError(t, tok3.ReadFromV2(&m))
	require.Equal(t, tok, tok3)
}

func TestToken(t *testing.T) {
	tok := bearertest.Token()
	require.True(t, tok.VerifySignature())
	require.NotEqual(t, tok, bearertest.Token())

	var tok2 bearer.Token
	require.NoError(t, tok2.Unmarshal(tok.Marshal()))
	require.Equal(t, tok, tok2)

	var m acl.BearerToken
	tok.WriteToV2(&m)
	var tok3 bearer.Token
	require.NoError(t, tok3.ReadFromV2(&m))
	require.Equal(t, tok, tok3)
}
