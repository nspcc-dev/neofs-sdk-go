package container

import (
	"crypto/sha256"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	v2netmap "github.com/nspcc-dev/neofs-api-go/v2/netmap"
	"github.com/nspcc-dev/neofs-api-go/v2/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

type Container struct {
	v2 container.Container
}

// New creates, initializes and returns blank Container instance.
//
// Defaults:
//  - token: nil;
//  - sig: nil;
//  - basicACL: Private;
//  - version: version.Current;
//  - nonce: random UUID;
//  - attr: nil;
//  - policy: nil;
//  - ownerID: nil.
func New(opts ...Option) *Container {
	cnrOptions := defaultContainerOptions()

	for i := range opts {
		opts[i](&cnrOptions)
	}

	cnr := new(Container)
	cnr.SetNonceUUID(cnrOptions.nonce)
	cnr.SetBasicACL(cnrOptions.acl)

	if cnrOptions.owner != nil {
		cnr.SetOwnerID(cnrOptions.owner)
	}

	if cnrOptions.policy != nil {
		cnr.SetPlacementPolicy(cnrOptions.policy)
	}

	cnr.SetAttributes(cnrOptions.attributes)
	ver := version.Current()
	cnr.SetVersion(&ver)

	return cnr
}

// ToV2 returns the v2 Container message.
//
// Nil Container converts to nil.
func (c *Container) ToV2() *container.Container {
	if c == nil {
		return nil
	}

	return &c.v2
}

// NewVerifiedFromV2 constructs Container from NeoFS API V2 Container message.
//
// Does not perform if message meets NeoFS API V2 specification. To do this
// use NewVerifiedFromV2 constructor.
func NewContainerFromV2(c *container.Container) *Container {
	cnr := new(Container)

	if c != nil {
		cnr.v2 = *c
	}

	return cnr
}

// CalculateID calculates container identifier
// based on its structure.
func CalculateID(c *Container) cid.ID {
	var id cid.ID
	id.SetSHA256(sha256.Sum256(c.ToV2().StableMarshal(nil)))

	return id
}

func (c *Container) Version() *version.Version {
	var ver version.Version
	if v2ver := c.v2.GetVersion(); v2ver != nil {
		ver.ReadFromV2(*c.v2.GetVersion())
	}
	return &ver
}

func (c *Container) SetVersion(v *version.Version) {
	var verV2 refs.Version
	v.WriteToV2(&verV2)
	c.v2.SetVersion(&verV2)
}

func (c *Container) OwnerID() *user.ID {
	m := c.v2.GetOwnerID()
	if m == nil {
		return nil
	}

	var id user.ID

	_ = id.ReadFromV2(*m)

	return &id
}

func (c *Container) SetOwnerID(v *user.ID) {
	var m refs.OwnerID
	v.WriteToV2(&m)

	c.v2.SetOwnerID(&m)
}

// Returns container nonce in UUID format.
//
// Returns error if container nonce is not a valid UUID.
func (c *Container) NonceUUID() (uuid.UUID, error) {
	return uuid.FromBytes(c.v2.GetNonce())
}

// SetNonceUUID sets container nonce as UUID.
func (c *Container) SetNonceUUID(v uuid.UUID) {
	data, _ := v.MarshalBinary()
	c.v2.SetNonce(data)
}

func (c *Container) BasicACL() (res acl.Basic) {
	res.FromBits(c.v2.GetBasicACL())
	return
}

func (c *Container) SetBasicACL(v acl.Basic) {
	c.v2.SetBasicACL(v.Bits())
}

func (c *Container) Attributes() Attributes {
	return NewAttributesFromV2(c.v2.GetAttributes())
}

func (c *Container) SetAttributes(v Attributes) {
	c.v2.SetAttributes(v.ToV2())
}

func (c *Container) PlacementPolicy() *netmap.PlacementPolicy {
	m := c.v2.GetPlacementPolicy()
	if m == nil {
		return nil
	}

	var p netmap.PlacementPolicy
	// FIXME(@cthulhu-rider): #225 handle error
	err := p.ReadFromV2(*m)
	if err != nil {
		panic(err)
	}

	return &p
}

func (c *Container) SetPlacementPolicy(v *netmap.PlacementPolicy) {
	var m *v2netmap.PlacementPolicy

	if v != nil {
		m = new(v2netmap.PlacementPolicy)
		v.WriteToV2(m)
	}

	c.v2.SetPlacementPolicy(m)
}

// Marshal marshals Container into a protobuf binary form.
func (c *Container) Marshal() ([]byte, error) {
	return c.v2.StableMarshal(nil), nil
}

// Unmarshal unmarshals protobuf binary representation of Container.
func (c *Container) Unmarshal(data []byte) error {
	return c.v2.Unmarshal(data)
}

// MarshalJSON encodes Container to protobuf JSON format.
func (c *Container) MarshalJSON() ([]byte, error) {
	return c.v2.MarshalJSON()
}

// UnmarshalJSON decodes Container from protobuf JSON format.
func (c *Container) UnmarshalJSON(data []byte) error {
	return c.v2.UnmarshalJSON(data)
}
