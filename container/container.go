package container

import (
	"crypto/sha256"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-api-go/v2/container"
	"github.com/nspcc-dev/neofs-sdk-go/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/owner"
	"github.com/nspcc-dev/neofs-sdk-go/session"
	"github.com/nspcc-dev/neofs-sdk-go/signature"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// Container represents in-memory descriptor of the NeoFS container.
//
// Container is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/container.Container
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration or via .
type Container struct {
	v2 container.Container

	token *session.Token

	sig *signature.Signature
}

// ReadFromV2 reads Container from the container.Container message.
//
// Does not verify if the message meets NeoFS API V2 specification.
//
// See also WriteToV2.
func (c *Container) ReadFromV2(m container.Container) {
	if c != nil {
		c.v2 = m
	}
}

// WriteToV2 writes Container to the container.Container message.
// The message must not be nil.
//
// See also ReadFromV2.
func (c Container) WriteToV2(m *container.Container) {
	*m = c.v2
}

// InitCreation creates, initializes and returns blank Container instance.
//
// Defaults:
//  - token: nil;
//  - sig: nil;
//  - basicACL: acl.PrivateBasicRule;
//  - version: version.Current;
//  - nonce: random UUID;
//  - attr: nil;
//  - policy: nil;
//  - ownerID: nil.
func InitCreation(opts ...Option) *Container {
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
	cnr.SetVersion(version.Current())

	return cnr
}

// CalculateID calculates container identifier
// based on its structure.
func (c Container) CalculateID() *cid.ID {
	var v2 container.Container
	c.WriteToV2(&v2)

	data, err := v2.StableMarshal(nil)
	if err != nil {
		panic(err)
	}

	var id cid.ID
	id.SetSHA256(sha256.Sum256(data))

	return &id
}

// Version returns version of the NeoFS API
// protocol by which the container was created.
//
// Zero Container has version.Current version.
//
// See also SetVersion.
func (c Container) Version() *version.Version {
	return version.NewFromV2(c.v2.GetVersion())
}

// SetVersion sets container's version.
// Version must not be nil.
//
// See also Version.
func (c *Container) SetVersion(v *version.Version) {
	c.v2.SetVersion(v.ToV2())
}

// OwnerID returns container's owner.
//
// Zero Container has nil value.
//
// See also SetOwnerID.
func (c Container) OwnerID() *owner.ID {
	return owner.NewIDFromV2(c.v2.GetOwnerID())
}

// SetOwnerID sets container's owner.
// Owner must not be nil.
//
// See also OwnerID.
func (c *Container) SetOwnerID(v *owner.ID) {
	c.v2.SetOwnerID(v.ToV2())
}

// NonceUUID returns container's nonce in UUID format.
//
// Returns error if container nonce is not a valid UUID.
//
// See also SetNonceUUID.
func (c Container) NonceUUID() (uuid.UUID, error) {
	return uuid.FromBytes(c.v2.GetNonce())
}

// SetNonceUUID sets container nonce as UUID.
//
// See also NonceUUID.
func (c *Container) SetNonceUUID(v uuid.UUID) {
	data, _ := v.MarshalBinary()
	c.v2.SetNonce(data)
}

// BasicACL returns container's basic ACL.
//
// Zero Container has PrivateBasicRule ALC.
//
// See also SetBasicACL.
func (c Container) BasicACL() acl.BasicACL {
	return acl.BasicACL(c.v2.GetBasicACL())
}

// SetBasicACL sets basic ALC rule.
//
// See also BasicACL.
func (c *Container) SetBasicACL(v acl.BasicACL) {
	c.v2.SetBasicACL(uint32(v))
}

// Attributes returns container's attributes.
//
// Zero Container has nil attributes.
//
// See also SetAttributes.
func (c Container) Attributes() Attributes {
	attrsV2 := c.v2.GetAttributes()
	if attrsV2 == nil {
		return nil
	}

	var attrs Attributes
	attrs.ReadFromV2(attrsV2)

	return attrs
}

// SetAttributes sets container's attributes.
// Attributes must not be nil.
//
// See also Attributes.
func (c *Container) SetAttributes(v Attributes) {
	var attrsV2 []container.Attribute
	v.WriteToV2(&attrsV2)

	c.v2.SetAttributes(attrsV2)
}

// PlacementPolicy returns container's placement policy.
//
// Zero Container has nil policy.
//
// See also SetPlacementPolicy.
func (c Container) PlacementPolicy() *netmap.PlacementPolicy {
	return netmap.NewPlacementPolicyFromV2(c.v2.GetPlacementPolicy())
}

// SetPlacementPolicy sets placement policy.
// Policy must not be nil.
//
// See also PlacementPolicy.
func (c *Container) SetPlacementPolicy(v *netmap.PlacementPolicy) {
	c.v2.SetPlacementPolicy(v.ToV2())
}

// SessionToken returns token of the session within
// which container was created.
//
// Zero Container has nil token.
//
// See also SetSessionToken.
func (c Container) SessionToken() *session.Token {
	return c.token
}

// SetSessionToken sets token of the session within
// which container was created.
//
// See also SessionToken.
func (c *Container) SetSessionToken(t *session.Token) {
	c.token = t
}

// Signature returns signature of the marshaled container.
//
// Zero Container has nil signature.
//
// See also SetSignature.
func (c Container) Signature() *signature.Signature {
	return c.sig
}

// SetSignature sets signature of the marshaled container.
//
// See also Signature.
func (c *Container) SetSignature(sig *signature.Signature) {
	c.sig = sig
}

// Marshal marshals Container into a protobuf binary form.
func (c Container) Marshal() ([]byte, error) {
	return c.v2.StableMarshal(nil)
}

// Unmarshal unmarshals protobuf binary representation of Container.
func (c *Container) Unmarshal(data []byte) error {
	return c.v2.Unmarshal(data)
}

// MarshalJSON encodes Container to protobuf JSON format.
func (c Container) MarshalJSON() ([]byte, error) {
	return c.v2.MarshalJSON()
}

// UnmarshalJSON decodes Container from protobuf JSON format.
func (c *Container) UnmarshalJSON(data []byte) error {
	return c.v2.UnmarshalJSON(data)
}
