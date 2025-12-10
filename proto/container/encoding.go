package container

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldContainerAttrKey
	fieldContainerAttrVal
)

// MarshaledSize returns size of the Container_Attribute in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *Container_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldContainerAttrKey, x.Key) +
			proto.SizeBytes(fieldContainerAttrVal, x.Value)
	}
	return sz
}

// MarshalStable writes the Container_Attribute in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [Container_Attribute.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *Container_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToBytes(b, fieldContainerAttrKey, x.Key)
		proto.MarshalToBytes(b[off:], fieldContainerAttrVal, x.Value)
	}
}

const (
	_ = iota
	fieldContainerVersion
	fieldContainerOwner
	fieldContainerNonce
	fieldContainerBasicACL
	fieldContainerAttributes
	fieldContainerPolicy
)

// MarshaledSize returns size of the Container in Protocol Buffers V3 format in
// bytes. MarshaledSize is NPE-safe.
func (x *Container) MarshaledSize() int {
	if x != nil {
		return proto.SizeEmbedded(fieldContainerVersion, x.Version) +
			proto.SizeEmbedded(fieldContainerOwner, x.OwnerId) +
			proto.SizeBytes(fieldContainerNonce, x.Nonce) +
			proto.SizeVarint(fieldContainerBasicACL, x.BasicAcl) +
			proto.SizeEmbedded(fieldContainerPolicy, x.PlacementPolicy) +
			proto.SizeRepeatedMessages(fieldContainerAttributes, x.Attributes)
	}
	return 0
}

// MarshalStable writes the Container in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [Container.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *Container) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldContainerVersion, x.Version)
		off += proto.MarshalToEmbedded(b[off:], fieldContainerOwner, x.OwnerId)
		off += proto.MarshalToBytes(b[off:], fieldContainerNonce, x.Nonce)
		off += proto.MarshalToVarint(b[off:], fieldContainerBasicACL, x.BasicAcl)
		off += proto.MarshalToRepeatedMessages(b[off:], fieldContainerAttributes, x.Attributes)
		proto.MarshalToEmbedded(b[off:], fieldContainerPolicy, x.PlacementPolicy)
	}
}

const (
	_ = iota
	fieldPutReqContainer
	fieldPutReqSignature
)

// MarshaledSize returns size of the PutRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldPutReqContainer, x.Container) +
			proto.SizeEmbedded(fieldPutReqSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the PutRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldPutReqContainer, x.Container)
		proto.MarshalToEmbedded(b[off:], fieldPutReqSignature, x.Signature)
	}
}

const (
	_ = iota
	fieldPutRespID
)

// MarshaledSize returns size of the PutResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldPutRespID, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the PutResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [PutResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldPutRespID, x.ContainerId)
	}
}

const (
	_ = iota
	fieldDeleteReqContainer
	fieldDeleteReqSignature
)

// MarshaledSize returns size of the DeleteRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldDeleteReqContainer, x.ContainerId) +
			proto.SizeEmbedded(fieldDeleteReqSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the DeleteRequest_Body in Protocol Buffers V3 format
// with ascending order of fields by number into b. MarshalStable uses exactly
// [DeleteRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldDeleteReqContainer, x.ContainerId)
		proto.MarshalToEmbedded(b[off:], fieldDeleteReqSignature, x.Signature)
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

const (
	_ = iota
	fieldGetReqContainer
)

// MarshaledSize returns size of the GetRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetReqContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldGetReqContainer, x.ContainerId)
	}
}

const (
	_ = iota
	fieldGetRespContainer
	fieldGetRespSignature
	fieldGetRespSession
)

// MarshaledSize returns size of the GetResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetRespContainer, x.Container) +
			proto.SizeEmbedded(fieldGetRespSignature, x.Signature) +
			proto.SizeEmbedded(fieldGetRespSession, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [GetResponse_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGetRespContainer, x.Container)
		off += proto.MarshalToEmbedded(b[off:], fieldGetRespSignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], fieldGetRespSession, x.SessionToken)
	}
}

const (
	_ = iota
	fieldListReqOwner
)

// MarshaledSize returns size of the ListRequest_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ListRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldListReqOwner, x.OwnerId)
	}
	return sz
}

// MarshalStable writes the ListRequest_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListRequest_Body.MarshaledSize] first bytes of b. MarshalStable is NPE-safe.
func (x *ListRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldListReqOwner, x.OwnerId)
	}
}

