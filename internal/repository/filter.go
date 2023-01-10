package repository

import (
	"time"
)

// Filter contains the parsed query_string
type Filters struct {
	Action         string
	ActorType      string
	ActorID        string
	EntityType     string
	StartTimestamp time.Time
	EndTimestamp   time.Time
	SortField      string
	SortDescending bool
	PageSize       int
	Page           int
}

// A Metadata provides extra info about the filtered, sorted and paginated
// log records returned on 'GET /v1/event-log?<query_string>'.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}
