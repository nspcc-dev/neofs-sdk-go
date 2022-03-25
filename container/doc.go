/*
Package container provides primitives to perform container operations in NeoFS.

Container type provides functionality to work with containers. For example
public container without eACL and with "key":"value" attribute:

	import "github.com/nspcc-dev/neofs-sdk-go/acl"
	...

	var c Container

	c.SetBasicACL(acl.PublicBasicRule) // setting ACL

	// setting custom attribute
	var customAttr Attribute
	customAttr.SetKey("key")
	customAttr.SetValue("value")

	c.SetAttributes(Attributes{customAttr})

Package also contains:

• types for internals of the Container (Attribute, Attributes);

• types for storage estimations (UsedSpaceAnnouncement);

• well-known container attributes (AttributeName, AttributeTimestamp)

• utils for working with containers (such as SetNativeNameWithZone,
GetNativeNameWithZone, etc.).

Container instances can be also used to process NeoFS API V2 protocol messages
(see neo.fs.v2.container package in https://github.com/nspcc-dev/neofs-api).

On client side:
	import "github.com/nspcc-dev/neofs-api-go/v2/container"

	msg := container.InitCreation()
	cnr.WriteToV2(&msg)

	// send msg

On server side:
	// recv msg

	msg := container.InitCreation()
	cnr.ReadFromV2(msg)

	// process cnr

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package container
