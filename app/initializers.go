package server

import (
	"github.com/aluo/gomono/edgecom/config"
	priceCtl "github.com/aluo/gomono/edgecom/controller/price"
	"github.com/aluo/gomono/edgecom/gateway"
	priceRepo "github.com/aluo/gomono/edgecom/repository/pricedata"
	"go.mongodb.org/mongo-driver/mongo"
)

type gateways struct {
	assetGateway gateway.AssetClient
}

func InitiateGateways(cfg *config.Config) (*gateways, error) {
	assetGateway, err := gateway.NewAssetClient(cfg)
	if err != nil {
		return nil, err
	}
	return &gateways{
		assetGateway: assetGateway,
	}, nil
}

type repos struct {
	PriceDataMongoRepo  priceRepo.PriceDataMongoRepo
	LastUpdateMongoRepo priceRepo.LastUpdateMongoRepo
}

func InitiateRepositories(mongoClient *mongo.Client, cfg *config.Config) (*repos, error) {
	priceDataMongoRepo, err := priceRepo.NewPriceDataMongoRepo(mongoClient)
	if err != nil {
		return nil, err
	}

	lastUpdateMongoRepo, err := priceRepo.NewLastRetrivalMongoRepo(mongoClient)
	if err != nil {
		return nil, err
	}

	return &repos{priceDataMongoRepo, lastUpdateMongoRepo}, nil
}

type controllers struct {
	priceConroller priceCtl.PriceDataController
}

func InitiateControllers(cfg *config.Config, rps *repos, gws *gateways) *controllers {
	priceConroller := priceCtl.NewPriceDataController(cfg,
		rps.PriceDataMongoRepo,
		rps.LastUpdateMongoRepo,
		gws.assetGateway,
	)

	return &controllers{priceConroller}
}
