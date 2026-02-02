package session

import (
	"bytes"
	"testing"
	"time"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestV2_CopyTo(t *testing.T) {
	sig := neofscrypto.NewSignatureFromRawKey(neofscrypto.ECDSA_SHA512, []byte("key"), []byte("sign"))

	usr := usertest.User()

	data := Token{
		Lifetime: Lifetime{
			iat: time.Unix(1, 0),
			nbf: time.Unix(2, 0),
			exp: time.Unix(3, 0),
		},
		version: 2,
		appdata: []byte("appdata"),
		issuer:  usr.UserID(),
		subjects: []Target{
			{nnsName: "alice.neofs"},
		},
		contexts: []Context{
			{
				container: cidtest.ID(),
				verbs:     []Verb{VerbObjectGet, VerbObjectPut},
			},
		},
		final: false,
		origin: &Token{
			Lifetime: Lifetime{
				iat: time.Unix(1, 0),
				nbf: time.Unix(2, 0),
				exp: time.Unix(3, 0),
			},
			issuer: usr.UserID(),
			subjects: []Target{
				{nnsName: "bob.neofs"},
			},
			sigSet: true,
			sig:    sig,
		},
		sigSet: true,
		sig:    sig,
	}

	t.Run("copy", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data, dst)
		require.Equal(t, data.Marshal(), dst.Marshal())
		require.Equal(t, data.issuer, dst.issuer)
	})

	t.Run("change version", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.version, dst.version)

		dst.version = 100

		require.NotEqual(t, data.version, dst.version)
	})

	t.Run("overwrite version", func(t *testing.T) {
		var local Token
		require.Zero(t, local.version)

		var dst Token
		dst.version = 100
		require.NotZero(t, dst.version)

		local.CopyTo(&dst)
		require.Zero(t, local.version)
		require.Zero(t, dst.version)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.version = 100
		require.Zero(t, local.version)
		require.NotZero(t, dst.version)

		require.NotEqual(t, local.version, dst.version)
	})

	t.Run("change app data", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.appdata, dst.appdata)

		require.NoError(t, dst.SetAppData([]byte{9, 8, 7}))

		require.NotEqual(t, data.appdata, dst.appdata)
	})

	t.Run("overwrite nonce", func(t *testing.T) {
		local := Token{}
		require.Zero(t, local.appdata)

		var dst Token
		require.NoError(t, dst.SetAppData([]byte{7}))
		require.NotZero(t, dst.appdata)

		local.CopyTo(&dst)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.Zero(t, local.appdata)
		require.Zero(t, dst.appdata)

		require.NoError(t, dst.SetAppData([]byte{8}))

		require.Zero(t, local.appdata)
		require.NotZero(t, dst.appdata)
	})

	t.Run("change final", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.final, dst.final)

		dst.SetFinal(true)

		require.NotEqual(t, data.final, dst.final)
		require.True(t, dst.final)
	})

	t.Run("overwrite final", func(t *testing.T) {
		local := Token{}
		require.False(t, local.final)

		var dst Token
		dst.SetFinal(true)
		require.True(t, dst.final)

		local.CopyTo(&dst)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.False(t, local.final)
		require.False(t, dst.final)

		dst.SetFinal(true)

		require.False(t, local.final)
		require.True(t, dst.final)
	})

	t.Run("change issuer", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.issuer, dst.issuer)

		dst.SetIssuer(usertest.OtherID(usr.ID))

		require.NotEqual(t, data.issuer, dst.issuer)
	})

	t.Run("overwrite issuer", func(t *testing.T) {
		var local Token
		require.Zero(t, local.issuer)

		var dst Token
		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.NotZero(t, dst.issuer)

		local.CopyTo(&dst)
		require.Zero(t, local.issuer)
		require.Zero(t, dst.issuer)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.Equal(t, local.issuer, dst.issuer)

		dst.SetIssuer(usertest.OtherID(usr.ID))
		require.Zero(t, local.issuer)
		require.NotZero(t, dst.issuer)

		require.NotEqual(t, local.issuer, dst.issuer)
	})

	t.Run("change subjects", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.subjects, dst.subjects)

		// Modify the copy's subjects
		dst.subjects = append(dst.subjects, Target{nnsName: "charlie.neofs"})

		require.NotEqual(t, len(data.subjects), len(dst.subjects))
	})

	t.Run("overwrite subjects", func(t *testing.T) {
		var local Token
		require.Nil(t, local.subjects)

		var dst Token
		dst.subjects = []Target{{nnsName: "test.neofs"}}
		require.NotNil(t, dst.subjects)

		local.CopyTo(&dst)
		require.Nil(t, local.subjects)
		require.Nil(t, dst.subjects)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.subjects = []Target{{nnsName: "test.neofs"}}
		require.Nil(t, local.subjects)
		require.NotNil(t, dst.subjects)
	})

	t.Run("change lifetime", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.iat, dst.iat)
		require.Equal(t, data.nbf, dst.nbf)
		require.Equal(t, data.exp, dst.exp)

		dst.SetExp(time.Unix(100, 0))
		dst.SetIat(time.Unix(200, 0))
		dst.SetNbf(time.Unix(300, 0))

		require.NotEqual(t, data.iat, dst.iat)
		require.NotEqual(t, data.nbf, dst.nbf)
		require.NotEqual(t, data.exp, dst.exp)
	})

	t.Run("overwrite lifetime", func(t *testing.T) {
		local := Token{}

		var dst Token
		dst.SetExp(time.Unix(100, 0))
		dst.SetIat(time.Unix(200, 0))
		dst.SetNbf(time.Unix(300, 0))

		local.CopyTo(&dst)

		require.Equal(t, local.Marshal(), dst.Marshal())

		require.Equal(t, local.iat, dst.iat)
		require.Equal(t, local.nbf, dst.nbf)
		require.Equal(t, local.exp, dst.exp)

		dst.SetExp(time.Unix(100, 0))
		dst.SetIat(time.Unix(200, 0))
		dst.SetNbf(time.Unix(300, 0))

		require.NotEqual(t, local.iat, dst.iat)
		require.NotEqual(t, local.nbf, dst.nbf)
		require.NotEqual(t, local.exp, dst.exp)
	})

	t.Run("change contexts", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, len(data.contexts), len(dst.contexts))

		dst.contexts = append(dst.contexts, Context{
			container: cidtest.ID(),
			verbs:     []Verb{VerbObjectDelete},
		})

		require.NotEqual(t, len(data.contexts), len(dst.contexts))
	})

	t.Run("overwrite contexts", func(t *testing.T) {
		var local Token
		require.Nil(t, local.contexts)

		var dst Token
		dst.contexts = []Context{{container: cidtest.ID()}}
		require.NotNil(t, dst.contexts)

		local.CopyTo(&dst)
		require.Nil(t, local.contexts)
		require.Nil(t, dst.contexts)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.contexts = []Context{{container: cidtest.ID()}}
		require.Nil(t, local.contexts)
		require.NotNil(t, dst.contexts)
	})

	t.Run("change origin", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		if data.origin != nil {
			require.NotNil(t, dst.origin)
		} else {
			require.Nil(t, dst.origin)
		}

		newOrigin := &Token{
			issuer:   usertest.OtherID(usr.ID),
			subjects: []Target{{nnsName: "charlie.neofs"}},
		}
		dst.origin = newOrigin

		if data.origin != nil {
			require.NotEqual(t, data.origin.issuer, dst.origin.issuer)
		}
	})

	t.Run("overwrite origin", func(t *testing.T) {
		var local Token
		require.Nil(t, local.origin)

		var dst Token
		dst.origin = &Token{
			issuer:   usr.UserID(),
			subjects: []Target{{nnsName: "test.neofs"}},
		}
		require.NotNil(t, dst.origin)

		local.CopyTo(&dst)
		require.Nil(t, local.origin)
		require.Nil(t, dst.origin)

		require.Equal(t, local.Marshal(), dst.Marshal())

		dst.origin = &Token{
			issuer:   usr.UserID(),
			subjects: []Target{{nnsName: "test.neofs"}},
		}
		require.Nil(t, local.origin)
		require.NotNil(t, dst.origin)
	})

	t.Run("change sig", func(t *testing.T) {
		var dst Token
		data.CopyTo(&dst)

		require.Equal(t, data.sig, dst.sig)

		dst.sig.SetPublicKeyBytes([]byte{1, 2, 3})
		dst.sig.SetScheme(100)
		dst.sig.SetValue([]byte{10, 11, 12})

		require.Equal(t, data.issuer, dst.issuer)
		require.NotEqual(t, data.sig.Scheme(), dst.sig.Scheme())
		require.False(t, bytes.Equal(data.sig.PublicKeyBytes(), dst.sig.PublicKeyBytes()))
		require.False(t, bytes.Equal(data.sig.Value(), dst.sig.Value()))
	})

	t.Run("overwrite sig", func(t *testing.T) {
		var local Token
		require.False(t, local.sigSet)

		var dst Token
		require.NoError(t, dst.Sign(usr))
		require.True(t, dst.sigSet)

		local.CopyTo(&dst)
		require.False(t, local.sigSet)
		require.False(t, dst.sigSet)
		require.Empty(t, local.sig)
		require.Empty(t, dst.sig)

		require.True(t, bytes.Equal(local.Marshal(), dst.Marshal()))

		require.NoError(t, dst.Sign(usr))
		require.False(t, local.sigSet)
		require.True(t, dst.sigSet)
	})
}

