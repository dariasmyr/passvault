package mocks_test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"passvault/internal/http-server/handlers/entry/get"
)

type MockEntryDeleter struct {
	mock.Mock
}

func (m *MockEntryDeleter) DeleteEntry(ctx context.Context, accountId int64, entryID int64) (*get.Entry, error) {
	args := m.Called(ctx, accountId, entryID)
	return args.Get(0).(*get.Entry), args.Error(1)
}

type mockConstructorTestingTEntryDeleter interface {
	mock.TestingT
	Cleanup(func())
}

func NewEntryDeleter(t mockConstructorTestingTEntryDeleter) *MockEntryDeleter {
	mock := &MockEntryDeleter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
