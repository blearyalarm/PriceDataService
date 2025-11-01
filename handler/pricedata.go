package handler

import (
	"context"

	priceDataApi "github.com/aluo/api/zeonology/price_data/v1"
	"github.com/aluo/gomono/zeonology/mapper"
)

// Create implements v1.PhotographerServiceServer.
func (u *priceDataApiServer) LoadData(ctx context.Context, req *priceDataApi.LoadDataRequest) (*priceDataApi.LoadDataResponse, error) {
	ctx, span := u.tracer.Start(ctx, "handler.LoadData")
	defer span.End()

	err := u.priceCtl.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &priceDataApi.LoadDataResponse{}, nil
}

func (u *priceDataApiServer) FindData(ctx context.Context, req *priceDataApi.FindDataRequest) (*priceDataApi.FindDataResponse, error) {
	ctx, span := u.tracer.Start(ctx, "handler.FindData")
	defer span.End()

	query, err := mapper.ToQueryModel(req.Query)
	if err != nil {
		return nil, err
	}
	entries, err := u.priceCtl.Find(ctx, query)
	if err != nil {
		return nil, err
	}

	return &priceDataApi.FindDataResponse{
		Prices: mapper.ToPriceDataProto(entries),
	}, nil
}
