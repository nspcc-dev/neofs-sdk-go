package container

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	protocontainer "github.com/nspcc-dev/neofs-sdk-go/proto/container"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
)

// various attributes.
const (
	sysAttrPrefix          = "__NEOFS__"
	sysAttrDisableHomohash = sysAttrPrefix + "DISABLE_HOMOMORPHIC_HASHING"
	sysAttrDomainName      = sysAttrPrefix + "NAME"
	sysAttrDomainZone      = sysAttrPrefix + "ZONE"
)

// Container represents descriptor of the NeoFS container. Container logically
// stores NeoFS objects. Container is one of the basic and at the same time
// necessary data storage units in the NeoFS. Container includes data about the
// owner, rules for placing objects and other information necessary for the
// system functioning.
//
// Container type instances can represent different container states in the
// system, depending on the context. To create new container in NeoFS zero
// instance SHOULD be declared, initialized using Init method and filled using
// dedicated methods. Once container is saved in the NeoFS network, it can't be
// changed: containers stored in the system are immutable, and NeoFS is a CAS
// of containers that are identified by a fixed length value (see cid.ID type).
// Instances for existing containers can be initialized using decoding methods
// (e.g Unmarshal).
//
// Container is mutually compatible with [protocontainer.Container]
// message. See [Container.FromProtoMessage] / [Container.ProtoMessage] methods.
type Container struct {
	version  *version.Version
	owner    user.ID
	nonce    uuid.UUID
	basicACL acl.Basic
	attrs    [][2]string
	policy   *netmap.PlacementPolicy
}

const (
	attributeName      = "Name"
	attributeTimestamp = "Timestamp"
)

// CopyTo writes deep copy of the [Container] to dst.
func (x Container) CopyTo(dst *Container) {
	dst.SetBasicACL(x.BasicACL())

	dst.owner = x.owner

	if x.version != nil {
		dst.version = new(version.Version)
		*dst.version = *x.version
	} else {
		dst.version = nil
	}

	// do we need to set the different nonce?
	dst.nonce = x.nonce

	if len(x.attrs) > 0 {
		dst.attrs = slices.Clone(x.attrs)
	}

	if x.policy != nil {
		dst.policy = new(netmap.PlacementPolicy)
		x.policy.CopyTo(dst.policy)
	} else {
		dst.policy = nil
	}
}

// reads Container from the container.Container message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field.
func (x *Container) fromProtoMessage(m *protocontainer.Container, checkFieldPresence bool) error {
	var err error

	if m.OwnerId != nil {
		err = x.owner.FromProtoMessage(m.OwnerId)
		if err != nil {
			return fmt.Errorf("invalid owner: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing owner")
	} else {
		x.owner = user.ID{}
	}

	if len(m.Nonce) > 0 {
		err = x.nonce.UnmarshalBinary(m.Nonce)
		if err != nil {
			return fmt.Errorf("invalid nonce: %w", err)
		} else if ver := x.nonce.Version(); ver != 4 {
			return fmt.Errorf("invalid nonce: wrong UUID version %d, expected 4", ver)
		}
	} else if checkFieldPresence {
		return errors.New("missing nonce")
	} else {
		x.nonce = uuid.Nil
	}

	if m.Version != nil {
		if x.version == nil {
			x.version = new(version.Version)
		}
		if err = x.version.FromProtoMessage(m.Version); err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing version")
	} else {
		x.version = nil
	}

	if m.PlacementPolicy != nil {
		if x.policy == nil {
			x.policy = new(netmap.PlacementPolicy)
		}
		err = x.policy.FromProtoMessage(m.PlacementPolicy)
		if err != nil {
			return fmt.Errorf("invalid placement policy: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing placement policy")
	} else {
		x.policy = nil
	}

	attrs := m.GetAttributes()
	mAttr := make(map[string]struct{}, len(attrs))
	var key, val string
	var was bool
	if attrs != nil {
		x.attrs = make([][2]string, len(attrs))
	}

	for i := range attrs {
		key = attrs[i].GetKey()
		if key == "" {
			return errors.New("empty attribute key")
		}

		_, was = mAttr[key]
		if was {
			return fmt.Errorf("duplicated attribute %s", key)
		}

		val = attrs[i].GetValue()
		if val == "" {
			return fmt.Errorf("empty %q attribute value", key)
		}

		switch key {
		case attributeTimestamp:
			_, err = strconv.ParseInt(val, 10, 64)
		}

		if err != nil {
			return fmt.Errorf("invalid attribute value %s: %s (%w)", key, val, err)
		}

		mAttr[key] = struct{}{}
		x.attrs[i] = [2]string{key, val}
	}

	x.basicACL.FromBits(m.BasicAcl)

	return nil
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [Container.ProtoMessage].
func (x *Container) FromProtoMessage(m *protocontainer.Container) error {
	return x.fromProtoMessage(m, true)
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [Container.FromProtoMessage].
func (x Container) ProtoMessage() *protocontainer.Container {
	m := &protocontainer.Container{
		BasicAcl: x.basicACL.Bits(),
	}
	if x.version != nil {
		m.Version = x.version.ProtoMessage()
	}
	if !x.owner.IsZero() {
		m.OwnerId = x.owner.ProtoMessage()
	}
	if x.nonce != uuid.Nil {
		m.Nonce = x.nonce[:]
	}
	if x.policy != nil {
		m.PlacementPolicy = x.policy.ProtoMessage()
	}
	if len(x.attrs) > 0 {
		m.Attributes = make([]*protocontainer.Container_Attribute, len(x.attrs))
		for i := range x.attrs {
			m.Attributes[i] = &protocontainer.Container_Attribute{Key: x.attrs[i][0], Value: x.attrs[i][1]}
		}
	}
	return m
}

// Marshal encodes Container into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x Container) Marshal() []byte {
	return neofsproto.Marshal(x)
}

// SignedData returns actual payload to sign.
//
// See also [Container.CalculateSignature].
func (x Container) SignedData() []byte {
	return x.Marshal()
}

// Unmarshal decodes NeoFS API protocol binary format into the Container
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *Container) Unmarshal(data []byte) error {
	return neofsproto.UnmarshalOptional(data, x, (*Container).fromProtoMessage)
}

// MarshalJSON encodes Container into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x Container) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(x)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the Container
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *Container) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSONOptional(data, x, (*Container).fromProtoMessage)
}

