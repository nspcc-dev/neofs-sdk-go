# Changelog

## [1.0.0-rc.11] - 2023-09-07

We've received some valuable feedback and improved some aspects of SDK API in
this release. Most of the changes are just extensions and only a few are
incompatible with the previous RC (except for scheduled deprecated Pool API
removal). Try it out and leave feedback. We don't currently have any API
changes planned, so this might become 1.0 as is (not counting documentation
improvements that will be done).

New features:
 * SignedData method for signed structures (#513)
 * CopyTo method for deep structure copies (#512)

Behaviour changes:
 * ObjectRangeInit and ObjectRangeInit unification between Client and Pool
   (#491)
 * removal of all deprecated Pool APIs (#491)
 * ObjectHead method now returns header instead of intermediate struct (#497)
 * ObjectListReader now returns error from Read instead of bool flag
 * Object's Sign/VerifySignature methods replacing
   CalculateAndSetSignature/VerifyIDSignature, harmonizing this API with other
   structures (#513)
 * System role is ignored in the eACL matcher now (#515)

Improvements:
 * extended integration tests (#467)
 * new well-known "Version" node attribute support (#490)
 * documentation and examples (#495, #511, #507, #509)
 * exported AttributeExpirationEpoch (#504)
 * methods to work with user's container and object attributes (#505, #508)
 * exported filter keys (#506)
 * Signature constructor (#508)
 * support for creation epoch/payload size search filters (#508)

Bugs fixed:
 * failing object upload via slicer in some cases (#486)
 * io.EOF returned from payload writer on Write instead of protocol-specific
   error (#498)
 * late session token updates leading to call failures in Pool (#503)

## [1.0.0-rc.10] - 2023-08-04

This RC is supposed to be the last one having this many changes. We don't have
any API changes planned now (except deprecated APIs removal), we're just
collecting feedback (for about a month) and improving the
documentation. Depending on the feedback we may release RC11 in September with
some cosmetic changes, but then a final 1.0 will be release. So please
consider RC10 as a real stable RC and provide feedback if any.

The most significant change in this new RC is unification of pool and client
APIs. For a long time they were developed pretty much independently leading to
differences in behavior and features, now that's fixed. Old pool APIs are
available in this release to simplify transition, but will be removed before
1.0. Another important thing to highlight is a new finalized slicer API that
can now do much more than the previous one and do it easier.

New features:
 * additional APIs for data get/set in session package (#447)
 * `waiter` package that implements waiting for asynchronous operations and
   works both with Client and Pool (#451)
 * statistics collection is available via callback both for Client and Pool
   now, `stat` package contains an old Pool statistics data collector (#453)
 * automatic session creation/use can be disabled for Pool now (#454)

Behaviour changes:
 * `container` package APIs using Container struct are now methods of the
   Container struct (#435)
 * container.CalculateIDFromBinary moved to container/id package (#435)
 * Pool provides a set of methods compliant with Client now, old methods are
   deprecated (#434, #450)
 * Client.ObjectPutInit now requires an object header to be provided (but you
   can write payload right after init, #434)
 * Client.ObjectGetInit immediately returns an object header now (#434)
 * `object` package APIs using Object struct are now methods of the
   Object struct (#457)
 * simplified `Pool` creation for flat node list (#456)
 * `ns` package is gone, please refer to neofs-contracts repository for new
   RPC bindings providing this and other functions (#460)
 * `relations` interface can work with both Client and Pool now (#455)
 * Signers are explicitly required in APIs that need them (#462)
 * Pool.SyncContainerWithNetwork is gone and client.SyncContainerWithNetwork
   can be used instead (#464)
 * ReadChunk method is no longer provided by the object's PayloadReader,
   please use plain Read (#473)
 * new user.Signer interface and structures binding Signer and user ID,
   removal of IDFromSigner and IDFromKey (#475, #481, #479)
 * client's ObjectWriter is compatible with io.WriteCloser now (#466)
 * completely reworked `slicer` that can be used in one-off mode with a simple
   function call or can reuse some context for a set of objects (#466, #481)

Improvements:
 * structures are no longer serialized when signing with StaticSigner (#436)
 * fail fast in Pool init if wrong node addresses are given (#471)
 * documentation updates (#469, #473)
 * optimized session creation in Pool (#468)
 * slicer performance optimizations (#466)
 * additional methods for Signature to avoid the need to use V2 structures in
   many cases (#479)
 * PublicKeyBytes helper for `crypto` package (#479)

Bugs fixed:
 * Pool.DeleteObject not working correctly (#440)
 * missing split info in the first child object produced by `slicer` (#463)
 * split info written into root object by `slicer` (#474)
 * incorrect min aggregation result in netmap placement function (#480)
 * excessive headers produced by `slicer` (#481)

## [1.0.0-rc.9] - 2023-05-29

A number of API simplifications (with ResolveNeoFSFailures being the most
important one) goes with this version of SDK as well as a very significant
change in the form of subnet support removal. Subnets are deprecated and to be
removed from the protocol (there are zero subnet users, so it doesn't affect
anyone), therefore SDK won't support them starting with this release.

Behaviour changes:
 * IsErr* functions are removed from client, please use standard errors.Is (#407)
 * ResolveNeoFSFailures is removed from the client, all protocol errors are
   translated into Go errors and returns as errors, (that can be used with
   errors.Is or errors.As to detect NeoFS API-level problems, #413)
 * client no longer panics on invalid parameters (#414, #422)
 * RFC6979 signer is mandatory for container-specific operations, so it's
   also mandatory for default client/pool use (but can be overridden for
   specific operation if needed, #418)
 * apistatus package now provides ErrorFromV2/ErrorToV2 functions to translate
   v2 API statuses into/from errors, IsSuccessful/ErrToStatus/ErrFromStatus
   functions are gone (and not needed after ResolveNeoFSFailures change, #419)
 * removal of subnet support (#420)
 * Client methods now accept mandatory paramerters as explicit arguments (not
   hidden in Prm* structures, #424, #425)
 * Client methods now return plain resulting structures where possible (#426)

Improvements:
 * oid.Address can be passed directly into PrmObject* structs (#408)
 * object.Range can be passed directly into the PrmObjectRange (#408)
 * proper handling of protocol violations in the client (#410)
 * updated tzhash dependency (#415)
 * better error handling for some conversions (#423)

Bugs fixed:
 * broken object metadata produced by slicer code in some cases (#427)

## [1.0.0-rc.8] - 2023-04-27

We've reworked the way we handle keys in the SDK and added an experimental API
simplifying object creation. Pools were also improved and a number of other
minor enhancements are included in this release as well. Try it in your
applications and leave your feedback in issues, we need it to make 1.0.0 a
really good release for all NeoFS use cases.

New features:
 * object/relations package for large (splitted) object data handling (#360, #348)
 * pool can now accept stream operations timeout parameter (#364)
 * pool can also accept callbacks to be invoked after every client response (#365)
 * additional marshaling/unmarshaling methods for accounting.Decimal,
   netmap.NetworkInfo and version.Version (#378)
 * it's possible to get a single client from a pool now (#388)
 * object.SearchFilters now provide more convenient APIs to search for object
   by their hashes (#386)
 * experimental client-side API for object creation (and slicing large inputs
   into split objects, #382)

Behaviour changes:
 * ns.ResolveContainerName changed to ns.ResolveContainerName with more
   specific input type (#356)
 * pool.ErrPoolClientUnhealthy is no longer exported (#358)
 * deprecated object.RawObject type is gone (#387)
 * pools no longer use default session tokens for read operations (they're not
   required technically, #363)
 * Go 1.18+ is required now (#394)
 * enumeration types in eacl and object packages now have
   EncodeToString/DecodeString methods replacing String/FromString for
   serialization; String is still available, but it's not guaranteed to be
   compatible across versions (#393)
 * signing APIs no longer require a specific key, it can be provided in a
   Signer wrapper (with different schemes) or a crypto.StaticSigner can be
   used for signatures calculated outside of the SDK (#398, #401)
 * (NetMap).PlacementVectors and (NetMap).ContainerNodes APIs now use more
   specific types simplifying their use (#396)

Improvements:
 * pool can work with some nodes being inaccessible now (#358)
 * unified error messages in the client/status package (#369)
 * documentation updates (#370, #392, #396)
 * updated NeoGo, ANTLR, zap, golang-lru dependencies (#372)
 * (*Client).ContainerSetEACL method now checks CID internally before
   generating a request to send to the network (#389)
 * neofs-api-go signature package is no longer required (#401)

Bugs fixed:
 * pool.DeleteObject now correctly handles large (splitted) objects (#360)
 * object structure problems are no longer considered as connection problems
   by the pool (#374)
 * private key used as a part of a map key in the pool (#401)

## pre-v1.0.0-rc.8

See git log.

[1.0.0-rc.11]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.10...v1.0.0-rc.11
[1.0.0-rc.10]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.9...v1.0.0-rc.10
[1.0.0-rc.9]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.8...v1.0.0-rc.9
[1.0.0-rc.8]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.7...v1.0.0-rc.8
