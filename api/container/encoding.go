package container

import "github.com/nspcc-dev/neofs-sdk-go/internal/proto"

const (
	_ = iota
	fieldContainerAttrKey
	fieldContainerAttrVal
)

func (x *Container_Attribute) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeBytes(fieldContainerAttrKey, x.Key) +
			proto.SizeBytes(fieldContainerAttrVal, x.Value)
	}
	return sz
}

func (x *Container_Attribute) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalBytes(b, fieldContainerAttrKey, x.Key)
		proto.MarshalBytes(b[off:], fieldContainerAttrVal, x.Value)
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

func (x *Container) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldContainerVersion, x.Version) +
			proto.SizeNested(fieldContainerOwner, x.OwnerId) +
			proto.SizeBytes(fieldContainerNonce, x.Nonce) +
			proto.SizeVarint(fieldContainerBasicACL, x.BasicAcl) +
			proto.SizeNested(fieldContainerPolicy, x.PlacementPolicy)
		for i := range x.Attributes {
			sz += proto.SizeNested(fieldContainerAttributes, x.Attributes[i])
		}
	}
	return sz
}

func (x *Container) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldContainerVersion, x.Version)
		off += proto.MarshalNested(b[off:], fieldContainerOwner, x.OwnerId)
		off += proto.MarshalBytes(b[off:], fieldContainerNonce, x.Nonce)
		off += proto.MarshalVarint(b[off:], fieldContainerBasicACL, x.BasicAcl)
		for i := range x.Attributes {
			off += proto.MarshalNested(b[off:], fieldContainerAttributes, x.Attributes[i])
		}
		proto.MarshalNested(b[off:], fieldContainerPolicy, x.PlacementPolicy)
	}
}

const (
	_ = iota
	fieldPutReqContainer
	fieldPutReqSignature
)

func (x *PutRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldPutReqContainer, x.Container) +
			proto.SizeNested(fieldPutReqSignature, x.Signature)
	}
	return sz
}

func (x *PutRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldPutReqContainer, x.Container)
		proto.MarshalNested(b[off:], fieldPutReqSignature, x.Signature)
	}
}

const (
	_ = iota
	fieldPutRespID
)

func (x *PutResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldPutRespID, x.ContainerId)
	}
	return sz
}

func (x *PutResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldPutRespID, x.ContainerId)
	}
}

const (
	_ = iota
	fieldDeleteReqContainer
	fieldDeleteReqSignature
)

func (x *DeleteRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldDeleteReqContainer, x.ContainerId) +
			proto.SizeNested(fieldDeleteReqSignature, x.Signature)
	}
	return sz
}

func (x *DeleteRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldDeleteReqContainer, x.ContainerId)
		proto.MarshalNested(b[off:], fieldDeleteReqSignature, x.Signature)
	}
}

func (x *DeleteResponse_Body) MarshaledSize() int   { return 0 }
func (x *DeleteResponse_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldGetReqContainer
)

func (x *GetRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetReqContainer, x.ContainerId)
	}
	return sz
}

func (x *GetRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldGetReqContainer, x.ContainerId)
	}
}

const (
	_ = iota
	fieldGetRespContainer
	fieldGetRespSignature
	fieldGetRespSession
)

func (x *GetResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetRespContainer, x.Container) +
			proto.SizeNested(fieldGetRespSignature, x.Signature) +
			proto.SizeNested(fieldGetRespSession, x.SessionToken)
	}
	return sz
}

func (x *GetResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldGetRespContainer, x.Container)
		off += proto.MarshalNested(b[off:], fieldGetRespSignature, x.Signature)
		proto.MarshalNested(b[off:], fieldGetRespSession, x.SessionToken)
	}
}

const (
	_ = iota
	fieldListReqOwner
)

func (x *ListRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldListReqOwner, x.OwnerId)
	}
	return sz
}

func (x *ListRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldListReqOwner, x.OwnerId)
	}
}

