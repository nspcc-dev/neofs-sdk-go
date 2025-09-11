package netmap

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"slices"
	"strconv"

	"github.com/nspcc-dev/hrw/v2"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	neofsproto "github.com/nspcc-dev/neofs-sdk-go/internal/proto"
	protonetmap "github.com/nspcc-dev/neofs-sdk-go/proto/netmap"
)

// NodeInfo groups information about NeoFS storage node which is reflected
// in the NeoFS network map. Storage nodes advertise this information when
// registering with the NeoFS network. After successful registration, information
// about the nodes is available to all network participants to work with the network
// map (mainly to comply with container storage policies).
//
// NodeInfo is mutually compatible with [protonetmap.NodeInfo] message. See
// [NodeInfo.FromProtoMessage] / [NodeInfo.ProtoMessage] methods.
//
// Instances can be created using built-in var declaration.
type NodeInfo struct {
	state protonetmap.NodeInfo_State
	pub   []byte
	addrs []string
	attrs [][2]string
}

// reads NodeInfo from netmap.NodeInfo message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field. Verifies format of any
// presented field according to NeoFS API V2 protocol.
func (x *NodeInfo) fromProtoMessage(m *protonetmap.NodeInfo, checkFieldPresence bool) error {
	if m.State < 0 {
		return fmt.Errorf("negative state %d", m.State)
	}
	var err error

	binPublicKey := m.GetPublicKey()
	if checkFieldPresence && len(binPublicKey) == 0 {
		return errors.New("missing public key")
	}

	if checkFieldPresence && len(m.Addresses) == 0 {
		return errors.New("missing network endpoints")
	}

	attributes := m.GetAttributes()
	mAttr := make(map[string]struct{}, len(attributes))
	attrs := make([][2]string, len(attributes))
	for i := range attributes {
		if attributes[i] == nil {
			return fmt.Errorf("nil attribute #%d", i)
		}
		key := attributes[i].GetKey()
		if key == "" {
			return fmt.Errorf("empty key of the attribute #%d", i)
		} else if _, ok := mAttr[key]; ok {
			return fmt.Errorf("duplicated attribute %s", key)
		}

		val := attributes[i].GetValue()
		switch {
		case key == attrCapacity:
			_, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid %s attribute: %w", attrCapacity, err)
			}
		case key == attrPrice:
			var err error
			_, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid %s attribute: %w", attrPrice, err)
			}
		default:
			if val == "" {
				return fmt.Errorf("empty %q attribute value", key)
			}
		}

		mAttr[key] = struct{}{}
		attrs[i][0], attrs[i][1] = key, val
	}

	x.state = m.State
	x.pub = m.PublicKey
	x.addrs = m.Addresses
	x.attrs = attrs

	return nil
}

// FromProtoMessage validates m according to the NeoFS API protocol and restores
// x from it.
//
// See also [NodeInfo.ProtoMessage].
func (x *NodeInfo) FromProtoMessage(m *protonetmap.NodeInfo) error {
	return x.fromProtoMessage(m, true)
}

// ProtoMessage converts x into message to transmit using the NeoFS API
// protocol.
//
// See also [NodeInfo.FromProtoMessage].
func (x NodeInfo) ProtoMessage() *protonetmap.NodeInfo {
	m := &protonetmap.NodeInfo{
		PublicKey: x.pub,
		Addresses: x.addrs,
		State:     x.state,
	}
	if len(x.attrs) > 0 {
		m.Attributes = make([]*protonetmap.NodeInfo_Attribute, len(x.attrs))
		for i := range x.attrs {
			m.Attributes[i] = &protonetmap.NodeInfo_Attribute{Key: x.attrs[i][0], Value: x.attrs[i][1]}
		}
	}
	return m
}

// Marshal encodes NodeInfo into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x NodeInfo) Marshal() []byte {
	return neofsproto.Marshal(x)
}

// Unmarshal decodes NeoFS API protocol binary format into the NodeInfo
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *NodeInfo) Unmarshal(data []byte) error {
	return neofsproto.UnmarshalOptional(data, x, (*NodeInfo).fromProtoMessage)
}

// MarshalJSON encodes NodeInfo into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x NodeInfo) MarshalJSON() ([]byte, error) {
	return neofsproto.MarshalJSON(x)
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the NodeInfo
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *NodeInfo) UnmarshalJSON(data []byte) error {
	return neofsproto.UnmarshalJSONOptional(data, x, (*NodeInfo).fromProtoMessage)
}

