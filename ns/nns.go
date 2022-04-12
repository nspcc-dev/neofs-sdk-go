package ns

import (
	"context"
	"errors"
	"fmt"

	nns "github.com/nspcc-dev/neo-go/examples/nft-nd-nns"
	neoclient "github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

// NNS looks up NeoFS names using Neo Name Service.
//
// Instances are created with a variable declaration. Before work, the connection
// to the NNS server MUST BE established using Dial method.
type NNS struct {
	nnsContract util.Uint160

	// neoclient.Client interface wrapper, needed for testing
	neoClient interface {
		invoke(contract util.Uint160, method string, prm []smartcontract.Parameter) (*result.Invoke, error)
	}
}

// client is a core implementation of internal NNS.neoClient which is used by NNS.Dial.
type client neoclient.WSClient

func (x *client) invoke(contract util.Uint160, method string, prm []smartcontract.Parameter) (*result.Invoke, error) {
	return (*neoclient.WSClient)(x).InvokeFunction(contract, method, prm, nil)
}

// Dial connects to the address of the NNS server. If fails, the instance
// SHOULD NOT be used.
func (n *NNS) Dial(address string) error {
	cli, err := neoclient.NewWS(context.Background(), address, neoclient.Options{})
	if err != nil {
		return fmt.Errorf("create neo client: %w", err)
	}

	if err = cli.Init(); err != nil {
		return fmt.Errorf("initialize neo client: %w", err)
	}

	nnsContract, err := cli.GetContractStateByID(1)
	if err != nil {
		return fmt.Errorf("get NNS contract state: %w", err)
	}

	n.neoClient = (*client)(cli)
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
func (n *NNS) ResolveContainerName(name string) (*cid.ID, error) {
	res, err := n.neoClient.invoke(n.nnsContract, "resolve", []smartcontract.Parameter{
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
		return nil, fmt.Errorf("invoke NNS contract: %w", err)
	}

	if res.State != vm.HaltState.String() {
		return nil, fmt.Errorf("NNS contract fault exception: %s", res.FaultException)
	} else if len(res.Stack) == 0 {
		return nil, errors.New("empty stack in invocation result")
	}

	itemArr, err := res.Stack[len(res.Stack)-1].Convert(stackitem.ArrayT) // top stack element is last in the array
	if err != nil {
		return nil, fmt.Errorf("convert stack item to %s", stackitem.ArrayT)
	}

	if _, ok := itemArr.(stackitem.Null); !ok {
		arr, ok := itemArr.Value().([]stackitem.Item)
		if !ok {
			// unexpected for types from stackitem package
			return nil, errors.New("invalid cast to stack item slice")
		}

		var id cid.ID

		for i := range arr {
			bs, err := arr[i].TryBytes()
			if err != nil {
				return nil, fmt.Errorf("convert array item to byte slice: %w", err)
			}

			err = id.Parse(string(bs))
			if err == nil {
				return &id, nil
			}
		}
	}

	return nil, errNotFound
}
