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
	"passvault/internal/http-server/handlers/utils"
	"testing"
	"time"
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

			entrySaverMock := mocks.NewEntrySaver(t)

			if tc.respError == "" || tc.mockError != nil {
				entrySaverMock.On("SaveEntry", mock.AnythingOfType("*context.timerCtx"), int64(123), tc.entryType, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			handler := save.New(slog.New(
				slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			), entrySaverMock, 5*time.Second)

			input := fmt.Sprintf(`{"entry_type": "%s", "entry_data": "%s"}`, tc.entryType, tc.entryData)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			//handler.ServeHTTP(rr, req) // Call the handler wrapped in the middleware instead
			utils.TestMiddleware(handler, rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
