package register

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"passvault/internal/clients/sso/grpc"
	resp "passvault/internal/lib/api/response"
	"passvault/internal/lib/logger/sl"
	"time"
)

type EntryRegister interface {
	RegisterClient(ctx context.Context, appName string, secret string, redirectUrl string) (int, error)
}

func New(log *slog.Logger, grpcClient *grpc.Client, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.client.register.New"

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		appName := chi.URLParam(r, "appName")
		secret := chi.URLParam(r, "secret")
		redirectUrl := chi.URLParam(r, "redirectUrl")

		select {
		case <-ctx.Done():
			log.Error("request context cancelled", sl.Err(ctx.Err()))
			w.WriteHeader(http.StatusRequestTimeout)
			render.JSON(w, r, resp.Error("request timed out"))
			return
		default:
		}

		entry, err := grpcClient.RegisterClient(ctx, appName, secret, redirectUrl)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				log.Error("request timeout", slog.String("appName", appName))
				w.WriteHeader(http.StatusGatewayTimeout)
				render.JSON(w, r, resp.Error("request timed out"))
				return
			}
			log.Error("failed to register client", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to retrieve entry"))
			return
		}

		log.Info("app registered", slog.String("appName", appName))
		render.JSON(w, r, entry)
	}
}
