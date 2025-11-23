package session_test

import (
	"testing"

	"github.com/google/uuid"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

func TestVerbV2(t *testing.T) {
	tests := []struct {
		name   string
		verb   session.VerbV2
		isObjV bool
		isCnrV bool
	}{
		{"Unspecified", session.VerbV2Unspecified, false, false},
		{"ObjectPut", session.VerbV2ObjectPut, true, false},
		{"ObjectGet", session.VerbV2ObjectGet, true, false},
		{"ObjectHead", session.VerbV2ObjectHead, true, false},
		{"ObjectSearch", session.VerbV2ObjectSearch, true, false},
		{"ObjectDelete", session.VerbV2ObjectDelete, true, false},
		{"ObjectRange", session.VerbV2ObjectRange, true, false},
		{"ObjectRangeHash", session.VerbV2ObjectRangeHash, true, false},
		{"ContainerPut", session.VerbV2ContainerPut, false, true},
		{"ContainerDelete", session.VerbV2ContainerDelete, false, true},
		{"ContainerSetEACL", session.VerbV2ContainerSetEACL, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isObjV, tt.verb.IsObjectVerb())
			require.Equal(t, tt.isCnrV, tt.verb.IsContainerVerb())
		})
	}
}

func TestTarget(t *testing.T) {
	t.Run("zero target", func(t *testing.T) {
		var target session.Target

		require.False(t, target.IsOwnerID())
		require.False(t, target.IsNNS())
		require.True(t, target.OwnerID().IsZero())
		require.Empty(t, target.NNSName())
	})

	t.Run("NewTarget", func(t *testing.T) {
		userID := usertest.ID()
		target := session.NewTarget(userID)

		require.True(t, target.IsOwnerID())
		require.False(t, target.IsNNS())
		require.Equal(t, userID, target.OwnerID())
		require.Empty(t, target.NNSName())
	})

	t.Run("NewTargetFromNNS", func(t *testing.T) {
		nnsName := "test.neo"
		target := session.NewTargetFromNNS(nnsName)

		require.False(t, target.IsOwnerID())
		require.True(t, target.IsNNS())
		require.Equal(t, nnsName, target.NNSName())
		require.True(t, target.OwnerID().IsZero())
	})

	t.Run("Equals", func(t *testing.T) {
		userID := usertest.ID()
		target1 := session.NewTarget(userID)
		target2 := session.NewTarget(userID)
		target3 := session.NewTarget(usertest.ID())

		require.True(t, target1.Equals(target2))
		require.False(t, target1.Equals(target3))

		nnsTarget1 := session.NewTargetFromNNS("test.neo")
		nnsTarget2 := session.NewTargetFromNNS("test.neo")
		nnsTarget3 := session.NewTargetFromNNS("other.neo")

		require.True(t, nnsTarget1.Equals(nnsTarget2))
		require.False(t, nnsTarget1.Equals(nnsTarget3))
		require.False(t, target1.Equals(nnsTarget1))

		var zeroTarget session.Target
		require.True(t, zeroTarget.Equals(session.Target{}))
		require.False(t, target1.Equals(zeroTarget))
	})

	t.Run("empty NNS", func(t *testing.T) {
		var target session.Target

		emptyNNSTarget := session.NewTargetFromNNS("")
		require.False(t, emptyNNSTarget.IsOwnerID())
		require.False(t, emptyNNSTarget.IsNNS()) // empty string means IsNNS returns false
		require.Empty(t, emptyNNSTarget.NNSName())

		// Zero target has expected behavior
		require.False(t, target.IsOwnerID())
		require.False(t, target.IsNNS())
	})
}

func TestDelegationInfo(t *testing.T) {
	subject := session.NewTarget(usertest.ID())
	iat := uint64(100)
	nbf := uint64(200)
	exp := uint64(300)
	lifetime := session.NewLifetime(iat, nbf, exp)
	verbs := []session.VerbV2{session.VerbV2ObjectGet, session.VerbV2ObjectPut}

	t.Run("NewDelegationInfo", func(t *testing.T) {
		del := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)

		require.True(t, del.Issuer().IsEmpty()) // issuer not set yet
		require.Equal(t, []session.Target{subject}, del.Subjects())
		require.Equal(t, lifetime, del.Lifetime)
		require.Equal(t, verbs, del.Verbs())
	})

	t.Run("Sign and Verify", func(t *testing.T) {
		del := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)
		signer := usertest.User()

		err := del.Sign(signer)
		require.NoError(t, err)

		// Verify issuer was set from signer
		require.True(t, del.Issuer().IsOwnerID())
		require.Equal(t, signer.UserID(), del.Issuer().OwnerID())

		require.True(t, del.VerifySignature())
	})

	t.Run("VerifySignature without signing", func(t *testing.T) {
		del := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)
		require.False(t, del.VerifySignature())
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("empty verbs", func(t *testing.T) {
			delEmptyVerbs := session.NewDelegationInfo([]session.Target{subject}, lifetime, []session.VerbV2{})
			require.Empty(t, delEmptyVerbs.Verbs())
		})

		t.Run("nil verbs", func(t *testing.T) {
			delNilVerbs := session.NewDelegationInfo([]session.Target{subject}, lifetime, nil)
			require.Empty(t, delNilVerbs.Verbs())
		})

		// Test signing and then verify after modification
		t.Run("signing and then verify after modification", func(t *testing.T) {
			del := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)
			signer := usertest.User()
			err := del.Sign(signer)
			require.NoError(t, err)
			require.True(t, del.VerifySignature())

			del2 := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)
			signer2 := usertest.User()
			err2 := del2.Sign(signer2)
			require.NoError(t, err2)

			require.True(t, del.VerifySignature())
			require.True(t, del2.VerifySignature())
		})

		t.Run("sign with zero ID signer", func(t *testing.T) {
			del := session.NewDelegationInfo([]session.Target{subject}, lifetime, verbs)
			zeroIDUser := user.NewSigner(usertest.User(), user.ID{})
			err := del.Sign(zeroIDUser)
			require.ErrorIs(t, err, user.ErrZeroID)
		})
	})
}

