package session_test

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofscryptotest "github.com/nspcc-dev/neofs-sdk-go/crypto/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	protosession "github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
	"github.com/stretchr/testify/require"
)

type nnsResolver struct {
	res map[string][]user.ID
}

func (r nnsResolver) HasUser(nnsName string, userID user.ID) (bool, error) {
	for _, uid := range r.res[nnsName] {
		if uid == userID {
			return true, nil
		}
	}
	return false, nil
}

func TestRandomNonce(t *testing.T) {
	t.Run("generates unique values", func(t *testing.T) {
		const iterations = 1000
		nonces := make(map[uint32]bool, iterations)

		for range iterations {
			nonces[session.RandomNonce()] = true
		}

		require.Greater(t, len(nonces), iterations*99/100)
	})
}

func TestTarget(t *testing.T) {
	t.Run("zero target", func(t *testing.T) {
		var target session.Target

		require.False(t, target.IsUserID())
		require.False(t, target.IsNNS())
		require.True(t, target.UserID().IsZero())
		require.Empty(t, target.NNSName())
	})

	t.Run("NewTargetUser", func(t *testing.T) {
		userID := usertest.ID()
		target := session.NewTargetUser(userID)

		require.True(t, target.IsUserID())
		require.False(t, target.IsNNS())
		require.Equal(t, userID, target.UserID())
		require.Empty(t, target.NNSName())
	})

	t.Run("NewTargetNamed", func(t *testing.T) {
		nnsName := "test.neo"
		target := session.NewTargetNamed(nnsName)

		require.False(t, target.IsUserID())
		require.True(t, target.IsNNS())
		require.Equal(t, nnsName, target.NNSName())
		require.True(t, target.UserID().IsZero())
	})

	t.Run("comparable", func(t *testing.T) {
		userID := usertest.ID()
		target1 := session.NewTargetUser(userID)
		target2 := session.NewTargetUser(userID)
		target3 := session.NewTargetUser(usertest.ID())

		require.True(t, target1 == target2)
		require.False(t, target1 == target3)

		nnsTarget1 := session.NewTargetNamed("test.neo")
		nnsTarget2 := session.NewTargetNamed("test.neo")
		nnsTarget3 := session.NewTargetNamed("other.neo")

		require.True(t, nnsTarget1 == nnsTarget2)
		require.False(t, nnsTarget1 == nnsTarget3)
		require.False(t, target1 == nnsTarget1)

		var zeroTarget session.Target
		require.True(t, zeroTarget == session.Target{})
		require.False(t, target1 == zeroTarget)
	})

	t.Run("empty NNS", func(t *testing.T) {
		var target session.Target

		emptyNNSTarget := session.NewTargetNamed("")
		require.False(t, emptyNNSTarget.IsUserID())
		require.False(t, emptyNNSTarget.IsNNS()) // empty string means IsNNS returns false
		require.Empty(t, emptyNNSTarget.NNSName())

		// Zero target has expected behavior
		require.False(t, target.IsUserID())
		require.False(t, target.IsNNS())
	})
}

func newValidToken(t *testing.T) session.Token {
	var tok session.Token
	tok.SetVersion(session.TokenCurrentVersion)
	tok.SetNonce(session.RandomNonce())
	tok.SetIssuer(usertest.ID())
	require.NoError(t, tok.AddSubject(session.NewTargetUser(usertest.ID())))
	tok.SetIat(time.Unix(100, 0))
	tok.SetNbf(time.Unix(200, 0))
	tok.SetExp(time.Unix(300, 0))
	ctx, err := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx))
	return tok
}

func newValidSignedToken(t *testing.T) session.Token {
	tok := newValidToken(t)
	signer := usertest.User()
	err := tok.Sign(signer)
	require.NoError(t, err)
	return tok
}

func newTokenForDelegation(t *testing.T, issuer user.ID, subject session.Target) session.Token {
	var tok session.Token
	tok.SetNonce(session.RandomNonce())
	tok.SetIssuer(issuer)
	require.NoError(t, tok.AddSubject(subject))
	tok.SetIat(time.Unix(100, 0))
	tok.SetNbf(time.Unix(200, 0))
	tok.SetExp(time.Unix(300, 0))
	ctx, err := session.NewContext(anyValidContainerID, []session.Verb{session.VerbObjectGet})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx))
	return tok
}

