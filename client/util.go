package client

import (
	"errors"
	"io"
)

// binaryMessageWriter represents stream of the binary messages.
type binaryMessageWriter interface {
	// Write writes given binary message into the stream. Returns [io.EOF] if server
	// side closed the stream or any other error encountered that prevented the
	// message to be delivered.
	//
	// Implementations must not modify the message.
	Write([]byte) error
}

// reads exactly len(b) continuous bytes from given [io.Reader]. If reading more
// than len(b) is appropriate, [io.ReadFull] should be used instead.
func readExact(r io.Reader, b []byte) error {
	var n, total int
	var err error
	for {
		n, err = r.Read(b[total:])
		total += n
		if total == len(b) || err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
