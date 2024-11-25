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
