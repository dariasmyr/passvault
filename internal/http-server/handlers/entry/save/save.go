package save

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
	EntryType string `json:"entry_type" validate:"required"`
	EntryData string `json:"entry_data" validate:"required"`
}

type Response struct {
	resp.Response
	ID int64 `json:"id"`
}

type EntrySaver interface {
	SaveEntry(ctx context.Context, entryType, entryData string) (int64, error)
}

func New(log *slog.Logger, entrySaver EntrySaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		const op = "handlers.entry.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "empty request",
			})

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "failed to decode request",
			})

			return
		}

		log.Info("request body decoded", slog.Any("req", req))

	}
}
