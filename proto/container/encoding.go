package container

import (
	protoencoding "github.com/nspcc-dev/neofs-sdk-go/proto/encoding"
)

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
		sz = protoencoding.SizeBytes(FieldContainerAttributeKey, x.Key) +
			protoencoding.SizeBytes(FieldContainerAttributeValue, x.Value)
	}
	return sz
}

// MarshalStable writes the Container_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [Container_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *Container_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToBytes(b, FieldContainerAttributeKey, x.Key)
		protoencoding.MarshalToBytes(b[off:], FieldContainerAttributeValue, x.Value)
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
		return protoencoding.SizeEmbedded(FieldContainerVersion, x.Version) +
			protoencoding.SizeEmbedded(FieldContainerOwnerID, x.OwnerId) +
			protoencoding.SizeBytes(FieldContainerNonce, x.Nonce) +
			protoencoding.SizeVarint(FieldContainerBasicACL, x.BasicAcl) +
			protoencoding.SizeEmbedded(FieldContainerPolicy, x.PlacementPolicy) +
			protoencoding.SizeRepeatedMessages(FieldContainerAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the Container in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Container.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Container) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldContainerVersion, x.Version)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldContainerOwnerID, x.OwnerId)
		off += protoencoding.MarshalToBytes(b[off:], FieldContainerNonce, x.Nonce)
		off += protoencoding.MarshalToVarint(b[off:], FieldContainerBasicACL, x.BasicAcl)
		off += protoencoding.MarshalToRepeatedMessages(b[off:], FieldContainerAttributes, x.Attributes)
		protoencoding.MarshalToEmbedded(b[off:], FieldContainerPolicy, x.PlacementPolicy)
	}
}

// Field numbers of [PutRequest_Body] message.
const (
	_ = iota
	FieldPutRequestBodyContainer
	FieldPutRequestBodySignature
	FieldPutRequestBodyEacl
	FieldPutRequestBodyEaclSignature
)

// MarshaledSize returns size of the PutRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = protoencoding.SizeEmbedded(FieldPutRequestBodyContainer, x.Container) +
			protoencoding.SizeEmbedded(FieldPutRequestBodySignature, x.Signature) +
			protoencoding.SizeEmbedded(FieldPutRequestBodyEacl, x.Eacl) +
			protoencoding.SizeEmbedded(FieldPutRequestBodyEaclSignature, x.EaclSignature)
	}
	return sz
}

// MarshalStable writes the PutRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldPutRequestBodyContainer, x.Container)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldPutRequestBodySignature, x.Signature)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldPutRequestBodyEacl, x.Eacl)
		protoencoding.MarshalToEmbedded(b[off:], FieldPutRequestBodyEaclSignature, x.EaclSignature)
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
		sz = protoencoding.SizeEmbedded(FieldPutResponseBodyID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the PutResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldPutResponseBodyID, x.ContainerId)
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
		sz = protoencoding.SizeEmbedded(FieldDeleteRequestBodyContainerID, x.ContainerId) +
			protoencoding.SizeEmbedded(FieldDeleteRequestBodySignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the DeleteRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldDeleteRequestBodyContainerID, x.ContainerId)
		protoencoding.MarshalToEmbedded(b[off:], FieldDeleteRequestBodySignature, x.Signature)
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
		sz = protoencoding.SizeEmbedded(FieldGetRequestBodyContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldGetRequestBodyContainer, x.ContainerId)
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
		sz = protoencoding.SizeEmbedded(FieldGetResponseBodyContainer, x.Container) +
			protoencoding.SizeEmbedded(FieldGetResponseBodySignature, x.Signature) +
			protoencoding.SizeEmbedded(FieldGetResponseBodySessionToken, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldGetResponseBodyContainer, x.Container)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldGetResponseBodySignature, x.Signature)
		protoencoding.MarshalToEmbedded(b[off:], FieldGetResponseBodySessionToken, x.SessionToken)
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
		sz = protoencoding.SizeEmbedded(FieldListRequestBodyOwner, x.OwnerId)
	}
	return sz
}

// MarshalStable writes the ListRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ListRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldListRequestBodyOwner, x.OwnerId)
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
		return protoencoding.SizeRepeatedMessages(FieldListResponseBodyContainerIDs, x.ContainerIds)
	}
	return 0
}

