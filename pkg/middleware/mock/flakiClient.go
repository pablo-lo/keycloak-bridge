// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/keycloak-bridge/api/flaki/fb (interfaces: FlakiClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	fb "github.com/cloudtrust/keycloak-bridge/api/flaki/fb"
	gomock "github.com/golang/mock/gomock"
	go0 "github.com/google/flatbuffers/go"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// FlakiClient is a mock of FlakiClient interface
type FlakiClient struct {
	ctrl     *gomock.Controller
	recorder *FlakiClientMockRecorder
}

// FlakiClientMockRecorder is the mock recorder for FlakiClient
type FlakiClientMockRecorder struct {
	mock *FlakiClient
}

// NewFlakiClient creates a new mock instance
func NewFlakiClient(ctrl *gomock.Controller) *FlakiClient {
	mock := &FlakiClient{ctrl: ctrl}
	mock.recorder = &FlakiClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *FlakiClient) EXPECT() *FlakiClientMockRecorder {
	return m.recorder
}

// NextID mocks base method
func (m *FlakiClient) NextID(arg0 context.Context, arg1 *go0.Builder, arg2 ...grpc.CallOption) (*fb.FlakiReply, error) {
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NextID", varargs...)
	ret0, _ := ret[0].(*fb.FlakiReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextID indicates an expected call of NextID
func (mr *FlakiClientMockRecorder) NextID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextID", reflect.TypeOf((*FlakiClient)(nil).NextID), varargs...)
}

// NextValidID mocks base method
func (m *FlakiClient) NextValidID(arg0 context.Context, arg1 *go0.Builder, arg2 ...grpc.CallOption) (*fb.FlakiReply, error) {
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NextValidID", varargs...)
	ret0, _ := ret[0].(*fb.FlakiReply)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextValidID indicates an expected call of NextValidID
func (mr *FlakiClientMockRecorder) NextValidID(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextValidID", reflect.TypeOf((*FlakiClient)(nil).NextValidID), varargs...)
}
