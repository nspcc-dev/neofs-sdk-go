package session_test

import (
	"math/rand"
	"testing"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	prototest "github.com/nspcc-dev/neofs-sdk-go/proto/internal/test"
	"github.com/nspcc-dev/neofs-sdk-go/proto/refs"
	"github.com/nspcc-dev/neofs-sdk-go/proto/session"
	"github.com/stretchr/testify/require"
)

func randTarget() *session.Target {
	v := &session.Target{}
	switch rand.Uint32() % 2 {
	case 0:
		v.Identifier = &session.Target_OwnerId{OwnerId: prototest.RandOwnerID()}
	case 1:
		v.Identifier = &session.Target_NnsName{NnsName: prototest.RandString()}
	}
	return v
}

func randTargets() []*session.Target { return prototest.RandRepeated(randTarget) }

func randVerbs() []session.Verb {
	verbs := make([]session.Verb, 1+rand.Uint32()%5)
	for i := range verbs {
		verbs[i] = prototest.RandInteger[session.Verb]()
	}
	return verbs
}

func randDelegationInfo() *session.DelegationInfo {
	return &session.DelegationInfo{
		Issuer:    randTarget(),
		Subjects:  randTargets(),
		Lifetime:  prototest.RandSessionTokenLifetime(),
		Verbs:     randVerbs(),
		Signature: prototest.RandSignature(),
	}
}

func randDelegationInfos() []*session.DelegationInfo {
	return prototest.RandRepeated(randDelegationInfo)
}

func randSessionContextV2() *session.SessionContextV2 {
	return &session.SessionContextV2{
		Container: prototest.RandContainerID(),
		Objects:   prototest.RandObjectIDs(),
		Verbs:     randVerbs(),
	}
}

func randSessionContextV2s() []*session.SessionContextV2 {
	return prototest.RandRepeated(randSessionContextV2)
}

func randSessionTokenV2Body() *session.SessionTokenV2_Body {
	return &session.SessionTokenV2_Body{
		Version:  prototest.RandUint32(),
		Id:       prototest.RandBytes(),
		Issuer:   randTarget(),
		Subjects: randTargets(),
		Lifetime: prototest.RandSessionTokenLifetime(),
		Contexts: randSessionContextV2s(),
	}
}

func randSessionTokenV2() *session.SessionTokenV2 {
	return &session.SessionTokenV2{
		Body:            randSessionTokenV2Body(),
		Signature:       prototest.RandSignature(),
		DelegationChain: randDelegationInfos(),
	}
}

func TestTarget_MarshalStable(t *testing.T) {
	t.Run("with OwnerId", func(t *testing.T) {
		src := &session.Target{
			Identifier: &session.Target_OwnerId{
				OwnerId: prototest.RandOwnerID(),
			},
		}

		var dst session.Target
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))
		require.NotNil(t, dst.GetOwnerId())
		require.Equal(t, src.GetOwnerId().GetValue(), dst.GetOwnerId().GetValue())
	})

	t.Run("with NnsName", func(t *testing.T) {
		src := &session.Target{
			Identifier: &session.Target_NnsName{
				NnsName: prototest.RandString(),
			},
		}

		var dst session.Target
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))
		require.Equal(t, src.GetNnsName(), dst.GetNnsName())
	})

	t.Run("nil identifier", func(t *testing.T) {
		src := &session.Target{}

		var dst session.Target
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))
	})

	t.Run("nil target", func(t *testing.T) {
		var src *session.Target
		data := neofsproto.MarshalMessage(src)
		require.Empty(t, data)
	})

	prototest.TestMarshalStable(t, []*session.Target{
		randTarget(),
		{Identifier: &session.Target_OwnerId{OwnerId: prototest.RandOwnerID()}},
		{Identifier: &session.Target_NnsName{NnsName: prototest.RandString()}},
	})
}

