package poolv2

import (
	"github.com/nspcc-dev/neofs-sdk-go/client"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

type Attribute struct {
	Name, Value string
}

type NodeParam struct {
	priority int
	address  string
	weight   float64
}

func NewNodeParam(address string, priority int, weight float64) NodeParam {
	return NodeParam{
		priority: priority,
		address:  address,
		weight:   weight,
	}
}

func (x *NodeParam) Priority() int {
	return x.priority
}

func (x *NodeParam) Address() string {
	return x.address
}

func (x *NodeParam) Weight() float64 {
	return x.weight
}

type Pool struct {
	signer  neofscrypto.Signer
	clients []*client.Client
}

func NewPool(nodes []NodeParam, signer neofscrypto.Signer) (*Pool, error) {
	init := client.NewPrmInit(signer)

	var clients []*client.Client
	for _, n := range nodes {
		cl, err := client.New(init)
		if err != nil {
			return nil, err
		}

		dial := client.NewPrmDial(n.Address())

		if err = cl.Dial(dial); err != nil {
			return nil, err
		}

		clients = append(clients, cl)
	}

	return &Pool{
		signer:  signer,
		clients: clients,
	}, nil
}

func (p *Pool) client() *client.Client {
	// some logic to determine which client should be returned

	return p.clients[0]
}
