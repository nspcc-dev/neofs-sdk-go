package container

import (
	"crypto/ecdsa"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
)

type (
	// Option represents container option setter.
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

// WithPublicBasicACL returns option to set
// public basic ACL rule.
//
// See also acl.PublicBasicRule.
func WithPublicBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.PublicBasicRule
	}
}

// WithReadOnlyBasicACL returns option to set
// read-only basic ACL rule.
//
// See acl.ReadOnlyBasicRule.
func WithReadOnlyBasicACL() Option {
	return func(option *containerOptions) {
		option.acl = acl.ReadOnlyBasicRule
	}
}

// WithCustomBasicACL returns option to set
// any custom basic ACL rule.
//
// See also documentation to the acl package.
func WithCustomBasicACL(acl acl.BasicACL) Option {
	return func(option *containerOptions) {
		option.acl = acl
	}
}

// WithNonce returns option to set UUID nonce
// to a container.
func WithNonce(nonce uuid.UUID) Option {
	return func(option *containerOptions) {
		option.nonce = nonce
	}
}

// WithOwnerID returns option to set owner
// of a container.
func WithOwnerID(id owner.ID) Option {
	return func(option *containerOptions) {
		option.owner = &id
	}
}

// WithOwnerPublicKey return option to set
// container owner's public key.
func WithOwnerPublicKey(pub *ecdsa.PublicKey) Option {
	return func(option *containerOptions) {
		if option.owner == nil {
			option.owner = new(owner.ID)
		}

		option.owner.SetPublicKey(pub)
	}
}

// WithPolicy returns option to set storage policy
// of a container.
//
// See also policy package.
func WithPolicy(policy *netmap.PlacementPolicy) Option {
	return func(option *containerOptions) {
		option.policy = policy
	}
}

// WithAttribute returns option to set (append if
// provided multiple WithAttribute options) container
// attributes.
func WithAttribute(key, value string) Option {
	return func(option *containerOptions) {
		index := len(option.attributes)
		option.attributes = append(option.attributes, Attribute{})
		option.attributes[index].SetKey(key)
		option.attributes[index].SetValue(value)
	}
}
