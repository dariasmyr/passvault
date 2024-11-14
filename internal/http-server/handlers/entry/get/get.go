package get

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	authrest "passvault/internal/http-server/middlewares/auth"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
	"strconv"
)

type Entry struct {
	ID        int64  `json:"id"`
	EntryType string `json:"entry_type"`
	EntryData string `json:"entry_data"`
}

type EntryGetter interface {
	GetEntry(accountId int64, entryID int64) (Entry, error)
}

func New(log *slog.Logger, entryGetter EntryGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.get.New"

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

		entryID := chi.URLParam(r, "entryID")
		id, err := strconv.ParseInt(entryID, 10, 64)
		if err != nil {
			log.Error("invalid entryID parameter", slog.String("entryID", entryID))
			render.JSON(w, r, resp.Response{
				Status: resp.StatusError,
				Error:  "invalid entryID",
			})
			return
		}

		entry, err := entryGetter.GetEntry(claims.AccountID, id)
		if err != nil {
			log.Error("failed to retrieve entry", slog.Int64("entryID", id), sl.Err(err))
			render.JSON(w, r, resp.Error("failed to retrieve entry"))
			return
		}

		log.Info("entry retrieved", slog.Int64("entryID", entry.ID))
		render.JSON(w, r, entry)
	}
}
