package metric

import (
	"context"

	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

// BusinessMetrics holds all custom business metric instruments
type BusinessMetrics struct {
	// Counters
	PriceDataLoadsTotal   syncint64.Counter
	PriceDataQueriesTotal syncint64.Counter
	PriceDataLoadErrors   syncint64.Counter
	PriceDataQueryErrors  syncint64.Counter

	// Histograms
	PriceDataLoadDuration  syncfloat64.Histogram
	PriceDataQueryDuration syncfloat64.Histogram
	RecordsFetched         syncint64.Histogram

	// Gauges (using UpDownCounter)
	LastUpdateTimestamp syncint64.UpDownCounter
}

var businessMetrics *BusinessMetrics

// InitBusinessMetrics initializes all custom business metrics
func InitBusinessMetrics() error {
	meter := global.Meter(MetricsPrefix + "business")

	var err error
	bm := &BusinessMetrics{}

	// Counters
	bm.PriceDataLoadsTotal, err = meter.SyncInt64().Counter(
		MetricsPrefix+"price_data_loads_total",
		instrument.WithDescription("Total number of price data loads"),
	)
	if err != nil {
		return err
	}

	bm.PriceDataQueriesTotal, err = meter.SyncInt64().Counter(
		MetricsPrefix+"price_data_queries_total",
		instrument.WithDescription("Total number of price data queries"),
	)
	if err != nil {
		return err
	}

	bm.PriceDataLoadErrors, err = meter.SyncInt64().Counter(
		MetricsPrefix+"price_data_load_errors_total",
		instrument.WithDescription("Total number of price data load errors"),
	)
	if err != nil {
		return err
	}

	bm.PriceDataQueryErrors, err = meter.SyncInt64().Counter(
		MetricsPrefix+"price_data_query_errors_total",
		instrument.WithDescription("Total number of price data query errors"),
	)
	if err != nil {
		return err
	}

	// Histograms
	bm.PriceDataLoadDuration, err = meter.SyncFloat64().Histogram(
		MetricsPrefix+"price_data_load_duration_seconds",
		instrument.WithDescription("Duration of price data load operations"),
	)
	if err != nil {
		return err
	}

	bm.PriceDataQueryDuration, err = meter.SyncFloat64().Histogram(
		MetricsPrefix+"price_data_query_duration_seconds",
		instrument.WithDescription("Duration of price data query operations"),
	)
	if err != nil {
		return err
	}

	bm.RecordsFetched, err = meter.SyncInt64().Histogram(
		MetricsPrefix+"price_data_records_fetched",
		instrument.WithDescription("Number of records fetched per query"),
	)
	if err != nil {
		return err
	}

	// Gauge (using UpDownCounter)
	bm.LastUpdateTimestamp, err = meter.SyncInt64().UpDownCounter(
		MetricsPrefix+"price_data_last_update_timestamp",
		instrument.WithDescription("Timestamp of last successful price data update"),
	)
	if err != nil {
		return err
	}

	businessMetrics = bm
	return nil
}

// GetBusinessMetrics returns the initialized business metrics
func GetBusinessMetrics() *BusinessMetrics {
	return businessMetrics
}

// RecordLoadSuccess records a successful price data load
func (bm *BusinessMetrics) RecordLoadSuccess(ctx context.Context, duration float64) {
	bm.PriceDataLoadsTotal.Add(ctx, 1)
	bm.PriceDataLoadDuration.Record(ctx, duration)
}

// RecordLoadError records a failed price data load
func (bm *BusinessMetrics) RecordLoadError(ctx context.Context) {
	bm.PriceDataLoadErrors.Add(ctx, 1)
}

// RecordQuerySuccess records a successful price data query
func (bm *BusinessMetrics) RecordQuerySuccess(ctx context.Context, duration float64, recordCount int64) {
	bm.PriceDataQueriesTotal.Add(ctx, 1)
	bm.PriceDataQueryDuration.Record(ctx, duration)
	bm.RecordsFetched.Record(ctx, recordCount)
}

// RecordQueryError records a failed price data query
func (bm *BusinessMetrics) RecordQueryError(ctx context.Context) {
	bm.PriceDataQueryErrors.Add(ctx, 1)
}

// UpdateLastUpdateTimestamp updates the last update timestamp
func (bm *BusinessMetrics) UpdateLastUpdateTimestamp(ctx context.Context, timestamp int64) {
	bm.LastUpdateTimestamp.Add(ctx, timestamp)
}
