package models

type Order struct {
	BaseModel
	CustomerName string  `gorm:"not null"`
	TotalAmount  float64 `gorm:"not null"`
	Status       string  `gorm:"not null"`
	Address      string
}

// TableName sets the name of the table
func (Order) TableName() string {
	return "orders.orders"
}