func TestToken_ValidateFields(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  string
		fn   func(tok *session.Token)
	}{
		{"wrong version", "depth 0: invalid fields: invalid token version: expected 0, got 2", func(tok *session.Token) {
			tok.SetVersion(2)
		}},
		{"missing issuer", "depth 0: invalid fields: issuer is not set", func(tok *session.Token) {
			tok.SetIssuer(user.ID{})
		}},
		{"no subjects", "depth 0: invalid fields: no subjects specified", func(tok *session.Token) {
			require.NoError(t, tok.SetSubjects([]session.Target{}))
		}},
		{"empty subject", "depth 0: invalid fields: subject at index 1 is empty", func(tok *session.Token) {
			require.NoError(t, tok.AddSubject(session.Target{}))
		}},
		{"missing iat", "depth 0: invalid fields: issued at (iat) is not set", func(tok *session.Token) {
			tok.SetIat(time.Time{})
		}},
		{"missing nbf", "depth 0: invalid fields: not valid before (nbf) is not set", func(tok *session.Token) {
			tok.SetNbf(time.Time{})
		}},
		{"missing exp", "depth 0: invalid fields: expiration (exp) is not set", func(tok *session.Token) {
			tok.SetExp(time.Time{})
		}},
		{"nbf after exp", "depth 0: invalid fields: not before (nbf) is after expiration (exp)", func(tok *session.Token) {
			tok.SetNbf(time.Unix(300, 0))
			tok.SetExp(time.Unix(200, 0))
		}},
		{"iat after exp", "depth 0: invalid fields: issued at (iat) is after expiration (exp)", func(tok *session.Token) {
			tok.SetIat(time.Unix(400, 0))
		}},
		{"no contexts", "depth 0: invalid fields: no contexts specified", func(tok *session.Token) {
			require.NoError(t, tok.SetContexts([]session.Context{}))
		}},
		{"empty context", "depth 0: invalid fields: context at index 0 has no verbs", func(tok *session.Token) {
			require.NoError(t, tok.SetContexts([]session.Context{{}}))
		}},
		{"multiple wildcard containers in contexts", "depth 0: invalid fields: duplicate container at index 1: 11111111111111111111111111111111", func(tok *session.Token) {
			ctx, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			ctx2, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx, ctx2}))
		}},
		{"multiple identical containers in contexts", "depth 0: invalid fields: duplicate container at index 1: " + anyValidContainerID.String(), func(tok *session.Token) {
			ctx, err := session.NewContext(anyValidContainerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			ctx2, err := session.NewContext(anyValidContainerID, []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx, ctx2}))
		}},
		{"verbs in context not in ascending order", "depth 0: invalid fields: context at index 0: verbs must be sorted in ascending order (verb 1 at index 1 <= verb 2 at index 0)", func(tok *session.Token) {
			ctx, err := session.NewContext(anyValidContainerID, []session.Verb{session.VerbObjectGet, session.VerbObjectPut})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx}))
		}},
		{"explicit container cannot have the same verbs as wildcard", "depth 0: invalid fields: context at index 1: explicit container cannot have the same verbs as wildcard", func(tok *session.Token) {
			ctx1, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbContainerPut, session.VerbContainerDelete})
			require.NoError(t, err)
			ctx2, err := session.NewContext(cidtest.ID(), []session.Verb{session.VerbContainerPut, session.VerbContainerDelete})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx1, ctx2}))
		}},
		{"unsigned token", "depth 0: invalid fields: token is not signed", func(tok *session.Token) {
			// Don't sign the token
		}},
		{"invalid delegation", "depth 0: invalid fields: token is not signed", func(tok *session.Token) {
			del := newValidToken(t)
			tok.SetOrigin(&del)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tok := newValidToken(t)
			tc.fn(&tok)

			err := tok.Validate()
			require.EqualError(t, err, tc.err)
		})
	}

	t.Run("valid token", func(t *testing.T) {
		tok := newValidSignedToken(t)
		err := tok.Validate()
		require.NoError(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		tok := newValidSignedToken(t)
		tok.SetExp(time.Unix(400, 0))

		err := tok.Validate()
		require.EqualError(t, err, "depth 0: invalid fields: token signature verification failed")
	})

	t.Run("containers in context not in ascending order", func(t *testing.T) {
		tok := newValidToken(t)
		cnr1 := cidtest.ID()
		cnr2 := cidtest.ID()
		if bytes.Compare(cnr1[:], cnr2[:]) < 0 {
			cnr1, cnr2 = cnr2, cnr1
		}

		ctx1, err := session.NewContext(cnr1, []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		ctx2, err := session.NewContext(cnr2, []session.Verb{session.VerbObjectPut})
		require.NoError(t, err)
		require.NoError(t, tok.SetContexts([]session.Context{ctx1, ctx2}))

		err = tok.Validate()
		require.EqualError(t, err, fmt.Sprintf("depth 0: invalid fields: contexts must be sorted by container ID: index 1 (%s) < previous index 0 (%s)", cnr2, cnr1))
	})

	t.Run("limits validation", func(t *testing.T) {
		t.Run("too many subjects", func(t *testing.T) {
			tok := newValidSignedToken(t)

			m := tok.ProtoMessage()
			m.Body.Subjects = make([]*protosession.Target, session.MaxSubjectsPerToken+1)
			for i := range m.Body.Subjects {
				userID := usertest.ID()
				m.Body.Subjects[i] = &protosession.Target{
					Identifier: &protosession.Target_OwnerId{
						OwnerId: userID.ProtoMessage(),
					},
				}
			}
			var invalidTok session.Token
			require.NoError(t, invalidTok.FromProtoMessage(m))

			err := invalidTok.Validate()
			require.ErrorContains(t, err, "too many subjects")
		})

		t.Run("too many contexts", func(t *testing.T) {
			tok := newValidSignedToken(t)

			m := tok.ProtoMessage()
			m.Body.Contexts = make([]*protosession.SessionContextV2, session.MaxContextsPerToken+1)
			for i := range m.Body.Contexts {
				ctx, _ := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
				m.Body.Contexts[i] = &protosession.SessionContextV2{
					Container: ctx.Container().ProtoMessage(),
					Verbs:     []protosession.Verb{protosession.Verb_OBJECT_GET},
				}
			}
			var invalidTok session.Token
			require.NoError(t, invalidTok.FromProtoMessage(m))

			err := invalidTok.Validate()
			require.ErrorContains(t, err, "too many contexts")
		})

		t.Run("too many verbs in context", func(t *testing.T) {
			tok := newValidSignedToken(t)

			m := tok.ProtoMessage()
			m.Body.Contexts[0].Verbs = make([]protosession.Verb, session.MaxVerbsPerContext+1)
			for i := range m.Body.Contexts[0].Verbs {
				m.Body.Contexts[0].Verbs[i] = protosession.Verb(i%11 + 1)
			}
			var invalidTok session.Token
			require.NoError(t, invalidTok.FromProtoMessage(m))

			err := invalidTok.Validate()
			require.ErrorContains(t, err, "too many verbs")
		})

		t.Run("too many objects in context", func(t *testing.T) {
			tok := newValidSignedToken(t)

			m := tok.ProtoMessage()
			m.Body.Contexts[0].Objects = make([]*refs.ObjectID, session.MaxObjectsPerContext+1)
			for i := range m.Body.Contexts[0].Objects {
				m.Body.Contexts[0].Objects[i] = oidtest.ID().ProtoMessage()
			}
			var invalidTok session.Token
			require.NoError(t, invalidTok.FromProtoMessage(m))

			err := invalidTok.Validate()
			require.ErrorContains(t, err, "too many objects")
		})
	})
}

func TestToken_ValidateDelegationChain(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		tok := newValidSignedToken(t)

		err := tok.Validate()
		require.NoError(t, err)
	})

	t.Run("valid single delegation", func(t *testing.T) {
		issuer := usertest.User()
		subject := usertest.User()

		origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
		require.NoError(t, origin.Sign(issuer))

		subject2 := session.NewTargetUser(usertest.ID())
		tok := newTokenForDelegation(t, subject.UserID(), subject2)
		tok.SetOrigin(&origin)
		require.NoError(t, tok.Sign(subject))

		err := tok.Validate()
		require.NoError(t, err)
	})

	t.Run("valid multi-level delegation chain", func(t *testing.T) {
		issuer := usertest.User()
		signer1 := usertest.User()
		signer2 := usertest.User()
		signer3 := usertest.User()
		intermediate1 := session.NewTargetUser(signer1.UserID())
		intermediate2 := session.NewTargetUser(signer2.UserID())
		finalSubject := session.NewTargetUser(signer3.UserID())

		tok := newTokenForDelegation(t, issuer.UserID(), intermediate1)
		require.NoError(t, tok.Sign(issuer))

		// First delegation: intermediate1 -> intermediate2
		del1 := newTokenForDelegation(t, intermediate1.UserID(), intermediate2)
		del1.SetOrigin(&tok)
		require.NoError(t, del1.Sign(signer1))

		// Second delegation: intermediate2 -> finalSubject
		del2 := newTokenForDelegation(t, intermediate2.UserID(), finalSubject)
		del2.SetOrigin(&del1)
		require.NoError(t, del2.Sign(signer2))

		require.NoError(t, del2.Validate())
	})

	t.Run("invalid fields", func(t *testing.T) {
		t.Run("delegation has empty subject", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()

			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			require.NoError(t, origin.AddSubject(session.Target{}))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 1: invalid fields: subject at index 1 is empty")
		})

		t.Run("origin token has invalid lifetime", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			origin.SetNbf(time.Time{})
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 1: invalid fields: not valid before (nbf) is not set")
		})

		t.Run("delegation with invalid signature", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			require.NoError(t, origin.Sign(issuer))
			// Modify origin to invalidate signature
			origin.SetExp(time.Unix(999, 0))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 1: invalid fields: token signature verification failed")
		})

		t.Run("origin token unsigned", func(t *testing.T) {
			subject := usertest.User()
			origin := newTokenForDelegation(t, usertest.ID(), session.NewTargetUser(subject.UserID()))
			// Don't sign origin

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 1: invalid fields: token is not signed")
		})

		t.Run("origin token has empty issuer", func(t *testing.T) {
			subject := usertest.User()
			origin := newTokenForDelegation(t, user.ID{}, session.NewTargetUser(subject.UserID()))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 1: invalid fields: issuer is not set")
		})
	})

	t.Run("lifetime narrowing", func(t *testing.T) {
		t.Run("delegation nbf before token nbf", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			origin.SetNbf(time.Unix(100, 0))
			origin.SetExp(time.Unix(300, 0))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetNbf(time.Unix(50, 0)) // origin.Nbf() > tok.Nbf()
			tok.SetExp(time.Unix(300, 0))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 0: origin token lifetime is outside this token's lifetime")
		})

		t.Run("delegation timestamp outside token lifetime", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			origin.SetNbf(time.Unix(100, 0))
			origin.SetExp(time.Unix(500, 0))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetNbf(time.Unix(200, 0))
			tok.SetExp(time.Unix(600, 0)) // origin.Exp() < tok.Exp()
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 0: origin token lifetime is outside this token's lifetime")
		})

		t.Run("valid exact lifetime match", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			origin.SetNbf(time.Unix(200, 0))
			origin.SetExp(time.Unix(300, 0))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetNbf(time.Unix(200, 0)) // Same nbf
			tok.SetExp(time.Unix(300, 0)) // Same exp
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.NoError(t, err)
		})

		t.Run("valid nested lifetime", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			origin.SetNbf(time.Unix(200, 0))
			origin.SetExp(time.Unix(300, 0))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetNbf(time.Unix(225, 0)) // Later nbf
			tok.SetExp(time.Unix(275, 0)) // Earlier exp
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.NoError(t, err)
		})
	})

	t.Run("context narrowing", func(t *testing.T) {
		t.Run("delegation with unauthorized verb", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containerID := cidtest.ID()

			// Origin token authorizes only ObjectGet
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token tries to use ObjectPut (not authorized)
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx, err = session.NewContext(containerID, []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+containerID.String()+", context 0: verb OBJECT_PUT not authorized by origin")
		})

		t.Run("delegation with authorized subset of verbs", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containerID := cidtest.ID()

			// Origin token authorizes ObjectGet and ObjectPut
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))

			ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectPut, session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token uses only ObjectGet (subset of authorized verbs)
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx2, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx2}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.NoError(t, err)
		})

		t.Run("delegation with wildcard container in origin", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containerID := cidtest.ID()

			// Origin token with wildcard container (zero container ID)
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token with specific container
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx2, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx2}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.NoError(t, err)
		})

		t.Run("delegation with different container", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containerID1 := cidtest.ID()
			containerID2 := cidtest.ID()

			// Origin token for containerID1
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx, err := session.NewContext(containerID1, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token tries to use containerID2
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx2, err := session.NewContext(containerID2, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx2}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+containerID2.String()+" at context 0 not found in origin")
		})

		t.Run("multi-level delegation with verb narrowing", func(t *testing.T) {
			issuer := usertest.User()
			signer1 := usertest.User()
			signer2 := usertest.User()
			containerID := cidtest.ID()

			// Level 0: Root token with Get, Put, Delete
			root := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(signer1.UserID()))
			ctx0, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
				session.VerbObjectDelete,
			})
			require.NoError(t, err)
			require.NoError(t, root.SetContexts([]session.Context{ctx0}))
			require.NoError(t, root.Sign(issuer))

			// Level 1: Narrows to Get, Put
			del1 := newTokenForDelegation(t, signer1.UserID(), session.NewTargetUser(signer2.UserID()))
			ctx1, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
			})
			require.NoError(t, err)
			require.NoError(t, del1.SetContexts([]session.Context{ctx1}))
			del1.SetOrigin(&root)
			require.NoError(t, del1.Sign(signer1))

			// Level 2: Narrows to just Get
			del2 := newTokenForDelegation(t, signer2.UserID(), session.NewTargetUser(usertest.ID()))
			ctx2, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, del2.SetContexts([]session.Context{ctx2}))
			del2.SetOrigin(&del1)
			require.NoError(t, del2.Sign(signer2))

			err = del2.Validate()
			require.NoError(t, err)
		})

		t.Run("multi-level delegation with unauthorized verb in chain", func(t *testing.T) {
			issuer := usertest.User()
			signer1 := usertest.User()
			signer2 := usertest.User()
			containerID := cidtest.ID()

			// Level 0: Root token with Get, Put
			root := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(signer1.UserID()))
			ctx0, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectGet,
				session.VerbObjectPut,
			})
			require.NoError(t, err)
			require.NoError(t, root.SetContexts([]session.Context{ctx0}))
			require.NoError(t, root.Sign(issuer))

			// Level 1: Tries to add Delete (not authorized)
			del1 := newTokenForDelegation(t, signer1.UserID(), session.NewTargetUser(signer2.UserID()))
			ctx1, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectGet,
				session.VerbObjectDelete, // Not authorized in root
			})
			require.NoError(t, err)
			require.NoError(t, del1.SetContexts([]session.Context{ctx1}))
			del1.SetOrigin(&root)
			require.NoError(t, del1.Sign(signer1))

			err = del1.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+containerID.String()+", context 0: verb OBJECT_DELETE not authorized by origin")
		})

		t.Run("deep chain verb validation", func(t *testing.T) {
			// Test that verb validation works recursively through the entire chain
			containerID := cidtest.ID()
			signer0 := usertest.User()
			signer1 := usertest.User()
			signer2 := usertest.User()
			signer3 := usertest.User()

			// Level 0: Root with Get, Put, Delete
			tok0 := newTokenForDelegation(t, signer0.UserID(), session.NewTargetUser(signer1.UserID()))
			ctx0, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectGet,
				session.VerbObjectPut,
				session.VerbObjectDelete,
			})
			require.NoError(t, err)
			require.NoError(t, tok0.SetContexts([]session.Context{ctx0}))
			require.NoError(t, tok0.Sign(signer0))

			// Level 1: Narrows to Get, Put
			tok1 := newTokenForDelegation(t, signer1.UserID(), session.NewTargetUser(signer2.UserID()))
			ctx1, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectGet,
				session.VerbObjectPut,
			})
			require.NoError(t, err)
			require.NoError(t, tok1.SetContexts([]session.Context{ctx1}))
			tok1.SetOrigin(&tok0)
			require.NoError(t, tok1.Sign(signer1))

			// Level 2: Tries to add Delete back (should fail - Delete not in tok1)
			tok2 := newTokenForDelegation(t, signer2.UserID(), session.NewTargetUser(signer3.UserID()))
			ctx2, err := session.NewContext(containerID, []session.Verb{
				session.VerbObjectGet,
				session.VerbObjectDelete, // Was in tok0 but not tok1
			})
			require.NoError(t, err)
			require.NoError(t, tok2.SetContexts([]session.Context{ctx2}))
			tok2.SetOrigin(&tok1)
			require.NoError(t, tok2.Sign(signer2))

			// Level 3: Uses Get (valid through entire chain)
			tok3 := newTokenForDelegation(t, signer3.UserID(), session.NewTargetUser(usertest.ID()))
			ctx3, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, tok3.SetContexts([]session.Context{ctx3}))
			tok3.SetOrigin(&tok2)
			require.NoError(t, tok3.Sign(signer3))

			// tok2 should fail because it tries to use Delete which tok1 doesn't authorize
			err = tok2.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+containerID.String()+", context 0: verb OBJECT_DELETE not authorized by origin")

			// tok3 should also fail because tok2 is invalid
			err = tok3.Validate()
			require.ErrorContains(t, err, "depth 1: invalid origin chain")
		})

		t.Run("delegation with new container when origin has wildcard, but unauthorized verb", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			wildcard := cid.ID{}
			containerID := cidtest.ID()

			// Origin token for all containers with Get, Put, Delete
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx, err := session.NewContext(wildcard, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
				session.VerbObjectDelete,
			})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token uses new containerID2 with unauthorized VerbContainerDelete
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx2, err := session.NewContext(containerID, []session.Verb{session.VerbContainerDelete})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx, ctx2}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+containerID.String()+", context 1: verb CONTAINER_DELETE not authorized by origin")
		})

		t.Run("delegation wildcard with unauthorized verb", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			wildcard := cid.ID{}

			// Origin token authorizes only ObjectGet
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx, err := session.NewContext(wildcard, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token tries to use ObjectPut (not authorized)
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			ctx, err = session.NewContext(wildcard, []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{ctx}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.EqualError(t, err, "depth 0: invalid origin chain: container "+wildcard.String()+", context 0: verb OBJECT_PUT not authorized by origin")
		})

		t.Run("wildcard + same container at the end", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containers := []cid.ID{cidtest.ID(), cidtest.ID(), cidtest.ID(), cidtest.ID()}
			sort.Slice(containers, func(i, j int) bool {
				return bytes.Compare(containers[i][:], containers[j][:]) < 0
			})
			wildcard := cid.ID{}
			t.Logf("containers: %s, %s, %s, %s", containers[0], containers[1], containers[2], containers[3])
			ctx0, err := session.NewContext(wildcard, []session.Verb{session.VerbObjectPut, session.VerbObjectGet})
			require.NoError(t, err)
			ctx1, err := session.NewContext(containers[0], []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			ctx2, err := session.NewContext(containers[2], []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			ctx3, err := session.NewContext(containers[1], []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)
			ctx4, err := session.NewContext(containers[3], []session.Verb{session.VerbObjectPut})
			require.NoError(t, err)

			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			require.NoError(t, origin.SetContexts([]session.Context{ctx0, ctx2, ctx4}))
			require.NoError(t, origin.Sign(issuer))

			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			require.NoError(t, tok.SetContexts([]session.Context{ctx1, ctx3, ctx4}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.NoError(t, err)
		})

		t.Run("delegation narrowing with multiple containers", func(t *testing.T) {
			issuer := usertest.User()
			subject := usertest.User()
			containerID1 := cidtest.ID()
			containerID2 := cidtest.ID()
			if bytes.Compare(containerID1[:], containerID2[:]) > 0 {
				containerID1, containerID2 = containerID2, containerID1
			}

			// Origin token authorizes 2 containers with Get, Put, Delete
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
			ctx1, err := session.NewContext(containerID1, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
				session.VerbObjectDelete,
			})
			require.NoError(t, err)
			ctx2, err := session.NewContext(containerID2, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
			})
			require.NoError(t, err)
			require.NoError(t, origin.SetContexts([]session.Context{ctx1, ctx2}))
			require.NoError(t, origin.Sign(issuer))

			// Delegated token narrows both containers:
			// - containerID1: Get, Put (excludes Delete)
			// - containerID2: Get only (excludes Put)
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			delCtx1, err := session.NewContext(containerID1, []session.Verb{
				session.VerbObjectPut,
				session.VerbObjectGet,
			})
			require.NoError(t, err)
			delCtx2, err := session.NewContext(containerID2, []session.Verb{session.VerbObjectGet})
			require.NoError(t, err)
			require.NoError(t, tok.SetContexts([]session.Context{delCtx1, delCtx2}))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err = tok.Validate()
			require.NoError(t, err)
		})
	})

	t.Run("issuer authorization", func(t *testing.T) {
		t.Run("issuer not in origin token subjects", func(t *testing.T) {
			issuer := usertest.User()

			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(usertest.ID()))
			require.NoError(t, origin.Sign(issuer))

			subject := usertest.User()
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.Validate()
			require.EqualError(t, err, "depth 0: token issuer is not in this origin token's subjects")
		})

		t.Run("issuer in origin token NNS subjects", func(t *testing.T) {
			issuer := usertest.User()
			origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetNamed("some.nns"))
			require.NoError(t, origin.Sign(issuer))

			subject := usertest.User() // corresponds to "some.nns" mapped user
			resolver := nnsResolver{map[string][]user.ID{"some.nns": {subject.UserID()}}}
			tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
			tok.SetOrigin(&origin)
			require.NoError(t, tok.Sign(subject))

			err := tok.ValidateWithNNS(resolver)
			require.NoError(t, err)
		})
	})

	t.Run("final token with delegation chain", func(t *testing.T) {
		issuer := usertest.User()
		subject := usertest.User()
		origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
		require.NoError(t, origin.Sign(issuer))

		tok := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
		tok.SetFinal(true)
		tok.SetOrigin(&origin)
		require.NoError(t, tok.Sign(subject))

		err := tok.Validate()
		require.NoError(t, err)
	})

	t.Run("final token cannot be further delegated", func(t *testing.T) {
		issuer := usertest.User()
		subject := usertest.User()
		finalTarget := session.NewTargetUser(usertest.ID())

		origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
		origin.SetFinal(true)
		require.NoError(t, origin.Sign(issuer))

		del := newTokenForDelegation(t, subject.UserID(), finalTarget)
		del.SetOrigin(&origin)
		require.NoError(t, del.Sign(subject))

		err := del.Validate()
		require.EqualError(t, err, "depth 1: final token cannot be used as origin (further delegated)")
	})

	t.Run("delegation depth limit", func(t *testing.T) {
		t.Run("exactly at limit", func(t *testing.T) {
			// Create a chain at exactly MaxDelegationDepth (4)
			var currentToken *session.Token
			signer := usertest.User()
			for i := range session.MaxDelegationDepth + 1 { // 5 tokens (4 delegations)
				nextSigner := usertest.User()

				newTok := newTokenForDelegation(t, signer.UserID(), session.NewTargetUser(nextSigner.UserID()))
				if i > 0 {
					newTok.SetOrigin(currentToken)
				}
				require.NoError(t, newTok.Sign(signer))
				currentToken = &newTok
				signer = nextSigner
			}

			require.NoError(t, currentToken.Validate())
		})

		t.Run("exceeds limit", func(t *testing.T) {
			// Create a chain exceeding MaxDelegationDepth
			var currentToken *session.Token
			signer := usertest.User()
			for i := range session.MaxDelegationDepth + 2 { // 6 tokens (5 delegations)
				nextSigner := usertest.User()

				newTok := newTokenForDelegation(t, signer.UserID(), session.NewTargetUser(nextSigner.UserID()))
				if i > 0 {
					newTok.SetOrigin(currentToken)
				}
				require.NoError(t, newTok.Sign(signer))
				currentToken = &newTok
				signer = nextSigner
			}

			err := currentToken.Validate()
			require.EqualError(t, err, "delegation chain exceeds maximum depth of 4")
		})
	})

	t.Run("cycle detection", func(t *testing.T) {
		t.Run("direct cycle - token references itself", func(t *testing.T) {
			issuer := usertest.User()

			tok := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(issuer.UserID()))
			require.NoError(t, tok.Sign(issuer))

			// Create a cycle: tok -> tok
			tok.SetOrigin(&tok)

			err := tok.Validate()
			require.EqualError(t, err, "delegation chain exceeds maximum depth of 4")
		})

		t.Run("simple cycle - A -> B -> A", func(t *testing.T) {
			signer1 := usertest.User()
			signer2 := usertest.User()

			tokA := newTokenForDelegation(t, signer1.UserID(), session.NewTargetUser(signer2.UserID()))
			require.NoError(t, tokA.Sign(signer1))

			tokB := newTokenForDelegation(t, signer2.UserID(), session.NewTargetUser(signer1.UserID()))
			require.NoError(t, tokB.Sign(signer2))

			tokB.SetOrigin(&tokA)
			tokA.SetOrigin(&tokB)

			err := tokB.Validate()
			require.EqualError(t, err, "delegation chain exceeds maximum depth of 4")
		})

		t.Run("longer cycle - A -> B -> C -> A", func(t *testing.T) {
			signer1 := usertest.User()
			signer2 := usertest.User()
			signer3 := usertest.User()

			tokA := newTokenForDelegation(t, signer1.UserID(), session.NewTargetUser(signer2.UserID()))
			require.NoError(t, tokA.Sign(signer1))

			// B -> A
			tokB := newTokenForDelegation(t, signer2.UserID(), session.NewTargetUser(signer3.UserID()))
			tokB.SetOrigin(&tokA)
			require.NoError(t, tokB.Sign(signer2))

			// C -> B
			tokC := newTokenForDelegation(t, signer3.UserID(), session.NewTargetUser(signer1.UserID()))
			tokC.SetOrigin(&tokB)
			require.NoError(t, tokC.Sign(signer3))

			// A -> C
			tokA.SetOrigin(&tokC)

			err := tokC.Validate()
			require.EqualError(t, err, "delegation chain exceeds maximum depth of 4")
		})
	})
}

