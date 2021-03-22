/*
Package acl provides primitives to perform handling basic ACL management in NeoFS.

BasicACL type provides functionality for managing container basic access-control list.
For example, setting public basic ACL that could not be extended with any eACL rules:

	import "github.com/nspcc-dev/neofs-sdk-go/container"
	...
		c := container.New()
		c.SetBasicACL(acl.PublicBasicRule)

Using package types in an application is recommended to potentially work with
different protocol versions with which these types are compatible.

	Basic ACL bits meaning:

	┌──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┐
	│31│30│29│28│27│26│25│24│23│22│21│20│19│18│17│16│ <- Bit
	├──┼──┼──┼──┼──┴──┴──┴──┼──┴──┴──┴──┼──┴──┴──┴──┤
	│  │  │  │  │ RANGEHASH │   RANGE   │   SEARCH  │ <- Object service method
	│  │  │  │  ├──┬──┬──┬──┼──┬──┬──┬──┼──┬──┬──┬──┤
	│  │  │ X│ F│ U│ S│ O│ B│ U│ S│ O│ B│ U│ S│ O│ B│ <- Rule
	├──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┼──┤
	│15│14│13│12│11│10│09│08│07│06│05│04│03│02│01│00│ <- Bit
	├──┴──┴──┴──┼──┴──┴──┴──┼──┴──┴──┴──┼──┴──┴──┴──┤
	│   DELETE  │    PUT    │   HEAD    │    GET    │ <- Object service method
	├──┬──┬──┬──┼──┬──┬──┬──┼──┬──┬──┬──┼──┬──┬──┬──┤
	│ U│ S│ O│ B│ U│ S│ O│ B│ U│ S│ O│ B│ U│ S│ O│ B│ <- Rule
	└──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┴──┘

	U - Allows access to the owner of the container.
	S - Allows access to Inner Ring and container nodes in the current version of network map.
	O - Clients that do not match any of the categories above.
	B - Allows using Bear Token ACL rules to replace eACL rules.
	F - Flag denying Extended ACL. If set Extended ACL is ignored.
	X - Flag denying different owners of the request and the object.

	Remaining bits are reserved and are not used.
*/
package acl
