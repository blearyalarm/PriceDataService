package metric

import (
	"context"
	"log"
	"net/http"

	prometheus2 "github.com/prometheus/client_golang/prometheus"

	"github.com/erich/pricetracking/config"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const MetricsPrefix = "price_data_service_"

var ctrl *controller.Controller

func InitMetrics(cfg *config.Config) (func(context.Context) error, error) {
	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.Metrics.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Server.AppVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create Prometheus exporter with controller
	config := prometheus.Config{
		Registry:   prometheus2.NewRegistry(),
		Registerer: prometheus2.DefaultRegisterer,
		Gatherer:   prometheus2.DefaultGatherer,
	}

	ctrl = controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(res),
	)

	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		return nil, err
	}

	// Set global MeterProvider
	global.SetMeterProvider(exporter.MeterProvider())

	// Start runtime metrics
	if err = runtimemetrics.Start(); err != nil {
		return nil, err
	}

	// Initialize business metrics
	if err = InitBusinessMetrics(); err != nil {
		return nil, err
	}

	// Set up metrics HTTP endpoint
	http.HandleFunc("/metrics", exporter.ServeHTTP)

	// Start HTTP server in goroutine with proper error handling
	go func() {
		log.Printf(
			"Metrics available URL: %s, ServiceName: %s",
			cfg.Metrics.URL,
			cfg.Metrics.ServiceName,
		)
		if err := http.ListenAndServe(cfg.Metrics.URL, nil); err != nil {
			log.Printf("Metrics HTTP server error: %v", err)
		}
	}()

	// Return shutdown function
	return func(ctx context.Context) error {
		return ctrl.Stop(ctx)
	}, nil
}
