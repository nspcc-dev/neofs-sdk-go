package container

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

// Field numbers of [Container_Attribute] message.
const (
	_ = iota
	FieldContainerAttributeKey
	FieldContainerAttributeValue
)

// MarshaledSize returns size of the Container_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Container_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(FieldContainerAttributeKey, x.Key) +
			proto.SizeBytes(FieldContainerAttributeValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Container_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [Container_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *Container_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, FieldContainerAttributeKey, x.Key)
		proto.MarshalToBytes(b[off:], FieldContainerAttributeValue, x.Value)
	}
}

// Field numbers of [Container] message.
const (
	_ = iota
	FieldContainerVersion
	FieldContainerOwnerID
	FieldContainerNonce
	FieldContainerBasicACL
	FieldContainerAttributes
	FieldContainerPolicy
)

// MarshaledSize returns size of the Container in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Container) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(FieldContainerVersion, x.Version) +
			proto.SizeEmbedded(FieldContainerOwnerID, x.OwnerId) +
			proto.SizeBytes(FieldContainerNonce, x.Nonce) +
			proto.SizeVarint(FieldContainerBasicACL, x.BasicAcl) +
			proto.SizeEmbedded(FieldContainerPolicy, x.PlacementPolicy) +
			proto.SizeRepeatedMessages(FieldContainerAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the Container in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Container.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Container) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldContainerVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], FieldContainerOwnerID, x.OwnerId)
		off += proto.MarshalToBytes(b[off:], FieldContainerNonce, x.Nonce)
		off += proto.MarshalToVarint(b[off:], FieldContainerBasicACL, x.BasicAcl)
		off += proto.MarshalToRepeatedMessages(b[off:], FieldContainerAttributes, x.Attributes)
		proto.MarshalToEmbedded(b[off:], FieldContainerPolicy, x.PlacementPolicy)
	}
}

// Field numbers of [PutRequest_Body] message.
const (
	_ = iota
	FieldPutRequestBodyContainer
	FieldPutRequestBodySignature
)

// MarshaledSize returns size of the PutRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldPutRequestBodyContainer, x.Container) +
			proto.SizeEmbedded(FieldPutRequestBodySignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the PutRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldPutRequestBodyContainer, x.Container)
		proto.MarshalToEmbedded(b[off:], FieldPutRequestBodySignature, x.Signature)
	}
}

// Field numbers of [PutResponse_Body] message.
const (
	_ = iota
	FieldPutResponseBodyID
)

// MarshaledSize returns size of the PutResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldPutResponseBodyID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the PutResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldPutResponseBodyID, x.ContainerId)
	}
}

// Field numbers of [DeleteRequest_Body] message.
const (
	_ = iota
	FieldDeleteRequestBodyContainerID
	FieldDeleteRequestBodySignature
)

// MarshaledSize returns size of the DeleteRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldDeleteRequestBodyContainerID, x.ContainerId) +
			proto.SizeEmbedded(FieldDeleteRequestBodySignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the DeleteRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldDeleteRequestBodyContainerID, x.ContainerId)
		proto.MarshalToEmbedded(b[off:], FieldDeleteRequestBodySignature, x.Signature)
	}
}

// MarshaledSize returns size of the DeleteResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteResponse_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the DeleteResponse_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteResponse_Body) MarshalStable([]byte) {}

// Field numbers of [GetRequest_Body] message.
const (
	_ = iota
	FieldGetRequestBodyContainer
)

// MarshaledSize returns size of the GetRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetRequestBodyContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldGetRequestBodyContainer, x.ContainerId)
	}
}

// Field numbers of [GetResponse_Body] message.
const (
	_ = iota
	FieldGetResponseBodyContainer
	FieldGetResponseBodySignature
	FieldGetResponseBodySessionToken
)

// MarshaledSize returns size of the GetResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetResponseBodyContainer, x.Container) +
			proto.SizeEmbedded(FieldGetResponseBodySignature, x.Signature) +
			proto.SizeEmbedded(FieldGetResponseBodySessionToken, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGetResponseBodyContainer, x.Container)
		off += proto.MarshalToEmbedded(b[off:], FieldGetResponseBodySignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], FieldGetResponseBodySessionToken, x.SessionToken)
	}
}

// Field numbers of [ListRequest_Body] message.
const (
	_ = iota
	FieldListRequestBodyOwner
)

