package save

import (
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
	SaveEntry(accountId int64, entryType, entryData string) (int64, error)
}

func New(log *slog.Logger, entrySaver EntrySaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Retrieve UserClaims from context
		claims, ok := r.Context().Value(authrest.UserClaimsKey).(*authrest.UserClaims)
		if !ok || claims == nil {
			log.Error("unauthorized access: user claims not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Log the AccountID for tracking purposes
		log = log.With(slog.Int64("account_id", claims.AccountID))

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

		if err := validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			if errors.As(err, &validateErr) {
				log.Error("invalid request", sl.Err(err))
				render.JSON(w, r, resp.ValidationError(validateErr))
				return
			}
		}

		// Use AccountID from claims if needed in saving process
		id, err := entrySaver.SaveEntry(claims.AccountID, req.EntryType, req.EntryData)
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
