package grpc_env

import (
	"context"

	"github.com/erich/pricetracking/config"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const CtxServerEnv = ServerEnvKey("server.grpc_env")

type ServerEnvKey string

type ServerEnv struct {
	Logger *zap.Logger
	Config *config.Config
}

// UnaryServerInterceptor returns a new unary server interceptors that sets the values for request tags.
func UnaryServerInterceptor(env ServerEnv) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := context.WithValue(ctx, CtxServerEnv, env)
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that sets the values for request tags.
func StreamServerInterceptor(env ServerEnv) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := context.WithValue(stream.Context(), CtxServerEnv, env)
		wrapped := &grpc_middleware.WrappedServerStream{ServerStream: stream, WrappedContext: newCtx}
		err := handler(srv, wrapped)
		return err
	}
}
