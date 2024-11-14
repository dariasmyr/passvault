package delete

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

type EntryDeleter interface {
	DeleteEntry(accountId int64, entryID int64) error
}

func New(log *slog.Logger, entryDeleter EntryDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.entry.delete.New"

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

		if err := entryDeleter.DeleteEntry(claims.AccountID, id); err != nil {
			log.Error("failed to delete entry", slog.Int64("entryID", id), sl.Err(err))
			render.JSON(w, r, resp.Error("failed to delete entry"))
			return
		}

		log.Info("entry deleted", slog.Int64("entryID", id))
		render.JSON(w, r, resp.Response{Status: resp.StatusOK})
	}
}
