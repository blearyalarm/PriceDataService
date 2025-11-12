package open_tel

import (
	"context"
	"fmt"
	"time"

	"github.com/erich/pricetracking/config"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InitOpenTelemetryCollector initializes a telemetry collector using GRPC or stdout based on the environment
func InitOpenTelemetryCollector(cfg *config.Config) (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Jaeger.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporter, err := initializeExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	smpl := sdktrace.TraceIDRatioBased(cfg.Jaeger.SamplingRatio)
	if cfg.Server.Mode != "Production" {
		smpl = sdktrace.AlwaysSample()
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(smpl),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func initializeExporter(ctx context.Context, cfg *config.Config) (sdktrace.SpanExporter, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.Jaeger.Host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return traceExporter, nil
}
