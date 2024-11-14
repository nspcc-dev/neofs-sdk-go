package netmap

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nspcc-dev/hrw/v2"
	"github.com/nspcc-dev/neofs-api-go/v2/netmap"
	neofscrypto "github.com/nspcc-dev/neofs-sdk-go/crypto"
)

// NodeInfo groups information about NeoFS storage node which is reflected
// in the NeoFS network map. Storage nodes advertise this information when
// registering with the NeoFS network. After successful registration, information
// about the nodes is available to all network participants to work with the network
// map (mainly to comply with container storage policies).
//
// NodeInfo is mutually compatible with github.com/nspcc-dev/neofs-api-go/v2/netmap.NodeInfo
// message. See ReadFromV2 / WriteToV2 methods.
//
// Instances can be created using built-in var declaration.
type NodeInfo struct {
	m netmap.NodeInfo
}

// reads NodeInfo from netmap.NodeInfo message. If checkFieldPresence is set,
// returns an error on absence of any protocol-required field. Verifies format of any
// presented field according to NeoFS API V2 protocol.
func (x *NodeInfo) readFromV2(m netmap.NodeInfo, checkFieldPresence bool) error {
	var err error

	binPublicKey := m.GetPublicKey()
	if checkFieldPresence && len(binPublicKey) == 0 {
		return errors.New("missing public key")
	}

	if checkFieldPresence && m.NumberOfAddresses() <= 0 {
		return errors.New("missing network endpoints")
	}

	attributes := m.GetAttributes()
	mAttr := make(map[string]struct{}, len(attributes))
	for i := range attributes {
		key := attributes[i].GetKey()
		if key == "" {
			return fmt.Errorf("empty key of the attribute #%d", i)
		} else if _, ok := mAttr[key]; ok {
			return fmt.Errorf("duplicated attribute %s", key)
		}

		switch {
		case key == attrCapacity:
			_, err = strconv.ParseUint(attributes[i].GetValue(), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid %s attribute: %w", attrCapacity, err)
			}
		case key == attrPrice:
			var err error
			_, err = strconv.ParseUint(attributes[i].GetValue(), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid %s attribute: %w", attrPrice, err)
			}
		default:
			if attributes[i].GetValue() == "" {
				return fmt.Errorf("empty value of the attribute %s", key)
			}
		}

		mAttr[key] = struct{}{}
	}

	x.m = m

	return nil
}

// ReadFromV2 reads NodeInfo from the netmap.NodeInfo message. Checks if the
// message conforms to NeoFS API V2 protocol.
//
// See also WriteToV2.
func (x *NodeInfo) ReadFromV2(m netmap.NodeInfo) error {
	return x.readFromV2(m, true)
}

// WriteToV2 writes NodeInfo to the netmap.NodeInfo message. The message MUST NOT
// be nil.
//
// See also ReadFromV2.
func (x NodeInfo) WriteToV2(m *netmap.NodeInfo) {
	*m = x.m
}

// Marshal encodes NodeInfo into a binary format of the NeoFS API protocol
// (Protocol Buffers with direct field order).
//
// See also Unmarshal.
func (x NodeInfo) Marshal() []byte {
	var m netmap.NodeInfo
	x.WriteToV2(&m)

	return m.StableMarshal(nil)
}

// Unmarshal decodes NeoFS API protocol binary format into the NodeInfo
// (Protocol Buffers with direct field order). Returns an error describing
// a format violation.
//
// See also Marshal.
func (x *NodeInfo) Unmarshal(data []byte) error {
	var m netmap.NodeInfo

	err := m.Unmarshal(data)
	if err != nil {
		return err
	}

	return x.readFromV2(m, false)
}

// MarshalJSON encodes NodeInfo into a JSON format of the NeoFS API protocol
// (Protocol Buffers JSON).
//
// See also UnmarshalJSON.
func (x NodeInfo) MarshalJSON() ([]byte, error) {
	var m netmap.NodeInfo
	x.WriteToV2(&m)

	return m.MarshalJSON()
}