// SetPublicKey sets binary-encoded public key bound to the node. The key
// authenticates the storage node, so it MUST be unique within the network.
//
// Argument MUST NOT be mutated, make a copy first.
//
// The key parameter is a serialized compressed public key. See [elliptic.MarshalCompressed].
//
// See also [NodeInfo.PublicKey].
func (x *NodeInfo) SetPublicKey(key []byte) {
	x.pub = key
}

// PublicKey returns value set using [NodeInfo.SetPublicKey].
//
// Zero [NodeInfo] has no public key, which is incorrect according to
// NeoFS system requirements.
//
// The resulting slice of bytes is a serialized compressed public key. See [elliptic.MarshalCompressed].
// Use [neofsecdsa.PublicKey.Decode] to decode it into a type-specific structure.
//
// The value returned shares memory with the structure itself, so changing it can lead to data corruption.
// Make a copy if you need to change it.
func (x NodeInfo) PublicKey() []byte {
	return x.pub
}

// StringifyPublicKey returns HEX representation of PublicKey.
func StringifyPublicKey(node NodeInfo) string {
	return neofscrypto.StringifyKeyBinary(node.PublicKey())
}

// SetNetworkEndpoints sets list to the announced node's network endpoints.
// Node MUSt have at least one announced endpoint. List MUST be unique.
// Endpoints are used for communication with the storage node within NeoFS
// network. It is expected that node serves storage node services on these
// endpoints (it also adds a wait on their network availability).
//
// Argument MUST NOT be mutated, make a copy first.
func (x *NodeInfo) SetNetworkEndpoints(v ...string) {
	x.addrs = v
}

// NumberOfNetworkEndpoints returns number of network endpoints announced by the node.
//
// See also SetNetworkEndpoints.
func (x NodeInfo) NumberOfNetworkEndpoints() int {
	return len(x.addrs)
}

// NetworkEndpoints returns an iterator that yields the network endpoints
// announced by the node.
func (x NodeInfo) NetworkEndpoints() iter.Seq[string] {
	return slices.Values(x.addrs)
}

// assert NodeInfo type provides hrw.Hasher required for HRW sorting.
var _ hrw.Hashable = NodeInfo{}

// Hash implements hrw.Hasher interface.
//
// Hash is needed to support weighted HRW therefore sort function sorts nodes
// based on their public key. Hash isn't expected to be used directly.
func (x NodeInfo) Hash() uint64 {
	return hrw.Hash(x.PublicKey())
}

func (x *NodeInfo) setNumericAttribute(key string, num uint64) {
	x.SetAttribute(key, strconv.FormatUint(num, 10))
}

// SetPrice sets the storage cost declared by the node. By default, zero
// price is announced.
func (x *NodeInfo) SetPrice(price uint64) {
	x.setNumericAttribute(attrPrice, price)
}

// Price returns price set using SetPrice.
//
// Zero NodeInfo has zero price.
func (x NodeInfo) Price() uint64 {
	val := x.Attribute(attrPrice)
	if val == "" {
		return 0
	}

	price, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("unexpected price parsing error %s: %v", val, err))
	}

	return price
}

// SetCapacity sets the storage capacity declared by the node. By default, zero
// capacity is announced.
func (x *NodeInfo) SetCapacity(capacity uint64) {
	x.setNumericAttribute(attrCapacity, capacity)
}

// SetVersion sets node's version. By default, version
// is not announced.
func (x *NodeInfo) SetVersion(version string) {
	x.SetAttribute(attrVersion, version)
}

// capacity returns capacity set using SetCapacity.
//
// Zero NodeInfo has zero capacity.
func (x NodeInfo) capacity() uint64 {
	val := x.Attribute(attrCapacity)
	if val == "" {
		return 0
	}

	capacity, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("unexpected capacity parsing error %s: %v", val, err))
	}

	return capacity
}

const (
	attrUNLOCODE    = "UN-LOCODE"
	attrCountryCode = "CountryCode"
	attrCountryName = "Country"
	attrLocation    = "Location"
	attrSubDivCode  = "SubDivCode"
	attrSubDivName  = "SubDiv"
	attrContinent   = "Continent"
)

// SetLOCODE specifies node's geographic location in UN/LOCODE format. Each
// storage node MUST declare it for entrance to the NeoFS network. Node MAY
// declare the code of the nearest location as needed, for example, when it is
// impossible to unambiguously attribute the node to any location from UN/LOCODE
// database.
//
// See also LOCODE.
func (x *NodeInfo) SetLOCODE(locode string) {
	x.SetAttribute(attrUNLOCODE, locode)
}