// MarshaledSize returns size of the ListRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ListRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldListRequestBodyOwner, x.OwnerId)
	}
	return sz
}

// MarshalStable writes the ListRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ListRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldListRequestBodyOwner, x.OwnerId)
	}
}

// Field numbers of [ListResponse_Body] message.
const (
	_ = iota
	FieldListResponseBodyContainerIDs
)

// MarshaledSize returns size of the ListResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ListResponse_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(FieldListResponseBodyContainerIDs, x.ContainerIds)
	}
	return 0
}

// MarshalStable writes the ListResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ListResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, FieldListResponseBodyContainerIDs, x.ContainerIds)
	}
}

// Field numbers of [SetExtendedACLRequest_Body] message.
const (
	_ = iota
	FieldSetExtendedACLRequestBodyEACL
	FieldSetExtendedACLRequestBodySignature
)

// MarshaledSize returns size of the SetExtendedACLRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldSetExtendedACLRequestBodyEACL, x.Eacl) +
			proto.SizeEmbedded(FieldSetExtendedACLRequestBodySignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSetExtendedACLRequestBodyEACL, x.Eacl)
		proto.MarshalToEmbedded(b[off:], FieldSetExtendedACLRequestBodySignature, x.Signature)
	}
}

// MarshaledSize returns size of the SetExtendedACLResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SetExtendedACLResponse_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the SetExtendedACLResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SetExtendedACLResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetExtendedACLResponse_Body) MarshalStable([]byte) {}

// Field numbers of [GetExtendedACLRequest_Body] message.
const (
	_ = iota
	FieldGetExtendedACLRequestBodyContainer
)

// MarshaledSize returns size of the GetExtendedACLRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetExtendedACLRequestBodyContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, FieldGetExtendedACLRequestBodyContainer, x.ContainerId)
	}
}

// Field numbers of [GetExtendedACLResponse_Body] message.
const (
	_ = iota
	FieldGetExtendedACLResponseBodyEACL
	FieldGetExtendedACLResponseBodySignature
	FieldGetExtendedACLResponseBodySessionToken
)

// MarshaledSize returns size of the GetExtendedACLResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetExtendedACLResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldGetExtendedACLResponseBodyEACL, x.Eacl) +
			proto.SizeEmbedded(FieldGetExtendedACLResponseBodySignature, x.Signature) +
			proto.SizeEmbedded(FieldGetExtendedACLResponseBodySessionToken, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldGetExtendedACLResponseBodyEACL, x.Eacl)
		off += proto.MarshalToEmbedded(b[off:], FieldGetExtendedACLResponseBodySignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], FieldGetExtendedACLResponseBodySessionToken, x.SessionToken)
	}
}

// Field numbers of [AnnounceUsedSpaceRequest_Body_Announcement] message.
const (
	_ = iota
	FieldAnnounceUsedSpaceRequestBodyAnnouncementEpoch
	FieldAnnounceUsedSpaceRequestBodyAnnouncementContainerID
	FieldAnnounceUsedSpaceRequestBodyAnnouncementUsedSpace
)

// MarshaledSize returns size of the AnnounceUsedSpaceRequest_Body_Announcement
// in Protocol Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body_Announcement) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(FieldAnnounceUsedSpaceRequestBodyAnnouncementEpoch, x.Epoch) +
			proto.SizeEmbedded(FieldAnnounceUsedSpaceRequestBodyAnnouncementContainerID, x.ContainerId) +
			proto.SizeVarint(FieldAnnounceUsedSpaceRequestBodyAnnouncementUsedSpace, x.UsedSpace)
	}
	return sz
}

// MarshalStable writes the AnnounceUsedSpaceRequest_Body_Announcement in
// Protocol Buffers V3 format with ascending order of fields by number into b.
// MarshalStable uses exactly
// [AnnounceUsedSpaceRequest_Body_Announcement.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body_Announcement) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToVarint(b, FieldAnnounceUsedSpaceRequestBodyAnnouncementEpoch, x.Epoch)
		off += proto.MarshalToEmbedded(b[off:], FieldAnnounceUsedSpaceRequestBodyAnnouncementContainerID, x.ContainerId)
		proto.MarshalToVarint(b[off:], FieldAnnounceUsedSpaceRequestBodyAnnouncementUsedSpace, x.UsedSpace)
	}
}