// Init initializes all internal data of the Container required by NeoFS API
// protocol. Init MUST be called when creating a new container. Init SHOULD NOT
// be called multiple times. Init SHOULD NOT be called if the Container instance
// is used for decoding only.
func (x *Container) Init() {
	ver := version.Current()
	x.version = &ver
	for {
		if x.nonce = uuid.New(); x.nonce != uuid.Nil {
			break
		}
	}
}

// SetOwner specifies the owner of the Container. Each Container has exactly
// one owner, so SetOwner MUST be called for instances to be saved in the
// NeoFS.
//
// See also Owner.
func (x *Container) SetOwner(owner user.ID) {
	x.owner = owner
}

// Owner returns owner of the Container set using SetOwner.
//
// Zero Container has no owner which is incorrect according to NeoFS API
// protocol.
func (x Container) Owner() user.ID {
	return x.owner
}

// SetBasicACL specifies basic part of the Container ACL. Basic ACL is used
// to control access inside container storage.
//
// See also BasicACL.
func (x *Container) SetBasicACL(basicACL acl.Basic) {
	x.basicACL = basicACL
}

// BasicACL returns basic ACL set using SetBasicACL.
//
// Zero Container has zero basic ACL which structurally correct but doesn't
// make sense since it denies any access to any party.
func (x Container) BasicACL() acl.Basic {
	return x.basicACL
}

// SetPlacementPolicy sets placement policy for the objects within the Container.
// NeoFS storage layer strives to follow the specified policy.
//
// See also PlacementPolicy.
func (x *Container) SetPlacementPolicy(policy netmap.PlacementPolicy) {
	x.policy = &policy
}

// PlacementPolicy returns placement policy set using SetPlacementPolicy.
//
// Zero Container has no placement policy which is incorrect according to
// NeoFS API protocol.
func (x Container) PlacementPolicy() netmap.PlacementPolicy {
	if x.policy != nil {
		return *x.policy
	}
	return netmap.PlacementPolicy{}
}

// SetAttribute sets Container attribute value by key. Both key and value
// MUST NOT be empty. Attributes set by the creator (owner) are most commonly
// ignored by the NeoFS system and used for application layer. Some attributes
// are so-called system or well-known attributes: they are reserved for system
// needs. System attributes SHOULD NOT be modified using SetAttribute, use
// corresponding methods/functions. List of the reserved keys is documented
// in the particular protocol version.
//
// SetAttribute overwrites existing attribute value.
//
// See also Attribute, IterateAttributes.
func (x *Container) SetAttribute(key, value string) {
	if key == "" {
		panic("empty attribute key")
	} else if value == "" {
		panic("empty attribute value")
	}

	for i := range x.attrs {
		if x.attrs[i][0] == key {
			x.attrs[i][1] = value
			return
		}
	}

	x.attrs = append(x.attrs, [2]string{key, value})
}

// Attribute reads value of the Container attribute by key. Empty result means
// attribute absence.
//
// See also SetAttribute, IterateAttributes.
func (x Container) Attribute(key string) string {
	for i := range x.attrs {
		if x.attrs[i][0] == key {
			return x.attrs[i][1]
		}
	}

	return ""
}

// IterateAttributes iterates over all Container attributes and passes them
// into f. The handler MUST NOT be nil.
//
// See also [Container.SetAttribute], [Container.Attribute], [Container.IterateUserAttributes].
func (x Container) IterateAttributes(f func(key, val string)) {
	for i := range x.attrs {
		f(x.attrs[i][0], x.attrs[i][1])
	}
}

// IterateUserAttributes iterates over user attributes of the Container and
// passes them into f. The handler MUST NOT be nil.
//
// See also [Container.SetAttribute], [Container.Attribute], [Container.IterateAttributes].
func (x Container) IterateUserAttributes(f func(key, val string)) {
	x.IterateAttributes(func(key, val string) {
		if !strings.HasPrefix(key, sysAttrPrefix) {
			f(key, val)
		}
	})
}

