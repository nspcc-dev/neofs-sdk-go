package clientutil

import (
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/client"
)

func createClient(endpoint string) (*client.Client, error) {
	var prmInit client.PrmInit

	c, err := client.New(prmInit)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	var prmDial client.PrmDial
	prmDial.SetServerURI(endpoint)

	if err = c.Dial(prmDial); err != nil {
		return nil, fmt.Errorf("endpoint dial: %w", err)
	}

	return c, nil
}
