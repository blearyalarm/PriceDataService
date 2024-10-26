package model

import "time"

// LastUpdateEntry represents the document structure for the last update time
type LastUpdateEntry struct {
	LastUpdateTime time.Time `bson:"lastUpdateTime"`
}
