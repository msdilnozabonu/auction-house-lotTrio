package model

import "time"

type CancelLot struct {
	LotID int64 `json:"lot_id"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type ReportLot struct {
	LotID int64 `json:"lot_id"`
	Reason string `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}
