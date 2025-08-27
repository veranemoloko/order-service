package model

import "time"

// Order represents a customer order
type Order struct {
	OrderUID          string    `json:"order_uid"          validate:"required" gorm:"type:varchar(255);primaryKey"`
	TrackNumber       string    `json:"track_number"       validate:"required,alphanumunicode" gorm:"type:varchar(255);not null"`
	Entry             string    `json:"entry"              validate:"required,alphanumunicode" gorm:"type:varchar(50);not null"`
	Locale            string    `json:"locale"             validate:"required" gorm:"type:varchar(10);not null"`
	InternalSignature string    `json:"internal_signature" validate:"-" gorm:"type:varchar(255)"`
	CustomerID        string    `json:"customer_id"        validate:"required,alphanumunicode,min=1,max=64" gorm:"type:varchar(255);not null"`
	DeliveryService   string    `json:"delivery_service"   validate:"required,ascii,max=50" gorm:"type:varchar(255);not null"`
	Shardkey          string    `json:"shardkey"           validate:"required,numeric" gorm:"type:varchar(50);not null"`
	SmID              int       `json:"sm_id"              validate:"required,gt=0" gorm:"not null"`
	DateCreated       time.Time `json:"date_created"       validate:"required" gorm:"not null"`
	OofShard          string    `json:"oof_shard"          validate:"required,numeric" gorm:"type:varchar(50);not null"`

	DeliveryID uint     `json:"-"`
	Delivery   Delivery `json:"delivery"                   gorm:"foreignKey:DeliveryID"`

	PaymentID uint    `json:"-"`
	Payment   Payment `json:"payment"                      gorm:"foreignKey:PaymentID"`

	Items []Item `json:"items" gorm:"many2many:order_items;foreignKey:OrderUID;constraint:OnDelete:CASCADE"`
}

// Delivery represents delivery details for an order
type Delivery struct {
	ID      uint   `json:"-"       validate:"-"                      gorm:"primaryKey;autoIncrement"`
	Name    string `json:"name"    validate:"required,min=2,max=255" gorm:"type:varchar(255);not null"`
	Phone   string `json:"phone"   validate:"required,e164"          gorm:"type:varchar(50);not null"`
	Zip     string `json:"zip"     validate:"required,min=4,max=12"  gorm:"type:varchar(50);not null"`
	City    string `json:"city"    validate:"required,ascii,max=255" gorm:"type:varchar(255);not null"`
	Address string `json:"address" validate:"required,max=255"       gorm:"type:varchar(255);not null"`
	Region  string `json:"region"  validate:"required,max=255"       gorm:"type:varchar(255);not null"`
	Email   string `json:"email"   validate:"required,email"         gorm:"type:varchar(255);not null"`
}

// Payment represents payment details for an order
type Payment struct {
	ID           uint   `json:"-"             validate:"-"                     gorm:"primaryKey;autoIncrement"`
	Transaction  string `json:"transaction"   validate:"required"              gorm:"type:varchar(255);not null"`
	RequestID    string `json:"request_id"    validate:"-"                     gorm:"type:varchar(255)"`
	Currency     string `json:"currency"      validate:"required"              gorm:"type:varchar(10);not null"`
	Provider     string `json:"provider"      validate:"required,ascii,max=50" gorm:"type:varchar(50);not null"`
	Amount       int    `json:"amount"        validate:"required,gt=0"         gorm:"not null"`
	PaymentDt    int64  `json:"payment_dt"    validate:"required,gt=0"         gorm:"not null"`
	Bank         string `json:"bank"          validate:"required,max=50"       gorm:"type:varchar(50);not null"`
	DeliveryCost int    `json:"delivery_cost" validate:"required,gte=0"        gorm:"not null"`
	GoodsTotal   int    `json:"goods_total"   validate:"required,gte=0"        gorm:"not null"`
	CustomFee    int    `json:"custom_fee"    validate:"gte=0"                 gorm:"not null"`
}

// Item represents an individual item within an order
type Item struct {
	ID          uint   `json:"-"            validate:"-"                                     gorm:"primaryKey;autoIncrement"`
	Rid         string `json:"rid"          validate:"required,alphanumunicode,min=5,max=64" gorm:"type:varchar(255)"`
	ChrtID      int    `json:"chrt_id"      validate:"required,gte=0"                        gorm:"not null"`
	TrackNumber string `json:"track_number" validate:"required,alphanumunicode,max=32"       gorm:"type:varchar(255);not null"`
	Price       int    `json:"price"        validate:"required,gt=0"                         gorm:"not null"`
	Name        string `json:"name"         validate:"required,ascii,max=255"                gorm:"type:varchar(255);not null"`
	Sale        int    `json:"sale"         validate:"gte=0,lte=100"                         gorm:"not null"`
	Size        string `json:"size"         validate:"-"                                     gorm:"type:varchar(50);not null"`
	TotalPrice  int    `json:"total_price"  validate:"required,gt=0"                         gorm:"not null"`
	NmID        int    `json:"nm_id"        validate:"required,gt=0"                         gorm:"not null"`
	Brand       string `json:"brand"        validate:"required,max=255"                      gorm:"type:varchar(255);not null"`
	Status      int    `json:"status"       validate:"required,gte=0"                        gorm:"not null"`
}
