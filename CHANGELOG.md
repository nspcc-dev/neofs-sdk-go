# Changelog

## [1.0.0-rc.17] - 2026-02-04

This version is compatible with API 2.21, it introduces session tokens v2,
container attribute management API with a number of new well-known names and
synchronous container operations.

New features:
 * Container attibute management and new well-known attributes (#758, #763, #767, #770, #772)
 * Session token v2 (#750, #764, #768, #769, #762, #771, #773, #774, #775)
 * Three-way comparison methods for OID, CID and Address (#766)

Behaviour changes:
 * Synchronous container API (#759)
 * pool uses sliding window for error counting (#765)
 * object.New is more tailored to create new objects, InitCreation is deleted (#766)

Improvements:
 * Updated NeoGo dependency to 0.116.0 (#754, #776)
 * Updated golang.org/x/crypto dependency from 0.42.0 to 0.45.0 (#756)

Bugs fixed:
 * Incorrect Incomplete status handling (#755)

## [1.0.0-rc.16] - 2025-11-14

This version is compatible with API 2.20, the most important change is erasure
coding support that has experimental status for now and APIs can be updated
if it requires some fixes.

New features:
 * EC container placement policies (#736)

Improvements:
 * MaxSearchObjectsCount constant (#745)
 * NodeInfo.Capacity method (#746)
 * Better documentation for NodeInfo.Capacity and NodeInfo.Price (#747)
 * Object PUT and GET examples (#752)

Bugs fixed:
 * Potential panic in NodeInfo.Capacity and NodeInfo.Price (#747)

## [1.0.0-rc.15] - 2025-09-26

This version makes SDK compatible with API 2.19, removes all deprecated APIs
and brings updates to use Go 1.24+.

New features:
 * QuotaExceeded API status (#734)
 * pool.ErrUnhealthy error (#737)
 * Incomplete API status (#742)
 * BadRequest API status (#742)
 * Busy API status (#742)

Behaviour changes:
 * Removed P2P estimations API which is deprecated and unused (#732)
 * Removed all previously deprecated APIs (#738)

Improvements:
 * Go 1.24+ is required now (#743)
 * Updated NeoGo dependency to 0.113.0 (#743)

Bugs fixed:
 * panic in NetMap.ContainerNodes methods when the number of SELECTs exceeds the number of REPs (#735)
 * nil EACL targets and target subjects were not preserved in CopyTo methods (#738)

## [1.0.0-rc.14] - 2025-08-07

This version makes SDK compatible with API 2.18, cleans up some APIs and
fixes a number of bugs.

New features:
 * New (iter.Seq-style) iterator APIs for parameters and attributes (#694)
 * Client constructor reusing existing GRPC connection (#702)
 * N3 witnesses in requests and objects (#708)
 * Unsigned GET/HEAD response support (#719)
 * DecodeString() support for version.Version (#720)
 * Associate attribute support (#721)

Behaviour changes:
 * ExternalAddresses deprecation following API changes (#700)
 * object.NewAttribute returns instance instead of a pointer making it easier to use other APIs (#701)
 * SearchV2 checks are adjusted according to API changes (#712)
 * storagegroup package is gone because it's deprecated in API 2.18 (#723)
 * API 2.18 is used now by default (#721)

Improvements:
 * SearchObjects documentation update (#698)
 * EACL processing code allows for lazy header evaluation now (#716)

Bugs fixed:
 * Client.Dial waiting for more time than allowed by timeout configuration (#696)
 * Client panic on Close() call before Dial() (#706)
 * Slicer overriding object owner based on bearer token data (#710)
 * Incorrect buffer pool use by Client (#713)
 * Old objects with incorrect session tokens (missing lifetime settings) couldn't be processed correctly (#715, #727)
 * Link objects created by slicer could have doubled payload (#725)

## [1.0.0-rc.13] - 2025-03-06

The key change is the removal of `github.com/nspcc-dev/neofs-api-go/v2` module (#667).
When updating, app code may break in places importing components of this module.
However, it is expected that this will only affect the system code: regular apps
following the lib recommendations will most likely not require edits.

Several code elements have been marked as deprecated. It is strongly
recommended to update all places. All deprecated parts will be removed in the
upcoming major release.

New features:
 * eACL rules may subject NeoFS accounts now (#592)
 * various instance constructors are available now (#593, #598, #601, #602, #605, #612)
 * constant sizes of various IDs are exported now (#598)
 * tokens' lifetime claims are exported now (#612, #626)
 * eACL presence marker (#626)
 * node attributes' getter and setter are available now (#635)
 * `client.Client/SearchObjects` method calling new `ObjectService.SearchV2` API (#676, #682)

Behaviour changes:
 * all ID types are declared as arrays (#598)
 * several types have been affirmed `comparable` (#598, #601)
 * IDs consisting only of zeros are now prohibited (#607)
 * `client/Client.ReplicateObject` now works with object signature field (#622)
 * container JSON fields are verified on decoding now (#628)
 * payload of irregular objects is checked against the NeoFS protocol now (#629)
 * zero range objects ops now operate with the full payload (#685)
 * several reasonable limits are now imposed on storage policies (#686)
 * `pool/Pool.Close` returns an error and implements `io.Closer` now (#689)
 * minimal required Go is 1.23 now (#618, #688)

Improvements:
 * intra-system `audit` package is no longer present in the lib (#590)
 * Go tests may now easier change ID instance or randomize several ones (#590)
 * test randomizers are now error-free and no longer accept `testing.TB` (#590)
 * crypto test randomizers now provide more stuff out of the box (#590)
 * Go-compiled NeoFS API proto files are in this lib now (#591)
 * stable marshalling functionality is in this lib now (#588, #597)
 * enum string mapping is now pure-functional, `fmt.Stringer` is stabilized (#605, #629)
 * session token randomizers return return instances with valid lifetime now (#610)
 * node shortage error when applying storage policy can now be checked (#615)
 * connection pool no longer opens redundant sessions with storage nodes (#623)
 * invalid URI errors are caught faster and their causes are clear now (#636)
 * `client.Client` no longer sends requests with invalid signature (#641)
 * `object.Object` type works with owner ID by value (#634)
 * updated github.com/nspcc-dev/hrw/v2 dependency to v2.0.3 (#688)
 * updated github.com/nspcc-dev/neo-go dependency to v0.108.1 (#688)
 * updated github.com/nspcc-dev/tzhash dependency to v1.8.2 (#621)
 * updated github.com/antlr4-go/antlr/v4 dependency to v4.13.1 (#621)
 * updated google.golang.org/grpc dependency to v1.70.0 (#688)
 * updated google.golang.org/protobuf dependency to v1.36.5 (#688)
 * updated github.com/google/go-cmp dependency to v0.7.0 (#688)
 * updated github.com/stretchr/testify dependency to v1.10.0 (#688)
 * updated github.com/testcontainers/testcontainers-go dependency to v0.35.0 (#688)
 * updated golang.org/x/exp dependency to v0.0.0-20250210185358-939b2ce775ac (#688)
 * updated golang.org/x/crypto dependency to v0.31.0 (#668)
 * updated golang.org/x/sys dependency to v0.28.0 (#668)
 * updated golang.org/x/text dependency to v0.21.0 (#668)
 * updated github.com/docker/docker dependency to v26.1.5+incompatible (#619)
 * updated many indirect dependencies to their latest secure versions (#609)

Bugs fixed:
 * unknown `enum` field values are no longer lost on conversions (#593, #605, #629)
 * ID fields with incorrect byte len are no longer passed (#595)
 * `user/ID.DecodeString` method no longer passes invalid byte sequences (#598)
 * validity of tokens with unset lifetime at zero epoch (#612, #626)
 * bit-len of session verb types (#612)
 * duplicated node attributes are no longer passed (#627)
 * invalid EigenTrust alpha network parameter is no longer passed (#627)
 * invalid duration in `client.Client` statistics (#636)
 * response fields unchecked by `client.Client` (#641)
 * thread-unsafe random number generator used by `pool.Pool` concurrently (#643)
 * unclosed connections after `pool.Pool` close (#689)
 * `object/Object.Version` method return for unset version (#634)

## [1.0.0-rc.12] - 2024-05-29

RC12 brings some protocol updates, new dependencies and solves a number of
bugs. There are some very minor API changes, mostly it's compatible with
RC11. No further API changes are planned.

New features:
 * verified node domains support (#523)
 * LOCODE getters for node info (#534)
 * binary object replication support (#535, #565)
 * numeric matches for SEARCH queries (#550)
 * numeric matches for EACLs (#554, #573)
 * new split object scheme support (#543, #560, #566, #575)
 * explicit issuer field in bearer token, new SetSignature method (#564)
 * additional methods for things that were previously only provided by
   neofs-api-go module (#570)
 * container version getter (#581)

Behaviour changes:
 * minimal required Go is 1.20 now (#522, #559)
 * parameterized constructor for attributes (#528)
 * header size limit (#558)
 * removal of notification functionality (#557)

Improvements:
 * updated github.com/hashicorp/golang-lru dependency to v2.0.7 (#522, #527)
 * updated go.uber.org/zap dependency to v1.27.0 (#522, #559)
 * updated github.com/nspcc-dev/tzhash dependency to v1.7.2 (#524, #559)
 * buffer reuse in slicer (#525)
 * per-key node session caches in pool (#527)
 * updated github.com/nspcc-dev/hrw/v2 dependency to v2.0.1 (#529, #559)
 * buffer size hints for slicer (#539)
 * updated github.com/nspcc-dev/neo-go dependency to v0.106.0 (#559, #584)
 * updated google.golang.org/grpc dependency to v1.62.0 (#559)
 * updated google.golang.org/protobuf dependency to v1.33.0 (#559, #569)
 * updated golang.org/x/net dependency to 0.23.0 (#579)

Bugs fixed:
 * incorrect number of objectGetStream, objectRangeStream, objectPutStream,
   objectSearchStream events counted (#521)
 * error not returned from object writer in some cases (#521)
 * pool failed to restart temporary unhealthy clients (#521)
 * panic in xheaders handler (#582)

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

[1.0.0-rc.17]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.16...v1.0.0-rc.17
[1.0.0-rc.16]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.15...v1.0.0-rc.16
[1.0.0-rc.15]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.14...v1.0.0-rc.15
[1.0.0-rc.14]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.13...v1.0.0-rc.14
[1.0.0-rc.13]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.12...v1.0.0-rc.13
[1.0.0-rc.12]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.11...v1.0.0-rc.12
[1.0.0-rc.11]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.10...v1.0.0-rc.11
[1.0.0-rc.10]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.9...v1.0.0-rc.10
[1.0.0-rc.9]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.8...v1.0.0-rc.9
[1.0.0-rc.8]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.7...v1.0.0-rc.8
