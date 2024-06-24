package container

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nspcc-dev/neofs-sdk-go/api/container"
	apinetmap "github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/api/refs"
	"github.com/nspcc-dev/neofs-sdk-go/container/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/netmap"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Container represents descriptor of the NeoFS container. Container logically
// stores NeoFS objects. Container is one of the basic and at the same time
// necessary data storage units in the NeoFS. Container includes data about the
// owner, rules for placing objects and other information necessary for the
// system functioning.
//
// Container type instances can represent different container states in the
// system, depending on the context. To create new container in NeoFS zero
// instance should be initialized using [New] and finalized using dedicated
// methods. Once container is saved in the NeoFS network, it can't be changed:
// containers stored in the system are immutable, and NeoFS is a CAS of
// containers that are identified by a fixed length value (see [cid.ID] type).
// Instances for existing containers can be initialized using decoding methods
// (e.g [Container.Unmarshal]).
//
// Container is mutually compatible with [container.Container] message. See
// [Container.ReadFromV2] / [Container.WriteToV2] methods.
type Container struct {
	versionSet bool
	version    version.Version

	nonceSet bool
	nonce    uuid.UUID

	ownerSet bool
	owner    user.ID

	basicACL acl.Basic

	policySet bool
	policy    netmap.PlacementPolicy

	attrs []*container.Container_Attribute
}

// Various well-known container attributes widely used by applications.
const (
	attributeName      = "Name"
	attributeTimestamp = "Timestamp"
)

// System container attributes.
const (
	sysAttributePrefix          = "__NEOFS__"
	sysAttributeDomainName      = sysAttributePrefix + "NAME"
	sysAttributeDomainZone      = sysAttributePrefix + "ZONE"
	sysAttributeDisableHomoHash = sysAttributePrefix + "DISABLE_HOMOMORPHIC_HASHING"
)

// New constructs new Container instance.
func New(owner user.ID, basicACL acl.Basic, policy netmap.PlacementPolicy) Container {
	return Container{
		versionSet: true,
		version:    version.Current,
		nonceSet:   true,
		nonce:      uuid.New(),
		ownerSet:   true,
		owner:      owner,
		basicACL:   basicACL,
		policySet:  true,
		policy:     policy,
	}
}

// CopyTo writes deep copy of the [Container] to dst.
func (x Container) CopyTo(dst *Container) {
	dst.versionSet = x.versionSet
	dst.version = x.version
	dst.nonceSet = x.nonceSet
	dst.nonce = x.nonce
	dst.ownerSet = x.ownerSet
	dst.owner = x.owner
	dst.SetBasicACL(x.BasicACL())
	dst.policySet = x.policySet
	x.policy.CopyTo(&dst.policy)

	if x.attrs != nil {
		dst.attrs = make([]*container.Container_Attribute, len(x.attrs))
		for i := range x.attrs {
			if x.attrs[i] != nil {
				dst.attrs[i] = &container.Container_Attribute{Key: x.attrs[i].Key, Value: x.attrs[i].Value}
			} else {
				dst.attrs[i] = nil
			}
		}
	} else {
		dst.attrs = nil
	}
}

