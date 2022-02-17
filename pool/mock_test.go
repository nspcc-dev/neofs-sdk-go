// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/nspcc-dev/neofs-sdk-go/pool (interfaces: Client)

// Package pool is a generated GoMock package.
package pool

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client0 "github.com/nspcc-dev/neofs-sdk-go/client"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// CreateSession mocks base method.
func (m *MockClient) SessionCreate(arg0 context.Context, arg1 client0.PrmSessionCreate) (*client0.ResSessionCreate, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "SessionCreate", varargs...)
	ret0, _ := ret[0].(*client0.ResSessionCreate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockClientMockRecorder) CreateSession(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SessionCreate", reflect.TypeOf((*MockClient)(nil).SessionCreate), varargs...)
}

// DeleteContainer mocks base method.
func (m *MockClient) ContainerDelete(arg0 context.Context, arg1 client0.PrmContainerDelete) (*client0.ResContainerDelete, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerDelete", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerDelete)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteContainer indicates an expected call of DeleteContainer.
func (mr *MockClientMockRecorder) DeleteContainer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerDelete", reflect.TypeOf((*MockClient)(nil).ContainerDelete), varargs...)
}

// ObjectDelete mocks base method.
func (m *MockClient) ObjectDelete(arg0 context.Context, arg1 client0.PrmObjectDelete) (*client0.ResObjectDelete, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObjectDelete", arg0, arg1)
	ret0, _ := ret[0].(*client0.ResObjectDelete)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteObject indicates an expected call of DeleteObject.
func (mr *MockClientMockRecorder) DeleteObject(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteObject", reflect.TypeOf((*MockClient)(nil).ObjectDelete), varargs...)
}

// EACL mocks base method.
func (m *MockClient) ContainerEACL(arg0 context.Context, arg1 client0.PrmContainerEACL) (*client0.ResContainerEACL, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerEACL", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerEACL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EACL indicates an expected call of EACL.
func (mr *MockClientMockRecorder) EACL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerEACL", reflect.TypeOf((*MockClient)(nil).ContainerEACL), varargs...)
}

// EndpointInfo mocks base method.
func (m *MockClient) EndpointInfo(arg0 context.Context, arg1 client0.PrmEndpointInfo) (*client0.ResEndpointInfo, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "EndpointInfo", varargs...)
	ret0, _ := ret[0].(*client0.ResEndpointInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EndpointInfo indicates an expected call of EndpointInfo.
func (mr *MockClientMockRecorder) EndpointInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndpointInfo", reflect.TypeOf((*MockClient)(nil).EndpointInfo), varargs...)
}

// GetBalance mocks base method.
func (m *MockClient) BalanceGet(arg0 context.Context, arg1 client0.PrmBalanceGet) (*client0.ResBalanceGet, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "BalanceGet", varargs...)
	ret0, _ := ret[0].(*client0.ResBalanceGet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockClientMockRecorder) GetBalance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BalanceGet", reflect.TypeOf((*MockClient)(nil).BalanceGet), varargs...)
}

// GetContainer mocks base method.
func (m *MockClient) ContainerGet(arg0 context.Context, arg1 client0.PrmContainerGet) (*client0.ResContainerGet, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerGet", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerGet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetContainer indicates an expected call of GetContainer.
func (mr *MockClientMockRecorder) GetContainer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerGet", reflect.TypeOf((*MockClient)(nil).ContainerGet), varargs...)
}

