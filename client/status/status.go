package apistatus

// Status defines a variety of NeoFS API status returns.
//
// All statuses are split into two disjoint subsets: successful and failed, and:
//   - statuses that implement the build-in error interface are considered failed statuses;
//   - all other value types are considered successes (nil is a default success).
//
// In Go code type of success can be determined by a type switch, failure - by a switch with [errors.As] calls.
// Nil should be considered as a success, and default switch section - as an unrecognized Status.
//
// It should be noted that using direct typecasting is not a compatible approach.
//
// To transport statuses using the NeoFS API V2 protocol, see [StatusV2] interface and [FromStatusV2] and [ToStatusV2] functions.
type Status any
