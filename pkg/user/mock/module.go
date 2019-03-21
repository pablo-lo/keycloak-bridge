// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cloudtrust/keycloak-bridge/pkg/user (interfaces: Module,KeycloakClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	keycloak_client "github.com/cloudtrust/keycloak-client"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// Module is a mock of Module interface
type Module struct {
	ctrl     *gomock.Controller
	recorder *ModuleMockRecorder
}

// ModuleMockRecorder is the mock recorder for Module
type ModuleMockRecorder struct {
	mock *Module
}

// NewModule creates a new mock instance
func NewModule(ctrl *gomock.Controller) *Module {
	mock := &Module{ctrl: ctrl}
	mock.recorder = &ModuleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *Module) EXPECT() *ModuleMockRecorder {
	return m.recorder
}

// GetUsers mocks base method
func (m *Module) GetUsers(arg0 context.Context, arg1 string) ([]string, error) {
	ret := m.ctrl.Call(m, "GetUsers", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers
func (mr *ModuleMockRecorder) GetUsers(arg0, arg1 interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*Module)(nil).GetUsers), arg0, arg1)
}

// KeycloakClient is a mock of KeycloakClient interface
type KeycloakClient struct {
	ctrl     *gomock.Controller
	recorder *KeycloakClientMockRecorder
}

// KeycloakClientMockRecorder is the mock recorder for KeycloakClient
type KeycloakClientMockRecorder struct {
	mock *KeycloakClient
}

// NewKeycloakClient creates a new mock instance
func NewKeycloakClient(ctrl *gomock.Controller) *KeycloakClient {
	mock := &KeycloakClient{ctrl: ctrl}
	mock.recorder = &KeycloakClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *KeycloakClient) EXPECT() *KeycloakClientMockRecorder {
	return m.recorder
}

// GetUsers mocks base method
func (m *KeycloakClient) GetUsers(arg0, arg1 string, arg2 ...string) ([]keycloak_client.UserRepresentation, error) {
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetUsers", varargs...)
	ret0, _ := ret[0].([]keycloak_client.UserRepresentation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers
func (mr *KeycloakClientMockRecorder) GetUsers(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*KeycloakClient)(nil).GetUsers), varargs...)
}
