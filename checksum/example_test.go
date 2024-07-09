package checksum

import (
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/nspcc-dev/tzhash/tz"
)

func ExampleNewSHA256() {
	data := []byte("Hello, world!")
	c := NewSHA256(sha256.Sum256(data))
	fmt.Println(c)
}

func ExampleNewTillichZemor() {
	data := []byte("Hello, world!")
	c := NewTillichZemor(tz.Sum(data))
	fmt.Println(c)
}

func ExampleNewFromHash() {
	data := []byte("Hello, world!")
	for _, tc := range []struct {
		typ     Type
		newHash func() hash.Hash
	}{
		{SHA256, sha256.New},
		{TillichZemor, tz.New},
	} {
		h := tc.newHash()
		h.Write(data)
		cs := NewFromHash(tc.typ, h)
		fmt.Println(cs)
	}
}

func ExampleNewFromData() {
	data := []byte("Hello, world!")
	for _, typ := range []Type{
		SHA256,
		TillichZemor,
	} {
		cs, err := NewFromData(typ, data)
		if err != nil {
			fmt.Printf("calculation for %v failed: %v\n", typ, err)
			return
		}
		fmt.Println(cs)
	}
}
