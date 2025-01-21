package grpc

import (
	"context"
	"fmt"
	ssov1 "github.com/dariasmyr/protos/gen/go/sso"
	"github.com/go-chi/chi/v5/middleware"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"passvault/config"
	"passvault/internal/lib/logger/sl"
	"strconv"
	"time"
)

type Client struct {
	authClient    ssov1.AuthClient
	sessionClient ssov1.SessionsClient
	log           *slog.Logger
}

func New(
	log *slog.Logger,
	grpcHost string,
	grpcPort int,
	timeout time.Duration,
	retriesCount int,
) (context.Context, *Client, error) {
	const op = "clients.sso.grpc.New"

	cfg := config.MustLoad()

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	defer cancelCtx()

	// options for grpcretry interceptor
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retriesCount)),
		grpcretry.WithPerRetryTimeout(timeout),
	}

	// options for grpclog interceptor
	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	grpcAddress := net.JoinHostPort(grpcHost, strconv.Itoa(grpcPort))

	cc, err := grpc.NewClient(
		grpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		))

	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	// Create gRPC-client SSO/Auth and SSO/Session
	authClient := ssov1.NewAuthClient(cc)
	sessionClient := ssov1.NewSessionsClient(cc)

	client := &Client{
		authClient:    authClient,
		sessionClient: sessionClient,
		log:           log,
	}

	return ctx, client, nil
}

// InterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (c *Client) RegisterClient(ctx context.Context, appName string, secret string, redirectUrl string) (appId int64, err error) {
	const op = "grpc.RegisterClient"

	c.log = c.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(ctx)),
	)

	resp, err := c.authClient.RegisterClient(ctx, &ssov1.RegisterClientRequest{
		AppName:     appName,
		Secret:      secret,
		RedirectUrl: redirectUrl,
	})
	if err != nil {
		c.log.Error("failed to register client", sl.Err(err))
		return 0, err
	}

	return resp.AppId, nil
}