const (
	_ = iota
	fieldListRespIDs
)

func (x *ListResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.ContainerIds {
			sz += proto.SizeNested(fieldListRespIDs, x.ContainerIds[i])
		}
	}
	return sz
}

func (x *ListResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.ContainerIds {
			off += proto.MarshalNested(b[off:], fieldListRespIDs, x.ContainerIds[i])
		}
	}
}

const (
	_ = iota
	fieldSetEACLReqTable
	fieldSetEACLReqSignature
)

func (x *SetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldSetEACLReqTable, x.Eacl) +
			proto.SizeNested(fieldSetEACLReqSignature, x.Signature)
	}
	return sz
}

func (x *SetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldSetEACLReqTable, x.Eacl)
		proto.MarshalNested(b[off:], fieldSetEACLReqSignature, x.Signature)
	}
}

func (x *SetExtendedACLResponse_Body) MarshaledSize() int   { return 0 }
func (x *SetExtendedACLResponse_Body) MarshalStable([]byte) {}

const (
	_ = iota
	fieldGetEACLReqContainer
)

func (x *GetExtendedACLRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetEACLReqContainer, x.ContainerId)
	}
	return sz
}

func (x *GetExtendedACLRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		proto.MarshalNested(b, fieldGetEACLReqContainer, x.ContainerId)
	}
}

const (
	_ = iota
	fieldGetEACLRespTable
	fieldGetEACLRespSignature
	fieldGetEACLRespSession
)

func (x *GetExtendedACLResponse_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeNested(fieldGetEACLRespTable, x.Eacl) +
			proto.SizeNested(fieldGetEACLRespSignature, x.Signature) +
			proto.SizeNested(fieldGetEACLRespSession, x.SessionToken)
	}
	return sz
}

func (x *GetExtendedACLResponse_Body) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalNested(b, fieldGetEACLRespTable, x.Eacl)
		off += proto.MarshalNested(b[off:], fieldGetEACLRespSignature, x.Signature)
		proto.MarshalNested(b[off:], fieldGetEACLRespSession, x.SessionToken)
	}
}

const (
	_ = iota
	fieldUsedSpaceEpoch
	fieldUsedSpaceContainer
	fieldUsedSpaceValue
)

func (x *AnnounceUsedSpaceRequest_Body_Announcement) MarshaledSize() int {
	var sz int
	if x != nil {
		sz = proto.SizeVarint(fieldUsedSpaceEpoch, x.Epoch) +
			proto.SizeNested(fieldUsedSpaceContainer, x.ContainerId) +
			proto.SizeVarint(fieldUsedSpaceValue, x.UsedSpace)
	}
	return sz
}

func (x *AnnounceUsedSpaceRequest_Body_Announcement) MarshalStable(b []byte) {
	if x != nil {
		off := proto.MarshalVarint(b, fieldUsedSpaceEpoch, x.Epoch)
		off += proto.MarshalNested(b[off:], fieldUsedSpaceContainer, x.ContainerId)
		proto.MarshalVarint(b[off:], fieldUsedSpaceValue, x.UsedSpace)
	}
}

const (
	_ = iota
	fieldAnnounceReqList
)

func (x *AnnounceUsedSpaceRequest_Body) MarshaledSize() int {
	var sz int
	if x != nil {
		for i := range x.Announcements {
			sz += proto.SizeNested(fieldAnnounceReqList, x.Announcements[i])
		}
	}
	return sz
}

func (x *AnnounceUsedSpaceRequest_Body) MarshalStable(b []byte) {
	if x != nil {
		var off int
		for i := range x.Announcements {
			off += proto.MarshalNested(b[off:], fieldAnnounceReqList, x.Announcements[i])
		}
	}
}

func (x *AnnounceUsedSpaceResponse_Body) MarshaledSize() int   { return 0 }
func (x *AnnounceUsedSpaceResponse_Body) MarshalStable([]byte) {}