func TestDelegationInfo_MarshalStable(t *testing.T) {
	t.Run("nil in repeated strings", func(t *testing.T) {
		src := &session.DelegationInfo{
			Verbs: []session.Verb{0, 1, 2},
		}

		var dst session.DelegationInfo
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		verbs := dst.GetVerbs()
		require.Len(t, verbs, 3)
		require.EqualValues(t, 0, verbs[0])
		require.EqualValues(t, 1, verbs[1])
		require.EqualValues(t, 2, verbs[2])
	})

	t.Run("nil delegationInfo", func(t *testing.T) {
		var src *session.DelegationInfo
		data := neofsproto.MarshalMessage(src)
		require.Empty(t, data)
	})

	prototest.TestMarshalStable(t, []*session.DelegationInfo{
		randDelegationInfo(),
		{
			Issuer:    randTarget(),
			Subjects:  randTargets(),
			Lifetime:  prototest.RandSessionTokenLifetime(),
			Verbs:     randVerbs(),
			Signature: prototest.RandSignature(),
		},
	})
}

func TestSessionContextV2_MarshalStable(t *testing.T) {
	t.Run("nil in repeated objects", func(t *testing.T) {
		src := &session.SessionContextV2{
			Objects: []*refs.ObjectID{nil, {}},
		}

		var dst session.SessionContextV2
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		objects := dst.GetObjects()
		require.Len(t, objects, 2)
		require.Equal(t, objects[0], new(refs.ObjectID))
		require.Equal(t, objects[1], new(refs.ObjectID))
	})

	t.Run("empty verbs", func(t *testing.T) {
		src := &session.SessionContextV2{
			Verbs:     []session.Verb{},
			Container: prototest.RandContainerID(),
		}

		var dst session.SessionContextV2
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		verbs := dst.GetVerbs()
		require.Empty(t, verbs)
	})

	t.Run("nil sessionContextV2", func(t *testing.T) {
		var src *session.SessionContextV2
		data := neofsproto.MarshalMessage(src)
		require.Empty(t, data)
	})

	prototest.TestMarshalStable(t, []*session.SessionContextV2{
		randSessionContextV2(),
		{
			Container: prototest.RandContainerID(),
			Objects:   prototest.RandObjectIDs(),
			Verbs: []session.Verb{
				session.Verb_VERB_UNSPECIFIED,
				session.Verb_OBJECT_PUT,
				session.Verb_OBJECT_GET,
			},
		},
	})
}

func TestSessionTokenV2_Body_MarshalStable(t *testing.T) {
	t.Run("nil in repeated subjects", func(t *testing.T) {
		src := &session.SessionTokenV2_Body{
			Subjects: []*session.Target{nil, {}},
		}

		var dst session.SessionTokenV2_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		subjects := dst.GetSubjects()
		require.Len(t, subjects, 2)
		require.Equal(t, subjects[0], new(session.Target))
		require.Equal(t, subjects[1], new(session.Target))
	})

	t.Run("nil in repeated contexts", func(t *testing.T) {
		src := &session.SessionTokenV2_Body{
			Contexts: []*session.SessionContextV2{nil, {}},
		}

		var dst session.SessionTokenV2_Body
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		contexts := dst.GetContexts()
		require.Len(t, contexts, 2)
		require.Equal(t, contexts[0], new(session.SessionContextV2))
		require.Equal(t, contexts[1], new(session.SessionContextV2))
	})

	t.Run("nil body", func(t *testing.T) {
		var src *session.SessionTokenV2_Body
		data := neofsproto.MarshalMessage(src)
		require.Empty(t, data)
	})

	prototest.TestMarshalStable(t, []*session.SessionTokenV2_Body{
		randSessionTokenV2Body(),
		{
			Version:  1,
			Id:       prototest.RandBytes(),
			Issuer:   randTarget(),
			Subjects: randTargets(),
			Lifetime: prototest.RandSessionTokenLifetime(),
			Contexts: randSessionContextV2s(),
		},
	})
}

func TestSessionTokenV2_MarshalStable(t *testing.T) {
	t.Run("nil in delegation chain", func(t *testing.T) {
		src := &session.SessionTokenV2{
			DelegationChain: []*session.DelegationInfo{nil, {}},
		}

		var dst session.SessionTokenV2
		require.NoError(t, neofsproto.UnmarshalMessage(neofsproto.MarshalMessage(src), &dst))

		chain := dst.GetDelegationChain()
		require.Len(t, chain, 2)
		require.Equal(t, chain[0], new(session.DelegationInfo))
		require.Equal(t, chain[1], new(session.DelegationInfo))
	})

	t.Run("nil token", func(t *testing.T) {
		var src *session.SessionTokenV2
		data := neofsproto.MarshalMessage(src)
		require.Empty(t, data)
	})

	prototest.TestMarshalStable(t, []*session.SessionTokenV2{
		randSessionTokenV2(),
		{
			Body:            randSessionTokenV2Body(),
			Signature:       prototest.RandSignature(),
			DelegationChain: randDelegationInfos(),
		},
	})
}

