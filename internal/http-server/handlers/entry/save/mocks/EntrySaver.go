package mocks_test

import (
	"github.com/stretchr/testify/mock"
)

type MockEntrySaver struct {
	mock.Mock
}

func (m *MockEntrySaver) SaveEntry(accountId int64, entryType, entryData string) (int64, error) {
	args := m.Called(accountId, entryType, entryData)
	return args.Get(0).(int64), args.Error(1)
}

type mockConstructorTestingTEntrySaver interface {
	mock.TestingT
	Cleanup(func())
}

func NewEntrySaver(t mockConstructorTestingTEntrySaver) *MockEntrySaver {
	mock := &MockEntrySaver{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
