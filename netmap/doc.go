/*
Package netmap provides functionality for working with information about the
NeoFS network, primarily a layer of storage nodes.

The package concentrates all the characteristics of NeoFS networks.

[NetMap] represents NeoFS network map - one of the main technologies used to
store data in the system. It is composed of information about all storage nodes
([NodeInfo] type) in a particular network. [NetMap] methods allow you to impose
container storage policies ([PlacementPolicy] type) on a fixed composition of
nodes for selecting nodes corresponding to the placement rules chosen by the
container creator.

[NetworkInfo] type is dedicated to descriptive characterization of network state
and settings.
*/
package netmap
