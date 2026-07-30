package model

import "time"

type LotsFilter struct {
	Search    string
	Category  string
	MinPrice  float64
	MaxPrice  float64
	Page      int
	Limit     int
	Status    string
	SellerID  int64
	DateField string
	DateFrom  *time.Time
	DateTo    *time.Time
}

type BidFilter struct {
	LotID  int64
	Limit  int
	Offset int
}