// LOCODE returns node's location code set using SetLOCODE.
//
// Zero NodeInfo has empty location code which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) LOCODE() string {
	return x.Attribute(attrUNLOCODE)
}

// SetCountryCode sets code of the country in ISO 3166-1_alpha-2 to which
// storage node belongs (or the closest one).
func (x *NodeInfo) SetCountryCode(countryCode string) {
	x.SetAttribute(attrCountryCode, countryCode)
}

// CountryCode returns node's country code set using SetCountryCode.
//
// Zero NodeInfo has empty country code which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) CountryCode() string {
	return x.Attribute(attrCountryCode)
}

// SetCountryName sets short name of the country in ISO-3166 format to which
// storage node belongs (or the closest one).
func (x *NodeInfo) SetCountryName(country string) {
	x.SetAttribute(attrCountryName, country)
}

// CountryName returns node's country name set using SetCountryName.
//
// Zero NodeInfo has empty country name which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) CountryName() string {
	return x.Attribute(attrCountryName)
}

// SetLocationName sets storage node's location name from "NameWoDiacritics"
// column in the UN/LOCODE record corresponding to the specified LOCODE.
func (x *NodeInfo) SetLocationName(location string) {
	x.SetAttribute(attrLocation, location)
}

// LocationName returns node's location set using SetLocationName.
//
// Zero NodeInfo has empty location which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) LocationName() string {
	return x.Attribute(attrLocation)
}

// SetSubdivisionCode sets storage node's subdivision code from "SubDiv" column in
// the UN/LOCODE record corresponding to the specified LOCODE.
func (x *NodeInfo) SetSubdivisionCode(subDiv string) {
	x.SetAttribute(attrSubDivCode, subDiv)
}

// SubdivisionCode returns node's subdivision code set using SetSubdivisionCode.
//
// Zero NodeInfo has subdivision code which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) SubdivisionCode() string {
	return x.Attribute(attrSubDivCode)
}

// SetSubdivisionName sets storage node's subdivision name in ISO 3166-2 format.
func (x *NodeInfo) SetSubdivisionName(subDiv string) {
	x.SetAttribute(attrSubDivName, subDiv)
}

// SubdivisionName returns node's subdivision name set using SetSubdivisionName.
//
// Zero NodeInfo has subdivision name which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) SubdivisionName() string {
	return x.Attribute(attrSubDivName)
}

// SetContinentName sets name of the storage node's continent from
// Seven-Continent model.
func (x *NodeInfo) SetContinentName(continent string) {
	x.SetAttribute(attrContinent, continent)
}

// ContinentName returns node's continent set using SetContinentName.
//
// Zero NodeInfo has continent which is invalid according to
// NeoFS API system requirement.
func (x NodeInfo) ContinentName() string {
	return x.Attribute(attrContinent)
}

// Enumeration of well-known attributes.
const (
	// attrPrice is a key to the node attribute that indicates the
	// price in GAS tokens for storing one GB of data during one Epoch.
	attrPrice = "Price"

	// attrVersion is a key to the node attribute that indicates node's
	// version.
	attrVersion = "Version"

	// attrCapacity is a key to the node attribute that indicates the
	// total available disk space in Gigabytes.
	attrCapacity = "Capacity"
)

// NumberOfAttributes returns number of attributes announced by the node.
//
// See also SetAttribute.
func (x NodeInfo) NumberOfAttributes() int {
	return len(x.attrs)
}

// Attributes returns an iterator that yields the node attributes.
func (x NodeInfo) Attributes() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for i := range x.attrs {
			if !yield(x.attrs[i][0], x.attrs[i][1]) {
				return
			}
		}
	}
}

// GetAttributes returns all the node attributes.
// Each attribute is a [2]string slice: {"key", "value"}.
//
// See also Attribute.
func (x NodeInfo) GetAttributes() [][2]string {
	return x.attrs
}

// SetAttributes sets list of node attributes.
// Each attribute is a [2]string slice: {"key", "value"}.
// Both key and value of attributes MUST NOT be empty.
//
// See also SetAttribute.
func (x *NodeInfo) SetAttributes(attrs [][2]string) {
	for _, attr := range attrs {
		if attr[0] == "" {
			panic("empty key in SetAttributes")
		}
		if attr[1] == "" {
			panic(fmt.Errorf("empty value in SetAttributes for key: %s", attr[0]))
		}
	}

	x.attrs = attrs
}

