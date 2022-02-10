// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/wetware/ww/pkg/cap/anchor (interfaces: Cluster)

// Package mock_anchor is a generated GoMock package.
package mock_anchor

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	cluster "github.com/wetware/casm/pkg/cluster"
)

// MockCluster is a mock of Cluster interface.
type MockCluster struct {
	ctrl     *gomock.Controller
	recorder *MockClusterMockRecorder
}

// MockClusterMockRecorder is the mock recorder for MockCluster.
type MockClusterMockRecorder struct {
	mock *MockCluster
}

// NewMockCluster creates a new mock instance.
func NewMockCluster(ctrl *gomock.Controller) *MockCluster {
	mock := &MockCluster{ctrl: ctrl}
	mock.recorder = &MockClusterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCluster) EXPECT() *MockClusterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockCluster) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockClusterMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockCluster)(nil).Close))
}

// String mocks base method.
func (m *MockCluster) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockClusterMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockCluster)(nil).String))
}

// View mocks base method.
func (m *MockCluster) View() cluster.View {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "View")
	ret0, _ := ret[0].(cluster.View)
	return ret0
}

// View indicates an expected call of View.
func (mr *MockClusterMockRecorder) View() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "View", reflect.TypeOf((*MockCluster)(nil).View))
}