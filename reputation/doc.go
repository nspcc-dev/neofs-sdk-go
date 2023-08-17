/*
Package reputation collects functionality related to the NeoFS reputation system.

The functionality is based on the system described in the NeoFS specification.

Trust type represents simple instances of trust values. PeerToPeerTrust extends
Trust to support the direction of trust, i.e. from whom to whom. GlobalTrust
is designed as a global measure of trust in a network member. See the docs
for each type for details.

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.
*/
package reputation