func newValidTokenV2() session.TokenV2 {
	var tok session.TokenV2
	tok.SetVersion(session.TokenV2CurrentVersion)
	tok.SetID(uuid.New())
	tok.SetIssuer(session.NewTarget(usertest.ID()))
	tok.AddSubject(session.NewTarget(usertest.ID()))
	tok.SetIat(100)
	tok.SetNbf(200)
	tok.SetExp(300)
	tok.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))
	return tok
}

func newValidSignedTokenV2(t *testing.T) session.TokenV2 {
	tok := newValidTokenV2()
	signer := usertest.User()
	err := tok.Sign(signer)
	require.NoError(t, err)
	return tok
}

// newTokenV2ForDelegation creates a token suitable for delegation chain testing.
func newTokenV2ForDelegation(issuer, subject session.Target) session.TokenV2 {
	var tok session.TokenV2
	tok.SetID(uuid.New())
	tok.SetIssuer(issuer)
	tok.AddSubject(subject)
	tok.SetIat(100)
	tok.SetNbf(200)
	tok.SetExp(300)
	tok.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))
	return tok
}

// newSignedDelegation creates a signed delegation for testing.
func newSignedDelegation(t *testing.T, subjects []session.Target, iat, nbf, exp uint64, verbs []session.VerbV2) session.DelegationInfo {
	lifetime := session.NewLifetime(iat, nbf, exp)
	del := session.NewDelegationInfo(subjects, lifetime, verbs)
	signer := usertest.User()
	err := del.Sign(signer)
	require.NoError(t, err)
	return del
}

func TestTokenV2_Validate(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  string
		fn   func(tok *session.TokenV2)
	}{
		{"wrong version", "invalid token version: expected 1, got 2", func(tok *session.TokenV2) {
			tok.SetVersion(2)
		}},
		{"missing token ID", "token ID is not set", func(tok *session.TokenV2) {
			tok.SetID(uuid.UUID{})
		}},
		{"missing issuer", "issuer is not set", func(tok *session.TokenV2) {
			tok.SetIssuer(session.Target{})
		}},
		{"no subjects", "no subjects specified", func(tok *session.TokenV2) {
			tok.SetSubjects([]session.Target{})
		}},
		{"empty subject", "subject at index 1 is empty", func(tok *session.TokenV2) {
			tok.AddSubject(session.Target{})
		}},
		{"missing iat", "issued at (iat) is not set", func(tok *session.TokenV2) {
			tok.SetIat(0)
		}},
		{"missing nbf", "not before (nbf) is not set", func(tok *session.TokenV2) {
			tok.SetNbf(0)
		}},
		{"missing exp", "expiration (exp) is not set", func(tok *session.TokenV2) {
			tok.SetExp(0)
		}},
		{"nbf after exp", "not before (nbf) is after expiration (exp)", func(tok *session.TokenV2) {
			tok.SetNbf(300)
			tok.SetExp(200)
		}},
		{"iat after exp", "issued at (iat) is after expiration (exp)", func(tok *session.TokenV2) {
			tok.SetIat(400)
		}},
		{"no contexts", "no contexts specified", func(tok *session.TokenV2) {
			tok.SetContexts([]session.ContextV2{})
		}},
		{"context with no verbs", "context at index 0 has no verbs", func(tok *session.TokenV2) {
			tok.SetContexts([]session.ContextV2{
				session.NewContextV2(cidtest.ID(), []session.VerbV2{}),
			})
		}},
		{"invalid delegation", "invalid delegation chain: delegation chain doesn't start from token issuer", func(tok *session.TokenV2) {
			del := session.NewDelegationInfo([]session.Target{session.NewTarget(usertest.ID())}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
			tok.AddDelegation(del)
		}},
		{"unsigned token", "token is not signed", func(tok *session.TokenV2) {
			// Don't sign the token
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok := newValidTokenV2()
			tc.fn(&tok)

			err := tok.Validate()
			require.EqualError(t, err, tc.err)
		})
	}

	t.Run("valid token", func(t *testing.T) {
		tok := newValidSignedTokenV2(t)
		err := tok.Validate()
		require.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		tok := newValidSignedTokenV2(t)
		tok.SetExp(400)

		err := tok.Validate()
		require.EqualError(t, err, "token signature verification failed")
	})
}