func TestToken_OriginalIssuer(t *testing.T) {
	t.Run("token without origin", func(t *testing.T) {
		issuer := usertest.User()
		subject := usertest.User()

		tok := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
		require.NoError(t, tok.Sign(issuer))

		originalIssuer := tok.OriginalIssuer()
		require.Equal(t, originalIssuer, issuer.UserID())
		require.Equal(t, originalIssuer, tok.Issuer())
	})

	t.Run("single level delegation", func(t *testing.T) {
		originalIssuer := usertest.User()
		delegatedIssuer := usertest.User()
		finalSubject := usertest.User()

		tok := newTokenForDelegation(t, originalIssuer.UserID(), session.NewTargetUser(delegatedIssuer.UserID()))
		require.NoError(t, tok.Sign(originalIssuer))

		del := newTokenForDelegation(t, delegatedIssuer.UserID(), session.NewTargetUser(finalSubject.UserID()))
		del.SetOrigin(&tok)
		require.NoError(t, del.Sign(delegatedIssuer))

		originalIssuerTarget := del.OriginalIssuer()
		require.Equal(t, originalIssuerTarget, originalIssuer.UserID())
		require.NotEqual(t, originalIssuerTarget, del.Issuer())
	})

	t.Run("multi-level delegation chain", func(t *testing.T) {
		issuer := usertest.User()
		signer1 := usertest.User()
		signer2 := usertest.User()
		signer3 := usertest.User()
		intermediate1 := session.NewTargetUser(signer1.UserID())
		intermediate2 := session.NewTargetUser(signer2.UserID())
		finalSubject := session.NewTargetUser(signer3.UserID())

		// Level 0: Original token (issuer -> intermediate1)
		tok := newTokenForDelegation(t, issuer.UserID(), intermediate1)
		require.NoError(t, tok.Sign(issuer))

		// Level 1: First delegation (intermediate1 -> intermediate2)
		del1 := newTokenForDelegation(t, intermediate1.UserID(), intermediate2)
		del1.SetOrigin(&tok)
		require.NoError(t, del1.Sign(signer1))

		// Level 2: Second delegation (intermediate2 -> finalSubject)
		del2 := newTokenForDelegation(t, intermediate2.UserID(), finalSubject)
		del2.SetOrigin(&del1)
		require.NoError(t, del2.Sign(signer2))

		// OriginalIssuer at each level should return the original issuer
		require.Equal(t, tok.OriginalIssuer(), issuer.UserID())
		require.Equal(t, del1.OriginalIssuer(), issuer.UserID())
		require.Equal(t, del2.OriginalIssuer(), issuer.UserID())

		// Verify that Issuer() returns different values at each level
		require.Equal(t, tok.Issuer(), issuer.UserID())
		require.Equal(t, del1.Issuer(), intermediate1.UserID())
		require.Equal(t, del2.Issuer(), intermediate2.UserID())
	})
}

