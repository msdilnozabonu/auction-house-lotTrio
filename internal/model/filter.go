package model

import "time"

type LotsFilter struct {
	Search   string
	Category string
	MinPrice float64
	MaxPrice float64
	Page     int
	Limit    int
	Status   string
	SellerID int64
	PageSize int
	DateFrom *time.Time
	DateTo   *time.Time
}