const (
	_ = iota
	fieldListRespIDs
)

// MarshaledSize returns size of the ListResponse_Body in Protocol Buffers V3
// format in bytes. MarshaledSize is NPE-safe.
func (x *ListResponse_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(fieldListRespIDs, x.ContainerIds)
	}
	return 0
}

// MarshalStable writes the ListResponse_Body in Protocol Buffers V3 format with
// ascending order of fields by number into b. MarshalStable uses exactly
// [ListResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *ListResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, fieldListRespIDs, x.ContainerIds)
	}
}

const (
	_ = iota
	fieldSetEACLReqTable
	fieldSetEACLReqSignature
)

// MarshaledSize returns size of the SetExtendedACLRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *SetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldSetEACLReqTable, x.Eacl) +
			proto.SizeEmbedded(fieldSetEACLReqSignature, x.Signature)
	}
	return sz
}

// MarshalStable writes the SetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [SetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSetEACLReqTable, x.Eacl)
		proto.MarshalToEmbedded(b[off:], fieldSetEACLReqSignature, x.Signature)
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

const (
	_ = iota
	fieldGetEACLReqContainer
)

// MarshaledSize returns size of the GetExtendedACLRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetEACLReqContainer, x.ContainerId)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldGetEACLReqContainer, x.ContainerId)
	}
}

const (
	_ = iota
	fieldGetEACLRespTable
	fieldGetEACLRespSignature
	fieldGetEACLRespSession
)

// MarshaledSize returns size of the GetExtendedACLResponse_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *GetExtendedACLResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldGetEACLRespTable, x.Eacl) +
			proto.SizeEmbedded(fieldGetEACLRespSignature, x.Signature) +
			proto.SizeEmbedded(fieldGetEACLRespSession, x.SessionToken)
	}
	return sz
}

// MarshalStable writes the GetExtendedACLResponse_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [GetExtendedACLResponse_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *GetExtendedACLResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldGetEACLRespTable, x.Eacl)
		off += proto.MarshalToEmbedded(b[off:], fieldGetEACLRespSignature, x.Signature)
		proto.MarshalToEmbedded(b[off:], fieldGetEACLRespSession, x.SessionToken)
	}
}

const (
	_ = iota
	fieldUsedSpaceEpoch
	fieldUsedSpaceContainer
	fieldUsedSpaceValue
)

// MarshaledSize returns size of the AnnounceUsedSpaceRequest_Body_Announcement
// in Protocol Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body_Announcement) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldUsedSpaceEpoch, x.Epoch) +
			proto.SizeEmbedded(fieldUsedSpaceContainer, x.ContainerId) +
			proto.SizeVarint(fieldUsedSpaceValue, x.UsedSpace)
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
		off := proto.MarshalToVarint(b, fieldUsedSpaceEpoch, x.Epoch)
		off += proto.MarshalToEmbedded(b[off:], fieldUsedSpaceContainer, x.ContainerId)
		proto.MarshalToVarint(b[off:], fieldUsedSpaceValue, x.UsedSpace)
	}
}

const (
	_ = iota
	fieldAnnounceReqList
)

// MarshaledSize returns size of the AnnounceUsedSpaceRequest_Body in Protocol
// Buffers V3 format in bytes. MarshaledSize is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body) MarshaledSize() int {
	if x != nil {
		return proto.SizeRepeatedMessages(fieldAnnounceReqList, x.Announcements)
	}
	return 0
}

// MarshalStable writes the AnnounceUsedSpaceRequest_Body in Protocol Buffers V3
// format with ascending order of fields by number into b. MarshalStable uses
// exactly [AnnounceUsedSpaceRequest_Body.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *AnnounceUsedSpaceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToRepeatedMessages(b, fieldAnnounceReqList, x.Announcements)
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

const (
	_ = iota
	fieldSetAttributePrmID
	fieldSetAttributePrmAttribute
	fieldSetAttributePrmValue
	fieldSetAttributePrmValidUntil
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SetAttributeRequest_Body_Parameters) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldSetAttributePrmID, x.ContainerId) +
			proto.SizeBytes(fieldSetAttributePrmAttribute, x.Attribute) +
			proto.SizeBytes(fieldSetAttributePrmValue, x.Value) +
			proto.SizeVarint(fieldSetAttributePrmValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *SetAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSetAttributePrmID, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], fieldSetAttributePrmAttribute, x.Attribute)
		off += proto.MarshalToBytes(b[off:], fieldSetAttributePrmValue, x.Value)
		proto.MarshalToVarint(b[off:], fieldSetAttributePrmValidUntil, x.ValidUntil)
	}
}

