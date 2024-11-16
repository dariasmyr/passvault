package list

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"passvault/internal/http-server/handlers/entry/get"
	authrest "passvault/internal/http-server/middlewares/auth"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
)

type EntryLister interface {
	ListEntries(accountId int64) ([]get.Entry, error)
}

func New(log *slog.Logger, entryLister EntryLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.list.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		claims, ok := r.Context().Value(authrest.UserClaimsKey).(*authrest.UserClaims)
		if !ok || claims == nil {
			log.Error("unauthorized access: user claims not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		entries, err := entryLister.ListEntries(claims.AccountID)
		if err != nil {
			log.Error("failed to retrieve entries", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to retrieve entries"))
			return
		}

		log.Info("entries retrieved", slog.Int("count", len(entries)))
		render.JSON(w, r, entries)
	}
}
