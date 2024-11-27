package get_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"passvault/internal/http-server/handlers/entry/get"
	mocks "passvault/internal/http-server/handlers/entry/get/mocks"
	authrest "passvault/internal/http-server/middlewares/auth"
	"passvault/internal/lib/jwt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name       string
		entryID    string
		mockEntry  get.Entry
		mockError  error
		respStatus int
	}{
		{
			name:    "Success",
			entryID: "1",
			mockEntry: get.Entry{
				ID:        1,
				EntryType: "password",
				EntryData: "supersecretpassword",
			},
			mockError:  nil,
			respStatus: http.StatusOK,
		},
		{
			name:       "Invalid Entry ID",
			entryID:    "invalid",
			mockEntry:  get.Entry{},
			mockError:  nil,
			respStatus: http.StatusBadRequest,
		},
		{
			name:       "Error while retrieving entry",
			entryID:    "1",
			mockEntry:  get.Entry{},
			mockError:  fmt.Errorf("failed to retrieve entry"),
			respStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockEntryGetter := mocks.NewEntryGetter(t)
			mockEntryGetter.On("GetEntry", int64(123), int64(1)).Return(tc.mockEntry, tc.mockError)

			handler := get.New(slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			), mockEntryGetter)

			req, err := http.NewRequest(http.MethodGet, "/"+tc.entryID, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			Middleware(handler, rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.respStatus, resp.StatusCode)

			if tc.respStatus == http.StatusOK {
				var responseEntry get.Entry
				err := json.NewDecoder(resp.Body).Decode(&responseEntry)
				require.NoError(t, err)
				assert.Equal(t, tc.mockEntry, responseEntry)
			}

			mockEntryGetter.AssertExpectations(t)
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
