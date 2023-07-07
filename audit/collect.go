package audit

import (
	"context"
	"errors"
	"fmt"

	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/client"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/object/relations"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	"github.com/nspcc-dev/tzhash/tz"
)

// CollectMembers creates new storage group structure and fills it
// with information about members collected via HeadReceiver.
//
// Resulting storage group consists of physically stored objects only.
func CollectMembers(
	ctx context.Context,
	collector relations.Executor,
	cnr cid.ID,
	members []oid.ID,
	tokens relations.Tokens,
	calcHomoHash bool,
	signer neofscrypto.Signer,
) (*storagegroup.StorageGroup, error) {
	var (
		err        error
		sumPhySize uint64
		phyMembers []oid.ID
		phyHashes  [][]byte
		addr       oid.Address
		sg         storagegroup.StorageGroup
	)

	addr.SetContainer(cnr)

	for i := range members {
		if phyMembers, _, err = relations.Get(ctx, collector, cnr, members[i], tokens, signer); err != nil {
			return nil, err
		}

		var prmHead client.PrmObjectHead
		for _, phyMember := range phyMembers {
			addr.SetObject(phyMember)
			leaf, err := collector.ObjectHead(ctx, addr.Container(), addr.Object(), signer, prmHead)
			if err != nil {
				return nil, fmt.Errorf("head phy member '%s': %w", phyMember.EncodeToString(), err)
			}

			var hdr object.Object
			if !leaf.ReadHeader(&hdr) {
				return nil, errors.New("header err")
			}

			sumPhySize += hdr.PayloadSize()
			cs, _ := hdr.PayloadHomomorphicHash()

			if calcHomoHash {
				phyHashes = append(phyHashes, cs.Value())
			}
		}
	}

	sg.SetMembers(phyMembers)
	sg.SetValidationDataSize(sumPhySize)

	if calcHomoHash {
		sumHash, err := tz.Concat(phyHashes)
		if err != nil {
			return nil, err
		}

		var cs checksum.Checksum
		tzHash := [64]byte{}
		copy(tzHash[:], sumHash)
		cs.SetTillichZemor(tzHash)

		sg.SetValidationDataHash(cs)
	}

	return &sg, nil
}
