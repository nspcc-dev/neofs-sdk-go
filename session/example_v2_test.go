package session_test

import (
	"fmt"

	"github.com/google/uuid"
	cidtest "github.com/nspcc-dev/neofs-sdk-go/container/id/test"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	oidtest "github.com/nspcc-dev/neofs-sdk-go/object/id/test"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	usertest "github.com/nspcc-dev/neofs-sdk-go/user/test"
)

// ExampleTokenV2_basic demonstrates basic TokenV2 creation and usage.
func ExampleTokenV2_basic() {
	var token session.TokenV2

	// Set basic token properties
	token.SetID(uuid.New())
	token.SetVersion(session.TokenV2CurrentVersion)

	// Account creating this token
	signer := usertest.User()
	// Set issuer, but it is not needed to do,
	// as Sign method sets it automatically
	issuer := session.NewTarget(signer.UserID())
	token.SetIssuer(issuer)

	// Add authorized subjects
	subject := session.NewTarget(usertest.ID())
	token.AddSubject(subject)

	// Set validity period
	token.SetIat(100)  // issued at epoch 100
	token.SetNbf(100)  // not valid before epoch 100
	token.SetExp(1000) // expires at epoch 1000

	// Add authorization context
	containerID := cidtest.ID()
	ctx := session.NewContextV2(containerID, []session.VerbV2{
		session.VerbV2ObjectGet,
		session.VerbV2ObjectPut,
	})
	token.AddContext(ctx)

	// Sign the token with the issuer's key
	_ = token.Sign(signer)

	// Verify the signature
	fmt.Println("Token valid:", token.VerifySignature())

	// Output:
	// Token valid: true
}

// ExampleTokenV2_nns demonstrates using NNS names as targets.
func ExampleTokenV2_nns() {
	var token session.TokenV2

	token.SetID(uuid.New())
	token.SetVersion(session.TokenV2CurrentVersion)

	// Use NNS name for issuer
	issuer := session.NewTargetFromNNS("issuer.neo")
	token.SetIssuer(issuer)

	// Add subjects using NNS names
	alice := session.NewTargetFromNNS("alice.neo")
	bob := session.NewTargetFromNNS("bob.neo")
	token.AddSubject(alice)
	token.AddSubject(bob)

	token.SetIat(100)
	token.SetNbf(100)
	token.SetExp(1000)

	ctx := session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet})
	token.AddContext(ctx)

	// Sign the token using the user that matches the nns name
	signer := usertest.User()
	_ = token.Sign(signer)

	// Check NNS target properties
	fmt.Println("Issuer is NNS:", issuer.IsNNS())
	fmt.Println("Issuer name:", issuer.NNSName())

	// Output:
	// Issuer is NNS: true
	// Issuer name: issuer.neo
}

// ExampleTokenV2_delegation demonstrates delegation chain.
func ExampleTokenV2_delegation() {
	var token session.TokenV2

	token.SetID(uuid.New())
	token.SetVersion(session.TokenV2CurrentVersion)

	// Create signer - this will be the token issuer
	signer := usertest.User()

	// Add subject1 - direct subject of the token
	subject1 := session.NewTarget(usertest.ID())
	token.AddSubject(subject1)

	token.SetIat(100)
	token.SetNbf(100)
	token.SetExp(1000)

	ctx := session.NewContextV2(cidtest.ID(), []session.VerbV2{session.VerbV2ObjectGet})
	token.AddContext(ctx)

	// Create delegation: the signer delegates to subject2
	intermediate := usertest.User()
	subject2 := session.NewTarget(intermediate.UserID())
	delegation := session.NewDelegationInfo(
		[]session.Target{subject2},
		session.NewLifetime(150, 150, 900),
		[]session.VerbV2{session.VerbV2ObjectGet},
	)
	// Add subject2 to token subjects because it must be direct subject
	token.AddSubject(subject2)

	// Sign the delegation - this sets the issuer to signer's UserID
	_ = delegation.Sign(signer)

	// Add delegation to token
	token.AddDelegation(delegation)

	// Create another delegation: subject2 delegates to subject3
	subject3 := session.NewTarget(usertest.ID())
	delegation2 := session.NewDelegationInfo(
		[]session.Target{subject3},
		session.NewLifetime(200, 200, 800),
		[]session.VerbV2{session.VerbV2ObjectGet},
	)

	// Sign with intermediate user that is subject2
	_ = delegation2.Sign(intermediate)

	token.AddDelegation(delegation2)

	// Sign the token - this sets the token issuer to signer's UserID
	_ = token.Sign(signer)

	// Check authority - issuer, subject2 and subject3 should be authorized
	fmt.Println("Issuer authorized:", token.AssertAuthority(token.Issuer()))
	fmt.Println("Subject1 authorized:", token.AssertAuthority(subject1))
	fmt.Println("Subject2 authorized:", token.AssertAuthority(subject2))
	fmt.Println("Subject3 authorized:", token.AssertAuthority(subject3))

	// Output:
	// Issuer authorized: false
	// Subject1 authorized: true
	// Subject2 authorized: true
	// Subject3 authorized: true
}

// ExampleTokenV2_objectRestrictions demonstrates restricting access to specific objects.
func ExampleTokenV2_objectRestrictions() {
	var token session.TokenV2

	token.SetID(uuid.New())
	token.SetVersion(session.TokenV2CurrentVersion)
	token.SetIssuer(session.NewTarget(usertest.ID()))
	token.AddSubject(session.NewTarget(usertest.ID()))
	token.SetIat(100)
	token.SetNbf(100)
	token.SetExp(1000)

	containerID := cidtest.ID()
	ctx := session.NewContextV2(containerID, []session.VerbV2{session.VerbV2ObjectGet})

	// Restrict to specific objects
	obj1 := oidtest.ID()
	obj2 := oidtest.ID()
	ctx.SetObjects([]oid.ID{obj1, obj2})

	token.AddContext(ctx)

	signer := usertest.User()
	_ = token.Sign(signer)

	// Check object access
	fmt.Println("Object 1 allowed:", token.AssertObject(session.VerbV2ObjectGet, containerID, obj1))
	fmt.Println("Object 2 allowed:", token.AssertObject(session.VerbV2ObjectGet, containerID, obj2))
	fmt.Println("Other object allowed:", token.AssertObject(session.VerbV2ObjectGet, containerID, oidtest.ID()))

	// Output:
	// Object 1 allowed: true
	// Object 2 allowed: true
	// Other object allowed: false
}
