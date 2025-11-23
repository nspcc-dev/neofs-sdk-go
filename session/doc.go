/*
Package session collects functionality of the NeoFS sessions.

Sessions are used in NeoFS as a mechanism for transferring the power of attorney
of actions to another network member.

Session tokens represent proof of trust. Each session has a limited lifetime and
scope related to some NeoFS service: [Object], [Container], etc.

Both parties agree on a secret (private session key), the possession of which
will be authenticated by a trusted person. The principal confirms his trust by
signing the public part of the secret (public session key).

The trusted member can perform operations on behalf of the trustee.

# Session Token V2

[TokenV2] is an enhanced version supporting multiple subjects, unified contexts
for both object and container operations, delegation chains, and NNS name resolution.

V2 tokens use unified [VerbV2] constants for operations. Use [VerbV2.IsObjectVerb]
and [VerbV2.IsContainerVerb] to distinguish operation types.

Key differences from V1:
  - V1 has separate [Object] and [Container] types; V2 has unified [ContextV2]
  - V1 authorizes single subject; V2 supports multiple subjects
  - V2 adds explicit delegation chain tracking
  - V2 supports NNS names via [NewTargetFromNNS]
  - V2 allows multiple contexts in a single token
*/
package session
