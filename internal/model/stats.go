package model

type PlatformStats struct {
	LotsByStatus  map[string]int    `json:"lots_by_status"`
	TotalRevenue  float64           `json:"total_revenue"`
	AverageCheck  float64           `json:"average_check"`
	TopCategories []CategoryStat    `json:"top_categories"`
}

type CategoryStat struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Revenue  float64 `json:"revenue"`
}
