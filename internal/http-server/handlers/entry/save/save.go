package save

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"io"
	"log/slog"
	"net/http"
	authrest "passvault/internal/http-server/middlewares/auth"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
	"time"
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
	SaveEntry(ctx context.Context, accountId int64, entryType, entryData string) (int64, error)
}

func New(log *slog.Logger, entrySaver EntrySaver, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.save.New"

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		if err := authrest.GetAuthErrorFromContext(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := authrest.GetUserClaimsFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Log the AccountID for tracking purposes
		log = log.With(slog.Int64("account_id", claims.AccountID))

		var req Request
		var decodeErr error
		decodeErr = render.DecodeJSON(r.Body, &req)
		if errors.Is(decodeErr, io.EOF) {
			log.Error("request body is empty")
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "empty request",
			})
			return
		}
		if decodeErr != nil {
			log.Error("failed to decode request body", sl.Err(decodeErr))
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "failed to decode request",
			})
			return
		}

		log.Info("request body decoded", slog.Any("req", req))

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			if errors.As(err, &validateErr) {
				log.Error("invalid request", sl.Err(err))
				render.JSON(w, r, resp.ValidationError(validateErr))
				return
			}
		}

		select {
		case <-ctx.Done():
			log.Error("request context cancelled", sl.Err(ctx.Err()))
			w.WriteHeader(http.StatusRequestTimeout)
			render.JSON(w, r, resp.Error("request timed out"))
			return
		default:
		}

		id, err := entrySaver.SaveEntry(ctx, claims.AccountID, req.EntryType, req.EntryData)
		if err != nil {
			log.Error("failed to save entry", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to save entry"))
			return
		}

		log.Info("entry saved", slog.Int64("id", id))
		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
