package ecommerce

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product in the catalog
type Product struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	SKU          string         `json:"sku" gorm:"size:100;uniqueIndex;not null"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	Slug         string         `json:"slug" gorm:"size:255;uniqueIndex;not null"`
	Description  string         `json:"description" gorm:"type:text"`
	Price        float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	ComparePrice float64        `json:"compare_price" gorm:"type:decimal(10,2)"`
	CostPrice    float64        `json:"cost_price" gorm:"type:decimal(10,2)"`
	CategoryID   *uint          `json:"category_id"`
	Brand        string         `json:"brand" gorm:"size:100"`
	Stock        int            `json:"stock" gorm:"default:0"`
	LowStock     int            `json:"low_stock" gorm:"default:5"`
	Images       string         `json:"images" gorm:"type:text"`                // JSON array
	Tags         string         `json:"tags" gorm:"type:text"`                  // JSON array
	Status       string         `json:"status" gorm:"size:20;default:'active'"` // active, inactive, out_of_stock
	Featured     bool           `json:"featured" gorm:"default:false"`
	Weight       float64        `json:"weight" gorm:"type:decimal(10,2)"` // in kg
	SEOTitle     string         `json:"seo_title" gorm:"size:255"`
	SEODesc      string         `json:"seo_description" gorm:"size:500"`
	ViewCount    int            `json:"view_count" gorm:"default:0"`
	SoldCount    int            `json:"sold_count" gorm:"default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Product) TableName() string {
	return "ec_products"
}

// ProductCategory represents a product category
type ProductCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"size:100;not null"`
	Slug        string         `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id"`
	Image       string         `json:"image" gorm:"size:500"`
	Order       int            `json:"order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (ProductCategory) TableName() string {
	return "ec_product_categories"
}

// Cart represents a shopping cart
type Cart struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    *uint          `json:"user_id"`
	SessionID string         `json:"session_id" gorm:"size:255;index"`
	Items     string         `json:"items" gorm:"type:text"` // JSON array
	Subtotal  float64        `json:"subtotal" gorm:"type:decimal(10,2)"`
	Tax       float64        `json:"tax" gorm:"type:decimal(10,2)"`
	Total     float64        `json:"total" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Cart) TableName() string {
	return "ec_carts"
}

// Order represents a customer order
type Order struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	OrderNumber     string         `json:"order_number" gorm:"size:50;uniqueIndex;not null"`
	UserID          uint           `json:"user_id"`
	Status          string         `json:"status" gorm:"size:50;default:'pending'"`         // pending, processing, shipped, delivered, cancelled
	PaymentStatus   string         `json:"payment_status" gorm:"size:50;default:'pending'"` // pending, paid, failed, refunded
	PaymentMethod   string         `json:"payment_method" gorm:"size:50"`
	ShippingMethod  string         `json:"shipping_method" gorm:"size:50"`
	Items           string         `json:"items" gorm:"type:text"` // JSON array
	Subtotal        float64        `json:"subtotal" gorm:"type:decimal(10,2)"`
	Tax             float64        `json:"tax" gorm:"type:decimal(10,2)"`
	ShippingCost    float64        `json:"shipping_cost" gorm:"type:decimal(10,2)"`
	Discount        float64        `json:"discount" gorm:"type:decimal(10,2)"`
	Total           float64        `json:"total" gorm:"type:decimal(10,2)"`
	CouponCode      string         `json:"coupon_code" gorm:"size:50"`
	CustomerName    string         `json:"customer_name" gorm:"size:255"`
	CustomerEmail   string         `json:"customer_email" gorm:"size:255"`
	CustomerPhone   string         `json:"customer_phone" gorm:"size:50"`
	ShippingAddress string         `json:"shipping_address" gorm:"type:text"`
	BillingAddress  string         `json:"billing_address" gorm:"type:text"`
	Notes           string         `json:"notes" gorm:"type:text"`
	TrackingNumber  string         `json:"tracking_number" gorm:"size:100"`
	PaidAt          *time.Time     `json:"paid_at"`
	ShippedAt       *time.Time     `json:"shipped_at"`
	DeliveredAt     *time.Time     `json:"delivered_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Order) TableName() string {
	return "ec_orders"
}

// Payment represents a payment transaction
type Payment struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	OrderID         uint           `json:"order_id"`
	TransactionID   string         `json:"transaction_id" gorm:"size:255"`
	PaymentMethod   string         `json:"payment_method" gorm:"size:50"`
	Amount          float64        `json:"amount" gorm:"type:decimal(10,2)"`
	Status          string         `json:"status" gorm:"size:50"` // pending, completed, failed, refunded
	PaymentGateway  string         `json:"payment_gateway" gorm:"size:50"`
	GatewayResponse string         `json:"gateway_response" gorm:"type:text"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Payment) TableName() string {
	return "ec_payments"
}

// Coupon represents a discount coupon
type Coupon struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Code        string         `json:"code" gorm:"size:50;uniqueIndex;not null"`
	Type        string         `json:"type" gorm:"size:20"` // percentage, fixed
	Value       float64        `json:"value" gorm:"type:decimal(10,2)"`
	MinAmount   float64        `json:"min_amount" gorm:"type:decimal(10,2)"`
	MaxDiscount float64        `json:"max_discount" gorm:"type:decimal(10,2)"`
	UsageLimit  int            `json:"usage_limit"`
	UsageCount  int            `json:"usage_count" gorm:"default:0"`
	UserLimit   int            `json:"user_limit" gorm:"default:1"`
	Status      string         `json:"status" gorm:"size:20;default:'active'"` // active, inactive, expired
	StartsAt    *time.Time     `json:"starts_at"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Coupon) TableName() string {
	return "ec_coupons"
}

// Review represents a product review
type Review struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id"`
	UserID    uint           `json:"user_id"`
	Rating    int            `json:"rating" gorm:"default:5"` // 1-5
	Title     string         `json:"title" gorm:"size:255"`
	Comment   string         `json:"comment" gorm:"type:text"`
	Status    string         `json:"status" gorm:"size:20;default:'pending'"` // pending, approved, rejected
	Helpful   int            `json:"helpful" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName specifies the table name
func (Review) TableName() string {
	return "ec_reviews"
}
