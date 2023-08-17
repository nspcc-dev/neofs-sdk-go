/*
Package bearer provides bearer token definition.

Bearer token is attached to the object service requests, and it overwrites
extended ACL of the container. Mainly it is used to provide access of private
data for specific user. Therefore, it must be signed by owner of the container.
*/
package bearer
