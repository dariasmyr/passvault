package grpc

import (
	"context"
	"fmt"
	ssov1 "github.com/dariasmyr/protos/gen/go/sso"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"passvault/config"
	"time"
)

type Client struct {
	Cfg           *config.Config
	AuthClient    ssov1.AuthClient
	SessionClient ssov1.SessionsClient
	log           *slog.Logger
}

func New(
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriesCount int,
) (context.Context, *Client, error) {
	const op = "clients.sso.grpc.New"

	cfg := config.MustLoad()

	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.Timeout)
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

	cc, err := grpc.NewClient(
		addr,
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
		Cfg:           cfg,
		AuthClient:    authClient,
		SessionClient: sessionClient,
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
