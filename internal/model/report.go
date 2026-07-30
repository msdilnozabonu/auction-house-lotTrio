package model

import "time"

type CancelLot struct {
	LotID  int64  `json:"lot_id"`
	Reason string `json:"reason"`
	Status string `json:"status"`
}

type ReportLot struct {
	ID         int64     `json:"id"`
	LotID      int64     `json:"lot_id"`
	ReporterID int64     `json:"reporter_id"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