func TestTokenV2_ValidateDelegationChain(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		tok := newValidTokenV2()

		err := tok.ValidateDelegationChain()
		require.NoError(t, err)
	})

	t.Run("valid single delegation", func(t *testing.T) {
		subject := session.NewTarget(usertest.ID())

		del := newSignedDelegation(t, []session.Target{subject}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectGet})

		tok := newTokenV2ForDelegation(del.Issuer(), subject)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.NoError(t, err)
	})

	t.Run("delegation chain doesn't start from issuer", func(t *testing.T) {
		issuer := session.NewTarget(usertest.ID())
		subject := session.NewTarget(usertest.ID())

		var tok session.TokenV2
		tok.SetID(uuid.New())
		tok.SetIssuer(issuer)
		tok.AddSubject(subject)
		tok.SetIat(100)
		tok.SetNbf(200)
		tok.SetExp(300)
		tok.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))

		// Create delegation with different issuer
		del := session.NewDelegationInfo([]session.Target{subject}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
		signer := usertest.User()
		err := del.Sign(signer)
		require.NoError(t, err)

		tok.AddDelegation(del)

		err = tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation chain doesn't start from token issuer")
	})

	t.Run("delegation has no subject", func(t *testing.T) {
		del := newSignedDelegation(t, []session.Target{}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectGet})
		tok := newTokenV2ForDelegation(del.Issuer(), session.NewTarget(usertest.ID()))
		tok.AddDelegation(del)
		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 has no subjects")
	})

	t.Run("delegation has empty subject", func(t *testing.T) {
		del := newSignedDelegation(t, []session.Target{session.NewTargetFromNNS("")}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectGet})
		tok := newTokenV2ForDelegation(del.Issuer(), session.NewTarget(usertest.ID()))
		tok.AddDelegation(del)
		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 has empty subject at index 0")
	})

	t.Run("unsigned delegation", func(t *testing.T) {
		subject := session.NewTarget(usertest.ID())

		// Create unsigned delegation
		del := session.NewDelegationInfo([]session.Target{subject}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
		issuer := session.NewTarget(usertest.ID())
		del.SetIssuer(issuer)

		tok := newTokenV2ForDelegation(issuer, subject)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 is not signed")
	})

	t.Run("delegation chain broken", func(t *testing.T) {
		intermediate := session.NewTarget(usertest.ID())
		finalSubject := session.NewTarget(usertest.ID())

		// First delegation
		del1 := newSignedDelegation(t, []session.Target{intermediate}, 220, 220, 220, []session.VerbV2{session.VerbV2ObjectGet})

		// Second delegation with wrong issuer (doesn't match previous subject)
		del2 := newSignedDelegation(t, []session.Target{finalSubject}, 240, 240, 240, []session.VerbV2{session.VerbV2ObjectGet})

		tok := newTokenV2ForDelegation(del1.Issuer(), finalSubject)
		tok.AddDelegation(del1)
		tok.AddDelegation(del2)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation chain broken at index 1: issuer doesn't match any previous subject")
	})

	t.Run("delegation timestamp outside token lifetime", func(t *testing.T) {
		subject := session.NewTarget(usertest.ID())

		// Delegation with timestamp after token expiration (300)
		del := newSignedDelegation(t, []session.Target{subject}, 400, 400, 400, []session.VerbV2{session.VerbV2ObjectGet})

		tok := newTokenV2ForDelegation(del.Issuer(), subject)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 lifetime is outside token lifetime")
	})

	t.Run("delegation non chronological timestamps", func(t *testing.T) {
		intermediate := usertest.User()
		subject := session.NewTarget(intermediate.UserID())
		del := newSignedDelegation(t, []session.Target{subject}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectGet})
		tok := newTokenV2ForDelegation(del.Issuer(), subject)
		tok.AddDelegation(del)

		subject2 := session.NewTarget(usertest.ID())
		del2 := session.NewDelegationInfo([]session.Target{subject2}, session.NewLifetime(220, 220, 220), []session.VerbV2{session.VerbV2ObjectGet})
		require.NoError(t, del2.Sign(intermediate))
		tok.AddDelegation(del2)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 1 lifetime extends beyond previous delegation lifetime")
	})

	t.Run("delegation with unauthorized verb", func(t *testing.T) {
		subject := session.NewTarget(usertest.ID())

		// Delegation tries to grant ObjectPut which wasn't available
		del := newSignedDelegation(t, []session.Target{subject}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectPut})

		// Token only allows ObjectGet
		tok := newTokenV2ForDelegation(del.Issuer(), subject)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 tries to delegate verb 1 which was not available")
	})

	t.Run("delegation with invalid signature", func(t *testing.T) {
		subject := session.NewTarget(usertest.ID())
		del := session.NewDelegationInfo([]session.Target{subject}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
		del.AttachSignature(neofscryptotest.Signature())
		issuer := session.NewTarget(usertest.ID())
		del.SetIssuer(issuer)

		tok := newTokenV2ForDelegation(issuer, subject)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.EqualError(t, err, "delegation 0 has invalid signature")
	})

	t.Run("first subject not in token subjects", func(t *testing.T) {
		subject1 := session.NewTarget(usertest.ID())
		subject2 := session.NewTarget(usertest.ID())

		// Delegation to subject2 which is not in token subjects
		del := newSignedDelegation(t, []session.Target{subject2}, 250, 250, 250, []session.VerbV2{session.VerbV2ObjectGet})

		tok := newTokenV2ForDelegation(del.Issuer(), subject1)
		tok.AddDelegation(del)

		err := tok.ValidateDelegationChain()
		require.Error(t, err)
		require.Contains(t, err.Error(), "first delegation subject")
		require.Contains(t, err.Error(), "is not in token subjects")
	})

	t.Run("valid multi-level delegation chain", func(t *testing.T) {
		issuer := usertest.User()
		signer1 := usertest.User()
		signer2 := usertest.User()
		signer3 := usertest.User()
		intermediate1 := session.NewTarget(signer1.UserID())
		intermediate2 := session.NewTarget(signer2.UserID())
		finalSubject := session.NewTarget(signer3.UserID())

		var tok session.TokenV2
		tok.SetID(uuid.New())
		tok.SetIssuer(session.NewTarget(issuer.UserID()))
		tok.AddSubject(intermediate1)
		tok.SetIat(100)
		tok.SetNbf(200)
		tok.SetExp(300)
		tok.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{
			session.VerbV2ObjectGet,
			session.VerbV2ObjectPut,
		}))

		// First delegation: issuer -> intermediate1 (Get + Put) - token lifetime is 200-300, use 210-290
		del1 := session.NewDelegationInfo([]session.Target{intermediate1}, session.NewLifetime(210, 210, 290), []session.VerbV2{
			session.VerbV2ObjectGet,
			session.VerbV2ObjectPut,
		})
		err := del1.Sign(issuer)
		require.NoError(t, err)
		tok.AddDelegation(del1)

		// Second delegation: intermediate1 -> intermediate2 (only Get, narrowing permissions) - within del1: 210-290, use 220-280
		del2 := session.NewDelegationInfo([]session.Target{intermediate2}, session.NewLifetime(220, 220, 280), []session.VerbV2{
			session.VerbV2ObjectGet,
		})
		// Sign as intermediate1
		err = del2.Sign(signer1)
		require.NoError(t, err)
		tok.AddDelegation(del2)

		// Third delegation: intermediate2 -> finalSubject (only Get) - within del2: 220-280, use 230-270
		del3 := session.NewDelegationInfo([]session.Target{finalSubject}, session.NewLifetime(230, 230, 270), []session.VerbV2{
			session.VerbV2ObjectGet,
		})
		err = del3.Sign(signer2)
		require.NoError(t, err)
		tok.AddDelegation(del3)

		err = tok.ValidateDelegationChain()
		require.NoError(t, err)
	})
}