// SetName sets human-readable name of the Container. Name MUST NOT be empty.
//
// See also Name.
func (x *Container) SetName(name string) {
	x.SetAttribute(attributeName, name)
}

// Name returns container name set using SetName.
//
// Zero Container has no name.
func (x Container) Name() string {
	return x.Attribute(attributeName)
}

// SetCreationTime writes container's creation time in Unix Timestamp format.
//
// See also CreatedAt.
func (x *Container) SetCreationTime(t time.Time) {
	x.SetAttribute(attributeTimestamp, strconv.FormatInt(t.Unix(), 10))
}

// CreatedAt returns container's creation time set using SetCreationTime.
//
// Zero Container has zero timestamp (in seconds).
func (x Container) CreatedAt() time.Time {
	var sec int64

	attr := x.Attribute(attributeTimestamp)
	if attr != "" {
		var err error

		sec, err = strconv.ParseInt(x.Attribute(attributeTimestamp), 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parse container timestamp: %v", err))
		}
	}

	return time.Unix(sec, 0)
}

const attributeHomoHashEnabled = "true"

// DisableHomomorphicHashing sets flag to disable homomorphic hashing of the
// Container data.
//
// See also IsHomomorphicHashingDisabled.
func (x *Container) DisableHomomorphicHashing() {
	x.SetAttribute(sysAttrDisableHomohash, attributeHomoHashEnabled)
}

// IsHomomorphicHashingDisabled checks if DisableHomomorphicHashing was called.
//
// Zero Container has enabled hashing.
func (x Container) IsHomomorphicHashingDisabled() bool {
	return x.Attribute(sysAttrDisableHomohash) == attributeHomoHashEnabled
}

// Domain represents information about container domain registered in the NNS
// contract deployed in the NeoFS network.
type Domain struct {
	name, zone string
}

// SetName sets human-friendly container domain name.
func (x *Domain) SetName(name string) {
	x.name = name
}

// Name returns name set using SetName.
//
// Zero Domain has zero name.
func (x Domain) Name() string {
	return x.name
}

// SetZone sets zone which is used as a TLD of a domain name in NNS contract.
func (x *Domain) SetZone(zone string) {
	x.zone = zone
}

// Zone returns domain zone set using SetZone.
//
// Zero Domain has "container" zone.
func (x Domain) Zone() string {
	if x.zone != "" {
		return x.zone
	}

	return "container"
}

// WriteDomain writes Domain into the Container. Name MUST NOT be empty.
func (x *Container) WriteDomain(domain Domain) {
	x.SetAttribute(sysAttrDomainName, domain.Name())
	x.SetAttribute(sysAttrDomainZone, domain.Zone())
}

// ReadDomain reads Domain from the Container. Returns value with empty name
// if domain is not specified.
func (x Container) ReadDomain() (res Domain) {
	name := x.Attribute(sysAttrDomainName)
	if name != "" {
		res.SetName(name)
		res.SetZone(x.Attribute(sysAttrDomainZone))
	}

	return
}

// CalculateSignature calculates signature of the [Container] using provided signer
// and writes it into dst. Signature instance MUST NOT be nil. CalculateSignature
// is expected to be called after all the [Container] data is filled and before
// saving the [Container] in the NeoFS network. Note that мany subsequent change
// will most likely break the signature. signer MUST be of
// [neofscrypto.ECDSA_DETERMINISTIC_SHA256] scheme, for example, [neofsecdsa.SignerRFC6979]
// can be used.
//
// See also [Container.VerifySignature], [Container.SignedData].
//
// Returned errors:
//   - [neofscrypto.ErrIncorrectSigner]
func (x Container) CalculateSignature(dst *neofscrypto.Signature, signer neofscrypto.Signer) error {
	if signer.Scheme() != neofscrypto.ECDSA_DETERMINISTIC_SHA256 {
		return fmt.Errorf("%w: expected ECDSA_DETERMINISTIC_SHA256 scheme", neofscrypto.ErrIncorrectSigner)
	}
	return dst.Calculate(signer, x.Marshal())
}

// VerifySignature verifies Container signature calculated using CalculateSignature.
// Result means signature correctness.
func (x Container) VerifySignature(sig neofscrypto.Signature) bool {
	return sig.Verify(x.Marshal())
}

// CalculateID encodes the given Container and passes the result into FromBinary.
//
// See also Container.Marshal, AssertID.
// Deprecated: use cid.NewFromMarshalledContainer(x.Marshal()) instead.
func (x Container) CalculateID(dst *cid.ID) {
	*dst = cid.NewFromMarshalledContainer(x.Marshal())
}

// AssertID checks if the given Container matches its identifier in CAS of the
// NeoFS containers.
//
// See also CalculateID.
func (x Container) AssertID(id cid.ID) bool {
	var id2 cid.ID
	x.CalculateID(&id2)

	return id2 == id
}

// Version returns the NeoFS API version this container was created with.
func (x Container) Version() version.Version {
	if x.version != nil {
		return *x.version
	}
	return version.Version{}
}
