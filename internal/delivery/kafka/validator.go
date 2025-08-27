package kafka

import (
	"errors"
	"fmt"

	"order/internal/model"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validation for locale field ("ru" or "en")
	_ = validate.RegisterValidation("locale", func(fl validator.FieldLevel) bool {
		locale := fl.Field().String()
		return locale == "ru" || locale == "en"
	})

	// Register custom validation for phone numbers in E.164 format
	_ = validate.RegisterValidation("e164", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if len(phone) < 10 || len(phone) > 16 {
			return false
		}
		if phone[0] != '+' {
			return false
		}
		for _, ch := range phone[1:] {
			if ch < '0' || ch > '9' {
				return false
			}
		}
		return true
	})
}

// ValidateStruct validates any struct using the registered rules
func ValidateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errs string
	for _, e := range err.(validator.ValidationErrors) {
		errs += fmt.Sprintf("Field '%s' is invalid: rules '%s'\n",
			e.Field(), e.ActualTag())
	}
	return errors.New(errs)
}

// ValidateOrder validates an Order struct
func ValidateOrder(o *model.Order) error {
	return ValidateStruct(o)
}

// ValidateDelivery validates a Delivery struct
func ValidateDelivery(d *model.Delivery) error {
	return ValidateStruct(d)
}

// ValidatePayment validates a Payment struct
func ValidatePayment(p *model.Payment) error {
	return ValidateStruct(p)
}

// ValidateItem validates an Item struct
func ValidateItem(i *model.Item) error {
	return ValidateStruct(i)
}
