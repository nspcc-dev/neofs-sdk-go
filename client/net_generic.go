//go:build !linux

package client

import (
	"syscall"
)

// TuneTCPConn does nothing.
func tuneTCPConn(network, address string, c syscall.RawConn) error {
	return nil
}
