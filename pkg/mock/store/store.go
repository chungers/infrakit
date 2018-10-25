// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/docker/infrakit/pkg/store (interfaces: Snapshot)

package store // import "github.com/docker/infrakit/pkg/mock/store"

import (
	gomock "github.com/golang/mock/gomock"
)

// Mock of Snapshot interface
type MockSnapshot struct {
	ctrl     *gomock.Controller
	recorder *_MockSnapshotRecorder
}

// Recorder for MockSnapshot (not exported)
type _MockSnapshotRecorder struct {
	mock *MockSnapshot
}

func NewMockSnapshot(ctrl *gomock.Controller) *MockSnapshot {
	mock := &MockSnapshot{ctrl: ctrl}
	mock.recorder = &_MockSnapshotRecorder{mock}
	return mock
}

func (_m *MockSnapshot) EXPECT() *_MockSnapshotRecorder {
	return _m.recorder
}

func (_m *MockSnapshot) Close() error {
	ret := _m.ctrl.Call(_m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSnapshotRecorder) Close() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Close")
}

func (_m *MockSnapshot) Load(_param0 interface{}) error {
	ret := _m.ctrl.Call(_m, "Load", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSnapshotRecorder) Load(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Load", arg0)
}

func (_m *MockSnapshot) Save(_param0 interface{}) error {
	ret := _m.ctrl.Call(_m, "Save", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSnapshotRecorder) Save(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Save", arg0)
}
