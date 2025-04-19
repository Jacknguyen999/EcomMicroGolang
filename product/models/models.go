package models

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	AccountID   int     `json:"accountID"`
	Category    string  `json:"category"`
}

type ProductDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	AccountID   int     `json:"accountID"`
	Category    string  `json:"category"`
}

type EventData struct {
	ID          *string  `json:"product_id"`
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	AccountID   *int     `json:"accountID"`
	Category    *string  `json:"category"`
}

type Event struct {
	Type string    `json:"type"`
	Data EventData `json:"data"`
}

type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