// Field numbers of [AnnounceUsedSpaceRequest_Body] message.
const (
	_ = iota
	FieldAnnounceUsedSpaceRequestBodyAnnouncements
)

// MarshaledSize returns size of the AnnounceUsedSpaceRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(FieldAnnounceUsedSpaceRequestBodyAnnouncements, x.Announcements)
	}
	return 0
}

// MarshalStable writes the AnnounceUsedSpaceRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceUsedSpaceRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, FieldAnnounceUsedSpaceRequestBodyAnnouncements, x.Announcements)
	}
}

// MarshaledSize returns size of the AnnounceUsedSpaceResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceUsedSpaceResponse_Body) MarshaledSize() int { return 0 }

// MarshalStable writes the AnnounceUsedSpaceResponse_Body in Protocol Buffers
// V3 format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceUsedSpaceResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceUsedSpaceResponse_Body) MarshalStable([]byte) {}

// Field numbers of [SetAttributeRequest_Body_Parameters] message.
const (
	_ = iota
	FieldSetAttributeRequestBodyParametersContainerID
	FieldSetAttributeRequestBodyParametersAttribute
	FieldSetAttributeRequestBodyParametersValue
	FieldSetAttributeRequestBodyParametersValidUntil
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SetAttributeRequest_Body_Parameters) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldSetAttributeRequestBodyParametersContainerID, x.ContainerId) +
			proto.SizeBytes(FieldSetAttributeRequestBodyParametersAttribute, x.Attribute) +
			proto.SizeBytes(FieldSetAttributeRequestBodyParametersValue, x.Value) +
			proto.SizeVarint(FieldSetAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSetAttributeRequestBodyParametersContainerID, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], FieldSetAttributeRequestBodyParametersAttribute, x.Attribute)
		off += proto.MarshalToBytes(b[off:], FieldSetAttributeRequestBodyParametersValue, x.Value)
		proto.MarshalToVarint(b[off:], FieldSetAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
}

// Field numbers of [SetAttributeRequest_Body] message.
const (
	_ = iota
	FieldSetAttributeRequestBodyParameters
	FieldSetAttributeRequestBodySignature
	FieldSetAttributeRequestBodySessionToken
	FieldSetAttributeRequestBodySessionTokenV1
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SetAttributeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldSetAttributeRequestBodyParameters, x.Parameters) +
			proto.SizeEmbedded(FieldSetAttributeRequestBodySignature, x.Signature) +
			proto.SizeEmbedded(FieldSetAttributeRequestBodySessionToken, x.SessionToken) +
			proto.SizeEmbedded(FieldSetAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SetAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldSetAttributeRequestBodyParameters, x.Parameters)
		off += proto.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySessionToken, x.SessionToken)
		proto.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
}

// Field numbers of [RemoveAttributeRequest_Body_Parameters] message.
const (
	_ = iota
	FieldRemoveAttributeRequestBodyParametersContainerID
	FieldRemoveAttributeRequestBodyParametersAttribute
	FieldRemoveAttributeRequestBodyParametersValidUntil
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *RemoveAttributeRequest_Body_Parameters) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldRemoveAttributeRequestBodyParametersContainerID, x.ContainerId) +
			proto.SizeBytes(FieldRemoveAttributeRequestBodyParametersAttribute, x.Attribute) +
			proto.SizeVarint(FieldRemoveAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RemoveAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRemoveAttributeRequestBodyParametersContainerID, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], FieldRemoveAttributeRequestBodyParametersAttribute, x.Attribute)
		proto.MarshalToVarint(b[off:], FieldRemoveAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
}

// Field numbers of [RemoveAttributeRequest_Body] message.
const (
	_ = iota
	FieldRemoveAttributeRequestBodyParameters
	FieldRemoveAttributeRequestBodySignature
	FieldRemoveAttributeRequestBodySessionTOken
	FieldRemoveAttributeRequestBodySessionTokenV1
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *RemoveAttributeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(FieldRemoveAttributeRequestBodyParameters, x.Parameters) +
			proto.SizeEmbedded(FieldRemoveAttributeRequestBodySignature, x.Signature) +
			proto.SizeEmbedded(FieldRemoveAttributeRequestBodySessionTOken, x.SessionToken) +
			proto.SizeEmbedded(FieldRemoveAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *RemoveAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, FieldRemoveAttributeRequestBodyParameters, x.Parameters)
		off += proto.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySessionTOken, x.SessionToken)
		proto.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
}
