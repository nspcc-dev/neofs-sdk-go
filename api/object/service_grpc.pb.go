// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: object/grpc/service.proto

package object

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ObjectService_Get_FullMethodName          = "/neo.fs.v2.object.ObjectService/Get"
	ObjectService_Put_FullMethodName          = "/neo.fs.v2.object.ObjectService/Put"
	ObjectService_Delete_FullMethodName       = "/neo.fs.v2.object.ObjectService/Delete"
	ObjectService_Head_FullMethodName         = "/neo.fs.v2.object.ObjectService/Head"
	ObjectService_Search_FullMethodName       = "/neo.fs.v2.object.ObjectService/Search"
	ObjectService_GetRange_FullMethodName     = "/neo.fs.v2.object.ObjectService/GetRange"
	ObjectService_GetRangeHash_FullMethodName = "/neo.fs.v2.object.ObjectService/GetRangeHash"
	ObjectService_Replicate_FullMethodName    = "/neo.fs.v2.object.ObjectService/Replicate"
)

// ObjectServiceClient is the client API for ObjectService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ObjectServiceClient interface {
	// Receive full object structure, including Headers and payload. Response uses
	// gRPC stream. First response message carries the object with the requested address.
	// Chunk messages are parts of the object's payload if it is needed. All
	// messages, except the first one, carry payload chunks. The requested object can
	// be restored by concatenation of object message payload and all chunks
	// keeping the receiving order.
	//
	// Extended headers can change `Get` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions (starting from `__NEOFS__NETMAP_EPOCH` if specified or
	//     the latest one otherwise) of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     read access to the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (ObjectService_GetClient, error)
	// Put the object into container. Request uses gRPC stream. First message
	// SHOULD be of PutHeader type. `ContainerID` and `OwnerID` of an object
	// SHOULD be set. Session token SHOULD be obtained before `PUT` operation (see
	// session package). Chunk messages are considered by server as a part of an
	// object payload. All messages, except first one, SHOULD be payload chunks.
	// Chunk messages SHOULD be sent in the direct order of fragmentation.
	//
	// Extended headers can change `Put` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully saved in the container;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     write access to the container is denied;
	//   - **LOCKED** (2050, SECTION_OBJECT): \
	//     placement of an object of type TOMBSTONE that includes at least one locked
	//     object is prohibited;
	//   - **LOCK_NON_REGULAR_OBJECT** (2051, SECTION_OBJECT): \
	//     placement of an object of type LOCK that includes at least one object of
	//     type other than REGULAR is prohibited;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object storage container not found;
	//   - **TOKEN_NOT_FOUND** (4096, SECTION_SESSION): \
	//     (for trusted object preparation) session private key does not exist or has
	//
	// been deleted;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Put(ctx context.Context, opts ...grpc.CallOption) (ObjectService_PutClient, error)
	// Delete the object from a container. There is no immediate removal
	// guarantee. Object will be marked for removal and deleted eventually.
	//
	// Extended headers can change `Delete` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully marked to be removed from the container;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     delete access to the object is denied;
	//   - **LOCKED** (2050, SECTION_OBJECT): \
	//     deleting a locked object is prohibited;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error)
	// Returns the object Headers without data payload. By default full header is
	// returned. If `main_only` request field is set, the short header with only
	// the very minimal information will be returned instead.
	//
	// Extended headers can change `Head` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object header has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation HEAD of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Head(ctx context.Context, in *HeadRequest, opts ...grpc.CallOption) (*HeadResponse, error)
	// Search objects in container. Search query allows to match by Object
	// Header's filed values. Please see the corresponding NeoFS Technical
	// Specification section for more details.
	//
	// Extended headers can change `Search` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     objects have been successfully selected;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation SEARCH of the object is denied;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     search container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (ObjectService_SearchClient, error)
	// Get byte range of data payload. Range is set as an (offset, length) tuple.
	// Like in `Get` method, the response uses gRPC stream. Requested range can be
	// restored by concatenation of all received payload chunks keeping the receiving
	// order.
	//
	// Extended headers can change `GetRange` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     data range of the object payload has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation RANGE of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted.
	//   - **OUT_OF_RANGE** (2053, SECTION_OBJECT): \
	//     the requested range is out of bounds;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	GetRange(ctx context.Context, in *GetRangeRequest, opts ...grpc.CallOption) (ObjectService_GetRangeClient, error)
	// Returns homomorphic or regular hash of object's payload range after
	// applying XOR operation with the provided `salt`. Ranges are set of (offset,
	// length) tuples. Hashes order in response corresponds to the ranges order in
	// the request. Note that hash is calculated for XORed data.
	//
	// Extended headers can change `GetRangeHash` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     data range of the object payload has been successfully hashed;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation RANGEHASH of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OUT_OF_RANGE** (2053, SECTION_OBJECT): \
	//     the requested range is out of bounds;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	GetRangeHash(ctx context.Context, in *GetRangeHashRequest, opts ...grpc.CallOption) (*GetRangeHashResponse, error)
	// Save replica of the object on the NeoFS storage node. Both client and
	// server must authenticate NeoFS storage nodes matching storage policy of
	// the container referenced by the replicated object. Thus, this operation is
	// purely system: regular users should not pay attention to it but use Put.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     the object has been successfully replicated;
	//   - **INTERNAL_SERVER_ERROR** (1024, SECTION_FAILURE_COMMON): \
	//     internal server error described in the text message;
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     the client does not authenticate any NeoFS storage node matching storage
	//     policy of the container referenced by the replicated object
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     the container to which the replicated object is associated was not found.
	Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error)
}

type objectServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewObjectServiceClient(cc grpc.ClientConnInterface) ObjectServiceClient {
	return &objectServiceClient{cc}
}

func (c *objectServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (ObjectService_GetClient, error) {
	stream, err := c.cc.NewStream(ctx, &ObjectService_ServiceDesc.Streams[0], ObjectService_Get_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &objectServiceGetClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ObjectService_GetClient interface {
	Recv() (*GetResponse, error)
	grpc.ClientStream
}

type objectServiceGetClient struct {
	grpc.ClientStream
}

func (x *objectServiceGetClient) Recv() (*GetResponse, error) {
	m := new(GetResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectServiceClient) Put(ctx context.Context, opts ...grpc.CallOption) (ObjectService_PutClient, error) {
	stream, err := c.cc.NewStream(ctx, &ObjectService_ServiceDesc.Streams[1], ObjectService_Put_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &objectServicePutClient{stream}
	return x, nil
}

type ObjectService_PutClient interface {
	Send(*PutRequest) error
	CloseAndRecv() (*PutResponse, error)
	grpc.ClientStream
}

type objectServicePutClient struct {
	grpc.ClientStream
}

func (x *objectServicePutClient) Send(m *PutRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *objectServicePutClient) CloseAndRecv() (*PutResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(PutResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectServiceClient) Delete(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*DeleteResponse, error) {
	out := new(DeleteResponse)
	err := c.cc.Invoke(ctx, ObjectService_Delete_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectServiceClient) Head(ctx context.Context, in *HeadRequest, opts ...grpc.CallOption) (*HeadResponse, error) {
	out := new(HeadResponse)
	err := c.cc.Invoke(ctx, ObjectService_Head_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectServiceClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (ObjectService_SearchClient, error) {
	stream, err := c.cc.NewStream(ctx, &ObjectService_ServiceDesc.Streams[2], ObjectService_Search_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &objectServiceSearchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ObjectService_SearchClient interface {
	Recv() (*SearchResponse, error)
	grpc.ClientStream
}

type objectServiceSearchClient struct {
	grpc.ClientStream
}

func (x *objectServiceSearchClient) Recv() (*SearchResponse, error) {
	m := new(SearchResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectServiceClient) GetRange(ctx context.Context, in *GetRangeRequest, opts ...grpc.CallOption) (ObjectService_GetRangeClient, error) {
	stream, err := c.cc.NewStream(ctx, &ObjectService_ServiceDesc.Streams[3], ObjectService_GetRange_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &objectServiceGetRangeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ObjectService_GetRangeClient interface {
	Recv() (*GetRangeResponse, error)
	grpc.ClientStream
}

type objectServiceGetRangeClient struct {
	grpc.ClientStream
}

func (x *objectServiceGetRangeClient) Recv() (*GetRangeResponse, error) {
	m := new(GetRangeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *objectServiceClient) GetRangeHash(ctx context.Context, in *GetRangeHashRequest, opts ...grpc.CallOption) (*GetRangeHashResponse, error) {
	out := new(GetRangeHashResponse)
	err := c.cc.Invoke(ctx, ObjectService_GetRangeHash_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *objectServiceClient) Replicate(ctx context.Context, in *ReplicateRequest, opts ...grpc.CallOption) (*ReplicateResponse, error) {
	out := new(ReplicateResponse)
	err := c.cc.Invoke(ctx, ObjectService_Replicate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ObjectServiceServer is the server API for ObjectService service.
// All implementations should embed UnimplementedObjectServiceServer
// for forward compatibility
type ObjectServiceServer interface {
	// Receive full object structure, including Headers and payload. Response uses
	// gRPC stream. First response message carries the object with the requested address.
	// Chunk messages are parts of the object's payload if it is needed. All
	// messages, except the first one, carry payload chunks. The requested object can
	// be restored by concatenation of object message payload and all chunks
	// keeping the receiving order.
	//
	// Extended headers can change `Get` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions (starting from `__NEOFS__NETMAP_EPOCH` if specified or
	//     the latest one otherwise) of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     read access to the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Get(*GetRequest, ObjectService_GetServer) error
	// Put the object into container. Request uses gRPC stream. First message
	// SHOULD be of PutHeader type. `ContainerID` and `OwnerID` of an object
	// SHOULD be set. Session token SHOULD be obtained before `PUT` operation (see
	// session package). Chunk messages are considered by server as a part of an
	// object payload. All messages, except first one, SHOULD be payload chunks.
	// Chunk messages SHOULD be sent in the direct order of fragmentation.
	//
	// Extended headers can change `Put` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully saved in the container;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     write access to the container is denied;
	//   - **LOCKED** (2050, SECTION_OBJECT): \
	//     placement of an object of type TOMBSTONE that includes at least one locked
	//     object is prohibited;
	//   - **LOCK_NON_REGULAR_OBJECT** (2051, SECTION_OBJECT): \
	//     placement of an object of type LOCK that includes at least one object of
	//     type other than REGULAR is prohibited;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object storage container not found;
	//   - **TOKEN_NOT_FOUND** (4096, SECTION_SESSION): \
	//     (for trusted object preparation) session private key does not exist or has
	//
	// been deleted;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Put(ObjectService_PutServer) error
	// Delete the object from a container. There is no immediate removal
	// guarantee. Object will be marked for removal and deleted eventually.
	//
	// Extended headers can change `Delete` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object has been successfully marked to be removed from the container;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     delete access to the object is denied;
	//   - **LOCKED** (2050, SECTION_OBJECT): \
	//     deleting a locked object is prohibited;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	// Returns the object Headers without data payload. By default full header is
	// returned. If `main_only` request field is set, the short header with only
	// the very minimal information will be returned instead.
	//
	// Extended headers can change `Head` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     object header has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation HEAD of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Head(context.Context, *HeadRequest) (*HeadResponse, error)
	// Search objects in container. Search query allows to match by Object
	// Header's filed values. Please see the corresponding NeoFS Technical
	// Specification section for more details.
	//
	// Extended headers can change `Search` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     objects have been successfully selected;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation SEARCH of the object is denied;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     search container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	Search(*SearchRequest, ObjectService_SearchServer) error
	// Get byte range of data payload. Range is set as an (offset, length) tuple.
	// Like in `Get` method, the response uses gRPC stream. Requested range can be
	// restored by concatenation of all received payload chunks keeping the receiving
	// order.
	//
	// Extended headers can change `GetRange` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     data range of the object payload has been successfully read;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation RANGE of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OBJECT_ALREADY_REMOVED** (2052, SECTION_OBJECT): \
	//     the requested object has been marked as deleted.
	//   - **OUT_OF_RANGE** (2053, SECTION_OBJECT): \
	//     the requested range is out of bounds;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	GetRange(*GetRangeRequest, ObjectService_GetRangeServer) error
	// Returns homomorphic or regular hash of object's payload range after
	// applying XOR operation with the provided `salt`. Ranges are set of (offset,
	// length) tuples. Hashes order in response corresponds to the ranges order in
	// the request. Note that hash is calculated for XORed data.
	//
	// Extended headers can change `GetRangeHash` behaviour:
	//   - __NEOFS__NETMAP_EPOCH \
	//     Will use the requsted version of Network Map for object placement
	//     calculation. DEPRECATED: header ignored by servers.
	//   - __NEOFS__NETMAP_LOOKUP_DEPTH \
	//     Will try older versions of Network Map to find an object until the depth
	//     limit is reached. DEPRECATED: header ignored by servers.
	//
	// Please refer to detailed `XHeader` description.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     data range of the object payload has been successfully hashed;
	//   - Common failures (SECTION_FAILURE_COMMON);
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     access to operation RANGEHASH of the object is denied;
	//   - **OBJECT_NOT_FOUND** (2049, SECTION_OBJECT): \
	//     object not found in container;
	//   - **OUT_OF_RANGE** (2053, SECTION_OBJECT): \
	//     the requested range is out of bounds;
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     object container not found;
	//   - **TOKEN_EXPIRED** (4097, SECTION_SESSION): \
	//     provided session token has expired.
	GetRangeHash(context.Context, *GetRangeHashRequest) (*GetRangeHashResponse, error)
	// Save replica of the object on the NeoFS storage node. Both client and
	// server must authenticate NeoFS storage nodes matching storage policy of
	// the container referenced by the replicated object. Thus, this operation is
	// purely system: regular users should not pay attention to it but use Put.
	//
	// Statuses:
	//   - **OK** (0, SECTION_SUCCESS): \
	//     the object has been successfully replicated;
	//   - **INTERNAL_SERVER_ERROR** (1024, SECTION_FAILURE_COMMON): \
	//     internal server error described in the text message;
	//   - **ACCESS_DENIED** (2048, SECTION_OBJECT): \
	//     the client does not authenticate any NeoFS storage node matching storage
	//     policy of the container referenced by the replicated object
	//   - **CONTAINER_NOT_FOUND** (3072, SECTION_CONTAINER): \
	//     the container to which the replicated object is associated was not found.
	Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error)
}

// UnimplementedObjectServiceServer should be embedded to have forward compatible implementations.
type UnimplementedObjectServiceServer struct {
}

func (UnimplementedObjectServiceServer) Get(*GetRequest, ObjectService_GetServer) error {
	return status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedObjectServiceServer) Put(ObjectService_PutServer) error {
	return status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (UnimplementedObjectServiceServer) Delete(context.Context, *DeleteRequest) (*DeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedObjectServiceServer) Head(context.Context, *HeadRequest) (*HeadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Head not implemented")
}
func (UnimplementedObjectServiceServer) Search(*SearchRequest, ObjectService_SearchServer) error {
	return status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedObjectServiceServer) GetRange(*GetRangeRequest, ObjectService_GetRangeServer) error {
	return status.Errorf(codes.Unimplemented, "method GetRange not implemented")
}
func (UnimplementedObjectServiceServer) GetRangeHash(context.Context, *GetRangeHashRequest) (*GetRangeHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRangeHash not implemented")
}
func (UnimplementedObjectServiceServer) Replicate(context.Context, *ReplicateRequest) (*ReplicateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Replicate not implemented")
}

// UnsafeObjectServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ObjectServiceServer will
// result in compilation errors.
type UnsafeObjectServiceServer interface {
	mustEmbedUnimplementedObjectServiceServer()
}

func RegisterObjectServiceServer(s grpc.ServiceRegistrar, srv ObjectServiceServer) {
	s.RegisterService(&ObjectService_ServiceDesc, srv)
}

func _ObjectService_Get_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectServiceServer).Get(m, &objectServiceGetServer{stream})
}

type ObjectService_GetServer interface {
	Send(*GetResponse) error
	grpc.ServerStream
}

type objectServiceGetServer struct {
	grpc.ServerStream
}

func (x *objectServiceGetServer) Send(m *GetResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _ObjectService_Put_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ObjectServiceServer).Put(&objectServicePutServer{stream})
}

type ObjectService_PutServer interface {
	SendAndClose(*PutResponse) error
	Recv() (*PutRequest, error)
	grpc.ServerStream
}

type objectServicePutServer struct {
	grpc.ServerStream
}

func (x *objectServicePutServer) SendAndClose(m *PutResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *objectServicePutServer) Recv() (*PutRequest, error) {
	m := new(PutRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ObjectService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ObjectService_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectServiceServer).Delete(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectService_Head_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectServiceServer).Head(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ObjectService_Head_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectServiceServer).Head(ctx, req.(*HeadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectService_Search_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SearchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectServiceServer).Search(m, &objectServiceSearchServer{stream})
}

type ObjectService_SearchServer interface {
	Send(*SearchResponse) error
	grpc.ServerStream
}

type objectServiceSearchServer struct {
	grpc.ServerStream
}

func (x *objectServiceSearchServer) Send(m *SearchResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _ObjectService_GetRange_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRangeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ObjectServiceServer).GetRange(m, &objectServiceGetRangeServer{stream})
}

type ObjectService_GetRangeServer interface {
	Send(*GetRangeResponse) error
	grpc.ServerStream
}

type objectServiceGetRangeServer struct {
	grpc.ServerStream
}

func (x *objectServiceGetRangeServer) Send(m *GetRangeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _ObjectService_GetRangeHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRangeHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectServiceServer).GetRangeHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ObjectService_GetRangeHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectServiceServer).GetRangeHash(ctx, req.(*GetRangeHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ObjectService_Replicate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReplicateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ObjectServiceServer).Replicate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ObjectService_Replicate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ObjectServiceServer).Replicate(ctx, req.(*ReplicateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ObjectService_ServiceDesc is the grpc.ServiceDesc for ObjectService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ObjectService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "neo.fs.v2.object.ObjectService",
	HandlerType: (*ObjectServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Delete",
			Handler:    _ObjectService_Delete_Handler,
		},
		{
			MethodName: "Head",
			Handler:    _ObjectService_Head_Handler,
		},
		{
			MethodName: "GetRangeHash",
			Handler:    _ObjectService_GetRangeHash_Handler,
		},
		{
			MethodName: "Replicate",
			Handler:    _ObjectService_Replicate_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Get",
			Handler:       _ObjectService_Get_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Put",
			Handler:       _ObjectService_Put_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "Search",
			Handler:       _ObjectService_Search_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "GetRange",
			Handler:       _ObjectService_GetRange_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "object/grpc/service.proto",
}
