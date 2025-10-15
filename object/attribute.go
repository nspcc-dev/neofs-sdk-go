package object

import (
	"errors"
	"fmt"
	"strconv"

	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	protoobject "github.com/nspcc-dev/neofs-sdk-go/proto/object"
)

// Various system attributes.
const (
	sysAttrPrefix = "__NEOFS__"
	// AttributeExpirationEpoch is a key to an object attribute that determines
	// after what epoch the object becomes expired. Objects that do not have this
	// attribute never expire.
	//
	// Reaction of NeoFS system components to the objects' 'expired' property may
	// vary. For example, in the basic scenario, expired objects are auto-deleted
	// from the storage. Detailed behavior can be found in the NeoFS Specification.
	//
	// Note that the value determines exactly the last epoch of the object's
	// relevance: for example, with the value N, the object is relevant in epoch N
	// and expired in any epoch starting from N+1.
	AttributeExpirationEpoch = sysAttrPrefix + "EXPIRATION_EPOCH"
	// AttributeAssociatedObject is a key to an object attribute that determines
	// associated object's ID. For [object.TypeTombstone] it defines object to
	// delete. For [object.TypeLock] it defines object to lock.
	AttributeAssociatedObject = sysAttrPrefix + "ASSOCIATE"
	// AttributeECPrefix is a prefix of EC object attributes.
	AttributeECPrefix = sysAttrPrefix + "EC_"
	// AttributeECRuleIndex is an attribute of EC part object specifying index of EC
	// rule in container storage policy according to which the part was created.
	// Value is base-10 integer.
	AttributeECRuleIndex = AttributeECPrefix + "RULE_IDX"
	// AttributeECPartIndex is an attribute of EC part object specifying index in
	// the EC parts into which the parent object is divided. Value is base-10
	// integer.
	AttributeECPartIndex = AttributeECPrefix + "PART_IDX"
)

// Attribute represents an object attribute.
type Attribute struct{ k, v string }

// NewAttribute creates and initializes new [Attribute].
func NewAttribute(key, value string) Attribute {
	return Attribute{key, value}
}

// Key returns key to the object attribute.
func (a *Attribute) Key() string {
	return a.k
}

// SetKey sets key to the object attribute.
func (a *Attribute) SetKey(v string) {
	a.k = v
}

// Value return value of the object attribute.
func (a *Attribute) Value() string {
	return a.v
}

// SetValue sets value of the object attribute.
func (a *Attribute) SetValue(v string) {
	a.v = v
}

// fromProtoMessage validates m according to the NeoFS API protocol and restores
// a from it.
func (a *Attribute) fromProtoMessage(m *protoobject.Header_Attribute, checkFieldPresence bool) error {
	if checkFieldPresence && m.Key == "" {
		return fmt.Errorf("missing key")
	}
	if checkFieldPresence && m.Value == "" {
		return errors.New("missing value")
	}
	switch m.Key {
	case AttributeExpirationEpoch:
		if _, err := strconv.ParseUint(m.Value, 10, 64); err != nil {
			return fmt.Errorf("invalid expiration epoch (must be a uint): %w", err)
		}
	case AttributeAssociatedObject:
		var id oid.ID
		if err := id.DecodeString(m.Value); err != nil {
			return fmt.Errorf("invalid associated object ID: %w", err)
		}
	}
	a.k, a.v = m.Key, m.Value
	return nil
}

// protoMessage converts a into message to transmit using the NeoFS API
// protocol.
func (a *Attribute) protoMessage() *protoobject.Header_Attribute {
	if a != nil {
		return &protoobject.Header_Attribute{Key: a.k, Value: a.v}
	}
	return nil
}

// Marshal marshals [Attribute] into a protobuf binary form.
//
// See also [Attribute.Unmarshal].
func (a *Attribute) Marshal() []byte {
	return neofsproto.MarshalMessage(a.protoMessage())
}

// Unmarshal unmarshals protobuf binary representation of [Attribute].
//
// See also [Attribute.Marshal].
func (a *Attribute) Unmarshal(data []byte) error {
	m := new(protoobject.Header_Attribute)
	if err := neofsproto.UnmarshalMessage(data, m); err != nil {
		return err
	}
	return a.fromProtoMessage(m, false)
}

// MarshalJSON encodes [Attribute] to protobuf JSON format.
//
// See also [Attribute.UnmarshalJSON].
func (a *Attribute) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalMessageJSON(a.protoMessage())
}

// UnmarshalJSON decodes [Attribute] from protobuf JSON format.
//
// See also [Attribute.MarshalJSON].
func (a *Attribute) UnmarshalJSON(data []byte) error {
	m := new(protoobject.Header_Attribute)
	if err := neofsproto.UnmarshalMessageJSON(data, m); err != nil {
		return err
	}
	return a.fromProtoMessage(m, false)
}