// MarshalStable writes the ListResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ListResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToRepeatedMessages(b, FieldListResponseBodyContainerIDs, x.ContainerIds)
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
		sz = protoencoding.SizeEmbedded(FieldSetExtendedACLRequestBodyEACL, x.Eacl) +
			protoencoding.SizeEmbedded(FieldSetExtendedACLRequestBodySignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSetExtendedACLRequestBodyEACL, x.Eacl)
		protoencoding.MarshalToEmbedded(b[off:], FieldSetExtendedACLRequestBodySignature, x.Signature)
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
		sz = protoencoding.SizeEmbedded(FieldGetExtendedACLRequestBodyContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToEmbedded(b, FieldGetExtendedACLRequestBodyContainer, x.ContainerId)
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
		sz = protoencoding.SizeEmbedded(FieldGetExtendedACLResponseBodyEACL, x.Eacl) +
			protoencoding.SizeEmbedded(FieldGetExtendedACLResponseBodySignature, x.Signature) +
			protoencoding.SizeEmbedded(FieldGetExtendedACLResponseBodySessionToken, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldGetExtendedACLResponseBodyEACL, x.Eacl)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldGetExtendedACLResponseBodySignature, x.Signature)
		protoencoding.MarshalToEmbedded(b[off:], FieldGetExtendedACLResponseBodySessionToken, x.SessionToken)
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
		sz = protoencoding.SizeVarint(FieldAnnounceUsedSpaceRequestBodyAnnouncementEpoch, x.Epoch) +
			protoencoding.SizeEmbedded(FieldAnnounceUsedSpaceRequestBodyAnnouncementContainerID, x.ContainerId) +
			protoencoding.SizeVarint(FieldAnnounceUsedSpaceRequestBodyAnnouncementUsedSpace, x.UsedSpace)
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
		off := protoencoding.MarshalToVarint(b, FieldAnnounceUsedSpaceRequestBodyAnnouncementEpoch, x.Epoch)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldAnnounceUsedSpaceRequestBodyAnnouncementContainerID, x.ContainerId)
		protoencoding.MarshalToVarint(b[off:], FieldAnnounceUsedSpaceRequestBodyAnnouncementUsedSpace, x.UsedSpace)
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
		return protoencoding.SizeRepeatedMessages(FieldAnnounceUsedSpaceRequestBodyAnnouncements, x.Announcements)
	}
	return 0
}

// MarshalStable writes the AnnounceUsedSpaceRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceUsedSpaceRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		protoencoding.MarshalToRepeatedMessages(b, FieldAnnounceUsedSpaceRequestBodyAnnouncements, x.Announcements)
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
		sz = protoencoding.SizeEmbedded(FieldSetAttributeRequestBodyParametersContainerID, x.ContainerId) +
			protoencoding.SizeBytes(FieldSetAttributeRequestBodyParametersAttribute, x.Attribute) +
			protoencoding.SizeBytes(FieldSetAttributeRequestBodyParametersValue, x.Value) +
			protoencoding.SizeVarint(FieldSetAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSetAttributeRequestBodyParametersContainerID, x.ContainerId)
		off += protoencoding.MarshalToBytes(b[off:], FieldSetAttributeRequestBodyParametersAttribute, x.Attribute)
		off += protoencoding.MarshalToBytes(b[off:], FieldSetAttributeRequestBodyParametersValue, x.Value)
		protoencoding.MarshalToVarint(b[off:], FieldSetAttributeRequestBodyParametersValidUntil, x.ValidUntil)
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
		sz = protoencoding.SizeEmbedded(FieldSetAttributeRequestBodyParameters, x.Parameters) +
			protoencoding.SizeEmbedded(FieldSetAttributeRequestBodySignature, x.Signature) +
			protoencoding.SizeEmbedded(FieldSetAttributeRequestBodySessionToken, x.SessionToken) +
			protoencoding.SizeEmbedded(FieldSetAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SetAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldSetAttributeRequestBodyParameters, x.Parameters)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySignature, x.Signature)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySessionToken, x.SessionToken)
		protoencoding.MarshalToEmbedded(b[off:], FieldSetAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
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
		sz = protoencoding.SizeEmbedded(FieldRemoveAttributeRequestBodyParametersContainerID, x.ContainerId) +
			protoencoding.SizeBytes(FieldRemoveAttributeRequestBodyParametersAttribute, x.Attribute) +
			protoencoding.SizeVarint(FieldRemoveAttributeRequestBodyParametersValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RemoveAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldRemoveAttributeRequestBodyParametersContainerID, x.ContainerId)
		off += protoencoding.MarshalToBytes(b[off:], FieldRemoveAttributeRequestBodyParametersAttribute, x.Attribute)
		protoencoding.MarshalToVarint(b[off:], FieldRemoveAttributeRequestBodyParametersValidUntil, x.ValidUntil)
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
		sz = protoencoding.SizeEmbedded(FieldRemoveAttributeRequestBodyParameters, x.Parameters) +
			protoencoding.SizeEmbedded(FieldRemoveAttributeRequestBodySignature, x.Signature) +
			protoencoding.SizeEmbedded(FieldRemoveAttributeRequestBodySessionTOken, x.SessionToken) +
			protoencoding.SizeEmbedded(FieldRemoveAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *RemoveAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := protoencoding.MarshalToEmbedded(b, FieldRemoveAttributeRequestBodyParameters, x.Parameters)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySignature, x.Signature)
		off += protoencoding.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySessionTOken, x.SessionToken)
		protoencoding.MarshalToEmbedded(b[off:], FieldRemoveAttributeRequestBodySessionTokenV1, x.SessionTokenV1)
	}
}
