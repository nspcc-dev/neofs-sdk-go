/*
Package waiter provides synchronous implementation of asynchronous commands of [client.Client] and [pool.Pool].

Supported operations:
  - Container put
  - Container setEacl
  - Container delete

The main component is [Waiter] type. It is using [client.Client] or [pool.Pool] as [Executor] implementation
for querying async operation and wait some time, to be sure it has effect like container created/deleted etc.
*/
package waiter
