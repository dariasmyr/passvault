package get_test

import (
	"errors"
	"github.com/gavv/httpexpect/v2"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"passvault/internal/http-server/handlers/entry/get"
	authrest "passvault/internal/http-server/middlewares/auth"
	"passvault/internal/lib/jwt"
	"testing"
)

type MockEntryGetter struct {
	mock.Mock
}

func (m *MockEntryGetter) GetEntry(accountId int64, entryID int64) (get.Entry, error) {
	args := m.Called(accountId, entryID)
	return args.Get(0).(get.Entry), args.Error(1)
}

func TestGetEntryHandler(t *testing.T) {
	logger := slog.Default()

	entryGetter := new(MockEntryGetter)
	mockEntry := get.Entry{
		ID:        1,
		EntryType: "password",
		EntryData: "secret_data",
	}
	entryGetter.On("GetEntry", int64(123), int64(1)).Return(mockEntry, nil)
	entryGetter.On("GetEntry", int64(123), int64(999)).Return(get.Entry{}, errors.New("entry not found"))

	secret := "test_secret"
	mockToken := jwt.CreateMockToken(secret)

	r := chi.NewRouter()
	r.Use(authrest.New(logger, secret))
	r.Get("/entries/{entryID}", get.New(logger, entryGetter))

	server := httptest.NewServer(r)
	defer server.Close()

	e := httpexpect.Default(t, server.URL)

	t.Run("Successful retrieval", func(t *testing.T) {
		e.GET("/entries/1").
			WithHeader("Authorization", "Bearer "+mockToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("id").
			HasValue("id", 1).
			HasValue("entry_type", "password").
			HasValue("entry_data", "secret_data")
	})

	t.Run("Invalid entryID", func(t *testing.T) {
		e.GET("/entries/invalid").
			WithHeader("Authorization", "Bearer "+mockToken).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error").
			HasValue("error", "invalid entryID")
	})

	t.Run("Entry not found", func(t *testing.T) {
		e.GET("/entries/999").
			WithHeader("Authorization", "Bearer "+mockToken).
			Expect().
			Status(http.StatusInternalServerError).
			JSON().Object().
			ContainsKey("error").
			HasValue("error", "failed to retrieve entry")
	})

	t.Run("Unauthorized access", func(t *testing.T) {
		e.GET("/entries/1").
			Expect().
			Status(http.StatusUnauthorized).
			Body().
			Contains("Unauthorized")
	})
}
