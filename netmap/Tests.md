## Language-agnostic testing
`json_tests` sub-folder contains netmap selection algorithm tests in the following format:

### File structure
Field|Description
---|---
`name`|Human readable name for a set of related tests operating on a single netmap.
`nodes`|List of `netmap.NodeInfo` structures coressponding to a nodes in the network.
`tests`|Map from a single test name to a `test` structure.

### Single test structure
Field|Description
---|---
`policy`|JSON representation of a netmap policy.
`pivot`|Optional pivot to use in container node selection.
`result`|List of lists of node-indices corresponding to each replica in the placement policy.
`error`|Error that should be raised for this specific test. The actual strings are used in SDK, other implementation may simply check that error has occurred.
`placement`|Optional field containing another test for selecting placement nodes for an object. Can contain `pivot`, `result` and `error` fields with the same meaning as above. Note that if `pivot` is omitted, empty value should be used.