func TestTarget_protoMessage(t *testing.T) {
	t.Run("zero target returns nil", func(t *testing.T) {
		var target Target
		require.Nil(t, target.protoMessage())
	})

	t.Run("ownerID target", func(t *testing.T) {
		userID := usertest.ID()
		target := NewTargetUser(userID)
		msg := target.protoMessage()
		require.NotNil(t, msg)
		require.NotNil(t, msg.GetOwnerId())
	})

	t.Run("NNS target", func(t *testing.T) {
		nnsName := "test.neo"
		target := NewTargetNamed(nnsName)
		msg := target.protoMessage()
		require.NotNil(t, msg)
		require.Equal(t, nnsName, msg.GetNnsName())
	})
}

func Test_findUnauthorizedVerb(t *testing.T) {
	for _, tt := range []struct {
		name      string
		required  []Verb
		available []Verb
		wantErr   bool
		verb      Verb
	}{
		{
			name:      "all verbs authorized",
			required:  []Verb{VerbObjectPut, VerbObjectHead, VerbObjectDelete},
			available: []Verb{VerbObjectPut, VerbObjectGet, VerbObjectHead, VerbObjectDelete},
			wantErr:   false,
		},
		{
			name:      "first verb not authorized",
			required:  []Verb{VerbObjectPut, VerbObjectDelete},
			available: []Verb{VerbObjectGet, VerbObjectDelete},
			wantErr:   true,
			verb:      VerbObjectPut,
		},
		{
			name:      "middle verb not authorized",
			required:  []Verb{VerbObjectPut, VerbObjectGet, VerbObjectDelete},
			available: []Verb{VerbObjectPut, VerbObjectHead, VerbObjectDelete},
			wantErr:   true,
			verb:      VerbObjectGet,
		},
		{
			name:      "last verb not authorized",
			required:  []Verb{VerbObjectGet, VerbObjectHead},
			available: []Verb{VerbObjectGet, VerbObjectDelete},
			wantErr:   true,
			verb:      VerbObjectHead,
		},
		{
			name:      "empty required",
			required:  []Verb{},
			available: []Verb{VerbObjectPut, VerbObjectGet},
			wantErr:   false,
		},
		{
			name:      "empty available",
			required:  []Verb{VerbObjectGet},
			available: []Verb{},
			wantErr:   true,
			verb:      VerbObjectGet,
		},
		{
			name:      "single verb authorized",
			required:  []Verb{VerbObjectGet},
			available: []Verb{VerbObjectPut, VerbObjectGet, VerbObjectDelete},
			wantErr:   false,
		},
		{
			name:      "single verb not authorized",
			required:  []Verb{VerbObjectSearch},
			available: []Verb{VerbObjectPut, VerbObjectGet},
			wantErr:   true,
			verb:      VerbObjectSearch,
		},
		{
			name:      "returns first unauthorized verb",
			required:  []Verb{VerbObjectGet, VerbObjectSearch, VerbObjectDelete},
			available: []Verb{VerbObjectHead, VerbObjectPut},
			wantErr:   true,
			verb:      VerbObjectGet,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := findUnauthorizedVerb(tt.required, tt.available)
			if tt.wantErr {
				require.NotNil(t, result)
				require.Equal(t, tt.verb, *result)
			} else {
				if result != nil {
					t.Fatalf("expected nil, got %v", *result)
				}
			}
		})
	}
}
