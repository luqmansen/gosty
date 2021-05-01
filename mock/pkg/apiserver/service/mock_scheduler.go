// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/apiserver/services/scheduler.go

// Package mock_scheduler is a generated GoMock package.
package mock_scheduler

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/luqmansen/gosty/pkg/apiserver/models"
)

// MockScheduler is a mock of Scheduler interface.
type MockScheduler struct {
	ctrl     *gomock.Controller
	recorder *MockSchedulerMockRecorder
}

// MockSchedulerMockRecorder is the mock recorder for MockScheduler.
type MockSchedulerMockRecorder struct {
	mock *MockScheduler
}

// NewMockScheduler creates a new mock instance.
func NewMockScheduler(ctrl *gomock.Controller) *MockScheduler {
	mock := &MockScheduler{ctrl: ctrl}
	mock.recorder = &MockSchedulerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScheduler) EXPECT() *MockSchedulerMockRecorder {
	return m.recorder
}

// CreateDashTask mocks base method.
func (m *MockScheduler) CreateDashTask(task *models.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateDashTask", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateDashTask indicates an expected call of CreateDashTask.
func (mr *MockSchedulerMockRecorder) CreateDashTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateDashTask", reflect.TypeOf((*MockScheduler)(nil).CreateDashTask), task)
}

// CreateMergeTask mocks base method.
func (m *MockScheduler) CreateMergeTask(task *models.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMergeTask", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMergeTask indicates an expected call of CreateMergeTask.
func (mr *MockSchedulerMockRecorder) CreateMergeTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMergeTask", reflect.TypeOf((*MockScheduler)(nil).CreateMergeTask), task)
}

// CreateSplitTask mocks base method.
func (m *MockScheduler) CreateSplitTask(video *models.Video) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSplitTask", video)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateSplitTask indicates an expected call of CreateSplitTask.
func (mr *MockSchedulerMockRecorder) CreateSplitTask(video interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSplitTask", reflect.TypeOf((*MockScheduler)(nil).CreateSplitTask), video)
}

// CreateTranscodeTask mocks base method.
func (m *MockScheduler) CreateTranscodeTask(task *models.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTranscodeTask", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateTranscodeTask indicates an expected call of CreateTranscodeTask.
func (mr *MockSchedulerMockRecorder) CreateTranscodeTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTranscodeTask", reflect.TypeOf((*MockScheduler)(nil).CreateTranscodeTask), task)
}

// DeleteTask mocks base method.
func (m *MockScheduler) DeleteTask(taskId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTask", taskId)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTask indicates an expected call of DeleteTask.
func (mr *MockSchedulerMockRecorder) DeleteTask(taskId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTask", reflect.TypeOf((*MockScheduler)(nil).DeleteTask), taskId)
}

// GetAllTaskProgress mocks base method.
func (m *MockScheduler) GetAllTaskProgress() []*models.TaskProgressResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllTaskProgress")
	ret0, _ := ret[0].([]*models.TaskProgressResponse)
	return ret0
}

// GetAllTaskProgress indicates an expected call of GetAllTaskProgress.
func (mr *MockSchedulerMockRecorder) GetAllTaskProgress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllTaskProgress", reflect.TypeOf((*MockScheduler)(nil).GetAllTaskProgress))
}

// ReadMessages mocks base method.
func (m *MockScheduler) ReadMessages() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ReadMessages")
}

// ReadMessages indicates an expected call of ReadMessages.
func (mr *MockSchedulerMockRecorder) ReadMessages() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadMessages", reflect.TypeOf((*MockScheduler)(nil).ReadMessages))
}