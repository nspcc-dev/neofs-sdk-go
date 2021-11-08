package test

import (
	"github.com/nspcc-dev/neofs-sdk-go/signature"
)

// Signature returns random pkg.Signature.
func Signature() *signature.Signature {
	x := signature.New()

	x.SetKey([]byte("key"))
	x.SetSign([]byte("sign"))

	return x
}
