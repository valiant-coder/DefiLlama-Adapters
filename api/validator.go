package api

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	re := regexp.MustCompile(`^[A-Za-z\d@$!%*?&]{8,}$`)
	containsLower := regexp.MustCompile(`[a-z]`)
	containsUpper := regexp.MustCompile(`[A-Z]`)
	containsDigit := regexp.MustCompile(`\d`)
	containsSpecial := regexp.MustCompile(`[@$!%*?&]`)

	return re.MatchString(password) &&
		containsLower.MatchString(password) &&
		containsUpper.MatchString(password) &&
		containsDigit.MatchString(password) &&
		containsSpecial.MatchString(password)

}



func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", PasswordValidator)
	}
}
