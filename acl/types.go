package acl

import (
	"fmt"
	"strconv"
	"strings"
)

// BasicACL is Access Control List that defines who can interact with containers and what exactly they can do.
type BasicACL uint32

// String returns BasicACL string representation
// in hexadecimal form with 0x prefix.
func (a BasicACL) String() string {
	return fmt.Sprintf("0x%08x", uint32(a))
}

const (
	// PublicBasicRule is a basic ACL value for final public-read-write container for which extended ACL CANNOT be set.
	PublicBasicRule BasicACL = 0x1FBFBFFF

	// PrivateBasicRule is a basic ACL value for final private container for which extended ACL CANNOT be set.
	PrivateBasicRule BasicACL = 0x1C8C8CCC

	// ReadOnlyBasicRule is a basic ACL value for final public-read container for which extended ACL CANNOT be set.
	ReadOnlyBasicRule BasicACL = 0x1FBF8CFF

	// PublicAppendRule is a basic ACL value for final public-append container for which extended ACL CANNOT be set.
	PublicAppendRule BasicACL = 0x1FBF9FFF

	// EACLPublicBasicRule is a basic ACL value for non-final public-read-write container for which extended ACL CAN be set.
	EACLPublicBasicRule BasicACL = 0x0FBFBFFF

	// EACLPrivateBasicRule is a basic ACL value for non-final private container for which extended ACL CAN be set.
	EACLPrivateBasicRule BasicACL = 0x0C8C8CCC

	// EACLReadOnlyBasicRule is a basic ACL value for non-final public-read container for which extended ACL CAN be set.
	EACLReadOnlyBasicRule BasicACL = 0x0FBF8CFF

	// EACLPublicAppendRule is a basic ACL value for non-final public-append container for which extended ACL CAN be set.
	EACLPublicAppendRule BasicACL = 0x0FBF9FFF
)

const (
	// PublicBasicName is a well-known name for 0x1FBFBFFF basic ACL.
	// It represents fully-public container without eACL.
	PublicBasicName = "public-read-write"

	// PrivateBasicName is a well-known name for 0x1C8C8CCC basic ACL.
	// It represents fully-private container without eACL.
	PrivateBasicName = "private"

	// ReadOnlyBasicName is a well-known name for 0x1FBF8CFF basic ACL.
	// It represents public read-only container without eACL.
	ReadOnlyBasicName = "public-read"

	// PublicAppendName is a well-known name for 0x1FBF9FFF basic ACL.
	// It represents fully-public container without eACL except DELETE operation is only allowed on the owner.
	PublicAppendName = "public-append"

	// EACLPublicBasicName is a well-known name for 0x0FBFBFFF basic ACL.
	// It represents fully-public container that allows eACL.
	EACLPublicBasicName = "eacl-public-read-write"

	// EACLPrivateBasicName is a well-known name for 0x0C8C8CCC basic ACL.
	// It represents fully-private container that allows eACL.
	EACLPrivateBasicName = "eacl-private"

	// EACLReadOnlyBasicName is a well-known name for 0x0FBF8CFF basic ACL.
	// It represents public read-only container that allows eACL.
	EACLReadOnlyBasicName = "eacl-public-read"

	// EACLPublicAppendName is a well-known name for 0x0FBF9FFF basic ACL.
	// It represents fully-public container that allows eACL except DELETE operation is only allowed on the owner.
	EACLPublicAppendName = "eacl-public-append"
)

// ParseBasicACL parse string ACL (well-known names or hex representation).
func ParseBasicACL(basicACL string) (BasicACL, error) {
	switch basicACL {
	case PublicBasicName:
		return PublicBasicRule, nil
	case PrivateBasicName:
		return PrivateBasicRule, nil
	case ReadOnlyBasicName:
		return ReadOnlyBasicRule, nil
	case PublicAppendName:
		return PublicAppendRule, nil
	case EACLPublicBasicName:
		return EACLPublicBasicRule, nil
	case EACLPrivateBasicName:
		return EACLPrivateBasicRule, nil
	case EACLReadOnlyBasicName:
		return EACLReadOnlyBasicRule, nil
	case EACLPublicAppendName:
		return EACLPublicAppendRule, nil
	default:
		basicACL = strings.TrimPrefix(strings.ToLower(basicACL), "0x")

		value, err := strconv.ParseUint(basicACL, 16, 32)
		if err != nil {
			return 0, fmt.Errorf("can't parse basic ACL: %s", basicACL)
		}

		return BasicACL(value), nil
	}
}
