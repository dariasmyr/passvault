package list

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"passvault/internal/http-server/handlers/entry/get"
	authrest "passvault/internal/http-server/middlewares/auth"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
	"time"
)

type EntryLister interface {
	ListEntries(ctx context.Context, accountId int64) ([]get.Entry, error)
}

func New(log *slog.Logger, entryLister EntryLister, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.list.New"

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

		select {
		case <-ctx.Done():
			log.Error("request context cancelled", sl.Err(ctx.Err()))
			w.WriteHeader(http.StatusRequestTimeout)
			render.JSON(w, r, resp.Error("request timed out"))
			return
		default:
		}

		entries, err := entryLister.ListEntries(ctx, claims.AccountID)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("request timeout", slog.Int64("accountID", claims.AccountID), sl.Err(err))
				w.WriteHeader(http.StatusGatewayTimeout)
				render.JSON(w, r, resp.Error("request timed out"))
				return
			}
			log.Error("failed to retrieve entries", slog.Int64("accountID", claims.AccountID), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to retrieve entries"))
			return
		}

		log.Info("entries retrieved", slog.Int("count", len(entries)))
		render.JSON(w, r, entries)
	}
}
