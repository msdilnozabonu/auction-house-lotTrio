package model

import "time"

type WatchItem struct {
	LotID        int64     `json:"lot_id"`
	LotTitle     string    `json:"lot_title"`
	CurrentPrice float64   `json:"current_price"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}
