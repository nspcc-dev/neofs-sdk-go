package session_test

import (
	"fmt"
	"time"

	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
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
	now := time.Now()
	token.SetIat(now)                     // issued at current time
	token.SetNbf(now)                     // not valid before current time
	token.SetExp(now.Add(24 * time.Hour)) // expires in 24 hours

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

	now := time.Now()
	token.SetIat(now)
	token.SetNbf(now)
	token.SetExp(now.Add(24 * time.Hour))

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

	resolver := nnsResolver{
		res: map[string][]user.ID{
			"friends.neo": {aliceUser, bobUser},
		},
	}

	aliceAssert, _ := token.AssertAuthority(aliceUser, resolver)
	fmt.Println("Is Alice have authority: ", aliceAssert)
	bobAssert, _ := token.AssertAuthority(bobUser, resolver)
	fmt.Println("Is Bob have authority: ", bobAssert)
	mikeUser := usertest.ID()
	mikeAssert, _ := token.AssertAuthority(mikeUser, resolver)
	fmt.Println("Is Mike have authority: ", mikeAssert)

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

	now := time.Now()
	origin.SetIat(now)
	origin.SetNbf(now)
	origin.SetExp(now.Add(24 * time.Hour))
	ctx, _ := session.NewContext(cidtest.ID(), []session.Verb{session.VerbObjectGet})
	_ = origin.AddContext(ctx)

	signer := usertest.User()
	_ = origin.Sign(signer)

	// Issuer delegates to subject1
	subject2 := usertest.ID()

	var del session.Token
	del.SetNonce(session.RandomNonce())
	del.SetVersion(session.TokenCurrentVersion)
	del.SetIat(now)
	del.SetNbf(now)
	del.SetExp(now.Add(24 * time.Hour))
	_ = del.AddContext(ctx)
	_ = del.AddSubject(session.NewTargetUser(subject2))

	del.SetOrigin(&origin)
	_ = del.Sign(subject1)

	// Check authority - only subject2 should be authorized because it is the direct subject of the delegation token
	issuerAssert, _ := del.AssertAuthority(origin.Issuer(), nil)
	subject1Assert, _ := del.AssertAuthority(subject1.UserID(), nil)
	subject2Assert, _ := del.AssertAuthority(subject2, nil)
	subject3 := usertest.ID()
	subject3Assert, _ := del.AssertAuthority(subject3, nil)

	fmt.Println("Issuer authorized:", issuerAssert)
	fmt.Println("Subject1 authorized:", subject1Assert)
	fmt.Println("Subject2 authorized:", subject2Assert)
	fmt.Println("Subject3 authorized:", subject3Assert)

	// Output:
	// Issuer authorized: false
	// Subject1 authorized: false
	// Subject2 authorized: true
	// Subject3 authorized: false
}