func (x *Container) readFromV2(m *container.Container, checkFieldPresence bool) error {
	var err error
	if x.ownerSet = m.OwnerId != nil; x.ownerSet {
		err = x.owner.ReadFromV2(m.OwnerId)
		if err != nil {
			return fmt.Errorf("invalid owner: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing owner")
	}

	if x.nonceSet = len(m.Nonce) > 0; x.nonceSet {
		err = x.nonce.UnmarshalBinary(m.Nonce)
		if err != nil {
			return fmt.Errorf("invalid nonce: %w", err)
		} else if ver := x.nonce.Version(); ver != 4 {
			return fmt.Errorf("invalid nonce: wrong UUID version %d", ver)
		}
	} else if checkFieldPresence {
		return errors.New("missing nonce")
	}

	if x.versionSet = m.Version != nil; x.versionSet {
		err = x.version.ReadFromV2(m.Version)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing version")
	}

	if x.policySet = m.PlacementPolicy != nil; x.policySet {
		err = x.policy.ReadFromV2(m.PlacementPolicy)
		if err != nil {
			return fmt.Errorf("invalid placement policy: %w", err)
		}
	} else if checkFieldPresence {
		return errors.New("missing placement policy")
	}

	attrs := m.GetAttributes()
	var key string
	for i := range attrs {
		key = attrs[i].GetKey()
		if key == "" {
			return fmt.Errorf("invalid attribute #%d: missing key", i)
		} // also prevents further NPE
		for j := 0; j < i; j++ {
			if attrs[j].Key == key {
				return fmt.Errorf("multiple attributes with key=%s", key)
			}
		}
		if attrs[i].Value == "" {
			return fmt.Errorf("invalid attribute #%d (%s): missing value", i, key)
		}
		switch key {
		case attributeTimestamp:
			_, err = strconv.ParseInt(attrs[i].Value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid timestamp attribute (#%d): invalid integer (%w)", i, err)
			}
		}
	}

	x.basicACL.FromBits(m.BasicAcl)
	x.attrs = attrs

	return nil
}

// ReadFromV2 reads Container from the [container.Container] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Container.WriteToV2].
func (x *Container) ReadFromV2(m *container.Container) error {
	return x.readFromV2(m, true)
}

// WriteToV2 writes Container to the [container.Container] message of the NeoFS
// API protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [Container.ReadFromV2].
func (x Container) WriteToV2(m *container.Container) {
	if x.versionSet {
		m.Version = new(refs.Version)
		x.version.WriteToV2(m.Version)
	} else {
		m.Version = nil
	}
	if x.ownerSet {
		m.OwnerId = new(refs.OwnerID)
		x.owner.WriteToV2(m.OwnerId)
	} else {
		m.OwnerId = nil
	}
	if x.nonceSet {
		m.Nonce = x.nonce[:]
	} else {
		m.Nonce = nil
	}
	if x.policySet {
		m.PlacementPolicy = new(apinetmap.PlacementPolicy)
		x.policy.WriteToV2(m.PlacementPolicy)
	} else {
		m.PlacementPolicy = nil
	}
	m.BasicAcl = x.basicACL.Bits()
	m.Attributes = x.attrs
}

// Marshal encodes Container into a binary format of the NeoFS API protocol
// (Protocol Buffers V3 with direct field order).
//
// See also [Container.Unmarshal].
func (x Container) Marshal() []byte {
	var m container.Container
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the Container. Returns
// an error describing a format violation of the specified fields. Unmarshal
// does not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [Container.Marshal].
func (x *Container) Unmarshal(data []byte) error {
	var m container.Container
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}
	return x.readFromV2(&m, false)
}

// MarshalJSON encodes Container into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [Container.UnmarshalJSON].
func (x Container) MarshalJSON() ([]byte, error) {
	var m container.Container
	x.WriteToV2(&m)
	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the Container
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [Container.MarshalJSON].
func (x *Container) UnmarshalJSON(data []byte) error {
	var m container.Container
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}

	return x.readFromV2(&m, false)
}

// SetOwner specifies the owner of the Container. Each Container has exactly one
// owner.
//
// See also [Container.Owner].
func (x *Container) SetOwner(owner user.ID) {
	x.owner, x.ownerSet = owner, true
}

// Owner returns owner of the Container.
//
// Zero Container has no owner which is incorrect according to NeoFS API
// protocol.
//
// See also [Container.SetOwner].
func (x Container) Owner() user.ID {
	if x.ownerSet {
		return x.owner
	}
	return user.ID{}
}

// SetBasicACL specifies basic part of the Container ACL. Basic ACL is used
// to control access inside container storage.
//
// See also [Container.BasicACL].
func (x *Container) SetBasicACL(basicACL acl.Basic) {
	x.basicACL = basicACL
}

// BasicACL returns basic ACL of the Container.
//
// Zero Container has zero basic ACL which structurally correct but doesn't
// make sense since it denies any access to any party.
//
// See also [Container.SetBasicACL].
func (x Container) BasicACL() acl.Basic {
	return x.basicACL
}

// SetPlacementPolicy sets placement policy for the objects within the Container.
// NeoFS storage layer strives to follow the specified policy.
//
// See also [Container.PlacementPolicy].
func (x *Container) SetPlacementPolicy(policy netmap.PlacementPolicy) {
	x.policy, x.policySet = policy, true
}

// PlacementPolicy returns placement policy for the objects within the
// Container.
//
// Zero Container has no placement policy which is incorrect according to
// NeoFS API protocol.
//
// See also [Container.SetPlacementPolicy].
func (x Container) PlacementPolicy() netmap.PlacementPolicy {
	if x.policySet {
		return x.policy
	}
	return netmap.PlacementPolicy{}
}

func (x *Container) resetAttribute(key string) {
	for i := 0; i < len(x.attrs); i++ { // do not use range, slice is changed inside
		if x.attrs[i].GetKey() == key {
			x.attrs = append(x.attrs[:i], x.attrs[i+1:]...)
			i--
		}
	}
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
// See also [Container.Attribute], [Container.IterateAttributes].
func (x *Container) SetAttribute(key, value string) {
	if key == "" {
		panic("empty attribute key")
	} else if value == "" {
		panic("empty attribute value")
	}

	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			x.attrs[i].Value = value
			return
		}
	}

	x.attrs = append(x.attrs, &container.Container_Attribute{Key: key, Value: value})
}

// Attribute reads value of the Container attribute by key. Empty result means
// attribute absence.
//
// See also [Container.SetAttribute], [Container.IterateAttributes].
func (x Container) Attribute(key string) string {
	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			return x.attrs[i].GetValue()
		}
	}

	return ""
}

// NumberOfAttributes returns number of all attributes specified for this
// Container.
//
// See also [NodeInfo.SetAttribute], [Container.IterateAttributes].
func (x Container) NumberOfAttributes() int {
	return len(x.attrs)
}