// UnmarshalJSON decodes NeoFS API protocol JSON format into the NodeInfo
// (Protocol Buffers JSON). Returns an error describing a format violation.
//
// See also MarshalJSON.
func (x *NodeInfo) UnmarshalJSON(data []byte) error {
	var m netmap.NodeInfo

	err := m.UnmarshalJSON(data)
	if err != nil {
		return err
	}

	return x.readFromV2(m, false)
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
	x.m.SetPublicKey(key)
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
	return x.m.GetPublicKey()
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
//
// See also IterateNetworkEndpoints.
func (x *NodeInfo) SetNetworkEndpoints(v ...string) {
	x.m.SetAddresses(v...)
}

// NumberOfNetworkEndpoints returns number of network endpoints announced by the node.
//
// See also SetNetworkEndpoints.
func (x NodeInfo) NumberOfNetworkEndpoints() int {
	return x.m.NumberOfAddresses()
}

// IterateNetworkEndpoints iterates over network endpoints announced by the
// node and pass them into f. Breaks iteration on f's true return. Handler
// MUST NOT be nil.
//
// Zero NodeInfo contains no endpoints which is incorrect according to
// NeoFS system requirements.
//
// See also SetNetworkEndpoints.
func (x NodeInfo) IterateNetworkEndpoints(f func(string) bool) {
	x.m.IterateAddresses(f)
}

// IterateNetworkEndpoints is an extra-sugared function over IterateNetworkEndpoints
// method which allows to unconditionally iterate over all node's network endpoints.
func IterateNetworkEndpoints(node NodeInfo, f func(string)) {
	node.IterateNetworkEndpoints(func(addr string) bool {
		f(addr)
		return false
	})
}

// assert NodeInfo type provides hrw.Hasher required for HRW sorting.
var _ hrw.Hashable = NodeInfo{}

// Hash implements hrw.Hasher interface.
//
// Hash is needed to support weighted HRW therefore sort function sorts nodes
// based on their public key. Hash isn't expected to be used directly.
func (x NodeInfo) Hash() uint64 {
	return hrw.Hash(x.m.GetPublicKey())
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

	// attrExternalAddr is a key for the attribute storing node external addresses.
	attrExternalAddr = "ExternalAddr"
	// sepExternalAddr is a separator for multi-value ExternalAddr attribute.
	sepExternalAddr = ","
)

// SetExternalAddresses sets multi-addresses to use
// to connect to this node from outside.
//
// Panics if addr is an empty list.
func (x *NodeInfo) SetExternalAddresses(addr ...string) {
	x.SetAttribute(attrExternalAddr, strings.Join(addr, sepExternalAddr))
}

// ExternalAddresses returns list of multi-addresses to use
// to connect to this node from outside.
func (x NodeInfo) ExternalAddresses() []string {
	a := x.Attribute(attrExternalAddr)
	if len(a) == 0 {
		return nil
	}

	return strings.Split(a, sepExternalAddr)
}

// NumberOfAttributes returns number of attributes announced by the node.
//
// See also SetAttribute.
func (x NodeInfo) NumberOfAttributes() int {
	return len(x.m.GetAttributes())
}

// IterateAttributes iterates over all node attributes and passes the into f.
// Handler MUST NOT be nil.
func (x NodeInfo) IterateAttributes(f func(key, value string)) {
	a := x.m.GetAttributes()
	for i := range a {
		f(a[i].GetKey(), a[i].GetValue())
	}
}

// GetAttributes returns all the node attributes.
// Each attribute is a [2]string slice: {"key", "value"}.
//
// See also Attribute, IterateAttributes.
func (x NodeInfo) GetAttributes() [][2]string {
	attrs := make([][2]string, len(x.m.GetAttributes()))
	for i, attr := range x.m.GetAttributes() {
		attrs[i] = [2]string{attr.GetKey(), attr.GetValue()}
	}
	return attrs
}

// SetAttributes sets list of node attributes.
// Each attribute is a [2]string slice: {"key", "value"}.
// Both key and value of attributes MUST NOT be empty.
//
// See also SetAttribute.
func (x *NodeInfo) SetAttributes(attrs [][2]string) {
	netmapAttrs := make([]netmap.Attribute, 0, len(attrs))
	for _, attr := range attrs {
		if attr[0] == "" {
			panic("empty key in SetAttributes")
		}
		if attr[1] == "" {
			panic(fmt.Errorf("empty value in SetAttributes for key: %s", attr[0]))
		}

		netmapAttrs = append(netmapAttrs, netmap.Attribute{})
		netmapAttrs[len(netmapAttrs)-1].SetKey(attr[0])
		netmapAttrs[len(netmapAttrs)-1].SetValue(attr[1])
	}

	x.m.SetAttributes(netmapAttrs)
}

// SetAttribute sets value of the node attribute value by the given key.
// Both key and value MUST NOT be empty.
func (x *NodeInfo) SetAttribute(key, value string) {
	if key == "" {
		panic("empty key in SetAttribute")
	} else if value == "" {
		panic("empty value in SetAttribute")
	}

	a := x.m.GetAttributes()
	for i := range a {
		if a[i].GetKey() == key {
			a[i].SetValue(value)
			return
		}
	}

	a = append(a, netmap.Attribute{})
	a[len(a)-1].SetKey(key)
	a[len(a)-1].SetValue(value)

	x.m.SetAttributes(a)
}

// Attribute returns value of the node attribute set using SetAttribute by the
// given key. Returns empty string if attribute is missing.
func (x NodeInfo) Attribute(key string) string {
	a := x.m.GetAttributes()
	for i := range a {
		if a[i].GetKey() == key {
			return a[i].GetValue()
		}
	}

	return ""
}

// SortAttributes sorts node attributes set using SetAttribute lexicographically.
// The method is only needed to make NodeInfo consistent, e.g. for signing.
func (x *NodeInfo) SortAttributes() {
	as := x.m.GetAttributes()
	if len(as) == 0 {
		return
	}

	sort.Slice(as, func(i, j int) bool {
		switch strings.Compare(as[i].GetKey(), as[j].GetKey()) {
		case -1:
			return true
		case 1:
			return false
		default:
			return as[i].GetValue() < as[j].GetValue()
		}
	})

	x.m.SetAttributes(as)
}

// SetOffline sets the state of the node to "offline". When a node updates
// information about itself in the network map, this action is interpreted as
// an intention to leave the network.
func (x *NodeInfo) SetOffline() {
	x.m.SetState(netmap.Offline)
}

// IsOffline checks if the node is in the "offline" state.
//
// Zero NodeInfo has undefined state which is not offline (note that it does not
// mean online).
//
// See also SetOffline.
func (x NodeInfo) IsOffline() bool {
	return x.m.GetState() == netmap.Offline
}

// SetOnline sets the state of the node to "online". When a node updates
// information about itself in the network map, this
// action is interpreted as an intention to enter the network.
//
// See also IsOnline.
func (x *NodeInfo) SetOnline() {
	x.m.SetState(netmap.Online)
}

// IsOnline checks if the node is in the "online" state.
//
// Zero NodeInfo has undefined state which is not online (note that it does not
// mean offline).
//
// See also SetOnline.
func (x NodeInfo) IsOnline() bool {
	return x.m.GetState() == netmap.Online
}

// SetMaintenance sets the state of the node to "maintenance". When a node updates
// information about itself in the network map, this
// state declares temporal unavailability for a node.
//
// See also IsMaintenance.
func (x *NodeInfo) SetMaintenance() {
	x.m.SetState(netmap.Maintenance)
}

// IsMaintenance checks if the node is in the "maintenance" state.
//
// Zero NodeInfo has undefined state.
//
// See also SetMaintenance.
func (x NodeInfo) IsMaintenance() bool {
	return x.m.GetState() == netmap.Maintenance
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
