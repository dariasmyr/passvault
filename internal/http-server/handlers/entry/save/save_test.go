package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"passvault/internal/http-server/handlers/entry/save"
	mocks "passvault/internal/http-server/handlers/entry/save/mocks"
	authrest "passvault/internal/http-server/middlewares/auth"
	"passvault/internal/lib/jwt"
	"testing"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name      string
		entryType string
		entryData string
		respError string
		mockError error
	}{
		{
			name:      "Success",
			entryType: "password",
			entryData: "supersecretpassword",
		},
		{
			name:      "Empty Data",
			entryType: "password",
			entryData: "",
			respError: "field EntryData is a required field",
		},
		{
			name:      "SaveEntry Error",
			entryType: "password",
			entryData: "supersecretpassword",
			respError: "failed to save entry",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewEntrySaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveEntry", int64(123), tc.entryType, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := save.New(slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			), urlSaverMock)

			input := fmt.Sprintf(`{"entry_type": "%s", "entry_data": "%s"}`, tc.entryType, tc.entryData)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			//handler.ServeHTTP(rr, req) // Call the handler wrapped in the middleware instead
			Middleware(handler, rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
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
