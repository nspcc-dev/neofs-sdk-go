/*
Package pool provides a wrapper for several NeoFS API clients.

The main component is Pool type. It is a virtual connection to the network
and provides methods for executing operations on the server. It also supports
a weighted random selection of the underlying client to make requests.

Pool has an auto-session mechanism for object operations. It is enabled by default.
The mechanism allows to manipulate objects like upload, download, delete, etc, without explicit session passing.
This behavior may be disabled per request by calling IgnoreSession() on the appropriate Prm* argument.
Note that if auto-session is disabled, the user MUST provide the appropriate session manually for PUT and DELETE object operations.
The user may provide session, for another object operations.
*/
package pool
