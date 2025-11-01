package mapper

import (
	"errors"
	"strconv"

	price_data_api "github.com/aluo/api/zeonology/price_data/v1"
	"github.com/aluo/gomono/zeonology/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToQueryModel(protoQuery *price_data_api.Query) (model.Query, error) {
	unit, interval, err := parse(protoQuery.Window)
	if err != nil {
		return model.Query{}, err
	}

	return model.Query{
		StartTime:      protoQuery.Start.AsTime(),
		EndTime:        protoQuery.End.AsTime(),
		WindowUnit:     unit,
		WindowInterval: interval,
		Aggregation:    toAggregationModel(protoQuery.Aggregation),
	}, nil
}

func toAggregationModel(aggregation price_data_api.Aggregation) model.Aggregation {
	switch aggregation {
	case price_data_api.Aggregation_AGGREGATION_MIN:
		return model.Aggregation_MIN
	case price_data_api.Aggregation_AGGREGATION_AVG:
		return model.Aggregation_AVG
	case price_data_api.Aggregation_AGGREGATION_MAX:
		return model.Aggregation_MAX
	case price_data_api.Aggregation_AGGREGATION_SUM:
		return model.Aggregation_SUM
	default:
		return model.Aggregation_INVALID
	}
}

func parse(s string) (model.TimeUnit, int, error) {
	unit := s[len(s)-1:]
	var timeUnit model.TimeUnit
	if unit == "m" {
		timeUnit = model.TimeUnit_MINUTE
	} else if unit == "d" {
		timeUnit = model.TimeUnit_DAY
	} else if unit == "h" {
		timeUnit = model.TimeUnit_HOUR
	} else {
		return model.TimeUnit_INVALID, 0, errors.New("invalid window unit")
	}
	interv := s[0 : len(s)-1]
	num, err := strconv.Atoi(interv)
	if err != nil {
		return model.TimeUnit_INVALID, 0, err
	}
	return timeUnit, num, nil
}

func ToPriceDataProto(entries []model.Entry) []*price_data_api.PriceData {
	protoPd := make([]*price_data_api.PriceData, len(entries))
	for i, v := range entries {
		protoPd[i] = &price_data_api.PriceData{
			Time:  timestamppb.New(v.Time),
			Value: float64(v.Value),
		}
	}
	return protoPd
}
