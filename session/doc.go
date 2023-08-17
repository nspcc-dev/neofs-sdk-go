/*
Package session collects functionality of the NeoFS sessions.

Sessions are used in NeoFS as a mechanism for transferring the power of attorney
of actions to another network member.

Session tokens represent proof of trust. Each session has a limited lifetime and
scope related to some NeoFS service: Object, Container, etc.

Both parties agree on a secret (private session key), the possession of which
will be authenticated by a trusted person. The principal confirms his trust by
signing the public part of the secret (public session key).

The trusted member can perform operations on behalf of the trustee.

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.
*/
package session