func TestContextV2(t *testing.T) {
	containerID := cidtest.ID()
	verbs := []session.VerbV2{session.VerbV2ObjectGet, session.VerbV2ObjectPut}

	t.Run("NewContextV2", func(t *testing.T) {
		ctx := session.NewContextV2(containerID, verbs)

		require.Equal(t, containerID, ctx.Container())
		require.Equal(t, verbs, ctx.Verbs())
		require.Empty(t, ctx.Objects())
	})

	t.Run("SetObjects", func(t *testing.T) {
		ctx := session.NewContextV2(containerID, verbs)
		objects := []oid.ID{oidtest.ID(), oidtest.ID()}

		ctx.SetObjects(objects)

		require.Equal(t, objects, ctx.Objects())
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("empty verbs", func(t *testing.T) {
			ctxEmptyVerbs := session.NewContextV2(containerID, []session.VerbV2{})
			require.Empty(t, ctxEmptyVerbs.Verbs())
			require.Equal(t, containerID, ctxEmptyVerbs.Container())
		})

		t.Run("nil verbs", func(t *testing.T) {
			ctxNilVerbs := session.NewContextV2(containerID, nil)
			require.Empty(t, ctxNilVerbs.Verbs())
		})

		t.Run("zero container ID", func(t *testing.T) {
			var zeroContainer cid.ID
			ctxZeroContainer := session.NewContextV2(zeroContainer, verbs)
			require.True(t, ctxZeroContainer.Container().IsZero())
			require.Equal(t, verbs, ctxZeroContainer.Verbs())
		})

		t.Run("SetObjects with nil slice", func(t *testing.T) {
			ctx := session.NewContextV2(containerID, verbs)
			ctx.SetObjects(nil)
			require.Empty(t, ctx.Objects())
		})

		t.Run("SetObjects with empty slice", func(t *testing.T) {
			ctx := session.NewContextV2(containerID, verbs)
			ctx.SetObjects([]oid.ID{})
			require.Empty(t, ctx.Objects())
		})
	})
}

