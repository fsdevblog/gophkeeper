package http

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// validateMaxBytes verify length of bytes (not length of rune as gin validator).
func validateMaxBytes(fl validator.FieldLevel) bool {
	param := fl.Param() // получаем значение из тега
	maxBytes, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	str, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return len([]byte(str)) <= maxBytes
}

func registerValidators() error {
	v, _ := binding.Validator.Engine().(*validator.Validate)
	if err := v.RegisterValidation("max_bytes", validateMaxBytes); err != nil {
		return fmt.Errorf("validator registration: %s", err.Error())
	}
	return nil
}
