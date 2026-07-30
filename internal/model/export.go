package model

import "time"

type LotsExport struct {
	ID int64 `json:"id"`
	Title string `json:"title"`
	Category string `json:"category"`
	Status string `json:"status"`
	SellerID int64 `json:"seller_id"`
	Price float64 `json:"price"`
	EndsAt time.Time `json:"ends_at"`
}
