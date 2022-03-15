/*
Package pool provides a wrapper for several NeoFS API clients.

The main component is Pool type. It is a virtual connection to the network
and provides methods for executing operations on the server. It also supports
a weighted random selection of the underlying client to make requests.

Create pool instance with 3 nodes connection.
This InitParameters will make pool use 192.168.130.71 node while it is healthy. Otherwise, it will make the pool use
192.168.130.72 for 90% of requests and 192.168.130.73 for remaining 10%.
:
	prm := pool.InitParameters{
		Key: key,
		NodeParams: []NodeParam{
			{ Priority: 1, Address: "192.168.130.71", Weight: 1 },
			{ Priority: 2, Address: "192.168.130.72", Weight: 9 },
			{ Priority: 2, Address: "192.168.130.73", Weight: 1 },
		// ...
	}

	p, err := pool.NewPool(prm)
	// ...

Connect to the NeoFS server:
	err := p.Dial(ctx)
	// ...

Execute NeoFS operation on the server:
	var prm pool.PrmContainerPut
	prm.SetContainer(cnr)
	// ...

	res, err := p.PutContainer(context.Background(), prm)
	// ...

Execute NeoFS operation on the server and check error:
	var prm pool.PrmObjectHead
	prm.SetAddress(addr)
	// ...

	res, err := p.HeadObject(context.Background(), prm)
	if client.IsErrObjectNotFound(err) {
		// ...
	}
	// ...

Close the connection:
	p.Close()

*/
package pool
