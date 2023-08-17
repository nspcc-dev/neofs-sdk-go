package session_test

import (
	apiGoSession "github.com/nspcc-dev/neofs-api-go/v2/session"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Both parties agree on a secret (private session key), the possession of which
// will be authenticated by a trusted person. The principal confirms his trust by
// signing the public part of the secret (public session key).
func ExampleContainer_Sign() {
	// import neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	// import "github.com/nspcc-dev/neofs-sdk-go/user"
	// import cid "github.com/nspcc-dev/neofs-sdk-go/container/id"

	// you private key/signer, to prove you are you
	var principalSigner user.Signer
	// trusted party, who can do action on behalf of you
	var trustedPubKey neofscrypto.PublicKey
	var cnr cid.ID

	var tok session.Object
	tok.ForVerb(session.VerbObjectPut)
	tok.SetAuthKey(trustedPubKey)
	tok.BindContainer(cnr)
	// ...

	_ = tok.Sign(principalSigner)

	// transfer the token to a trusted party
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api) on client side.
func ExampleObject_WriteToV2() {
	// import apiGoSession "github.com/nspcc-dev/neofs-api-go/v2/session"

	var tok session.Object
	var msg apiGoSession.Token

	tok.WriteToV2(&msg)

	// send msg
}

// Instances can be also used to process NeoFS API V2 protocol messages
// (see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api) on server side.
func ExampleObject_ReadFromV2() {
	// import apiGoSession "github.com/nspcc-dev/neofs-api-go/v2/session"

	// recv msg

	var tok session.Object
	var msg apiGoSession.Token

	_ = tok.ReadFromV2(msg)

	// process cnr
}
