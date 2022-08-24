/*
Package version provides functionality for NeoFS versioning.

NeoFS uses NeoFS API versioning scheme. It uses major and minor version of
the API.

In most of the cases it will be enough to use the latest supported NeoFS API
version in SDK:

	ver := version.Current()

It is possible to specify arbitrary version by setting major and minor numbers:

	var ver version.Version
	ver.SetMajor(2)
	ver.SetMinor(5)
*/
package version
