package model

import "time"

type Lots struct {
	ID           int64
	SellerID     int64
	Title        string
	Description  string
	Category     string
	StartPrice   float64
	CurrentPrice float64
	WinnerID     int64
	Status       string
	Photo        string
	ModerationStatus string
	RejectionReason  string
	StartAt      time.Time
	EndAt        time.Time
	CreatedAt    time.Time
}