func TestTokenV2_Setters(t *testing.T) {
	var tok session.TokenV2

	t.Run("SetVersion", func(t *testing.T) {
		tok.SetVersion(2)
		require.Equal(t, uint32(2), tok.Version())
	})

	t.Run("SetID", func(t *testing.T) {
		id := uuid.New()
		tok.SetID(id)
		require.Equal(t, id, tok.ID())
	})

	t.Run("SetIssuer", func(t *testing.T) {
		issuer := session.NewTarget(usertest.ID())
		tok.SetIssuer(issuer)
		require.True(t, issuer.Equals(tok.Issuer()))
	})

	t.Run("SetSubjects and AddSubject", func(t *testing.T) {
		subject1 := session.NewTarget(usertest.ID())
		subject2 := session.NewTarget(usertest.ID())
		tok.SetSubjects([]session.Target{subject1})
		require.Len(t, tok.Subjects(), 1)

		tok.AddSubject(subject2)
		require.Len(t, tok.Subjects(), 2)
	})

	t.Run("SetIat", func(t *testing.T) {
		tok.SetIat(100)
		require.Equal(t, uint64(100), tok.Iat())
	})

	t.Run("SetNbf", func(t *testing.T) {
		tok.SetNbf(200)
		require.Equal(t, uint64(200), tok.Nbf())
	})

	t.Run("SetExp", func(t *testing.T) {
		tok.SetExp(300)
		require.Equal(t, uint64(300), tok.Exp())
	})

	t.Run("SetContexts and AddContext", func(t *testing.T) {
		ctx1 := session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet})
		ctx2 := session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectPut})

		tok.SetContexts([]session.ContextV2{ctx1})
		require.Len(t, tok.Contexts(), 1)

		tok.AddContext(ctx2)
		require.Len(t, tok.Contexts(), 2)
	})

	t.Run("SetDelegationChain and AddDelegation", func(t *testing.T) {
		lifetime := session.NewLifetime(100, 100, 100)
		del1 := session.NewDelegationInfo(
			[]session.Target{session.NewTarget(usertest.ID())},
			lifetime,
			[]session.VerbV2{session.VerbV2ObjectGet},
		)
		del2 := session.NewDelegationInfo(
			[]session.Target{session.NewTarget(usertest.ID())},
			lifetime,
			[]session.VerbV2{session.VerbV2ObjectPut},
		)

		tok.SetDelegationChain([]session.DelegationInfo{del1})
		require.Len(t, tok.DelegationChain(), 1)

		tok.AddDelegation(del2)
		require.Len(t, tok.DelegationChain(), 2)
	})

	t.Run("error cases", func(t *testing.T) {
		var errTok session.TokenV2

		// Test SetSubjects with nil slice
		errTok.SetSubjects(nil)
		require.Empty(t, errTok.Subjects())

		// Test SetSubjects with empty slice
		errTok.SetSubjects([]session.Target{})
		require.Empty(t, errTok.Subjects())

		// Test SetContexts with nil slice
		errTok.SetContexts(nil)
		require.Empty(t, errTok.Contexts())

		// Test SetContexts with empty slice
		errTok.SetContexts([]session.ContextV2{})
		require.Empty(t, errTok.Contexts())

		// Test SetDelegationChain with nil slice
		errTok.SetDelegationChain(nil)
		require.Empty(t, errTok.DelegationChain())

		// Test SetDelegationChain with empty slice
		errTok.SetDelegationChain([]session.DelegationInfo{})
		require.Empty(t, errTok.DelegationChain())

		// Test with zero values
		var zeroTok session.TokenV2
		require.Zero(t, zeroTok.Version())
		require.Empty(t, zeroTok.Subjects())
		require.Empty(t, zeroTok.Contexts())
		require.Empty(t, zeroTok.DelegationChain())
		require.Zero(t, zeroTok.Iat())
		require.Zero(t, zeroTok.Nbf())
		require.Zero(t, zeroTok.Exp())
	})
}

func TestTokenV2_ValidAt(t *testing.T) {
	var tok session.TokenV2
	tok.SetIat(100)
	tok.SetNbf(200)
	tok.SetExp(300)

	require.False(t, tok.ValidAt(150))
	require.True(t, tok.ValidAt(250))
	require.False(t, tok.ValidAt(350))

	t.Run("edge cases", func(t *testing.T) {
		require.False(t, tok.ValidAt(199)) // Just before nbf
		require.True(t, tok.ValidAt(200))  // Exactly at nbf
		require.True(t, tok.ValidAt(300))  // Exactly at exp
		require.False(t, tok.ValidAt(301)) // Just after exp

		// Test with zero values
		var zeroTok session.TokenV2
		require.True(t, zeroTok.ValidAt(0))
		require.False(t, zeroTok.ValidAt(1)) // After exp (which is 0)
	})
}