func TestContext(t *testing.T) {
	containerID := cidtest.ID()
	verbs := []session.Verb{session.VerbObjectGet, session.VerbObjectPut}

	t.Run("NewContext", func(t *testing.T) {
		ctx, err := session.NewContext(containerID, verbs)
		require.NoError(t, err)

		require.Equal(t, containerID, ctx.Container())
		require.Equal(t, verbs, ctx.Verbs())
		require.Empty(t, ctx.Objects())
	})

	t.Run("SetObjects", func(t *testing.T) {
		ctx, err := session.NewContext(containerID, verbs)
		require.NoError(t, err)
		objects := []oid.ID{oidtest.ID(), oidtest.ID()}

		require.NoError(t, ctx.SetObjects(objects))

		require.Equal(t, objects, ctx.Objects())
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("empty verbs", func(t *testing.T) {
			ctxEmptyVerbs, err := session.NewContext(containerID, []session.Verb{})
			require.EqualError(t, err, "no verbs specified")
			require.Empty(t, ctxEmptyVerbs.Verbs())
			require.Empty(t, ctxEmptyVerbs.Container())
		})

		t.Run("nil verbs", func(t *testing.T) {
			ctxNilVerbs, err := session.NewContext(containerID, nil)
			require.EqualError(t, err, "no verbs specified")
			require.Empty(t, ctxNilVerbs.Verbs())
		})

		t.Run("zero container ID", func(t *testing.T) {
			var zeroContainer cid.ID
			ctxZeroContainer, err := session.NewContext(zeroContainer, verbs)
			require.NoError(t, err)
			require.True(t, ctxZeroContainer.Container().IsZero())
			require.Equal(t, verbs, ctxZeroContainer.Verbs())
		})

		t.Run("SetObjects with nil slice", func(t *testing.T) {
			ctx, err := session.NewContext(containerID, verbs)
			require.NoError(t, err)
			require.NoError(t, ctx.SetObjects(nil))
			require.Empty(t, ctx.Objects())
		})

		t.Run("SetObjects with empty slice", func(t *testing.T) {
			ctx, err := session.NewContext(containerID, verbs)
			require.NoError(t, err)
			require.NoError(t, ctx.SetObjects([]oid.ID{}))
			require.Empty(t, ctx.Objects())
		})
	})

	t.Run("verbs limit", func(t *testing.T) {
		t.Run("exactly at limit", func(t *testing.T) {
			verbsAtLimit := make([]session.Verb, session.MaxVerbsPerContext)
			for i := range verbsAtLimit {
				verbsAtLimit[i] = session.Verb(i%11 + 1)
			}
			ctx, err := session.NewContext(containerID, verbsAtLimit)
			require.NoError(t, err)
			require.Len(t, ctx.Verbs(), session.MaxVerbsPerContext)
		})

		t.Run("exceeds limit", func(t *testing.T) {
			verbsOverLimit := make([]session.Verb, session.MaxVerbsPerContext+1)
			for i := range verbsOverLimit {
				verbsOverLimit[i] = session.Verb(i%11 + 1)
			}
			_, err := session.NewContext(containerID, verbsOverLimit)
			require.EqualError(t, err, fmt.Sprintf("too many verbs: expected max %d, got %d", session.MaxVerbsPerContext, session.MaxVerbsPerContext+1))
		})
	})

	t.Run("objects limit", func(t *testing.T) {
		ctx, err := session.NewContext(containerID, verbs)
		require.NoError(t, err)

		t.Run("exactly at limit", func(t *testing.T) {
			objects := make([]oid.ID, session.MaxObjectsPerContext)
			for i := range objects {
				objects[i] = oidtest.ID()
			}
			err = ctx.SetObjects(objects)
			require.NoError(t, err)
			require.Len(t, ctx.Objects(), session.MaxObjectsPerContext)
		})

		t.Run("exceeds limit", func(t *testing.T) {
			ctx2, err := session.NewContext(containerID, verbs)
			require.NoError(t, err)

			objects := make([]oid.ID, session.MaxObjectsPerContext+1)
			for i := range objects {
				objects[i] = oidtest.ID()
			}
			err = ctx2.SetObjects(objects)
			require.EqualError(t, err, fmt.Sprintf("too many objects: expected max %d, got %d", session.MaxObjectsPerContext, session.MaxObjectsPerContext+1))
		})
	})
}

