[![codecov](https://codecov.io/gh/nspcc-dev/neofs-sdk-go/branch/master/graph/badge.svg)](https://codecov.io/gh/nspcc-dev/neofs-sdk-go)
[![Report](https://goreportcard.com/badge/github.com/nspcc-dev/neofs-sdk-go)](https://goreportcard.com/report/github.com/nspcc-dev/neofs-sdk-go)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/nspcc-dev/neofs-sdk-go?sort=semver)
![License](https://img.shields.io/github/license/nspcc-dev/neofs-sdk-go.svg?style=popout)

# neofs-sdk-go
Go implementation of NeoFS SDK. It contains high-level version-independent wrappers
for structures from [proto](https://github.com/nspcc-dev/neofs-sdk-go/proto) packages as well as
helper functions for simplifying node/dApp implementations.

## Repository structure

### accounting
Contains fixed-point `Decimal` type for performing balance calculations.

### eacl
Contains Extended ACL types for fine-grained access control.
There is also a reference implementation of checking algorithm which is used in NeoFS node.

### checksum
Contains `Checksum` type encapsulating checksum as well as it's kind.
Currently Sha256 and [Tillich-Zemor hashsum](https://github.com/nspcc-dev/tzhash) are in use.

### bearer
Contains Bearer token type with several NeoFS-specific methods.

### session
To help lightweight clients interact with NeoFS without sacrificing trust, NeoFS has a concept
of session token. It is signed by client and allows any node with which a session is established
to perform certain actions on behalf of the user.

### client
Contains client for working with NeoFS.
```go
var prmInit client.PrmInit
prmInit.SetDefaultPrivateKey(key) // private key for request signing

c, err := client.New(prmInit)
if err != nil {
    return
}

var prmDial client.PrmDial
prmDial.SetServerURI("grpcs://localhost:40005") // endpoint address

err := c.Dial(prmDial)
if err != nil {
    return
}
    
ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
defer cancel()

var prm client.PrmBalanceGet
prm.SetAccount(acc)

res, err := c.BalanceGet(ctx, prm)
if err != nil {
    return
}

fmt.Printf("Balance for %s: %v\n", acc, res.Amount())
```

#### Response status
In NeoFS every operation can fail on multiple levels, so a single `error` doesn't suffice,
e.g. consider a case when object was put on 4 out of 5 replicas. Thus, all request execution
details are contained in `Status` returned from every RPC call. dApp can inspect them
if needed and perform any desired action. In the case above we may want to report
these details to the user as well as retry an operation, possibly with different parameters.
Status wire-format is extendable and each node can report any set of details it wants.
The set of reserved status codes can be found in
[NeoFS API](https://github.com/nspcc-dev/neofs-api/blob/master/status/types.proto).

### netmap
Contains CRUSH-like implementation of container node selection algorithm. Relevant details
are described in this paper http://ceur-ws.org/Vol-2344/short10.pdf . Note that it can be
outdated in some details.

`netmap/json_tests` subfolder contains language-agnostic tests for selection algorithm. 

```go
import (
    "github.com/nspcc-dev/neofs-sdk-go/netmap"
    "github.com/nspcc-dev/neofs-sdk-go/object"
)

func placementNodes(addr *object.Address, p *netmap.PlacementPolicy, neofsNodes []netmap.NodeInfo) {
    // Convert list of nodes in NeoFS API format to the intermediate representation.
    nodes := netmap.NodesFromInfo(nodes)

    // Create new netmap (errors are skipped for the sake of clarity). 
    nm, _ := NewNetmap(nodes)

    // Calculate nodes of container.
    cn, _ := nm.GetContainerNodes(p, addr.ContainerID().ToV2().GetValue())

    // Return list of nodes for each replica to place object on in the order of priority.
    return nm.GetPlacementVectors(cn, addr.ObjectID().ToV2().GetValue())
}
```

### pool
Simple pool for managing connections to NeoFS nodes.

### checksum, version
Contain simple API wrappers.

## Development

### NeoFS API protocol codegen
Go code for NeoFS protocol messages, client and server is in `api` directory.
To compile source files from https://github.com/nspcc-dev/neofs-api repository,
clone it first and then exec:
```
$ ./scripts/genapi.sh /path/to/neofs-api
```
