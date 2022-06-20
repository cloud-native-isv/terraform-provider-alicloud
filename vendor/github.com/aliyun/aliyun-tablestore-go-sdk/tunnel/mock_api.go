// Code generated by MockGen. DO NOT EDIT.
// Source: api.go

// Package tunnel is a generated GoMock package.
package tunnel

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockTunnelMetaApi is a mock of TunnelMetaApi interface
type MockTunnelMetaApi struct {
	ctrl     *gomock.Controller
	recorder *MockTunnelMetaApiMockRecorder
}

// MockTunnelMetaApiMockRecorder is the mock recorder for MockTunnelMetaApi
type MockTunnelMetaApiMockRecorder struct {
	mock *MockTunnelMetaApi
}

// NewMockTunnelMetaApi creates a new mock instance
func NewMockTunnelMetaApi(ctrl *gomock.Controller) *MockTunnelMetaApi {
	mock := &MockTunnelMetaApi{ctrl: ctrl}
	mock.recorder = &MockTunnelMetaApiMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTunnelMetaApi) EXPECT() *MockTunnelMetaApiMockRecorder {
	return m.recorder
}

// CreateTunnel mocks base method
func (m *MockTunnelMetaApi) CreateTunnel(req *CreateTunnelRequest) (*CreateTunnelResponse, error) {
	ret := m.ctrl.Call(m, "CreateTunnel", req)
	ret0, _ := ret[0].(*CreateTunnelResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTunnel indicates an expected call of CreateTunnel
func (mr *MockTunnelMetaApiMockRecorder) CreateTunnel(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTunnel", reflect.TypeOf((*MockTunnelMetaApi)(nil).CreateTunnel), req)
}

// ListTunnel mocks base method
func (m *MockTunnelMetaApi) ListTunnel(req *ListTunnelRequest) (*ListTunnelResponse, error) {
	ret := m.ctrl.Call(m, "ListTunnel", req)
	ret0, _ := ret[0].(*ListTunnelResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListTunnel indicates an expected call of ListTunnel
func (mr *MockTunnelMetaApiMockRecorder) ListTunnel(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTunnel", reflect.TypeOf((*MockTunnelMetaApi)(nil).ListTunnel), req)
}

// DescribeTunnel mocks base method
func (m *MockTunnelMetaApi) DescribeTunnel(req *DescribeTunnelRequest) (*DescribeTunnelResponse, error) {
	ret := m.ctrl.Call(m, "DescribeTunnel", req)
	ret0, _ := ret[0].(*DescribeTunnelResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeTunnel indicates an expected call of DescribeTunnel
func (mr *MockTunnelMetaApiMockRecorder) DescribeTunnel(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeTunnel", reflect.TypeOf((*MockTunnelMetaApi)(nil).DescribeTunnel), req)
}

// DeleteTunnel mocks base method
func (m *MockTunnelMetaApi) DeleteTunnel(req *DeleteTunnelRequest) (*DeleteTunnelResponse, error) {
	ret := m.ctrl.Call(m, "DeleteTunnel", req)
	ret0, _ := ret[0].(*DeleteTunnelResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteTunnel indicates an expected call of DeleteTunnel
func (mr *MockTunnelMetaApiMockRecorder) DeleteTunnel(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTunnel", reflect.TypeOf((*MockTunnelMetaApi)(nil).DeleteTunnel), req)
}

// GetRpo mocks base method
func (m *MockTunnelMetaApi) GetRpo(req *GetRpoRequest) (*GetRpoResponse, error) {
	ret := m.ctrl.Call(m, "GetRpo", req)
	ret0, _ := ret[0].(*GetRpoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRpo indicates an expected call of GetRpo
func (mr *MockTunnelMetaApiMockRecorder) GetRpo(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRpo", reflect.TypeOf((*MockTunnelMetaApi)(nil).GetRpo), req)
}

// GetRpoByOffset mocks base method
func (m *MockTunnelMetaApi) GetRpoByOffset(req *GetRpoRequest) (*GetRpoResponse, error) {
	ret := m.ctrl.Call(m, "GetRpoByOffset", req)
	ret0, _ := ret[0].(*GetRpoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRpoByOffset indicates an expected call of GetRpoByOffset
func (mr *MockTunnelMetaApiMockRecorder) GetRpoByOffset(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRpoByOffset", reflect.TypeOf((*MockTunnelMetaApi)(nil).GetRpoByOffset), req)
}

// Schedule mocks base method
func (m *MockTunnelMetaApi) Schedule(req *ScheduleRequest) (*ScheduleResponse, error) {
	ret := m.ctrl.Call(m, "Schedule", req)
	ret0, _ := ret[0].(*ScheduleResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Schedule indicates an expected call of Schedule
func (mr *MockTunnelMetaApiMockRecorder) Schedule(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Schedule", reflect.TypeOf((*MockTunnelMetaApi)(nil).Schedule), req)
}

// NewTunnelWorker mocks base method
func (m *MockTunnelMetaApi) NewTunnelWorker(tunnelId string, workerConfig *TunnelWorkerConfig) (TunnelWorker, error) {
	ret := m.ctrl.Call(m, "NewTunnelWorker", tunnelId, workerConfig)
	ret0, _ := ret[0].(TunnelWorker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NewTunnelWorker indicates an expected call of NewTunnelWorker
func (mr *MockTunnelMetaApiMockRecorder) NewTunnelWorker(tunnelId, workerConfig interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewTunnelWorker", reflect.TypeOf((*MockTunnelMetaApi)(nil).NewTunnelWorker), tunnelId, workerConfig)
}

// MockTunnelWorker is a mock of TunnelWorker interface
type MockTunnelWorker struct {
	ctrl     *gomock.Controller
	recorder *MockTunnelWorkerMockRecorder
}

// MockTunnelWorkerMockRecorder is the mock recorder for MockTunnelWorker
type MockTunnelWorkerMockRecorder struct {
	mock *MockTunnelWorker
}

// NewMockTunnelWorker creates a new mock instance
func NewMockTunnelWorker(ctrl *gomock.Controller) *MockTunnelWorker {
	mock := &MockTunnelWorker{ctrl: ctrl}
	mock.recorder = &MockTunnelWorkerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTunnelWorker) EXPECT() *MockTunnelWorkerMockRecorder {
	return m.recorder
}

// ConnectAndWorking mocks base method
func (m *MockTunnelWorker) ConnectAndWorking() error {
	ret := m.ctrl.Call(m, "ConnectAndWorking")
	ret0, _ := ret[0].(error)
	return ret0
}

// ConnectAndWorking indicates an expected call of ConnectAndWorking
func (mr *MockTunnelWorkerMockRecorder) ConnectAndWorking() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectAndWorking", reflect.TypeOf((*MockTunnelWorker)(nil).ConnectAndWorking))
}

// Shutdown mocks base method
func (m *MockTunnelWorker) Shutdown() {
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown
func (mr *MockTunnelWorkerMockRecorder) Shutdown() *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockTunnelWorker)(nil).Shutdown))
}

// MocktunnelDataApi is a mock of tunnelDataApi interface
type MocktunnelDataApi struct {
	ctrl     *gomock.Controller
	recorder *MocktunnelDataApiMockRecorder
}

// MocktunnelDataApiMockRecorder is the mock recorder for MocktunnelDataApi
type MocktunnelDataApiMockRecorder struct {
	mock *MocktunnelDataApi
}

// NewMocktunnelDataApi creates a new mock instance
func NewMocktunnelDataApi(ctrl *gomock.Controller) *MocktunnelDataApi {
	mock := &MocktunnelDataApi{ctrl: ctrl}
	mock.recorder = &MocktunnelDataApiMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocktunnelDataApi) EXPECT() *MocktunnelDataApiMockRecorder {
	return m.recorder
}

// ReadRecords mocks base method
func (m *MocktunnelDataApi) ReadRecords(req *ReadRecordRequest) (*ReadRecordResponse, error) {
	ret := m.ctrl.Call(m, "ReadRecords", req)
	ret0, _ := ret[0].(*ReadRecordResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadRecords indicates an expected call of ReadRecords
func (mr *MocktunnelDataApiMockRecorder) ReadRecords(req interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadRecords", reflect.TypeOf((*MocktunnelDataApi)(nil).ReadRecords), req)
}