func TestToken_Setters(t *testing.T) {
	var tok session.Token

	t.Run("SetVersion", func(t *testing.T) {
		tok.SetVersion(2)
		require.Equal(t, uint32(2), tok.Version())
	})

	t.Run("SetNonce", func(t *testing.T) {
		nonce := uint32(2)
		tok.SetNonce(nonce)
		require.Equal(t, nonce, tok.Nonce())
	})

	t.Run("SetIssuer", func(t *testing.T) {
		issuer := usertest.ID()
		tok.SetIssuer(issuer)
		require.True(t, issuer == tok.Issuer())
	})

	t.Run("SetSubjects and AddSubject", func(t *testing.T) {
		subject1 := session.NewTargetUser(usertest.ID())
		subject2 := session.NewTargetUser(usertest.ID())
		require.NoError(t, tok.SetSubjects([]session.Target{subject1}))
		require.Len(t, tok.Subjects(), 1)

		require.NoError(t, tok.AddSubject(subject2))
		require.Len(t, tok.Subjects(), 2)

		t.Run("subjects limit", func(t *testing.T) {
			var tok2 session.Token

			t.Run("exactly at limit", func(t *testing.T) {
				subjects := make([]session.Target, session.MaxSubjectsPerToken)
				for i := range subjects {
					subjects[i] = session.NewTargetUser(usertest.ID())
				}
				err := tok2.SetSubjects(subjects)
				require.NoError(t, err)
				require.Len(t, tok2.Subjects(), session.MaxSubjectsPerToken)
			})

			t.Run("exceeds limit via SetSubjects", func(t *testing.T) {
				subjects := make([]session.Target, session.MaxSubjectsPerToken+1)
				for i := range subjects {
					subjects[i] = session.NewTargetUser(usertest.ID())
				}
				err := tok2.SetSubjects(subjects)
				require.EqualError(t, err, fmt.Sprintf("too many subjects: expected max %d, got %d", session.MaxSubjectsPerToken, session.MaxSubjectsPerToken+1))
			})

			t.Run("exceeds limit via AddSubject", func(t *testing.T) {
				var tok3 session.Token
				// Add maximum subjects
				for range session.MaxSubjectsPerToken {
					err := tok3.AddSubject(session.NewTargetUser(usertest.ID()))
					require.NoError(t, err)
				}
				// Try to add one more
				err := tok3.AddSubject(session.NewTargetUser(usertest.ID()))
				require.EqualError(t, err, fmt.Sprintf("cannot add subject: already at maximum of %d", session.MaxSubjectsPerToken))
			})
		})
	})

	t.Run("SetIat", func(t *testing.T) {
		iat := time.Unix(100, 0)
		tok.SetIat(iat)
		require.Equal(t, iat, tok.Iat())
	})

	t.Run("SetNbf", func(t *testing.T) {
		nbf := time.Unix(200, 0)
		tok.SetNbf(nbf)
		require.Equal(t, nbf, tok.Nbf())
	})

	t.Run("SetExp", func(t *testing.T) {
		exp := time.Unix(300, 0)
		tok.SetExp(exp)
		require.Equal(t, exp, tok.Exp())
	})

	t.Run("SetContexts and AddContext", func(t *testing.T) {
		ctx1, err := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		ctx2, err := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectPut})
		require.NoError(t, err)

		require.NoError(t, tok.SetContexts([]session.Context{ctx1}))
		require.Len(t, tok.Contexts(), 1)

		require.NoError(t, tok.AddContext(ctx2))
		require.Len(t, tok.Contexts(), 2)

		t.Run("contexts limit", func(t *testing.T) {
			var tok2 session.Token
			containerID := cidtest.ID()

			t.Run("exactly at limit", func(t *testing.T) {
				contexts := make([]session.Context, session.MaxContextsPerToken)
				for i := range contexts {
					ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
					require.NoError(t, err)
					contexts[i] = ctx
				}
				err := tok2.SetContexts(contexts)
				require.NoError(t, err)
				require.Len(t, tok2.Contexts(), session.MaxContextsPerToken)
			})

			t.Run("exceeds limit via SetContexts", func(t *testing.T) {
				contexts := make([]session.Context, session.MaxContextsPerToken+1)
				for i := range contexts {
					ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
					require.NoError(t, err)
					contexts[i] = ctx
				}
				err := tok2.SetContexts(contexts)
				require.EqualError(t, err, fmt.Sprintf("too many contexts: expected max %d, got %d", session.MaxContextsPerToken, session.MaxContextsPerToken+1))
			})

			t.Run("exceeds limit via AddContext", func(t *testing.T) {
				var tok3 session.Token
				// Add maximum contexts
				for range session.MaxContextsPerToken {
					ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
					require.NoError(t, err)
					err = tok3.AddContext(ctx)
					require.NoError(t, err)
				}
				// Try to add one more
				ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
				require.NoError(t, err)
				err = tok3.AddContext(ctx)
				require.EqualError(t, err, fmt.Sprintf("cannot add context: already at maximum of %d", session.MaxContextsPerToken))
			})
		})
	})

	t.Run("SetOrigin", func(t *testing.T) {
		origin := newValidToken(t)
		tok.SetOrigin(&origin)

		require.Equal(t, tok.Origin(), &origin)
	})

	t.Run("empty values", func(t *testing.T) {
		var errTok session.Token

		// Test SetSubjects with nil slice
		require.NoError(t, errTok.SetSubjects(nil))
		require.Empty(t, errTok.Subjects())

		// Test SetSubjects with empty slice
		require.NoError(t, errTok.SetSubjects([]session.Target{}))
		require.Empty(t, errTok.Subjects())

		// Test SetContexts with nil slice
		require.NoError(t, errTok.SetContexts(nil))
		require.Empty(t, errTok.Contexts())

		// Test SetContexts with empty slice
		require.NoError(t, errTok.SetContexts([]session.Context{}))
		require.Empty(t, errTok.Contexts())

		// Test SetOrigin with nil
		errTok.SetOrigin(nil)
		require.Empty(t, errTok.Origin())

		// Test with zero values
		var zeroTok session.Token
		require.Zero(t, zeroTok.Version())
		require.Empty(t, zeroTok.Subjects())
		require.Empty(t, zeroTok.Contexts())
		require.Empty(t, zeroTok.Origin())
		require.Zero(t, zeroTok.Iat())
		require.Zero(t, zeroTok.Nbf())
		require.Zero(t, zeroTok.Exp())
	})
}

