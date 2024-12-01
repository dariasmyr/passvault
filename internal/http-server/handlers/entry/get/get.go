package get

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	authrest "passvault/internal/http-server/middlewares/auth"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
	"strconv"
	"time"
)

type Entry struct {
	ID        int64  `json:"id"`
	AccountId int64  `json:"account_id"`
	EntryType string `json:"entry_type"`
	EntryData string `json:"entry_data"`
}

type EntryGetter interface {
	GetEntry(ctx context.Context, accountID int64, entryID int64) (*Entry, error)
}

func New(log *slog.Logger, entryGetter EntryGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.get.New"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

		entryID := chi.URLParam(r, "entryID")
		id, err := strconv.ParseInt(entryID, 10, 64)
		if err != nil {
			log.Error("invalid entryID parameter", slog.String("entryID", entryID))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "invalid entryID",
			})
			return
		}

		select {
		case <-ctx.Done():
			log.Error("request context cancelled", sl.Err(ctx.Err()))
			w.WriteHeader(http.StatusRequestTimeout)
			render.JSON(w, r, resp.Error("request timed out"))
			return
		default:
		}

		entry, err := entryGetter.GetEntry(ctx, claims.AccountID, id)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("request timeout", slog.Int64("entryID", id))
				w.WriteHeader(http.StatusGatewayTimeout)
				render.JSON(w, r, resp.Error("request timed out"))
				return
			}
			log.Error("failed to retrieve entry", slog.Int64("entryID", id), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to retrieve entry"))
			return
		}

		log.Info("entry retrieved", slog.Int64("entryID", entry.ID))
		render.JSON(w, r, entry)
	}
}
