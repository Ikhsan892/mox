package models

type OrderItem struct {
	BaseModel
	ProductId   string `gorm:"not null"`
	ProductName string `gorm:"not null"`
	Price       float64
}

// TableName sets the name of the table
func (OrderItem) TableName() string {
	return "orders.order_items"
}