func TestToken_ValidAt(t *testing.T) {
	var tok session.Token
	tok.SetIat(time.Unix(100, 0))
	tok.SetNbf(time.Unix(200, 0))
	tok.SetExp(time.Unix(300, 0))

	require.False(t, tok.ValidAt(time.Unix(150, 0)))
	require.True(t, tok.ValidAt(time.Unix(250, 0)))
	require.False(t, tok.ValidAt(time.Unix(350, 0)))

	t.Run("edge cases", func(t *testing.T) {
		require.False(t, tok.ValidAt(time.Unix(199, 0))) // Just before nbf
		require.True(t, tok.ValidAt(time.Unix(200, 0)))  // Exactly at nbf
		require.True(t, tok.ValidAt(time.Unix(300, 0)))  // Exactly at exp
		require.False(t, tok.ValidAt(time.Unix(301, 0))) // Just after exp

		// Test with zero values
		var zeroTok session.Token
		require.True(t, zeroTok.ValidAt(time.Time{}))
		require.False(t, zeroTok.ValidAt(time.Unix(1, 0))) // After exp (which is 0)
	})
}

func TestToken_Sign_and_Verify(t *testing.T) {
	tok := newValidToken(t)

	signer := usertest.User()
	err := tok.Sign(signer)
	require.NoError(t, err)

	require.True(t, tok.VerifySignature())

	// Modify token and verify signature fails
	tok.SetExp(time.Unix(400, 0))
	require.False(t, tok.VerifySignature())

	t.Run("unsigned token", func(t *testing.T) {
		unsignedTok := newValidToken(t)
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
		subject := usertest.User()

		origin := newTokenForDelegation(t, issuer.UserID(), session.NewTargetUser(subject.UserID()))
		require.NoError(t, origin.Sign(issuer))

		require.True(t, origin.VerifySignature())

		del := newTokenForDelegation(t, subject.UserID(), session.NewTargetUser(usertest.ID()))
		del.SetOrigin(&origin)
		require.NoError(t, del.Sign(subject))
		require.True(t, del.VerifySignature())
	})
}

