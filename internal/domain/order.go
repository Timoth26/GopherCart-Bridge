package domain

import "time"

type Order struct {
	ID          int         `json:"id"`
	Items       []OrderItem `json:"items"`
	TotalPrice  float64     `json:"total_price"`
	Status      string      `json:"status"`
	DateCreated time.Time   `json:"date_created"`
}

type OrderItem struct {
	ProductID int     `json:"product_id"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	LineTotal float64 `json:"line_total"`
}
