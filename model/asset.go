package model

import "time"

type Entry struct {
	Time  time.Time `bson:"timestamp"`
	Value float64   `bson:"price"`
}