// ObjectGetInitmocks base method.
func (m *MockClient) ObjectGetInit(arg0 context.Context, arg1 client0.PrmObjectGet) (*client0.ObjectReader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObject", arg0, arg1)
	ret0, _ := ret[0].(*client0.ObjectReader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetObject indicates an expected call of GetObject.
func (mr *MockClientMockRecorder) GetObject(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObjectGetInit", reflect.TypeOf((*MockClient)(nil).ObjectGetInit), varargs...)
}

// ObjectHead mocks base method.
func (m *MockClient) ObjectHead(arg0 context.Context, arg1 client0.PrmObjectHead) (*client0.ResObjectHead, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HeadObject", arg0, arg1)
	ret0, _ := ret[0].(*client0.ResObjectHead)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HeadObject indicates an expected call of HeadObject.
func (mr *MockClientMockRecorder) HeadObject(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HeadObject", reflect.TypeOf((*MockClient)(nil).ObjectHead), varargs...)
}

// ListContainers mocks base method.
func (m *MockClient) ContainerList(arg0 context.Context, arg1 client0.PrmContainerList) (*client0.ResContainerList, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerList", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListContainers indicates an expected call of ListContainers.
func (mr *MockClientMockRecorder) ListContainers(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerList", reflect.TypeOf((*MockClient)(nil).ContainerList), varargs...)
}

// NetworkInfo mocks base method.
func (m *MockClient) NetworkInfo(arg0 context.Context, arg1 client0.PrmNetworkInfo) (*client0.ResNetworkInfo, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "NetworkInfo", varargs...)
	ret0, _ := ret[0].(*client0.ResNetworkInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NetworkInfo indicates an expected call of NetworkInfo.
func (mr *MockClientMockRecorder) NetworkInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NetworkInfo", reflect.TypeOf((*MockClient)(nil).NetworkInfo), varargs...)
}

// ObjectRangeInit mocks base method.
func (m *MockClient) ObjectRangeInit(arg0 context.Context, arg1 client0.PrmObjectRange) (*client0.ObjectRangeReader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObjectRangeInit", arg0, arg1)
	ret0, _ := ret[0].(*client0.ObjectRangeReader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObjectRange indicates an expected call of ObjectRangeInit.
func (mr *MockClientMockRecorder) ObjectRange(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObjectRangeInit", reflect.TypeOf((*MockClient)(nil).ObjectRangeInit), varargs...)
}

// PutContainer mocks base method.
func (m *MockClient) ContainerPut(arg0 context.Context, arg1 client0.PrmContainerPut) (*client0.ResContainerPut, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerPut", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerPut)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PutContainer indicates an expected call of PutContainer.
func (mr *MockClientMockRecorder) PutContainer(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerPut", reflect.TypeOf((*MockClient)(nil).ContainerPut), varargs...)
}

// ObjectPutInitmocks base method.
func (m *MockClient) ObjectPutInit(arg0 context.Context, arg1 client0.PrmObjectPutInit) (*client0.ObjectWriter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutObject", arg0)
	ret0, _ := ret[0].(*client0.ObjectWriter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PutObject indicates an expected call of PutObject.
func (mr *MockClientMockRecorder) PutObject(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1})
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObjectPutInit", reflect.TypeOf((*MockClient)(nil).ObjectPutInit), varargs...)
}

// ObjectSearchInitmocks base method.
func (m *MockClient) ObjectSearchInit(arg0 context.Context, arg1 client0.PrmObjectSearch) (*client0.ObjectListReader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObjectSearchInit", arg0, arg1)
	ret0, _ := ret[0].(*client0.ObjectListReader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchObjects indicates an expected call of SearchObjects.
func (mr *MockClientMockRecorder) SearchObjects(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchObjects", reflect.TypeOf((*MockClient)(nil).ObjectSearchInit), varargs...)
}

// SetEACL mocks base method.
func (m *MockClient) ContainerSetEACL(arg0 context.Context, arg1 client0.PrmContainerSetEACL) (*client0.ResContainerSetEACL, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	ret := m.ctrl.Call(m, "ContainerSetEACL", varargs...)
	ret0, _ := ret[0].(*client0.ResContainerSetEACL)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetEACL indicates an expected call of SetEACL.
func (mr *MockClientMockRecorder) SetEACL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContainerSetEACL", reflect.TypeOf((*MockClient)(nil).ContainerSetEACL), varargs...)
}