// SetAttribute sets value of the node attribute value by the given key.
// Both key and value MUST NOT be empty.
func (x *NodeInfo) SetAttribute(key, value string) {
	if key == "" {
		panic("empty key in SetAttribute")
	} else if value == "" {
		panic("empty value in SetAttribute")
	}

	for i := range x.attrs {
		if x.attrs[i][0] == key {
			x.attrs[i][1] = value
			return
		}
	}

	x.attrs = append(x.attrs, [2]string{key, value})
}

// Attribute returns value of the node attribute set using SetAttribute by the
// given key. Returns empty string if attribute is missing.
func (x NodeInfo) Attribute(key string) string {
	for i := range x.attrs {
		if x.attrs[i][0] == key {
			return x.attrs[i][1]
		}
	}

	return ""
}

// SortAttributes sorts node attributes set using SetAttribute lexicographically.
// The method is only needed to make NodeInfo consistent, e.g. for signing.
func (x *NodeInfo) SortAttributes() {
	if len(x.attrs) == 0 {
		return
	}

	slices.SortFunc(x.attrs, func(a, b [2]string) int {
		if c := cmp.Compare(a[0], b[0]); c != 0 {
			return c
		}
		return cmp.Compare(a[1], b[1])
	})
}

// SetOffline sets the state of the node to "offline". When a node updates
// information about itself in the network map, this action is interpreted as
// an intention to leave the network.
func (x *NodeInfo) SetOffline() {
	x.state = protonetmap.NodeInfo_OFFLINE
}

// IsOffline checks if the node is in the "offline" state.
//
// Zero NodeInfo has undefined state which is not offline (note that it does not
// mean online).
//
// See also SetOffline.
func (x NodeInfo) IsOffline() bool {
	return x.state == protonetmap.NodeInfo_OFFLINE
}

// SetOnline sets the state of the node to "online". When a node updates
// information about itself in the network map, this
// action is interpreted as an intention to enter the network.
//
// See also IsOnline.
func (x *NodeInfo) SetOnline() {
	x.state = protonetmap.NodeInfo_ONLINE
}

// IsOnline checks if the node is in the "online" state.
//
// Zero NodeInfo has undefined state which is not online (note that it does not
// mean offline).
//
// See also SetOnline.
func (x NodeInfo) IsOnline() bool {
	return x.state == protonetmap.NodeInfo_ONLINE
}

// SetMaintenance sets the state of the node to "maintenance". When a node updates
// information about itself in the network map, this
// state declares temporal unavailability for a node.
//
// See also IsMaintenance.
func (x *NodeInfo) SetMaintenance() {
	x.state = protonetmap.NodeInfo_MAINTENANCE
}

// IsMaintenance checks if the node is in the "maintenance" state.
//
// Zero NodeInfo has undefined state.
//
// See also SetMaintenance.
func (x NodeInfo) IsMaintenance() bool {
	return x.state == protonetmap.NodeInfo_MAINTENANCE
}

const attrVerifiedNodesDomain = "VerifiedNodesDomain"

// SetVerifiedNodesDomain sets optional NeoFS NNS domain name to be used to
// confirm admission to a storage nodes' group on registration in the NeoFS
// network of a storage node submitting this NodeInfo about itself. If domain is
// specified, the storage node requesting entry into the NeoFS network map with
// must be included in the access list located on the specified domain. The
// access list is represented by a set of TXT records: Neo script hashes from
// public keys. To be admitted to the network, script hash of the
// [NodeInfo.PublicKey] must be present in domain records. Otherwise,
// registration will be denied. By default, this check is not carried out.
//
// Value MUST be a valid NeoFS NNS domain name.
//
// See also [NodeInfo.VerifiedNodesDomain].
func (x *NodeInfo) SetVerifiedNodesDomain(domain string) {
	x.SetAttribute(attrVerifiedNodesDomain, domain)
}

// VerifiedNodesDomain returns optional NeoFS NNS domain name to be used to
// confirm admission to a storage nodes' group on registration in the NeoFS
// network of a storage node submitting this NodeInfo about itself. Returns zero
// value if domain is not specified.
//
// See also [NodeInfo.SetVerifiedNodesDomain].
func (x NodeInfo) VerifiedNodesDomain() string {
	return x.Attribute(attrVerifiedNodesDomain)
}