// IterateAttributes iterates over all Container attributes and passes them
// into f. The handler MUST NOT be nil.
//
// See also [Container.SetAttribute], [Container.Attribute],
// [Container.NumberOfAttributes], [Container.IterateUserAttributes].
func (x Container) IterateAttributes(f func(key, val string)) {
	for i := range x.attrs {
		f(x.attrs[i].GetKey(), x.attrs[i].GetValue())
	}
}

// IterateUserAttributes iterates over user attributes of the Container and
// passes them into f. The handler MUST NOT be nil.
//
// See also [Container.SetAttribute], [Container.Attribute], [Container.IterateAttributes].
func (x Container) IterateUserAttributes(f func(key, val string)) {
	x.IterateAttributes(func(key, val string) {
		if !strings.HasPrefix(key, sysAttributePrefix) {
			f(key, val)
		}
	})
}

// SetName sets human-readable name of the Container. Name MUST NOT be empty.
//
// See also [Container.Name].
func (x *Container) SetName(name string) {
	x.SetAttribute(attributeName, name)
}

// Name returns human-readable container name.
//
// Zero Container has no name.
//
// See also [Container.SetName].
func (x Container) Name() string {
	return x.Attribute(attributeName)
}

// SetCreationTime writes container's creation time in Unix Timestamp format.
//
// See also [Container.CreatedAt].
func (x *Container) SetCreationTime(t time.Time) {
	x.SetAttribute(attributeTimestamp, strconv.FormatInt(t.Unix(), 10))
}

// CreatedAt returns container's creation time in Unix Timestamp format.
//
// Zero Container has zero timestamp (in seconds).
//
// See also [Container.SetCreationTime].
func (x Container) CreatedAt() time.Time {
	var sec int64
	if s := x.Attribute(attributeTimestamp); s != "" {
		var err error
		sec, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("parse timestamp attribute: %v", err))
		}
	}
	return time.Unix(sec, 0)
}

const attributeHomoHashEnabled = "true"

// SetHomomorphicHashingDisabled sets flag indicating whether homomorphic
// hashing of the Container objects in the network is disabled.
//
// See also [Container.HomomorphicHashingDisabled].
func (x *Container) SetHomomorphicHashingDisabled(v bool) {
	if v {
		x.SetAttribute(sysAttributeDisableHomoHash, attributeHomoHashEnabled)
	} else {
		x.resetAttribute(sysAttributeDisableHomoHash)
	}
}

// HomomorphicHashingDisabled returns flag indicating whether homomorphic
// hashing of the Container objects in the network is disabled.
//
// Zero Container has enabled hashing.
//
// See also [Container.SetHomomorphicHashingDisabled].
func (x Container) HomomorphicHashingDisabled() bool {
	return x.Attribute(sysAttributeDisableHomoHash) == attributeHomoHashEnabled
}

// Domain represents information about container domain registered in the NNS
// contract deployed in the NeoFS network.
type Domain struct {
	name, zone string
}

const defaultDomainZone = "container"

// SetName sets human-friendly container domain name.
//
// See also [Domain.Name].
func (x *Domain) SetName(name string) {
	x.name = name
}

// Name returns human-friendly container domain name.
//
// Zero Domain has zero name.
//
// See also [Domain.SetName].
func (x Domain) Name() string {
	return x.name
}

// SetZone sets zone which is used as a TLD of a domain name in NNS contract.
//
// See also [Domain.Zone].
func (x *Domain) SetZone(zone string) {
	x.zone = zone
}

// Zone returns zone which is used as a TLD of a domain name in NNS contract.
//
// Zero Domain has "container" zone.
//
// See also [Domain.SetZone].
func (x Domain) Zone() string {
	if x.zone != "" {
		return x.zone
	}
	return defaultDomainZone
}

// SetDomain specifies Domain associated with the Container. Name MUST NOT be
// empty.
//
// See also [Container.Domain].
func (x *Container) SetDomain(domain Domain) {
	x.SetAttribute(sysAttributeDomainName, domain.Name())
	if domain.zone != "" && domain.zone != defaultDomainZone {
		x.SetAttribute(sysAttributeDomainZone, domain.zone)
	} else {
		x.resetAttribute(sysAttributeDomainZone)
	}
}

// Domain returns Domain associated with the Container. Returns value with empty
// name if domain is not specified.
//
// See also [Container.SetDomain].
func (x Container) Domain() Domain {
	var res Domain
	name := x.Attribute(sysAttributeDomainName)
	if name != "" {
		res.SetName(name)
		res.SetZone(x.Attribute(sysAttributeDomainZone))
	}
	return res
}

// CalculateID calculates and returns CAS ID for the given container.
func CalculateID(cnr Container) cid.ID {
	return sha256.Sum256(cnr.Marshal())
}

// Version returns the NeoFS API version this container was created with.
func (x Container) Version() version.Version {
	if x.versionSet {
		return x.version
	}
	return version.Version{}
}
