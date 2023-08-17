package client_test

import (
	"context"
	"fmt"
	"time"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	netmapv2 "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	rpcClient "github.com/nspcc-dev/neofs-api-go/v2/rpc/client"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/common"
	"github.com/nspcc-dev/neofs-api-go/v2/rpc/grpc"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
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

// Put a new container into NeoFS.
func ExampleClient_ContainerPut() {
	ctx := context.Background()
	var accountID user.ID

	// The account was taken from https://github.com/nspcc-dev/neofs-aio
	key, err := keys.NEP2Decrypt("6PYM8VdX2BSm7BSXKzV4Fz6S3R9cDLLWNrD9nMjxW352jEv3fsC8N3wNLY", "one", keys.NEP2ScryptParams())
	if err != nil {
		panic(err)
	}

	signer := user.NewAutoIDSignerRFC6979(key.PrivateKey)
	// take account from user's signer
	accountID = signer.UserID()

	// prepare client
	var prmInit client.PrmInit

	c, err := client.New(prmInit)
	if err != nil {
		panic(fmt.Errorf("New: %w", err))
	}

	// connect to NeoFS gateway
	var prmDial client.PrmDial
	prmDial.SetServerURI("grpc://localhost:8080") // endpoint address
	prmDial.SetTimeout(15 * time.Second)
	prmDial.SetStreamTimeout(15 * time.Second)

	if err = c.Dial(prmDial); err != nil {
		panic(fmt.Errorf("dial %v", err))
	}

	// describe new container
	cont := container.Container{}
	// set version and nonce
	cont.Init()
	cont.SetOwner(accountID)
	cont.SetBasicACL(acl.PublicRW)

	// set reserved attributes
	cont.SetName("name-1")
	cont.SetCreationTime(time.Now().UTC())

	// init placement policy
	var containerID cid.ID
	var placementPolicyV2 netmapv2.PlacementPolicy
	var replicas []netmapv2.Replica

	replica := netmapv2.Replica{}
	replica.SetCount(1)
	replicas = append(replicas, replica)
	placementPolicyV2.SetReplicas(replicas)

	var placementPolicy netmap.PlacementPolicy
	if err = placementPolicy.ReadFromV2(placementPolicyV2); err != nil {
		panic(fmt.Errorf("ReadFromV2 %w", err))
	}

	placementPolicy.SetContainerBackupFactor(1)
	cont.SetPlacementPolicy(placementPolicy)

	containerID, err = c.ContainerPut(ctx, cont, signer, client.PrmContainerPut{})
	if err != nil {
		panic(fmt.Errorf("ContainerPut %w", err))
	}

	// containerID already exists
	fmt.Println(containerID)
	// example output: 76wa5UNiT8gk8Q5rdCVCV4pKuZSmYsifh6g84BcL6Hqs

	// but container creation is async operation. We should wait some time or make polling to ensure container created
	// for simplifying we just wait
	<-time.After(2 * time.Second)

	contRes, err := c.ContainerGet(ctx, containerID, client.PrmContainerGet{})
	if err != nil {
		panic(fmt.Errorf("ContainerGet %w", err))
	}

	jsonData, err := contRes.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("MarshalJSON %w", err))
	}

	fmt.Println(string(jsonData))
	// example output:
	/*
		{
		    "version": {
		        "major": 2,
		        "minor": 13
		    },
		    "ownerID": {
		        "value": "Ne6eoiwn40vQFI/EEI4I906PUEiy8ZXKcw=="
		    },
		    "nonce": "rPVd/iw2RW6Q6d66FVnIqg==",
		    "basicACL": 532660223,
		    "attributes": [
		        {
		            "key": "Name",
		            "value": "name-1"
		        },
		        {
		            "key": "Timestamp",
		            "value": "1681738627"
		        }
		    ],
		    "placementPolicy": {
		        "replicas": [
		            {
		                "count": 1,
		                "selector": ""
		            }
		        ],
		        "containerBackupFactor": 1,
		        "selectors": [],
		        "filters": [],
		        "subnetId": {
		            "value": 0
		        }
		    }
		}
	*/
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
