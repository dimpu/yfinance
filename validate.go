package yahoofinance

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validateOptions(opts interface{}) error {
	return validate.Struct(opts)
}
