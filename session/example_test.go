package session_test

import (
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

// Both parties agree on a secret (private session key), the possession of which
// will be authenticated by a trusted person. The principal confirms his trust by
// signing the public part of the secret (public session key).
func ExampleContainer() {
	// import neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	// import "github.com/nspcc-dev/neofs-sdk-go/user"
	// import cid "github.com/nspcc-dev/neofs-sdk-go/container/id"

	// you private key/signer, to prove you are you
	var principalSigner user.Signer
	// trusted party, who can do action on behalf of you. Usually the key maybe taken from Client.SessionCreate.
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

// Instances can be also used to process NeoFS API V2 protocol messages with [https://github.com/nspcc-dev/neofs-api] package.
func ExampleObject_marshalling() {
	// On the client side.

	var tok session.Object

	msg := tok.ProtoMessage()
	// *send message*

	// On the server side.

	_ = tok.FromProtoMessage(msg)
}
