/*
Package ns provides functionality of NeoFS name system.

DNS type is designed to resolve NeoFS-related names using Domain Name System:

	const containerName = "some-container"

	var dns DNS

	containerID, err := dns.ResolveContainerName(containerName)
	// ...

NNS type is designed to resolve NeoFS-related names using Neo Name Service:

	var nns NNS

	err := nns.Dial(nnsServerAddress)
	// ...

	containerID, err := nns.ResolveContainerName(containerName)
	// ...
*/
package ns
