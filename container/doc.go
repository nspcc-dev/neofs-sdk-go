/*
Package container provides functionality related to the NeoFS containers.

The base type is Container. To create new container in the NeoFS network
Container instance should be initialized
  var cnr Container
  cnr.Init()
  // fill all the fields

  // encode cnr and send

After the container is persisted in the NeoFS network, applications can process
it using the instance of Container types
  // recv binary container

  var cnr Container

  err := cnr.Unmarshal(bin)
  // ...

  // process the container data

Instances can be also used to process NeoFS API V2 protocol messages
(see neo.fs.v2.container package in https://github.com/nspcc-dev/neofs-api).

On client side:
	import "github.com/nspcc-dev/neofs-api-go/v2/container"

	var msg container.Container
	cnr.WriteToV2(&msg)

	// send msg

On server side:
	// recv msg

	var cnr Container
	cnr.ReadFromV2(msg)

	// process cnr

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package container
