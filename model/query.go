package model

import "time"

type Query struct {
	StartTime time.Time
	EndTime   time.Time

	WindowUnit     TimeUnit
	WindowInterval int
	Aggregation    Aggregation
}

type TimeUnit string

const (
	TimeUnit_INVALID TimeUnit = "INVALID"
	TimeUnit_MINUTE  TimeUnit = "minute"
	TimeUnit_HOUR    TimeUnit = "hour"
	TimeUnit_DAY     TimeUnit = "day"
)

type Aggregation string

const (
	Aggregation_INVALID Aggregation = "INVALID"
	Aggregation_MIN     Aggregation = "min"
	Aggregation_MAX     Aggregation = "max"
	Aggregation_AVG     Aggregation = "avg"
	Aggregation_SUM     Aggregation = "sum"
)
