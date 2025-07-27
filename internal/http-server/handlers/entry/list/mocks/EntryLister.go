package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
	"passvault/internal/http-server/handlers/entry/get"
)

type MockEntryLister struct {
	mock.Mock
}

func (m *MockEntryLister) ListEntries(ctx context.Context, accountId int64) ([]get.Entry, error) {
	args := m.Called(ctx, accountId)
	return args.Get(0).([]get.Entry), args.Error(1)
}

type mockConstructorTestingTEntryLister interface {
	mock.TestingT
	Cleanup(func())
}

func NewEntryLister(t mockConstructorTestingTEntryLister) *MockEntryLister {
	mock := &MockEntryLister{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
