// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/nitrictech/nitric/pkg/plugins/secret/key_vault (interfaces: KeyVaultClient)

// Package mock_key_vault is a generated GoMock package.
package mock_key_vault

import (
	context "context"
	reflect "reflect"

	keyvault "github.com/Azure/azure-sdk-for-go/services/keyvault/v7.1/keyvault"
	gomock "github.com/golang/mock/gomock"
)

// MockKeyVaultClient is a mock of KeyVaultClient interface.
type MockKeyVaultClient struct {
	ctrl     *gomock.Controller
	recorder *MockKeyVaultClientMockRecorder
}

// MockKeyVaultClientMockRecorder is the mock recorder for MockKeyVaultClient.
type MockKeyVaultClientMockRecorder struct {
	mock *MockKeyVaultClient
}

// NewMockKeyVaultClient creates a new mock instance.
func NewMockKeyVaultClient(ctrl *gomock.Controller) *MockKeyVaultClient {
	mock := &MockKeyVaultClient{ctrl: ctrl}
	mock.recorder = &MockKeyVaultClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyVaultClient) EXPECT() *MockKeyVaultClientMockRecorder {
	return m.recorder
}

// GetSecret mocks base method.
func (m *MockKeyVaultClient) GetSecret(arg0 context.Context, arg1, arg2, arg3 string) (keyvault.SecretBundle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSecret", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(keyvault.SecretBundle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSecret indicates an expected call of GetSecret.
func (mr *MockKeyVaultClientMockRecorder) GetSecret(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).GetSecret), arg0, arg1, arg2, arg3)
}

// SetSecret mocks base method.
func (m *MockKeyVaultClient) SetSecret(arg0 context.Context, arg1, arg2 string, arg3 keyvault.SecretSetParameters) (keyvault.SecretBundle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetSecret", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(keyvault.SecretBundle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetSecret indicates an expected call of SetSecret.
func (mr *MockKeyVaultClientMockRecorder) SetSecret(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSecret", reflect.TypeOf((*MockKeyVaultClient)(nil).SetSecret), arg0, arg1, arg2, arg3)
}
