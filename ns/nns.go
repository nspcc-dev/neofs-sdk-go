package ns

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/nspcc-dev/neo-go/pkg/core/state"
	neoclient "github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-contract/nns"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// NNS looks up NeoFS names using Neo Name Service.
//
// Instances are created with a variable declaration. Before work, the connection
// to the NNS server MUST be established using Dial method.
type NNS struct {
	nnsContract util.Uint160

	neoClient neoClient
}

// represents virtual connection to Neo network used by NNS.Dial.
type neoClient interface {
	// calls specified method of the Neo smart contract with provided parameters.
	call(contract util.Uint160, method string, prm []smartcontract.Parameter) (*result.Invoke, error)
}

// implements neoClient using Neo HTTP client.
//
// note: see NNS.Dial to realize why this isn't defined as type wrapper like neoWebSocket.
type neoHTTP struct {
	*neoclient.Client
}

func (x *neoHTTP) call(contract util.Uint160, method string, prm []smartcontract.Parameter) (*result.Invoke, error) {
	return x.Client.InvokeFunction(contract, method, prm, nil)
}

// implements neoClient using Neo WebSocket client.
type neoWebSocket neoclient.WSClient

func (x *neoWebSocket) call(contract util.Uint160, method string, prm []smartcontract.Parameter) (*result.Invoke, error) {
	return (*neoclient.WSClient)(x).InvokeFunction(contract, method, prm, nil)
}

// Dial connects to the address of the NNS server. If fails, the instance
// MUST NOT be used.
//
// If URL address scheme is 'ws' or 'wss', then WebSocket protocol is used,
// otherwise HTTP.
func (n *NNS) Dial(address string) error {
	// multiSchemeClient unites neoClient and common interface of
	// neoclient.Client and neoclient.WSClient. Interface is anonymous
	// according to assumption that common interface of these client types
	// is not required by design and may diverge with changes.
	var multiSchemeClient interface {
		neoClient
		// Init turns client to "ready-to-work" state.
		Init() error
		// GetContractStateByID returns state of the NNS contract on 1 input.
		GetContractStateByID(int32) (*state.Contract, error)
	}

	uri, err := url.Parse(address)
	if err == nil && (uri.Scheme == "ws" || uri.Scheme == "wss") {
		cWebSocket, err := neoclient.NewWS(context.Background(), address, neoclient.Options{})
		if err != nil {
			return fmt.Errorf("create Neo WebSocket client: %w", err)
		}

		multiSchemeClient = (*neoWebSocket)(cWebSocket)
	} else {
		cHTTP, err := neoclient.New(context.Background(), address, neoclient.Options{})
		if err != nil {
			return fmt.Errorf("create Neo HTTP client: %w", err)
		}

		// if neoHTTP is defined as type wrapper
		//   type neoHTTP neoclient.Client
		// then next assignment causes compilation error
		//   multiSchemeClient = (*neoHTTP)(cHTTP)
		multiSchemeClient = &neoHTTP{
			Client: cHTTP,
		}
	}

	if err = multiSchemeClient.Init(); err != nil {
		return fmt.Errorf("initialize Neo client: %w", err)
	}

	nnsContract, err := multiSchemeClient.GetContractStateByID(1)
	if err != nil {
		return fmt.Errorf("get NNS contract state: %w", err)
	}

	n.neoClient = multiSchemeClient
	n.nnsContract = nnsContract.Hash

	return nil
}

// ResolveContainerName looks up for NNS TXT records for the given container name
// by calling `resolve` method of NNS contract. Returns the first record which represents
// valid container ID in a string format. Otherwise, returns an error.
//
// ResolveContainerName MUST NOT be called before successful Dial.
//
// See also https://docs.neo.org/docs/en-us/reference/nns.html.
func (n *NNS) ResolveContainerName(name string) (cid.ID, error) {
	res, err := n.neoClient.call(n.nnsContract, "resolve", []smartcontract.Parameter{
		{
			Type:  smartcontract.StringType,
			Value: name + ".container",
		},
		{
			Type:  smartcontract.IntegerType,
			Value: int64(nns.TXT),
		},
	})
	if err != nil {
		return cid.ID{}, fmt.Errorf("invoke NNS contract: %w", err)
	}

	if res.State != vm.HaltState.String() {
		return cid.ID{}, fmt.Errorf("NNS contract fault exception: %s", res.FaultException)
	} else if len(res.Stack) == 0 {
		return cid.ID{}, errors.New("empty stack in invocation result")
	}

	itemArr, err := res.Stack[len(res.Stack)-1].Convert(stackitem.ArrayT) // top stack element is last in the array
	if err != nil {
		return cid.ID{}, fmt.Errorf("convert stack item to %s", stackitem.ArrayT)
	}

	if _, ok := itemArr.(stackitem.Null); !ok {
		arr, ok := itemArr.Value().([]stackitem.Item)
		if !ok {
			// unexpected for types from stackitem package
			return cid.ID{}, errors.New("invalid cast to stack item slice")
		}

		var id cid.ID

		for i := range arr {
			bs, err := arr[i].TryBytes()
			if err != nil {
				return cid.ID{}, fmt.Errorf("convert array item to byte slice: %w", err)
			}

			err = id.DecodeString(string(bs))
			if err == nil {
				return id, nil
			}
		}
	}

	return cid.ID{}, errNotFound
}
