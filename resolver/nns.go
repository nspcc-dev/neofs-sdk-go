package resolver

import (
	"context"
	"fmt"

	nns "github.com/nspcc-dev/neo-go/examples/nft-nd-nns"
	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/rpc/response/result"
	"github.com/nspcc-dev/neo-go/pkg/smartcontract"
	"github.com/nspcc-dev/neo-go/pkg/util"
	"github.com/nspcc-dev/neo-go/pkg/vm"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
)

type NNSResolver interface {
	// ResolveContainerName get record for domain name+'.container' from NNS contract.
	ResolveContainerName(name string) (*cid.ID, error)
}

type nnsResolver struct {
	rpc     *client.Client
	nnsHash util.Uint160
}

const (
	resolve         = "resolve"
	containerDomain = ".container"
)

// NewNNSResolver creates resolver that can get records from NNS contract.
func NewNNSResolver(ctx context.Context, address string) (NNSResolver, error) {
	rpc, err := client.New(ctx, address, client.Options{})
	if err != nil {
		return nil, err
	}
	if err = rpc.Init(); err != nil {
		return nil, err
	}

	nnsContract, err := rpc.GetContractStateByID(1)
	if err != nil {
		return nil, err
	}

	return &nnsResolver{
		rpc:     rpc,
		nnsHash: nnsContract.Hash,
	}, nil
}

func (n *nnsResolver) ResolveContainerName(name string) (*cid.ID, error) {
	res, err := n.rpc.InvokeFunction(n.nnsHash, resolve, []smartcontract.Parameter{
		{
			Type:  smartcontract.StringType,
			Value: name + containerDomain,
		},
		{
			Type:  smartcontract.IntegerType,
			Value: int64(nns.TXT),
		},
	}, nil)
	if err != nil {
		return nil, err
	}
	if err = getInvocationError(res); err != nil {
		return nil, err
	}

	arr, err := getArrString(res.Stack)
	if err != nil {
		return nil, err
	}

	cnrID := cid.New()
	for _, rec := range arr {
		if err = cnrID.Parse(rec); err != nil {
			continue
		}
		return cnrID, nil
	}

	return nil, fmt.Errorf("not found")
}

func getArrString(st []stackitem.Item) ([]string, error) {
	array, err := getArray(st)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(array))
	for i, item := range array {
		bs, err := item.TryBytes()
		if err != nil {
			return nil, err
		}
		res[i] = string(bs)
	}

	return res, nil
}

func getArray(st []stackitem.Item) ([]stackitem.Item, error) {
	index := len(st) - 1 // top stack element is last in the array
	arr, err := st[index].Convert(stackitem.ArrayT)
	if err != nil {
		return nil, err
	}
	if _, ok := arr.(stackitem.Null); ok {
		return nil, nil
	}

	iterator, ok := arr.Value().([]stackitem.Item)
	if !ok {
		return nil, fmt.Errorf("bad conversion")
	}
	return iterator, nil
}

func getInvocationError(result *result.Invoke) error {
	if result.State != vm.HaltState.String() {
		return fmt.Errorf("invocation failed: %s", result.FaultException)
	}
	if len(result.Stack) == 0 {
		return fmt.Errorf("result stack is empty")
	}
	return nil
}
