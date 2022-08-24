/*
Package session collects functionality of the NeoFS sessions.

Sessions are used in NeoFS as a mechanism for transferring the power of attorney
of actions to another network member.

Session tokens represent proof of trust. Each session has a limited lifetime and
scope related to some NeoFS service: Object, Container, etc.

Both parties agree on a secret (private session key), the possession of which
will be authenticated by a trusted person. The principal confirms his trust by
signing the public part of the secret (public session key).

	var tok Container
	tok.ForVerb(VerbContainerDelete)
	tok.SetAuthKey(trustedKey)
	// ...

	err := tok.Sign(principalKey)
	// ...

	// transfer the token to a trusted party

The trusted member can perform operations on behalf of the trustee.

Instances can be also used to process NeoFS API V2 protocol messages
(see neo.fs.v2.accounting package in https://github.com/nspcc-dev/neofs-api).

On client side:

	import "github.com/nspcc-dev/neofs-api-go/v2/session"

	var msg session.Token
	tok.WriteToV2(&msg)

	// send msg

On server side:

	// recv msg

	var tok session.Container
	tok.ReadFromV2(msg)

	// process cnr

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.
*/
package session
