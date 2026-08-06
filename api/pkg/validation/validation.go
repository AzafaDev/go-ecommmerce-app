package validation

import (
	"regexp"
	"strconv"

	"github.com/go-playground/validator"
)

var priceRegex = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

func New() *validator.Validate {
	v := validator.New()
	_ = v.RegisterValidation("price", validatePrice)
	return v
}

func validatePrice(fl validator.FieldLevel) bool {
	s := fl.Field().String()

	if !priceRegex.MatchString(s) {
		return false
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}
	return val > 0
}
