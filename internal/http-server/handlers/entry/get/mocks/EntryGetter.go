package mocks_test

import (
	"context"
	"github.com/stretchr/testify/mock"
	"passvault/internal/http-server/handlers/entry/get"
)

type MockEntryGetter struct {
	mock.Mock
}

func (m *MockEntryGetter) GetEntry(ctx context.Context, accountId int64, entryID int64) (*get.Entry, error) {
	args := m.Called(ctx, accountId, entryID)
	return args.Get(0).(*get.Entry), args.Error(1)
}

type mockConstructorTestingTEntryGetter interface {
	mock.TestingT
	Cleanup(func())
}

func NewEntryGetter(t mockConstructorTestingTEntryGetter) *MockEntryGetter {
	mock := &MockEntryGetter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
