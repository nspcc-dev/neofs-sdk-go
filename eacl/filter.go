package eacl

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nspcc-dev/neofs-api-go/v2/acl"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
	"github.com/nspcc-dev/neofs-sdk-go/version"
	"github.com/nspcc-dev/tzhash/tz"
)

// Filter describes a binary property of an access-controlled NeoFS resource
// according to meta information about it. The meta information is represented
// by a set of key-value headers of various types.
type Filter struct {
	hdrType HeaderType
	matcher Matcher
	key     string
	value   string
}

// NewFilter returns Filter describing a condition that is suitable only for
// access-controlled resources with a header of the given type whose value
// corresponds to the specified value.
//
// Both typ and matcher MUST be supported values from corresponding enum.
// Key MUST be non-empty.
//
// See also other helper constructors.
func NewFilter(typ HeaderType, key string, matcher Matcher, value string) Filter {
	f := Filter{
		hdrType: typ,
		matcher: matcher,
		key:     key,
		value:   value,
	}
	if msg := f.validate(); msg != "" {
		panic(msg)
	}
	return f
}

// Various reserved types of header related to the NeoFS objects.
const (
	FilterObjectVersion                    = acl.FilterObjectVersion
	FilterObjectID                         = acl.FilterObjectID
	FilterObjectContainerID                = acl.FilterObjectContainerID
	FilterObjectOwnerID                    = acl.FilterObjectOwnerID
	FilterObjectCreationEpoch              = acl.FilterObjectCreationEpoch
	FilterObjectPayloadSize                = acl.FilterObjectPayloadLength
	FilterObjectPayloadChecksum            = acl.FilterObjectPayloadHash
	FilterObjectType                       = acl.FilterObjectType
	FilterObjectPayloadHomomorphicChecksum = acl.FilterObjectHomomorphicHash
)

// returns message about docs violation or zero if everything is OK.
func (f Filter) validate() string {
	switch {
	case f.hdrType <= 0 || f.hdrType >= lastHeaderType:
		return fmt.Sprintf("forbidden value %v from the enum %T", f.hdrType, f.hdrType)
	case f.matcher <= 0 || f.matcher >= lastMatcher:
		return fmt.Sprintf("forbidden value %v from the enum %T", f.matcher, f.matcher)
	case f.key == "":
		return "empty key"
	}
	return ""
}

// copyTo writes deep copy of the [Filter] to dst.
func (f Filter) copyTo(dst *Filter) {
	*dst = f
}

// HeaderValue returns value of the access-controlled resource's header to
// match.
func (f Filter) HeaderValue() string {
	return f.value
}

// Matcher returns operator to match the header.
func (f Filter) Matcher() Matcher {
	return f.matcher
}

// HeaderKey returns key to the access-controlled resource's header to match.
func (f Filter) HeaderKey() string {
	return f.key
}

// HeaderType returns type of access-controlled resource's header to match.
func (f Filter) HeaderType() HeaderType {
	return f.hdrType
}

// readFromV2 reads Filter from the [acl.HeaderFilter] message. Returns an error
// if the message is malformed according to the NeoFS API V2 protocol. Behavior
// is forward-compatible: unknown enum values are not considered invalid.
func (f *Filter) readFromV2(m acl.HeaderFilter) error {
	key := m.GetKey()
	if key == "" {
		return errors.New("empty header key")
	}

	f.hdrType = HeaderType(m.GetHeaderType())
	f.matcher = Matcher(m.GetMatchType())
	f.key = key
	f.value = m.GetValue()

	return nil
}

// writeToV2 writes Filter to the [acl.HeaderFilter] message of the NeoFS API V2
// protocol.
func (f Filter) writeToV2(m *acl.HeaderFilter) {
	m.SetValue(f.value)
	m.SetKey(f.key)
	m.SetMatchType(acl.MatchType(f.matcher))
	m.SetHeaderType(acl.HeaderType(f.hdrType))
}

func newFilterObject(key string, matcher Matcher, value string) Filter {
	return NewFilter(HeaderFromObject, key, matcher, value)
}

// NewFilterObjectAttribute constructs Filter to object attribute. Key MUST NOT
// start with reserved '$Object:' prefix.
//
// See also [NewFilter].
func NewFilterObjectAttribute(key string, matcher Matcher, value string) Filter {
	if strings.HasPrefix(key, acl.ObjectFilterPrefix) {
		panic(fmt.Sprintf("reserved prefix in '%s'", key))
	}

	return newFilterObject(key, matcher, value)
}

// NewFilterObjectVersion constructs Filter to match protocol version of the
// object.
//
// See also [NewFilter].
func NewFilterObjectVersion(matcher Matcher, v version.Version) Filter {
	return newFilterObject(FilterObjectVersion, matcher, version.EncodeToString(v))
}

// NewFilterObjectID constructs Filter to match objects' identifier.
//
// See also [NewFilter].
func NewFilterObjectID(matcher Matcher, id oid.ID) Filter {
	return newFilterObject(FilterObjectID, matcher, id.EncodeToString())
}

// NewFilterContainerID constructs Filter to match objects' container.
//
// See also [NewFilter].
func NewFilterContainerID(matcher Matcher, cnr cid.ID) Filter {
	return newFilterObject(FilterObjectContainerID, matcher, cnr.EncodeToString())
}

// NewFilterOwnerID constructs Filter to match objects' owner.
//
// See also [NewFilter].
func NewFilterOwnerID(matcher Matcher, owner user.ID) Filter {
	return newFilterObject(FilterObjectOwnerID, matcher, owner.EncodeToString())
}

// NewFilterObjectCreationEpoch constructs Filter to match creation epoch of the
// objects.
//
// See also [NewFilter].
func NewFilterObjectCreationEpoch(matcher Matcher, epoch uint64) Filter {
	return newFilterObject(FilterObjectCreationEpoch, matcher, strconv.FormatUint(epoch, 10))
}

// NewFilterObjectPayloadSize constructs Filter to match payload size of the
// objects.
//
// See also [NewFilter].
func NewFilterObjectPayloadSize(matcher Matcher, size uint64) Filter {
	return newFilterObject(FilterObjectPayloadSize, matcher, strconv.FormatUint(size, 10))
}

// NewFilterObjectType constructs Filter to match type of the objects.
//
// See also [NewFilter].
func NewFilterObjectType(matcher Matcher, typ object.Type) Filter {
	return newFilterObject(FilterObjectType, matcher, typ.EncodeToString())
}

// NewFilterObjectPayloadChecksum constructs Filter to match payload checksum of
// the objects.
//
// See also [NewFilter].
func NewFilterObjectPayloadChecksum(matcher Matcher, cs [sha256.Size]byte) Filter {
	return newFilterObject(FilterObjectPayloadChecksum, matcher, hex.EncodeToString(cs[:]))
}

// NewFilterObjectPayloadHomomorphicChecksum constructs Filter to match
// payload's homomorphic checksum of the objects.
//
// See also [NewFilter].
func NewFilterObjectPayloadHomomorphicChecksum(matcher Matcher, cs [tz.Size]byte) Filter {
	return newFilterObject(FilterObjectPayloadHomomorphicChecksum, matcher, hex.EncodeToString(cs[:]))
}
