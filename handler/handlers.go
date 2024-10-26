package handler

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	priceDataApi "github.com/aluo/api/edgecom/price_data/v1"
	"github.com/aluo/gomono/edgecom/config"
	"github.com/aluo/gomono/edgecom/controller/price"
)

type priceDataApiServer struct {
	cfg      *config.Config
	priceCtl price.PriceDataController
	tracer   trace.Tracer
}

// Auth controller constructor
func NewPriceDataApiServer(cfg *config.Config, priceCtl price.PriceDataController) priceDataApi.PriceDataServiceServer {
	return &priceDataApiServer{
		cfg:      cfg,
		priceCtl: priceCtl,
		tracer:   otel.Tracer(cfg.GetTracerName())}
}
