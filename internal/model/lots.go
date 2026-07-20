package model

import "time"

type Lots struct {
	ID           int
	SellerID     int
	Title        string
	StartPrice   int
	CurrentPrice int
	Status       string
	StartDate    time.Time
	EndDate      time.Time
}
