# Changelog

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

[1.0.0-rc.8]: https://github.com/nspcc-dev/neofs-sdk-go/compare/v1.0.0-rc.7...v1.0.0-rc.8
