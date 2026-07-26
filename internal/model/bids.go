package model

import "time"

type Bid struct {
	LotID     int64     `json:"lot_id"`
	LotTitle  string    `json:"lot_title"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}