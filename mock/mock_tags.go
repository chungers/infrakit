// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/codedellemc/infrakit.rackhd/monorail (interfaces: TagIface)

package mock

import (
	tags "github.com/codedellemc/gorackhd/client/tags"
	runtime "github.com/go-openapi/runtime"
	gomock "github.com/golang/mock/gomock"
)

// Mock of TagIface interface
type MockTagIface struct {
	ctrl     *gomock.Controller
	recorder *_MockTagIfaceRecorder
}

// Recorder for MockTagIface (not exported)
type _MockTagIfaceRecorder struct {
	mock *MockTagIface
}

func NewMockTagIface(ctrl *gomock.Controller) *MockTagIface {
	mock := &MockTagIface{ctrl: ctrl}
	mock.recorder = &_MockTagIfaceRecorder{mock}
	return mock
}

func (_m *MockTagIface) EXPECT() *_MockTagIfaceRecorder {
	return _m.recorder
}

func (_m *MockTagIface) DeleteNodesIdentifierTagsTagname(_param0 *tags.DeleteNodesIdentifierTagsTagnameParams, _param1 runtime.ClientAuthInfoWriter) (*tags.DeleteNodesIdentifierTagsTagnameOK, error) {
	ret := _m.ctrl.Call(_m, "DeleteNodesIdentifierTagsTagname", _param0, _param1)
	ret0, _ := ret[0].(*tags.DeleteNodesIdentifierTagsTagnameOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockTagIfaceRecorder) DeleteNodesIdentifierTagsTagname(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteNodesIdentifierTagsTagname", arg0, arg1)
}

func (_m *MockTagIface) GetNodesIdentifierTags(_param0 *tags.GetNodesIdentifierTagsParams, _param1 runtime.ClientAuthInfoWriter) (*tags.GetNodesIdentifierTagsOK, error) {
	ret := _m.ctrl.Call(_m, "GetNodesIdentifierTags", _param0, _param1)
	ret0, _ := ret[0].(*tags.GetNodesIdentifierTagsOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockTagIfaceRecorder) GetNodesIdentifierTags(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetNodesIdentifierTags", arg0, arg1)
}

func (_m *MockTagIface) PatchNodesIdentifierTags(_param0 *tags.PatchNodesIdentifierTagsParams, _param1 runtime.ClientAuthInfoWriter) (*tags.PatchNodesIdentifierTagsOK, error) {
	ret := _m.ctrl.Call(_m, "PatchNodesIdentifierTags", _param0, _param1)
	ret0, _ := ret[0].(*tags.PatchNodesIdentifierTagsOK)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockTagIfaceRecorder) PatchNodesIdentifierTags(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PatchNodesIdentifierTags", arg0, arg1)
}
