package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	server "github.com/aluo/gomono/zeonology/app"
	"github.com/aluo/gomono/zeonology/config"
	"github.com/aluo/gomono/zeonology/helper/logger"
	"github.com/aluo/gomono/zeonology/helper/metric"
	"github.com/aluo/gomono/zeonology/helper/mongo"
	"github.com/aluo/gomono/zeonology/helper/open_tel"
)

// Initiate all the external dependencies in main()
// - zap.logger
// - postgresql Client
// - redis Client
// - mongo Client
// - Jaeger tracing
// - prometheus metrics
func main() {
	log.Println("Starting auth microservice......")

	//load config
	log.Println("Load configuration......")
	cfg, err := config.GetServiceConfig()
	if err != nil {
		log.Fatalf("Loading config failed: %v", err)
		return
	}

	//initiate zapLogger
	zapLogger := logger.NewLogger(cfg)
	defer logger.SyncLogger(zapLogger)
	log.Println("Success initiated zap logger")

	//init mongodb client
	mongoClient, err := mongo.NewMongoClient(cfg)
	if err != nil {
		log.Fatalf("MongoDB init failed: %v", err)
		return
	}
	defer mongo.Close(mongoClient)
	log.Println("MongoDB connected")

	// initiate tracing
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := open_tel.InitOpenTelemetryCollector(cfg)
	if err != nil {
		log.Fatalf("failed to initiate open telemetry %v", err)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider %v", err)
		}
	}()
	log.Println("OpenTelemetry tracing connected")

	//init metrics
	if err = metric.InitMetrics(cfg); err != nil {
		log.Fatalf("failed to initiate open telemetry metrics %v", err)
	}
	log.Println("OpenTelemetry metrics connected")

	//start grpc server
	log.Printf(
		"AppVersion: %s, LogLevel: %s, Mode: %s, Port:%v",
		cfg.Server.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Mode,
		cfg.Server.Port,
	)
	myServer := server.NewGrpcServer(zapLogger, cfg, mongoClient)
	log.Fatal(myServer.Run())
}
