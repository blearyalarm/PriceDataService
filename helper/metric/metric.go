package metric

import (
	"log"
	"net/http"

	prometheus2 "github.com/prometheus/client_golang/prometheus"

	"github.com/aluo/gomono/edgecom/config"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

const MetricsPrefix = "user_mgmt_"

func InitMetrics(cfg *config.Config) error {
	exporter, err := initializeMetricsExporter()
	if err != nil {
		return err
	}

	if err = runtimemetrics.Start(); err != nil {
		return err
	}

	http.HandleFunc("/metrics", exporter.ServeHTTP)

	go func() {
		log.Fatal(http.ListenAndServe(cfg.Metrics.URL, nil))
	}()

	log.Printf(
		"Metrics available URL: %s, ServiceName: %s",
		cfg.Metrics.URL,
		cfg.Metrics.ServiceName,
	)

	return nil
}

func initializeMetricsExporter() (*prometheus.Exporter, error) {
	config := prometheus.Config{
		Registry:   prometheus2.NewRegistry(),
		Registerer: prometheus2.DefaultRegisterer,
		Gatherer:   prometheus2.DefaultGatherer,
	}

	ctrl := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)

	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		return nil, err
	}

	global.SetMeterProvider(exporter.MeterProvider())

	return exporter, err
}