func TestDelegationInfo_MarshaledSize(t *testing.T) {
	tests := []*session.DelegationInfo{
		nil,
		{},
		randDelegationInfo(),
		{
			Issuer:    randTarget(),
			Subjects:  randTargets(),
			Lifetime:  prototest.RandSessionTokenLifetime(),
			Verbs:     randVerbs(),
			Signature: prototest.RandSignature(),
		},
	}

	for _, test := range tests {
		size := test.MarshaledSize()
		data := neofsproto.MarshalMessage(test)
		require.Equal(t, size, len(data), "MarshaledSize should match actual marshaled data length")
	}
}

func TestTarget_MarshaledSize(t *testing.T) {
	tests := []*session.Target{
		nil,
		{},
		randTarget(),
		{Identifier: &session.Target_OwnerId{OwnerId: prototest.RandOwnerID()}},
		{Identifier: &session.Target_NnsName{NnsName: "example.neofs"}},
	}

	for _, test := range tests {
		size := test.MarshaledSize()
		data := neofsproto.MarshalMessage(test)
		require.Equal(t, size, len(data), "MarshaledSize should match actual marshaled data length")
	}
}

func TestSessionContextV2_MarshaledSize(t *testing.T) {
	tests := []*session.SessionContextV2{
		nil,
		{},
		randSessionContextV2(),
		{
			Verbs:   []session.Verb{session.Verb_OBJECT_PUT, session.Verb_OBJECT_GET},
			Objects: prototest.RandObjectIDs(),
			Container: &refs.ContainerID{
				Value: []byte{1, 2, 3, 4, 5},
			},
		},
		{
			Verbs:     []session.Verb{},
			Objects:   []*refs.ObjectID{},
			Container: &refs.ContainerID{},
		},
	}

	for _, test := range tests {
		size := test.MarshaledSize()
		data := neofsproto.MarshalMessage(test)
		require.Equal(t, size, len(data), "MarshaledSize should match actual marshaled data length")
	}
}

func TestSessionTokenV2_Body_MarshaledSize(t *testing.T) {
	tests := []*session.SessionTokenV2_Body{
		nil,
		{},
		randSessionTokenV2Body(),
		{
			Version:  1,
			Id:       []byte{1, 2, 3, 4},
			Issuer:   randTarget(),
			Subjects: randTargets(),
			Lifetime: prototest.RandSessionTokenLifetime(),
			Contexts: randSessionContextV2s(),
		},
	}

	for _, test := range tests {
		size := test.MarshaledSize()
		data := neofsproto.MarshalMessage(test)
		require.Equal(t, size, len(data), "MarshaledSize should match actual marshaled data length")
	}
}

func TestSessionTokenV2_MarshaledSize(t *testing.T) {
	tests := []*session.SessionTokenV2{
		nil,
		{},
		randSessionTokenV2(),
		{
			Body:            randSessionTokenV2Body(),
			Signature:       prototest.RandSignature(),
			DelegationChain: randDelegationInfos(),
		},
	}

	for _, test := range tests {
		size := test.MarshaledSize()
		data := neofsproto.MarshalMessage(test)
		require.Equal(t, size, len(data), "MarshaledSize should match actual marshaled data length")
	}
}

func TestDelegationInfo_RoundTrip(t *testing.T) {
	original := randDelegationInfo()

	data := neofsproto.MarshalMessage(original)

	var decoded session.DelegationInfo
	require.NoError(t, neofsproto.UnmarshalMessage(data, &decoded))

	require.Equal(t, original.GetLifetime(), decoded.GetLifetime())
	require.Equal(t, original.GetVerbs(), decoded.GetVerbs())
}

func TestSessionTokenV2_RoundTrip(t *testing.T) {
	original := randSessionTokenV2()

	data := neofsproto.MarshalMessage(original)

	var decoded session.SessionTokenV2
	require.NoError(t, neofsproto.UnmarshalMessage(data, &decoded))

	require.Equal(t, original.GetBody().GetVersion(), decoded.GetBody().GetVersion())
	require.Equal(t, original.GetBody().GetId(), decoded.GetBody().GetId())
}