func TestToken_AttachSignature(t *testing.T) {
	tok := newValidToken(t)

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

func TestToken_AssertAuthority(t *testing.T) {
	subject1 := usertest.User()
	subject1Target := session.NewTargetUser(subject1.UserID())
	subject2 := usertest.ID()
	subject3 := usertest.ID()
	signer := usertest.User()

	tok := newTokenForDelegation(t, signer.UserID(), subject1Target)
	require.NoError(t, tok.AddSubject(session.NewTargetUser(subject2)))

	require.NoError(t, tok.Sign(signer))

	ok, err := tok.AssertAuthority(subject1.UserID(), nil)
	require.True(t, ok)
	require.NoError(t, err)

	ok, err = tok.AssertAuthority(subject2, nil)
	require.True(t, ok)
	require.NoError(t, err)

	ok, err = tok.AssertAuthority(subject3, nil)
	require.False(t, ok)
	require.NoError(t, err)

	t.Run("with delegation chain", func(t *testing.T) {
		del := newTokenForDelegation(t, subject1.UserID(), session.NewTargetUser(subject3))
		del.SetOrigin(&tok)
		require.NoError(t, del.Sign(subject1))

		ok, err = del.AssertAuthority(subject1.UserID(), nil)
		require.False(t, ok)
		require.NoError(t, err)

		ok, err = del.AssertAuthority(subject2, nil)
		require.False(t, ok)
		require.NoError(t, err)

		ok, err = del.AssertAuthority(subject3, nil)
		require.True(t, ok) // direct subject
		require.NoError(t, err)
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("empty subjects", func(t *testing.T) {
			var emptyTok session.Token
			ok, err = emptyTok.AssertAuthority(subject1.UserID(), nil)
			require.False(t, ok)
			require.NoError(t, err)
		})

		t.Run("zero user", func(t *testing.T) {
			ok, err = tok.AssertAuthority(user.ID{}, nil)
			require.False(t, ok)
			require.NoError(t, err)
		})
	})

	t.Run("AssertAuthority with NNS", func(t *testing.T) {
		t.Run("nns subject with user mapping", func(t *testing.T) {
			issuer := usertest.User()
			directUser := usertest.ID()
			mappedUser := usertest.ID()
			otherUser := usertest.ID()

			tok := newTokenForDelegation(t, issuer.UserID(), session.NewTargetNamed("example.neo"))
			require.NoError(t, tok.AddSubject(session.NewTargetUser(directUser)))
			require.NoError(t, tok.Sign(issuer))

			resolver := nnsResolver{map[string][]user.ID{"example.neo": {mappedUser}}}

			ok, err = tok.AssertAuthority(mappedUser, resolver)
			require.True(t, ok)
			require.NoError(t, err)

			ok, err = tok.AssertAuthority(directUser, resolver)
			require.True(t, ok)
			require.NoError(t, err)

			ok, err = tok.AssertAuthority(otherUser, resolver)
			require.False(t, ok)
			require.NoError(t, err)
		})

		t.Run("multiple nns subjects", func(t *testing.T) {
			issuer := usertest.User()
			user11 := usertest.ID()
			user12 := usertest.ID()
			user2 := usertest.ID()
			user3 := usertest.ID()

			tok := newTokenForDelegation(t, issuer.UserID(), session.NewTargetNamed("domain1.neo"))
			require.NoError(t, tok.AddSubject(session.NewTargetNamed("domain2.neo")))
			require.NoError(t, tok.Sign(issuer))

			resolver := nnsResolver{map[string][]user.ID{
				"domain1.neo": {user11, user12}, "domain2.neo": {user2},
			}}

			ok, err = tok.AssertAuthority(user11, resolver)
			require.True(t, ok)
			require.NoError(t, err)

			ok, err = tok.AssertAuthority(user12, resolver)
			require.True(t, ok)
			require.NoError(t, err)

			ok, err = tok.AssertAuthority(user2, resolver)
			require.True(t, ok)
			require.NoError(t, err)

			ok, err = tok.AssertAuthority(user3, resolver)
			require.False(t, ok)
			require.NoError(t, err)
		})
	})
}

func testVerbAssertionCommonCases(t *testing.T, containerID cid.ID, verb1, verb2, verb3 session.Verb, makeAssert func(session.Token, session.Verb, cid.ID) bool) {
	t.Run("empty contexts", func(t *testing.T) {
		var emptyTok session.Token
		require.False(t, makeAssert(emptyTok, verb1, containerID))
	})

	t.Run("unspecified verb", func(t *testing.T) {
		var tok session.Token
		ctx, err := session.NewContext(containerID, []session.Verb{verb1})
		require.NoError(t, err)
		require.NoError(t, tok.AddContext(ctx))
		require.False(t, makeAssert(tok, session.VerbUnspecified, containerID))
	})

	t.Run("multiple contexts", func(t *testing.T) {
		ctx1, err := session.NewContext(containerID, []session.Verb{verb1})
		require.NoError(t, err)
		ctx2, err := session.NewContext(containerID, []session.Verb{verb2})
		require.NoError(t, err)
		var multiCtxTok session.Token
		require.NoError(t, multiCtxTok.AddContext(ctx1))
		require.NoError(t, multiCtxTok.AddContext(ctx2))
		require.True(t, makeAssert(multiCtxTok, verb1, containerID))
		require.True(t, makeAssert(multiCtxTok, verb2, containerID))
		require.False(t, makeAssert(multiCtxTok, verb3, containerID))
	})
}

func TestToken_AssertVerb(t *testing.T) {
	containerID := cidtest.ID()
	otherContainerID := cidtest.ID()

	var tok session.Token
	ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet, session.VerbObjectPut})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx))

	require.True(t, tok.AssertVerb(session.VerbObjectGet, containerID))
	require.True(t, tok.AssertVerb(session.VerbObjectPut, containerID))
	require.False(t, tok.AssertVerb(session.VerbObjectDelete, containerID))
	require.False(t, tok.AssertVerb(session.VerbObjectGet, otherContainerID))

	ctx2, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbObjectHead})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx2))

	require.True(t, tok.AssertVerb(session.VerbObjectHead, containerID))
	require.True(t, tok.AssertVerb(session.VerbObjectHead, otherContainerID))

	testVerbAssertionCommonCases(t, containerID, session.VerbObjectGet, session.VerbObjectPut, session.VerbObjectDelete,
		func(tok session.Token, verb session.Verb, cid cid.ID) bool { return tok.AssertVerb(verb, cid) })
}

