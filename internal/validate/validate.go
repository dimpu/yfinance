package validate

import "github.com/go-playground/validator/v10"

var V = validator.New()

func Struct(v interface{}) error {
	return V.Struct(v)
}
