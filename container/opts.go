package container

import (
	"crypto/ecdsa"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

type (
	Option func(*containerOptions)

	containerOptions struct {
		acl        acl.BasicACL
		policy     *netmap.PlacementPolicy
		attributes Attributes
		owner      *owner.ID
		nonce      uuid.UUID
	}
)

func defaultContainerOptions() containerOptions {
	rand, err := uuid.NewRandom()
	if err != nil {
		panic("can't create new random " + err.Error())
	}

	return containerOptions{
		acl:   acl.PrivateBasicRule,
		nonce: rand,
	}
}

func WithPublicBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.PublicBasicRule
	}
}

func WithReadOnlyBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.ReadOnlyBasicRule
	}
}

func WithCustomBasicACL(acl acl.BasicACL) Option {
	return func(option *containerOptions) {
		option.acl = acl
	}
}

func WithNonce(nonce uuid.UUID) Option {
	return func(option *containerOptions) {
		option.nonce = nonce
	}
}

func WithOwnerID(id *owner.ID) Option {
	return func(option *containerOptions) {
		option.owner = id
	}
}

func WithOwnerPublicKey(pub *ecdsa.PublicKey) Option {
	return func(option *containerOptions) {
		if option.owner == nil {
			option.owner = new(owner.ID)
		}

		option.owner.SetPublicKey(pub)
	}
}

func WithPolicy(policy *netmap.PlacementPolicy) Option {
	return func(option *containerOptions) {
		option.policy = policy
	}
}

func WithAttribute(key, value string) Option {
	return func(option *containerOptions) {
		index := len(option.attributes)
		option.attributes = append(option.attributes, Attribute{})
		option.attributes[index].SetKey(key)
		option.attributes[index].SetValue(value)
	}
}
