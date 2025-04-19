package models

import "time"

type Order struct {
	ID            uint `gorm:"primaryKey;autoIncrement"`
	CreatedAt     time.Time
	TotalPrice    float64
	AccountID     string
	ProductsInfos []ProductsInfo    `gorm:"foreignKey:OrderID"`
	Products      []*OrderedProduct `gorm:"-"`
}

type ProductsInfo struct {
	ID        uint `gorm:"primaryKey;autoIncrement"`
	OrderID   uint
	ProductID string
	Quantity  int
}

func (ProductsInfo) TableName() string {
	return "order_products"
}

type OrderedProduct struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    uint32
}

type EventData struct {
	AccountId int    `json:"user_id"`
	ProductId string `json:"product_id"`
}

type Event struct {
	Type      string    `json:"type"`
	EventData EventData `json:"data"`
}
