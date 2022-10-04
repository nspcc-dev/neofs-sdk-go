package clientutil

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

func createClient(endpoint string) (*client.Client, error) {
	var prmInit client.PrmInit
	prmInit.ResolveNeoFSFailures()

	var c client.Client
	c.Init(prmInit)

	var prmDial client.PrmDial
	prmDial.SetServerURI(endpoint)

	err := c.Dial(prmDial)
	if err != nil {
		return nil, fmt.Errorf("endpoint dial: %w", err)
	}

	return &c, nil
}
