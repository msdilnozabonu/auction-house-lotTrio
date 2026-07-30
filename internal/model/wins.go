package model

import "time"

type Win struct {
	LotID      int64     `json:"lot_id"`
	LotTitle   string    `json:"lot_title"`
	WinningBid float64   `json:"winning_bid"`
	ClosedAt   time.Time `json:"closed_at"`
}
