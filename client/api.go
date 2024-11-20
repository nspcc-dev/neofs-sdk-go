package client

import (
	"fmt"
)

// unifies errors of all RPC.
func rpcErr(e error) error {
	return fmt.Errorf("rpc failure: %w", e)
}
