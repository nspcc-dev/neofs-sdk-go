package client_test

import (
	"time"

	rpcClient "github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/client"
)

func ExampleClient_createInstance() {
	// Create client instance
	var prm client.PrmInit
	c, err := client.New(prm)
	_ = err

	// Connect to the NeoFS server
	var prmDial client.PrmDial
	prmDial.SetServerURI("grpc://localhost:8080") // endpoint address
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	_ = c.Dial(prmDial)
}

type CustomRPCRequest struct {
}

type CustomRPCResponse struct {
}

func (a *CustomRPCRequest) ToGRPCMessage() grpc.Message {
	return nil
}

func (a *CustomRPCRequest) FromGRPCMessage(grpc.Message) error {
	return nil
}

func (a *CustomRPCResponse) ToGRPCMessage() grpc.Message {
	return nil
}

func (a *CustomRPCResponse) FromGRPCMessage(grpc.Message) error {
	return nil
}

// Consume custom service of the server.
func Example_customService() {
	// syntax = "proto3";
	//
	// service CustomService {
	// 	rpc CustomRPC(CustomRPCRequest) returns (CustomRPCResponse);
	// }

	// import "github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	// import "github.com/nspcc-dev/neofs-api-go/v2/rpc/common"

	var prmInit client.PrmInit
	// ...

	c, _ := client.New(prmInit)

	req := &CustomRPCRequest{}
	resp := &CustomRPCResponse{}

	err := c.ExecRaw(func(c *rpcClient.Client) error {
		return rpcClient.SendUnary(c, common.CallMethodInfo{
			Service: "CustomService",
			Name:    "CustomRPC",
		}, req, resp)
	})

	_ = err

	// ...

	// Close the connection
	_ = c.Close()

	// Note that it's not allowed to override Client behaviour directly: the parameters
	// for the all operations are write-only and the results of the all operations are
	// read-only. To be able to override client behavior (e.g. for tests), abstract it
	// with an interface:
	//
	// import "github.com/nspcc-dev/neofs-sdk-go/client"
	//
	// type NeoFSClient interface {
	// // Operations according to the application needs
	// CreateContainer(context.Context, container.Container) error
	// // ...
	// }
	//
	// type client struct {
	// 	c *client.Client
	// }
	//
	// func (x *client) CreateContainer(context.Context, container.Container) error {
	// // ...
	// }
}