func TestTokenV2_Sign_and_Verify(t *testing.T) {
	tok := newValidTokenV2()

	signer := usertest.User()
	err := tok.Sign(signer)
	require.NoError(t, err)

	require.True(t, tok.VerifySignature())

	// Modify token and verify signature fails
	tok.SetExp(400)
	require.False(t, tok.VerifySignature())

	t.Run("unsigned token", func(t *testing.T) {
		unsignedTok := newValidTokenV2()
		require.False(t, unsignedTok.VerifySignature())

		_, ok := unsignedTok.Signature()
		require.False(t, ok)

		signer2 := usertest.User()
		err := unsignedTok.Sign(signer2)
		require.NoError(t, err)
		sig, ok := unsignedTok.Signature()
		require.True(t, ok)
		require.NotNil(t, sig)
	})

	t.Run("sign with zero ID signer", func(t *testing.T) {
		zeroIDUser := user.NewSigner(usertest.User(), user.ID{})
		err := tok.Sign(zeroIDUser)
		require.ErrorIs(t, err, user.ErrZeroID)
	})

	t.Run("sign and add delegation", func(t *testing.T) {
		issuer := usertest.User()
		subject := session.NewTarget(usertest.ID())

		lifetime := session.NewLifetime(250, 250, 250)
		del := session.NewDelegationInfo([]session.Target{subject}, lifetime, []session.VerbV2{session.VerbV2ObjectGet})
		require.NoError(t, del.Sign(issuer))

		tok := newTokenV2ForDelegation(del.Issuer(), subject)
		require.NoError(t, tok.Sign(issuer))

		require.True(t, tok.VerifySignature())

		tok.AddDelegation(del)
		require.True(t, tok.VerifySignature())
	})
}

func TestTokenV2_AttachSignature(t *testing.T) {
	tok := newValidTokenV2()

	signer := neofscryptotest.Signer()
	var sig neofscrypto.Signature
	err := sig.Calculate(signer, tok.SignedData())
	require.NoError(t, err)

	tok.AttachSignature(sig)

	retrievedSig, ok := tok.Signature()
	require.True(t, ok)
	require.Equal(t, sig.Scheme(), retrievedSig.Scheme())
	require.Equal(t, sig.PublicKeyBytes(), retrievedSig.PublicKeyBytes())
	require.Equal(t, sig.Value(), retrievedSig.Value())
}

func TestTokenV2_AssertAuthority(t *testing.T) {
	subject1 := session.NewTarget(usertest.ID())
	subject2 := session.NewTarget(usertest.ID())
	subject3 := session.NewTarget(usertest.ID())

	var tok session.TokenV2
	tok.SetID(uuid.New())
	tok.SetIssuer(session.NewTarget(usertest.ID()))
	tok.SetIat(100)
	tok.SetNbf(200)
	tok.SetExp(300)
	tok.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))
	tok.AddSubject(subject1)
	tok.AddSubject(subject2)

	signer := usertest.User()
	err := tok.Sign(signer)
	require.NoError(t, err)

	require.True(t, tok.AssertAuthority(subject1))
	require.True(t, tok.AssertAuthority(subject2))
	require.False(t, tok.AssertAuthority(subject3))

	t.Run("with valid delegation chain", func(t *testing.T) {
		delegateSubject := session.NewTarget(usertest.ID())

		del := session.NewDelegationInfo([]session.Target{delegateSubject}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
		delegateSigner := usertest.User()
		err := del.Sign(delegateSigner)
		require.NoError(t, err)

		var tokWithDel session.TokenV2
		tokWithDel.SetID(uuid.New())
		tokWithDel.SetIssuer(del.Issuer())
		tokWithDel.SetIat(100)
		tokWithDel.SetNbf(200)
		tokWithDel.SetExp(300)
		tokWithDel.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))
		tokWithDel.AddSubject(delegateSubject)
		tokWithDel.AddDelegation(del)

		err = tokWithDel.Sign(delegateSigner)
		require.NoError(t, err)

		require.True(t, tokWithDel.AssertAuthority(delegateSubject))
	})

	t.Run("with invalid delegation chain", func(t *testing.T) {
		var tokInvalidDel session.TokenV2
		tokInvalidDel.SetID(uuid.New())
		tokInvalidDel.SetIssuer(session.NewTarget(usertest.ID()))
		tokInvalidDel.SetIat(100)
		tokInvalidDel.SetNbf(200)
		tokInvalidDel.SetExp(300)
		tokInvalidDel.AddContext(session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet}))
		tokInvalidDel.AddSubject(subject1)

		unsignedDel := session.NewDelegationInfo([]session.Target{subject3}, session.NewLifetime(250, 250, 250), []session.VerbV2{session.VerbV2ObjectGet})
		tokInvalidDel.AddDelegation(unsignedDel)

		err := tokInvalidDel.Sign(usertest.User())
		require.NoError(t, err)

		// Should return false because delegation chain validation fails
		require.False(t, tokInvalidDel.AssertAuthority(subject3))
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("empty subjects", func(t *testing.T) {
			var emptyTok session.TokenV2
			require.False(t, emptyTok.AssertAuthority(subject1))
		})

		t.Run("zero target", func(t *testing.T) {
			var zeroTarget session.Target
			require.False(t, tok.AssertAuthority(zeroTarget))
		})

		t.Run("nns target subjects", func(t *testing.T) {
			nnsTarget1 := session.NewTargetFromNNS("test.neo")
			nnsTarget2 := session.NewTargetFromNNS("test.neo")
			nnsTarget3 := session.NewTargetFromNNS("other.neo")

			var nnsTok session.TokenV2
			nnsTok.AddSubject(nnsTarget1)
			require.True(t, nnsTok.AssertAuthority(nnsTarget2))
			require.False(t, nnsTok.AssertAuthority(nnsTarget3))
		})
	})
}

