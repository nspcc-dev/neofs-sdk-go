package object

import (
	"bytes"
	"slices"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/session"
)

func cloneSignature(src *neofscrypto.Signature) *neofscrypto.Signature {
	if src == nil {
		return nil
	}
	s := neofscrypto.NewSignatureFromRawKey(src.Scheme(), bytes.Clone(src.PublicKeyBytes()), src.Value())
	return &s
}

func cloneChecksum(src *checksum.Checksum) *checksum.Checksum {
	if src == nil {
		return nil
	}
	s := checksum.New(src.Type(), bytes.Clone(src.Value()))
	return &s
}

func (x split) copyTo(dst *split) {
	dst.parID = x.parID
	dst.prev = x.prev
	dst.id = x.id
	dst.children = slices.Clone(x.children)
	dst.first = x.first
	dst.parSig = cloneSignature(x.parSig)
	if x.parHdr != nil {
		if dst.parHdr == nil {
			dst.parHdr = new(header)
		}
		x.parHdr.copyTo(dst.parHdr)
	} else {
		dst.parHdr = nil
	}
}

func (x header) copyTo(dst *header) {
	dst.cnr = x.cnr
	dst.owner = x.owner
	dst.created = x.created
	dst.payloadLn = x.payloadLn
	dst.typ = x.typ
	dst.attrs = slices.Clone(x.attrs)
	dst.pldHash = cloneChecksum(x.pldHash)
	dst.pldHomoHash = cloneChecksum(x.pldHomoHash)
	x.split.copyTo(&dst.split)
	if x.version != nil {
		ver := *x.version
		dst.version = &ver
	} else {
		dst.version = nil
	}
	if x.session != nil {
		if dst.session == nil {
			dst.session = new(session.Object)
		}
		x.session.CopyTo(dst.session)
	} else {
		dst.session = nil
	}
}
