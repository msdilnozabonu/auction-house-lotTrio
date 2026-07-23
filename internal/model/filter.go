package model

type LotsFilter struct {
	Search   string
	Category string
	MinPrice float64
	MaxPrice float64
	Page     int
	Limit    int
}
