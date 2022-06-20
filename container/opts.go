package container

import (
	"crypto/ecdsa"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

type (
	Option func(*containerOptions)

	containerOptions struct {
		acl        acl.Basic
		policy     *netmap.PlacementPolicy
		attributes Attributes
		owner      *user.ID
		nonce      uuid.UUID
	}
)

func defaultContainerOptions() containerOptions {
	rand, err := uuid.NewRandom()
	if err != nil {
		panic("can't create new random " + err.Error())
	}

	return containerOptions{
		acl:   acl.Private,
		nonce: rand,
	}
}

func WithPublicBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.PublicRW
	}
}

func WithReadOnlyBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.PublicRO
	}
}

func WithCustomBasicACL(acl acl.Basic) Option {
	return func(option *containerOptions) {
		option.acl = acl
	}
}

func WithNonce(nonce uuid.UUID) Option {
	return func(option *containerOptions) {
		option.nonce = nonce
	}
}

func WithOwnerID(id *user.ID) Option {
	return func(option *containerOptions) {
		option.owner = id
	}
}

func WithOwnerPublicKey(pub *ecdsa.PublicKey) Option {
	return func(option *containerOptions) {
		if option.owner == nil {
			option.owner = new(user.ID)
		}

		user.IDFromKey(option.owner, *pub)
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