func testVerbAssertionCommonCases(t *testing.T, containerID cid.ID, verb1, verb2, verb3 session.VerbV2, makeAssert func(session.TokenV2, session.VerbV2, cid.ID) bool) {
	t.Run("empty contexts", func(t *testing.T) {
		var emptyTok session.TokenV2
		require.False(t, makeAssert(emptyTok, verb1, containerID))
	})

	t.Run("unspecified verb", func(t *testing.T) {
		var tok session.TokenV2
		ctx := session.NewContextV2(containerID, []session.VerbV2{verb1})
		tok.AddContext(ctx)
		require.False(t, makeAssert(tok, session.VerbV2Unspecified, containerID))
	})

	t.Run("context with empty verbs", func(t *testing.T) {
		ctxEmptyVerbs := session.NewContextV2(containerID, []session.VerbV2{})
		var tokEmptyVerbs session.TokenV2
		tokEmptyVerbs.AddContext(ctxEmptyVerbs)
		require.False(t, makeAssert(tokEmptyVerbs, verb1, containerID))
	})

	t.Run("multiple contexts", func(t *testing.T) {
		ctx1 := session.NewContextV2(containerID, []session.VerbV2{verb1})
		ctx2 := session.NewContextV2(containerID, []session.VerbV2{verb2})
		var multiCtxTok session.TokenV2
		multiCtxTok.AddContext(ctx1)
		multiCtxTok.AddContext(ctx2)
		require.True(t, makeAssert(multiCtxTok, verb1, containerID))
		require.True(t, makeAssert(multiCtxTok, verb2, containerID))
		require.False(t, makeAssert(multiCtxTok, verb3, containerID))
	})
}

func TestTokenV2_AssertVerb(t *testing.T) {
	containerID := cidtest.ID()
	otherContainerID := cidtest.ID()

	var tok session.TokenV2
	ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ObjectGet, session.VerbV2ObjectPut})
	tok.AddContext(ctx)

	require.True(t, tok.AssertVerb(session.VerbV2ObjectGet, containerID))
	require.True(t, tok.AssertVerb(session.VerbV2ObjectPut, containerID))
	require.False(t, tok.AssertVerb(session.VerbV2ObjectDelete, containerID))
	require.False(t, tok.AssertVerb(session.VerbV2ObjectGet, otherContainerID))

	ctx2 := session.NewContextV2(cid.ID{}, []session.VerbV2{session.VerbV2ObjectHead})
	tok.AddContext(ctx2)

	require.True(t, tok.AssertVerb(session.VerbV2ObjectHead, containerID))
	require.True(t, tok.AssertVerb(session.VerbV2ObjectHead, otherContainerID))

	testVerbAssertionCommonCases(t, containerID, session.VerbV2ObjectGet, session.VerbV2ObjectPut, session.VerbV2ObjectDelete,
		func(tok session.TokenV2, verb session.VerbV2, cid cid.ID) bool { return tok.AssertVerb(verb, cid) })
}

func TestTokenV2_AssertObject(t *testing.T) {
	containerID := cidtest.ID()
	objectID1 := oidtest.ID()
	objectID2 := oidtest.ID()
	objectID3 := oidtest.ID()

	var tok session.TokenV2

	t.Run("no specific objects", func(t *testing.T) {
		ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ObjectGet})
		tok.AddContext(ctx)

		require.True(t, tok.AssertObject(session.VerbV2ObjectGet, containerID, objectID1))
		require.True(t, tok.AssertObject(session.VerbV2ObjectGet, containerID, objectID2))
		require.False(t, tok.AssertObject(session.VerbV2ObjectPut, containerID, objectID1))
	})

	t.Run("with specific objects", func(t *testing.T) {
		tok = session.TokenV2{}
		ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ObjectGet})
		ctx.SetObjects([]oid.ID{objectID1, objectID2})
		tok.AddContext(ctx)

		require.True(t, tok.AssertObject(session.VerbV2ObjectGet, containerID, objectID1))
		require.True(t, tok.AssertObject(session.VerbV2ObjectGet, containerID, objectID2))
		require.False(t, tok.AssertObject(session.VerbV2ObjectGet, containerID, objectID3))
	})

	t.Run("container mismatch", func(t *testing.T) {
		otherContainerID := cidtest.ID()
		require.False(t, tok.AssertObject(session.VerbV2ObjectGet, otherContainerID, objectID1))
	})

	t.Run("empty contexts", func(t *testing.T) {
		var emptyTok session.TokenV2
		require.False(t, emptyTok.AssertObject(session.VerbV2ObjectGet, containerID, objectID1))
	})

	t.Run("zero object ID", func(t *testing.T) {
		var zeroObjID oid.ID
		ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ObjectGet})
		ctx.SetObjects([]oid.ID{objectID1})
		var testTok session.TokenV2
		testTok.AddContext(ctx)
		require.False(t, testTok.AssertObject(session.VerbV2ObjectGet, containerID, zeroObjID))
	})

	t.Run("zero container ID wildcard", func(t *testing.T) {
		zeroCtx := session.NewContextV2(cid.ID{}, []session.VerbV2{session.VerbV2ObjectGet})
		var wildcardTok session.TokenV2
		wildcardTok.AddContext(zeroCtx)
		require.True(t, wildcardTok.AssertObject(session.VerbV2ObjectGet, containerID, objectID1))
		require.True(t, wildcardTok.AssertObject(session.VerbV2ObjectGet, cidtest.ID(), objectID2))
	})

	t.Run("container verbs rejected", func(t *testing.T) {
		ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ContainerPut})
		var containerVerbTok session.TokenV2
		containerVerbTok.AddContext(ctx)
		require.False(t, containerVerbTok.AssertObject(session.VerbV2ContainerPut, containerID, objectID1))
		require.False(t, containerVerbTok.AssertObject(session.VerbV2ContainerDelete, containerID, objectID1))
		require.False(t, containerVerbTok.AssertObject(session.VerbV2ContainerSetEACL, containerID, objectID1))
	})
}

