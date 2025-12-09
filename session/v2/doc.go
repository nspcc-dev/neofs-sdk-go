/*
Package session collects functionality of the NeoFS session token v2.

[Token] is an enhanced version supporting multiple subjects, unified contexts
for both object and container operations, delegation chains, and NNS name resolution.

V2 tokens use unified [Verb] constants for operations.

Key differences from V1:
- V2 has unified [Context]
- V2 supports multiple subjects
- V2 adds explicit delegation chain tracking
- V2 supports NNS names via [NewTargetNamed]
- V2 allows multiple contexts in a single token
- V2 uses [time.Time] for lifetime claims
*/
package session
