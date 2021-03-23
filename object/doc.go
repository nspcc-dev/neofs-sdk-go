/*
Package object provides primitives to perform object operations in NeoFS.

Object type provides functionality to work with user object. For example:

	var obj Object
	obj.SetPayload([]byte("payload")) // set raw object payload

	// set object attribute
	var customAttr Attribute
	customAttr.SetKey("key")
	customAttr.SetValue("val")
	obj.SetAttributes(customAttr)

	id, _ := CalculateID(&obj) // calculate object ID

Package also contains:

• types for internals of the Object (such as Attribute, NotificationInfo, Range,
SplitID);

• types of non-regular object and their identification mechanisms (Lock, Tombstone
and Type);

• types and constants for object search operation (such as SearchFilter,
SearchMatchType, etc.);

• well-known object attributes (such as AttributeName, AttributeFileName, etc.)

• object related errors (such as SplitInfoError, ErrNotificationNotSet, etc.);

• utils for working with objects (such as CalculateAndSetID, VerifyID,
CalculateAndSetSignature, etc.).

Object instances can be also used to process NeoFS API V2 protocol messages
(see neo.fs.v2.object package in https://github.com/nspcc-dev/neofs-api).

On client side:
	import "github.com/nspcc-dev/neofs-api-go/v2/object"

	var msg object.Object
	obj.WriteToV2(&msg)

	// send msg

On server side:
	// recv msg

	var obj Object
	obj.ReadFromV2(msg)

	// process obj

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

*/
package object