func TestTokenV2_AssertContainer(t *testing.T) {
	containerID := cidtest.ID()
	otherContainerID := cidtest.ID()

	var tok session.TokenV2
	ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ContainerPut, session.VerbV2ContainerDelete})
	tok.AddContext(ctx)

	require.True(t, tok.AssertContainer(session.VerbV2ContainerPut, containerID))
	require.True(t, tok.AssertContainer(session.VerbV2ContainerDelete, containerID))
	require.False(t, tok.AssertContainer(session.VerbV2ContainerSetEACL, containerID))
	require.False(t, tok.AssertContainer(session.VerbV2ContainerPut, otherContainerID))

	ctx2 := session.NewContextV2(cid.ID{}, []session.VerbV2{session.VerbV2ContainerSetEACL})
	tok.AddContext(ctx2)

	require.True(t, tok.AssertContainer(session.VerbV2ContainerSetEACL, containerID))
	require.True(t, tok.AssertContainer(session.VerbV2ContainerSetEACL, otherContainerID))

	t.Run("object verbs rejected", func(t *testing.T) {
		require.False(t, tok.AssertContainer(session.VerbV2ObjectGet, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectPut, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectDelete, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectHead, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectSearch, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectRange, containerID))
		require.False(t, tok.AssertContainer(session.VerbV2ObjectRangeHash, containerID))
	})

	testVerbAssertionCommonCases(t, containerID, session.VerbV2ContainerPut, session.VerbV2ContainerDelete, session.VerbV2ContainerSetEACL,
		func(tok session.TokenV2, verb session.VerbV2, cid cid.ID) bool { return tok.AssertContainer(verb, cid) })

	t.Run("mixed object and container verbs", func(t *testing.T) {
		mixedCtx := session.NewContextV2(containerID, []session.VerbV2{
			session.VerbV2ObjectGet,
			session.VerbV2ContainerPut,
		})
		var mixedTok session.TokenV2
		mixedTok.AddContext(mixedCtx)

		require.True(t, mixedTok.AssertContainer(session.VerbV2ContainerPut, containerID))
		require.False(t, mixedTok.AssertContainer(session.VerbV2ObjectGet, containerID))
	})
}

func TestTokenV2_ProtoMessage(t *testing.T) {
	require.Equal(t, validV2Proto, validV2Token.ProtoMessage())
}

func TestTokenV2_FromProtoMessage(t *testing.T) {
	var val session.TokenV2
	require.NoError(t, val.FromProtoMessage(validV2Proto))

	t.Run("valid", func(t *testing.T) {
		checkTokenFields(t, val)

		require.Equal(t, validV2Token, val)
	})

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidProtoV2TokenCommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				err := new(session.TokenV2).FromProtoMessage(m)
				require.EqualError(t, err, tc.err)
			})
		}
	})
}

func TestTokenV2_Marshal(t *testing.T) {
	require.Equal(t, validBinV2Token, validV2Token.Marshal())
}

func TestTokenV2_Unmarshal(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var val session.TokenV2
		err := val.Unmarshal(validBinV2Token)
		require.NoError(t, err)

		checkTokenFields(t, val)
		require.Equal(t, validV2Token, val)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("invalid proto", func(t *testing.T) {
			err := new(session.TokenV2).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range invalidBinV2CommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.TokenV2).Unmarshal(tc.b), tc.err)
			})
		}
	})
}

func TestTokenV2_MarshalJSON(t *testing.T) {
	b, err := validV2Token.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, validJSONV2Token, string(b))
}
func TestTokenV2_UnmarshalJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var val session.TokenV2
		err := val.UnmarshalJSON([]byte(validJSONV2Token))
		require.NoError(t, err)

		checkTokenFields(t, val)
		require.Equal(t, validV2Token, val)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("invalid JSON", func(t *testing.T) {
			err := new(session.TokenV2).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range invalidJSONV2CommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.TokenV2).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})
}
