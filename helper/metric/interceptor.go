package metric

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// UnaryServerMetricsInterceptor records metrics for unary RPC calls
func UnaryServerMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Record metrics based on method
		metrics := GetBusinessMetrics()
		if metrics != nil {
			duration := time.Since(startTime).Seconds()

			// You can add method-specific logic here
			switch info.FullMethod {
			case "/pricedata.price_data.v1.PriceDataService/LoadData":
				if err != nil {
					metrics.RecordLoadError(ctx)
				} else {
					metrics.RecordLoadSuccess(ctx, duration)
				}
			case "/pricedata.price_data.v1.PriceDataService/FindData":
				if err != nil {
					metrics.RecordQueryError(ctx)
				} else {
					// For queries, we'd ideally extract record count from response
					// For now, just record the query with duration
					metrics.RecordQuerySuccess(ctx, duration, 0)
				}
			}
		}

		return resp, err
	}
}

// StreamServerMetricsInterceptor records metrics for streaming RPC calls
func StreamServerMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		startTime := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Record metrics
		metrics := GetBusinessMetrics()
		if metrics != nil {
			duration := time.Since(startTime).Seconds()

			switch info.FullMethod {
			case "/pricedata.price_data.v1.PriceDataService/StreamData":
				if err != nil {
					metrics.RecordQueryError(ss.Context())
				} else {
					metrics.RecordQuerySuccess(ss.Context(), duration, 0)
				}
			}
		}

		return err
	}
}
