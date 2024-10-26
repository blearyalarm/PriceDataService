package price

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/aluo/gomono/edgecom/config"
	"github.com/aluo/gomono/edgecom/gateway"
	"github.com/aluo/gomono/edgecom/model"
	priceData "github.com/aluo/gomono/edgecom/repository/pricedata"
)

type priceDataController struct {
	cfg            *config.Config
	priceRepo      priceData.PriceDataMongoRepo
	lastUpdateRepo priceData.LastUpdateMongoRepo
	assetGateway   gateway.AssetClient
	tracer         trace.Tracer
}

type PriceDataController interface {
	Load(ctx context.Context) error
	Find(ctx context.Context, query model.Query) ([]model.Entry, error)
}

func NewPriceDataController(cfg *config.Config,
	priceRepo priceData.PriceDataMongoRepo,
	lastUpdateRepo priceData.LastUpdateMongoRepo,
	assetGateway gateway.AssetClient,
) PriceDataController {
	return &priceDataController{
		cfg:            cfg,
		priceRepo:      priceRepo,
		lastUpdateRepo: lastUpdateRepo,
		assetGateway:   assetGateway,
		tracer:         otel.Tracer(cfg.GetTracerName()),
	}
}

// Create implements PhotoController.
func (p *priceDataController) Load(ctx context.Context) error {
	ctx, span := p.tracer.Start(ctx, "photoController.FindById")
	defer span.End()

	start, err := p.lastUpdateRepo.Get(ctx)
	if err != nil {
		return err
	}

	end := time.Now()
	if start.IsZero() {
		start = end.AddDate(-2, 0, 0) //go back 2 years as bootstrap data
	}

	//1. load from asset gateway
	assets, err := p.assetGateway.Load(ctx, start, end)
	if err != nil {
		return err
	}
	//2. persist into mongo
	if len(assets) > 0 {
		err = p.priceRepo.Create(ctx, assets)
		if err != nil {
			return err
		}
		//3. update lastUpdate
		return p.lastUpdateRepo.Update(ctx, end)
	}
	return nil
}

// FindByEventTypes implements PhotoController.
func (p *priceDataController) Find(ctx context.Context, query model.Query) ([]model.Entry, error) {
	ctx, span := p.tracer.Start(ctx, "photoController.FindById")
	defer span.End()
	return p.priceRepo.Find(ctx, query)
}
