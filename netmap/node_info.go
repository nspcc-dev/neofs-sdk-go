package netmap

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nspcc-dev/hrw/v2"
	"github.com/nspcc-dev/neofs-sdk-go/api/netmap"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// NodeInfo groups information about NeoFS storage node which is reflected
// in the NeoFS network map. Storage nodes advertise this information when
// registering with the NeoFS network. After successful registration, information
// about the nodes is available to all network participants to work with the network
// map (mainly to comply with container storage policies).
//
// NodeInfo is mutually compatible with [netmap.NodeInfo] message. See
// [NodeInfo.ReadFromV2] / [NodeInfo.WriteToV2] methods.
//
// Instances can be created using built-in var declaration.
type NodeInfo struct {
	state     netmap.NodeInfo_State
	pubKey    []byte
	endpoints []string
	attrs     []*netmap.NodeInfo_Attribute
}

func isEmptyNodeInfo(n NodeInfo) bool {
	return n.state == 0 && len(n.pubKey) == 0 && len(n.endpoints) == 0 && len(n.attrs) == 0
}

// reads NodeInfo from netmap.NodeInfo message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field. Verifies format of any
// presented field according to NeoFS API V2 protocol.
func (x *NodeInfo) readFromV2(m *netmap.NodeInfo, checkFieldPresence bool) error {
	var err error

	if checkFieldPresence && len(m.PublicKey) == 0 {
		return errors.New("missing public key")
	}

	if checkFieldPresence && len(m.Addresses) == 0 {
		return errors.New("missing network endpoints")
	}

	for i := range m.Addresses {
		if m.Addresses[i] == "" {
			return fmt.Errorf("empty network endpoint #%d", i)
		}
	}

	attributes := m.GetAttributes()
	for i := range attributes {
		key := attributes[i].GetKey()
		if key == "" {
			return fmt.Errorf("invalid attribute #%d: missing key", i)
		} // also prevents further NPE
		for j := 0; j < i; j++ {
			if attributes[j].Key == key {
				return fmt.Errorf("multiple attributes with key=%s", key)
			}
		}
		if attributes[i].Value == "" {
			return fmt.Errorf("invalid attribute #%d (%s): missing value", i, key)
		}
		switch key {
		case attrCapacity:
			_, err = strconv.ParseUint(attributes[i].Value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid capacity attribute (#%d): invalid integer (%w)", i, err)
			}
		case attrPrice:
			_, err = strconv.ParseUint(attributes[i].Value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid price attribute (#%d): invalid integer (%w)", i, err)
			}
		}
	}

	x.pubKey = m.PublicKey
	x.endpoints = m.Addresses
	x.state = m.State
	x.attrs = attributes

	return nil
}

// ReadFromV2 reads NodeInfo from the [netmap.NodeInfo] message. Returns an
// error if the message is malformed according to the NeoFS API V2 protocol. The
// message must not be nil.
//
// ReadFromV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [NodeInfo.WriteToV2].
func (x *NodeInfo) ReadFromV2(m *netmap.NodeInfo) error {
	return x.readFromV2(m, true)
}

// WriteToV2 writes NodeInfo to the [netmap.NodeInfo] message of the NeoFS API
// protocol.
//
// WriteToV2 is intended to be used by the NeoFS API V2 client/server
// implementation only and is not expected to be directly used by applications.
//
// See also [NodeInfo.ReadFromV2].
func (x NodeInfo) WriteToV2(m *netmap.NodeInfo) {
	m.Attributes = x.attrs
	m.PublicKey = x.pubKey
	m.Addresses = x.endpoints
	m.State = x.state
}

// Marshal encodes NodeInfo into a binary format of the NeoFS API
// protocol (Protocol Buffers V3 with direct field order).
//
// See also [NodeInfo.Unmarshal].
func (x NodeInfo) Marshal() []byte {
	var m netmap.NodeInfo
	x.WriteToV2(&m)
	b := make([]byte, m.MarshaledSize())
	m.MarshalStable(b)
	return b
}

// Unmarshal decodes Protocol Buffers V3 binary data into the NodeInfo. Returns
// an error describing a format violation of the specified fields. Unmarshal
// does not check presence of the required fields and, at the same time, checks
// format of presented fields.
//
// See also [NodeInfo.Marshal].
func (x *NodeInfo) Unmarshal(data []byte) error {
	var m netmap.NodeInfo
	err := proto.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protobuf: %w", err)
	}

	return x.readFromV2(&m, false)
}

// MarshalJSON encodes NodeInfo into a JSON format of the NeoFS API protocol
// (Protocol Buffers V3 JSON).
//
// See also [NodeInfo.UnmarshalJSON].
func (x NodeInfo) MarshalJSON() ([]byte, error) {
	var m netmap.NodeInfo
	x.WriteToV2(&m)

	return protojson.Marshal(&m)
}

// UnmarshalJSON decodes NeoFS API protocol JSON data into the NodeInfo
// (Protocol Buffers V3 JSON). Returns an error describing a format violation.
// UnmarshalJSON does not check presence of the required fields and, at the same
// time, checks format of presented fields.
//
// See also [NodeInfo.MarshalJSON].
func (x *NodeInfo) UnmarshalJSON(data []byte) error {
	var m netmap.NodeInfo
	err := protojson.Unmarshal(data, &m)
	if err != nil {
		return fmt.Errorf("decode protojson: %w", err)
	}

	return x.readFromV2(&m, false)
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
	x.pubKey = key
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
	return x.pubKey
}

// StringifyPublicKey returns HEX representation of [NodeInfo.PublicKey].
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
//
// See also [NodeInfo.IterateNetworkEndpoints].
func (x *NodeInfo) SetNetworkEndpoints(v []string) {
	x.endpoints = v
}

// NetworkEndpoints sets list to the announced node's network endpoints.
//
// See also [NodeInfo.SetNetworkEndpoints].
func (x NodeInfo) NetworkEndpoints() []string {
	return x.endpoints
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

// less declares "less than" comparison between two NodeInfo instances:
// x1 is less than x2 if it has less Hash().
//
// Method is needed for internal placement needs.
func less(x1, x2 NodeInfo) bool {
	return x1.Hash() < x2.Hash()
}

func (x *NodeInfo) setNumericAttribute(key string, num uint64) {
	x.SetAttribute(key, strconv.FormatUint(num, 10))
}

// SetPrice sets the storage cost declared by the node. By default, zero
// price is announced.
//
// See also [NodeInfo.Price].
func (x *NodeInfo) SetPrice(price uint64) {
	x.setNumericAttribute(attrPrice, price)
}

// Price returns price set using [NodeInfo.SetPrice].
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
//
// See also [NodeInfo.Capacity].
func (x *NodeInfo) SetCapacity(capacity uint64) {
	x.setNumericAttribute(attrCapacity, capacity)
}

// Capacity returns capacity set using [NodeInfo.SetCapacity].
//
// Zero NodeInfo has zero capacity.
func (x NodeInfo) Capacity() uint64 {
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

// SetVersion sets node's version. By default, version is not announced.
//
// See also [NodeInfo.Version].
func (x *NodeInfo) SetVersion(version string) {
	x.SetAttribute(attrVersion, version)
}

// Version returns announced node version set using [NodeInfo.SetVersion].
//
// Zero NodeInfo has no announced version.
func (x NodeInfo) Version() string {
	return x.Attribute(attrVersion)
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
// See also [NodeInfo.LOCODE].
func (x *NodeInfo) SetLOCODE(locode string) {
	x.SetAttribute(attrUNLOCODE, locode)
}

// LOCODE returns node's location code set using [NodeInfo.SetLOCODE].
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

	// attrExternalAddr is a key for the attribute storing node external addresses.
	attrExternalAddr = "ExternalAddr"
	// sepExternalAddr is a separator for multi-value ExternalAddr attribute.
	sepExternalAddr = ","
)

// SetExternalAddresses sets multi-addresses to use
// to connect to this node from outside.
//
// Panics if addr is an empty list.
//
// See also [NodeInfo.ExternalAddresses].
func (x *NodeInfo) SetExternalAddresses(addr []string) {
	x.SetAttribute(attrExternalAddr, strings.Join(addr, sepExternalAddr))
}

// ExternalAddresses returns list of multi-addresses to use
// to connect to this node from outside.
//
// See also [NodeInfo.SetExternalAddresses].
func (x NodeInfo) ExternalAddresses() []string {
	a := x.Attribute(attrExternalAddr)
	if len(a) == 0 {
		return nil
	}

	return strings.Split(a, sepExternalAddr)
}

// NumberOfAttributes returns number of attributes announced by the node.
//
// See also [NodeInfo.SetAttribute].
func (x NodeInfo) NumberOfAttributes() int {
	return len(x.attrs)
}

// IterateAttributes iterates over all node attributes and passes the into f.
// Handler MUST NOT be nil.
//
// See also [NodeInfo.SetAttribute].
func (x NodeInfo) IterateAttributes(f func(key, value string)) {
	for i := range x.attrs {
		f(x.attrs[i].GetKey(), x.attrs[i].GetValue())
	}
}

// SetAttribute sets value of the node attribute value by the given key.
// Both key and value MUST NOT be empty.
//
// See also [NodeInfo.NumberOfAttributes], [NodeInfo.IterateAttributes].
func (x *NodeInfo) SetAttribute(key, value string) {
	if key == "" {
		panic("empty key in SetAttribute")
	} else if value == "" {
		panic("empty value in SetAttribute")
	}

	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			x.attrs[i].Value = value
			return
		}
	}

	x.attrs = append(x.attrs, &netmap.NodeInfo_Attribute{
		Key:   key,
		Value: value,
	})
}

// Attribute returns value of the node attribute set using
// [NodeInfo.SetAttribute] by the given key. Returns empty string if attribute
// is missing.
func (x NodeInfo) Attribute(key string) string {
	for i := range x.attrs {
		if x.attrs[i].GetKey() == key {
			return x.attrs[i].GetValue()
		}
	}

	return ""
}

// SortAttributes sorts node attributes set using [NodeInfo.SetAttribute]
// lexicographically. The method is only needed to make NodeInfo consistent,
// e.g. for signing.
func (x *NodeInfo) SortAttributes() {
	if len(x.attrs) == 0 {
		return
	}

	sort.Slice(x.attrs, func(i, j int) bool {
		switch strings.Compare(x.attrs[i].GetKey(), x.attrs[j].GetKey()) {
		case -1:
			return true
		case 1:
			return false
		default:
			return x.attrs[i].GetValue() < x.attrs[j].GetValue()
		}
	})
}

// SetOffline sets the state of the node to "offline". When a node updates
// information about itself in the network map, this action is interpreted as
// an intention to leave the network.
//
// See also [NodeInfo.IsOffline].
func (x *NodeInfo) SetOffline() {
	x.state = netmap.NodeInfo_OFFLINE
}

// IsOffline checks if the node is in the "offline" state.
//
// Zero NodeInfo has undefined state which is not offline (note that it does not
// mean online).
//
// See also [NodeInfo.SetOffline].
func (x NodeInfo) IsOffline() bool {
	return x.state == netmap.NodeInfo_OFFLINE
}

// SetOnline sets the state of the node to "online". When a node updates
// information about itself in the network map, this
// action is interpreted as an intention to enter the network.
//
// See also [NodeInfo.IsOnline].
func (x *NodeInfo) SetOnline() {
	x.state = netmap.NodeInfo_ONLINE
}

// IsOnline checks if the node is in the "online" state.
//
// Zero NodeInfo has undefined state which is not online (note that it does not
// mean offline).
//
// See also [NodeInfo.SetOnline].
func (x NodeInfo) IsOnline() bool {
	return x.state == netmap.NodeInfo_ONLINE
}

// SetMaintenance sets the state of the node to "maintenance". When a node
// updates information about itself in the network map, this state declares
// temporal unavailability for a node.
//
// See also [NodeInfo.IsMaintenance].
func (x *NodeInfo) SetMaintenance() {
	x.state = netmap.NodeInfo_MAINTENANCE
}

// IsMaintenance checks if the node is in the "maintenance" state.
//
// Zero NodeInfo has undefined state.
//
// See also [NodeInfo.SetMaintenance].
func (x NodeInfo) IsMaintenance() bool {
	return x.state == netmap.NodeInfo_MAINTENANCE
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