func TestToken_AssertObject(t *testing.T) {
	containerID := cidtest.ID()
	objectID1 := oidtest.ID()
	objectID2 := oidtest.ID()
	objectID3 := oidtest.ID()

	var tok session.Token

	t.Run("no specific objects", func(t *testing.T) {
		ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		require.NoError(t, tok.AddContext(ctx))

		require.True(t, tok.AssertObject(session.VerbObjectGet, containerID, objectID1))
		require.True(t, tok.AssertObject(session.VerbObjectGet, containerID, objectID2))
		require.False(t, tok.AssertObject(session.VerbObjectPut, containerID, objectID1))
	})

	t.Run("with specific objects", func(t *testing.T) {
		tok = session.Token{}
		ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		require.NoError(t, ctx.SetObjects([]oid.ID{objectID1, objectID2}))
		require.NoError(t, tok.AddContext(ctx))

		require.True(t, tok.AssertObject(session.VerbObjectGet, containerID, objectID1))
		require.True(t, tok.AssertObject(session.VerbObjectGet, containerID, objectID2))
		require.False(t, tok.AssertObject(session.VerbObjectGet, containerID, objectID3))
	})

	t.Run("container mismatch", func(t *testing.T) {
		otherContainerID := cidtest.ID()
		require.False(t, tok.AssertObject(session.VerbObjectGet, otherContainerID, objectID1))
	})

	t.Run("empty contexts", func(t *testing.T) {
		var emptyTok session.Token
		require.False(t, emptyTok.AssertObject(session.VerbObjectGet, containerID, objectID1))
	})

	t.Run("zero object ID", func(t *testing.T) {
		var zeroObjID oid.ID
		ctx, err := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		require.NoError(t, ctx.SetObjects([]oid.ID{objectID1}))
		var testTok session.Token
		require.NoError(t, testTok.AddContext(ctx))
		require.False(t, testTok.AssertObject(session.VerbObjectGet, containerID, zeroObjID))
	})

	t.Run("zero container ID wildcard", func(t *testing.T) {
		zeroCtx, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbObjectGet})
		require.NoError(t, err)
		var wildcardTok session.Token
		require.NoError(t, wildcardTok.AddContext(zeroCtx))
		require.True(t, wildcardTok.AssertObject(session.VerbObjectGet, containerID, objectID1))
		require.True(t, wildcardTok.AssertObject(session.VerbObjectGet, cidtest.ID(), objectID2))
	})

	t.Run("container verbs rejected", func(t *testing.T) {
		ctx, err := session.NewContext(containerID, []session.Verb{session.VerbContainerPut})
		require.NoError(t, err)
		var containerVerbTok session.Token
		require.NoError(t, containerVerbTok.AddContext(ctx))
		require.False(t, containerVerbTok.AssertObject(session.VerbContainerPut, containerID, objectID1))
		require.False(t, containerVerbTok.AssertObject(session.VerbContainerDelete, containerID, objectID1))
		require.False(t, containerVerbTok.AssertObject(session.VerbContainerSetEACL, containerID, objectID1))
	})
}

func TestToken_AssertContainer(t *testing.T) {
	containerID := cidtest.ID()
	otherContainerID := cidtest.ID()

	var tok session.Token
	ctx, err := session.NewContext(containerID, []session.Verb{session.VerbContainerPut, session.VerbContainerDelete})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx))

	require.True(t, tok.AssertContainer(session.VerbContainerPut, containerID))
	require.True(t, tok.AssertContainer(session.VerbContainerDelete, containerID))
	require.False(t, tok.AssertContainer(session.VerbContainerSetEACL, containerID))
	require.False(t, tok.AssertContainer(session.VerbContainerPut, otherContainerID))

	ctx2, err := session.NewContext(cid.ID{}, []session.Verb{session.VerbContainerSetEACL})
	require.NoError(t, err)
	require.NoError(t, tok.AddContext(ctx2))

	require.True(t, tok.AssertContainer(session.VerbContainerSetEACL, containerID))
	require.True(t, tok.AssertContainer(session.VerbContainerSetEACL, otherContainerID))

	t.Run("object verbs rejected", func(t *testing.T) {
		require.False(t, tok.AssertContainer(session.VerbObjectGet, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectPut, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectDelete, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectHead, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectSearch, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectRange, containerID))
		require.False(t, tok.AssertContainer(session.VerbObjectRangeHash, containerID))
	})

	testVerbAssertionCommonCases(t, containerID, session.VerbContainerPut, session.VerbContainerDelete, session.VerbContainerSetEACL,
		func(tok session.Token, verb session.Verb, cid cid.ID) bool { return tok.AssertContainer(verb, cid) })

	t.Run("mixed object and container verbs", func(t *testing.T) {
		mixedCtx, err := session.NewContext(containerID, []session.Verb{
			session.VerbObjectGet,
			session.VerbContainerPut,
		})
		require.NoError(t, err)
		var mixedTok session.Token
		require.NoError(t, mixedTok.AddContext(mixedCtx))

		require.True(t, mixedTok.AssertContainer(session.VerbContainerPut, containerID))
		require.False(t, mixedTok.AssertContainer(session.VerbObjectGet, containerID))
	})
}

func TestToken_ProtoMessage(t *testing.T) {
	require.Equal(t, validProto, validToken.ProtoMessage())
}

func TestToken_FromProtoMessage(t *testing.T) {
	var val session.Token
	require.NoError(t, val.FromProtoMessage(validProto))

	t.Run("valid", func(t *testing.T) {
		checkTokenFields(t, val)

		require.Equal(t, validToken, val)
	})

	t.Run("invalid", func(t *testing.T) {
		for _, tc := range invalidProtoTokenCommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				st := val
				m := st.ProtoMessage()
				tc.corrupt(m)
				err := new(session.Token).FromProtoMessage(m)
				require.EqualError(t, err, tc.err)
			})
		}
	})
}

func TestToken_Marshal(t *testing.T) {
	require.Equal(t, validBinToken, validToken.Marshal())
}

func TestToken_Unmarshal(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var val session.Token
		err := val.Unmarshal(validBinToken)
		require.NoError(t, err)

		checkTokenFields(t, val)
		require.Equal(t, validToken, val)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("invalid proto", func(t *testing.T) {
			err := new(session.Token).Unmarshal([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "cannot parse invalid wire-format data")
		})
		for _, tc := range invalidBinCommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Token).Unmarshal(tc.b), tc.err)
			})
		}
	})
}

func TestToken_MarshalJSON(t *testing.T) {
	b, err := validToken.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, validJSONToken, string(b))
}
func TestToken_UnmarshalJSON(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		var val session.Token
		err := val.UnmarshalJSON([]byte(validJSONToken))
		require.NoError(t, err)

		checkTokenFields(t, val)
		require.Equal(t, validToken, val)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("invalid JSON", func(t *testing.T) {
			err := new(session.Token).UnmarshalJSON([]byte("Hello, world!"))
			require.ErrorContains(t, err, "proto")
			require.ErrorContains(t, err, "syntax error")
		})
		for _, tc := range invalidJSONCommonTestcases {
			t.Run(tc.name, func(t *testing.T) {
				require.EqualError(t, new(session.Token).UnmarshalJSON([]byte(tc.j)), tc.err)
			})
		}
	})
}
