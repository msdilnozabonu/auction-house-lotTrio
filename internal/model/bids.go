package model

import "time"

type Bid struct {
	LotID     int64     `json:"lot_id"`
	LotTitle  string    `json:"lot_title"`
	BidderID  int64     `json:"bidder_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
