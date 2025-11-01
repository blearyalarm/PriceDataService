package server

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	priceDataApi "github.com/aluo/api/zeonology/price_data/v1"
	"github.com/aluo/gomono/zeonology/helper/grpc_env"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/aluo/gomono/zeonology/config"
	"github.com/aluo/gomono/zeonology/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Server is struct to hold server
type Server struct {
	logger      *zap.Logger
	cfg         *config.Config
	mongoClient *mongo.Client
}

// NewGrpcServer is to create Server constructor
func NewGrpcServer(logger *zap.Logger, cfg *config.Config, mongoClient *mongo.Client) *Server {
	return &Server{
		logger:      logger,
		cfg:         cfg,
		mongoClient: mongoClient,
	}
}

// Run server
func (s *Server) Run() error {

	//initiate gateway modules
	gws, err := InitiateGateways(s.cfg)
	if err != nil {
		return err
	}

	//initiate persistence modules
	rps, err := InitiateRepositories(s.mongoClient, s.cfg)
	if err != nil {
		return err
	}

	//initiate controller modules
	ctls := InitiateControllers(s.cfg, rps, gws)

	//initiate GRPC server
	serverEnv := grpc_env.ServerEnv{
		Logger: s.logger,
		Config: s.cfg,
	}

	server := s.initGrpcServer(serverEnv)

	//register server metrics
	grpcPrometheus.Register(server)

	//register health api
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())

	//register reflection api for non-production environment, so that GRPC clients can be used.
	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	//register User api
	authServer := handler.NewPriceDataApiServer(s.cfg, ctls.priceConroller)
	priceDataApi.RegisterPriceDataServiceServer(server, authServer)

	return s.startGrpcServer(server, func() {
		err := ctls.priceConroller.Load(context.TODO())
		if err != nil {
			log.Println("bootstrap data failed")
		}
	})
}

func (s *Server) initGrpcServer(serverEnv grpc_env.ServerEnv) *grpc.Server {
	opts := []grpcZap.Option{
		grpcZap.WithDecider(func(fullMethodName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && fullMethodName == "/grpc.health.v1.Health/Check" {
				return false
			}

			// by default everything will be logged
			return true
		}),
	}
	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Time * time.Minute,
	}),
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(
			grpc_env.StreamServerInterceptor(serverEnv),
			grpcCtxTags.StreamServerInterceptor(),
			otelgrpc.StreamServerInterceptor(),
			grpcPrometheus.StreamServerInterceptor,
			grpcZap.StreamServerInterceptor(s.logger, opts...),
			grpcAuth.StreamServerInterceptor(AuthInterceptorFunc),
			grpcRecovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			grpc_env.UnaryServerInterceptor(serverEnv),
			grpcCtxTags.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			grpcPrometheus.UnaryServerInterceptor,
			grpcZap.UnaryServerInterceptor(s.logger, opts...),
			grpcAuth.UnaryServerInterceptor(AuthInterceptorFunc),
			grpcRecovery.UnaryServerInterceptor(),
		)),
	)
	return server
}

func (s *Server) startGrpcServer(server *grpc.Server, postBootFunc func()) error {
	listener, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		log.Printf("zeonology GRPC Server is listening on port: %v", s.cfg.Server.Port)

		go func() {
			// Wait a brief moment to ensure the server is ready.
			time.Sleep(100 * time.Millisecond)
			postBootFunc()
		}()

		if err := server.Serve(listener); err != nil {
			log.Fatalf("Server failed to start. %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	server.GracefulStop()
	log.Println("Server Exited Properly")
	return nil
}

// AuthInterceptorFunc is used by grpc_auth.UnaryServerInterceptor for authorization.
// this method is a dummy method, it is overwritten by method in httpHandler func (u *authService) AuthFuncOverride()
func AuthInterceptorFunc(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
