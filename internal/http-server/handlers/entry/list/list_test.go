package list_test

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"passvault/internal/http-server/handlers/entry/get"
	"passvault/internal/http-server/handlers/entry/list"
	mocks "passvault/internal/http-server/handlers/entry/list/mocks"
	authrest "passvault/internal/http-server/middlewares/auth"
	"passvault/internal/lib/jwt"
	"testing"
	"time"
)

func TestListHandler(t *testing.T) {
	cases := []struct {
		name        string
		mockEntries []get.Entry
		mockError   error
		respStatus  int
	}{
		{
			name: "Success",
			mockEntries: []get.Entry{
				{ID: 1, EntryType: "password", EntryData: "secret1"},
				{ID: 2, EntryType: "note", EntryData: "note content"},
			},
			mockError:  nil,
			respStatus: http.StatusOK,
		},
		{
			name:        "Empty List",
			mockEntries: []get.Entry{},
			mockError:   nil,
			respStatus:  http.StatusOK,
		},
		{
			name:        "Error while retrieving entries",
			mockEntries: nil,
			mockError:   fmt.Errorf("failed to retrieve entries"),
			respStatus:  http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockEntryLister := mocks.NewEntryLister(t)

			mockEntryLister.On("ListEntries", mock.AnythingOfType("*context.timerCtx"), int64(123)).Return(tc.mockEntries, tc.mockError)

			router := chi.NewRouter()
			handler := list.New(slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			), mockEntryLister, 5*time.Second)
			router.Get("/", handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			Middleware(router, rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.respStatus, resp.StatusCode)

			if tc.respStatus == http.StatusOK {
				var responseEntries []get.Entry
				err := json.NewDecoder(resp.Body).Decode(&responseEntries)
				require.NoError(t, err)
				assert.Equal(t, tc.mockEntries, responseEntries)
			}

			mockEntryLister.AssertExpectations(t)
		})
	}
}

func Middleware(handler http.Handler, w http.ResponseWriter, r *http.Request) {
	secret := "test_secret"
	mockToken := jwt.CreateMockToken(secret)

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", mockToken))

	authMiddleware := authrest.New(slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	), secret)

	handler = authMiddleware(handler)

	handler.ServeHTTP(w, r)
}
