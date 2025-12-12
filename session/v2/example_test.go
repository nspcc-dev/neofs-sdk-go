package session_test

import (
	"fmt"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session/v2"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// ExampleToken_basic demonstrates basic Token creation and usage.
func ExampleToken_basic() {
	var token session.Token

	// Set basic token properties
	token.SetNonce(session.RandomNonce())
	token.SetVersion(session.TokenCurrentVersion)

	// Account creating this token
	signer := usertest.User()

	// Add authorized subjects
	subject := session.NewTargetUser(usertest.ID())
	_ = token.AddSubject(subject)

	// Set validity period
	token.SetIat(100)  // issued at epoch 100
	token.SetNbf(100)  // not valid before epoch 100
	token.SetExp(1000) // expires at epoch 1000

	// Add authorization context
	containerID := cidtest.ID()
	ctx, _ := session.NewContext(containerID, []session.Verb{
		session.VerbObjectGet,
		session.VerbObjectPut,
	})
	_ = token.AddContext(ctx)

	// Sign the token with the issuer's key
	_ = token.Sign(signer)

	// Verify the signature
	fmt.Println("Token valid:", token.VerifySignature())

	// Output:
	// Token valid: true
}

// ExampleToken_nns demonstrates using NNS names as targets.
func ExampleToken_nns() {
	var token session.Token

	token.SetNonce(session.RandomNonce())
	token.SetVersion(session.TokenCurrentVersion)

	// Add subjects using NNS names
	friends := session.NewTargetNamed("friends.neo")
	_ = token.AddSubject(friends)

	token.SetIat(100)
	token.SetNbf(100)
	token.SetExp(1000)

	ctx, _ := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
	_ = token.AddContext(ctx)

	// Sign the token using the user
	signer := usertest.User()
	_ = token.Sign(signer)

	// Check NNS target properties
	fmt.Println("Subject 0 is NNS:", token.Subjects()[0].IsNNS())
	fmt.Println("Subject 0 name:", token.Subjects()[0].NNSName())

	aliceUser := usertest.ID()
	bobUser := usertest.ID()
	hasUserInNNS := func(name string, id user.ID) bool {
		switch name {
		case "friends.neo":
			return id == aliceUser || id == bobUser
		default:
			return false
		}
	}

	fmt.Println("Is Alice have authority: ", token.AssertAuthority(aliceUser, hasUserInNNS))
	fmt.Println("Is Bob have authority: ", token.AssertAuthority(bobUser, hasUserInNNS))
	mikeUser := usertest.ID()
	fmt.Println("Is Mike have authority: ", token.AssertAuthority(mikeUser, hasUserInNNS))

	// Output:
	// Subject 0 is NNS: true
	// Subject 0 name: friends.neo
	// Is Alice have authority:  true
	// Is Bob have authority:  true
	// Is Mike have authority:  false
}

// ExampleToken_delegation demonstrates delegation chain.
func ExampleToken_delegation() {
	var origin session.Token

	origin.SetNonce(session.RandomNonce())
	origin.SetVersion(session.TokenCurrentVersion)

	subject1 := usertest.User()
	_ = origin.AddSubject(session.NewTargetUser(subject1.UserID()))

	origin.SetIat(100)
	origin.SetNbf(100)
	origin.SetExp(1000)
	ctx, _ := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
	_ = origin.AddContext(ctx)

	signer := usertest.User()
	_ = origin.Sign(signer)

	// Create delegation: the subject1 delegates to subject2
	subject2 := usertest.ID()

	var del session.Token
	del.SetNonce(session.RandomNonce())
	del.SetVersion(session.TokenCurrentVersion)
	del.SetIat(100)
	del.SetNbf(100)
	del.SetExp(1000)
	_ = del.AddContext(ctx)
	_ = del.AddSubject(session.NewTargetUser(subject2))

	del.SetOrigin(&origin)
	_ = del.Sign(subject1)

	// Check authority - only subject3 should be authorized because it is the direct subject of the delegation token
	fmt.Println("Issuer authorized:", del.AssertAuthority(origin.Issuer(), nil))
	fmt.Println("Subject1 authorized:", del.AssertAuthority(subject1.UserID(), nil))
	fmt.Println("Subject2 authorized:", del.AssertAuthority(subject2, nil))
	subject3 := usertest.ID()
	fmt.Println("Subject3 authorized:", del.AssertAuthority(subject3, nil))

	// Output:
	// Issuer authorized: false
	// Subject1 authorized: false
	// Subject2 authorized: true
	// Subject3 authorized: false
}

// ExampleToken_objectRestrictions demonstrates restricting access to specific objects.
func ExampleToken_objectRestrictions() {
	var token session.Token

	token.SetNonce(session.RandomNonce())
	token.SetVersion(session.TokenCurrentVersion)
	_ = token.AddSubject(session.NewTargetUser(usertest.ID()))
	token.SetIat(100)
	token.SetNbf(100)
	token.SetExp(1000)

	containerID := cidtest.ID()
	ctx, _ := session.NewContext(containerID, []session.Verb{session.VerbObjectGet})

	// Restrict to specific objects
	obj1 := oidtest.ID()
	obj2 := oidtest.ID()
	_ = ctx.SetObjects([]oid.ID{obj1, obj2})

	_ = token.AddContext(ctx)

	signer := usertest.User()
	_ = token.Sign(signer)

	// Check object access
	fmt.Println("Object 1 allowed:", token.AssertObject(session.VerbObjectGet, containerID, obj1))
	fmt.Println("Object 2 allowed:", token.AssertObject(session.VerbObjectGet, containerID, obj2))
	fmt.Println("Other object allowed:", token.AssertObject(session.VerbObjectGet, containerID, oidtest.ID()))

	// Output:
	// Object 1 allowed: true
	// Object 2 allowed: true
	// Other object allowed: false
}
