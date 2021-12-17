package acl

const (
	// PublicBasicRule is a basic ACL value for final public-read-write container for which extended ACL CANNOT be set.
	PublicBasicRule = 0x1FBFBFFF

	// PrivateBasicRule is a basic ACL value for final private container for which extended ACL CANNOT be set.
	PrivateBasicRule = 0x1C8C8CCC

	// ReadOnlyBasicRule is a basic ACL value for final public-read container for which extended ACL CANNOT be set.
	ReadOnlyBasicRule = 0x1FBF8CFF

	// PublicAppendRule is a basic ACL value for final public-append container for which extended ACL CANNOT be set.
	PublicAppendRule = 0x1FBF9FFF

	// EACLPublicBasicRule is a basic ACL value for non-final public-read-write container for which extended ACL CAN be set.
	EACLPublicBasicRule = 0x0FBFBFFF

	// EACLPrivateBasicRule is a basic ACL value for non-final private container for which extended ACL CAN be set.
	EACLPrivateBasicRule = 0x0C8C8CCC

	// EACLReadOnlyBasicRule is a basic ACL value for non-final public-read container for which extended ACL CAN be set.
	EACLReadOnlyBasicRule = 0x0FBF8CFF

	// EACLPublicAppendRule is a basic ACL value for non-final public-append container for which extended ACL CAN be set.
	EACLPublicAppendRule = 0x0FBF9FFF
)