const (
	_ = iota
	fieldSetAttributeReqParams
	fieldSetAttributeReqSignature
	fieldSetAttributeReqSession
	fieldSetAttributeReqSessionV1
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SetAttributeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldSetAttributeReqParams, x.Parameters) +
			proto.SizeEmbedded(fieldSetAttributeReqSignature, x.Signature) +
			proto.SizeEmbedded(fieldSetAttributeReqSession, x.SessionToken) +
			proto.SizeEmbedded(fieldSetAttributeReqSessionV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SetAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldSetAttributeReqParams, x.Parameters)
		off += proto.MarshalToEmbedded(b[off:], fieldSetAttributeReqSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], fieldSetAttributeReqSession, x.SessionToken)
		proto.MarshalToEmbedded(b[off:], fieldSetAttributeReqSessionV1, x.SessionTokenV1)
	}
}

const (
	_ = iota
	fieldSetAttributeRespStatus
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *SetAttributeResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldSetAttributeRespStatus, x.Status)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [SetAttributeResponse_Body.MarshaledSize] first bytes of b. MarshalStable is
// NPE-safe.
func (x *SetAttributeResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldSetAttributeRespStatus, x.Status)
	}
}

const (
	_ = iota
	fieldRemoveAttributePrmID
	fieldRemoveAttributePrmAttribute
	fieldRemoveAttributePrmValidUntil
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *RemoveAttributeRequest_Body_Parameters) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldRemoveAttributePrmID, x.ContainerId) +
			proto.SizeBytes(fieldRemoveAttributePrmAttribute, x.Attribute) +
			proto.SizeVarint(fieldRemoveAttributePrmValidUntil, x.ValidUntil)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body_Parameters.MarshaledSize] first bytes of b.
// MarshalStable is NPE-safe.
func (x *RemoveAttributeRequest_Body_Parameters) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRemoveAttributePrmID, x.ContainerId)
		off += proto.MarshalToBytes(b[off:], fieldRemoveAttributePrmAttribute, x.Attribute)
		proto.MarshalToVarint(b[off:], fieldRemoveAttributePrmValidUntil, x.ValidUntil)
	}
}

const (
	_ = iota
	fieldRemoveAttributeReqParams
	fieldRemoveAttributeReqSignature
	fieldRemoveAttributeReqSession
	fieldRemoveAttributeReqSessionV1
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *RemoveAttributeRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldRemoveAttributeReqParams, x.Parameters) +
			proto.SizeEmbedded(fieldRemoveAttributeReqSignature, x.Signature) +
			proto.SizeEmbedded(fieldRemoveAttributeReqSession, x.SessionToken) +
			proto.SizeEmbedded(fieldRemoveAttributeReqSessionV1, x.SessionTokenV1)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeRequest_Body.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *RemoveAttributeRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalToEmbedded(b, fieldRemoveAttributeReqParams, x.Parameters)
		off += proto.MarshalToEmbedded(b[off:], fieldRemoveAttributeReqSignature, x.Signature)
		off += proto.MarshalToEmbedded(b[off:], fieldRemoveAttributeReqSession, x.SessionToken)
		proto.MarshalToEmbedded(b[off:], fieldRemoveAttributeReqSessionV1, x.SessionTokenV1)
	}
}

const (
	_ = iota
	fieldRemoveAttributeRespStatus
)

// MarshaledSize returns size of the x in Protocol Buffers V3 format in bytes.
// MarshaledSize is NPE-safe.
func (x *RemoveAttributeResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeEmbedded(fieldRemoveAttributeRespStatus, x.Status)
	}
	return sz
}

// MarshalStable writes x in Protocol Buffers V3 format with ascending order of
// fields by number into b. MarshalStable uses exactly
// [RemoveAttributeResponse_Body.MarshaledSize] first bytes of b. MarshalStable
// is NPE-safe.
func (x *RemoveAttributeResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalToEmbedded(b, fieldRemoveAttributeRespStatus, x.Status)
	}
}
