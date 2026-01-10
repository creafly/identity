package validator

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var (
	validate *validator.Validate
	once     sync.Once
)

func Init() {
	once.Do(func() {
		validate = validator.New()

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		_ = validate.RegisterValidation("uuid", validateUUID)
		_ = validate.RegisterValidation("tenant_role", validateTenantRole)
		_ = validate.RegisterValidation("password", validatePassword)
		_ = validate.RegisterValidation("username", validateUsername)

		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation("uuid", validateUUID)
			_ = v.RegisterValidation("tenant_role", validateTenantRole)
			_ = v.RegisterValidation("password", validatePassword)
			_ = v.RegisterValidation("username", validateUsername)
		}
	})
}

func Get() *validator.Validate {
	Init()
	return validate
}

func Validate(s interface{}) error {
	Init()
	return validate.Struct(s)
}

func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

func validateTenantRole(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	validRoles := []string{"owner", "admin", "member", "viewer"}
	for _, r := range validRoles {
		if role == r {
			return true
		}
	}
	return false
}

var commonPasswords = map[string]bool{
	"password": true, "123456": true, "12345678": true, "qwerty": true, "abc123": true,
	"monkey": true, "1234567": true, "letmein": true, "trustno1": true, "dragon": true,
	"baseball": true, "iloveyou": true, "master": true, "sunshine": true, "ashley": true,
	"bailey": true, "shadow": true, "123123": true, "654321": true, "superman": true,
	"qazwsx": true, "michael": true, "football": true, "password1": true, "password123": true,
	"batman": true, "login": true, "admin": true, "welcome": true, "hello": true,
	"charlie": true, "donald": true, "password12": true, "qwerty123": true, "admin123": true,
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 || len(password) > 72 {
		return false
	}

	if commonPasswords[strings.ToLower(password)] {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;':\",./<>?`~", c):
			hasSpecial = true
		}
	}

	typesCount := 0
	if hasUpper {
		typesCount++
	}
	if hasLower {
		typesCount++
	}
	if hasNumber {
		typesCount++
	}
	if hasSpecial {
		typesCount++
	}

	return typesCount >= 3
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	for _, c := range username {
		isLower := c >= 'a' && c <= 'z'
		isUpper := c >= 'A' && c <= 'Z'
		isDigit := c >= '0' && c <= '9'
		isAllowed := isLower || isUpper || isDigit || c == '_' || c == '-'
		if !isAllowed {
			return false
		}
	}
	return true
}

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Error       string       `json:"error"`
	Code        string       `json:"code"`
	FieldErrors []FieldError `json:"fieldErrors"`
}

func FormatValidationErrors(err error) []FieldError {
	var errors []FieldError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, FieldError{
				Field:   e.Field(),
				Code:    e.Tag(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

func HandleBindError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		c.JSON(http.StatusBadRequest, ValidationErrorResponse{
			Error:       "Validation failed",
			Code:        "VALIDATION_ERROR",
			FieldErrors: FormatValidationErrors(validationErrors),
		})
		return true
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"error": "Invalid request body",
		"code":  "INVALID_REQUEST_BODY",
	})
	return true
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	case "uuid":
		return "Invalid UUID format"
	case "password":
		return "Password must be 8-72 characters with at least 3 of: uppercase, lowercase, number, special character. Common passwords are not allowed."
	case "username":
		return "Username must be 3-30 characters and contain only letters, numbers, underscores and hyphens"
	case "tenant_role":
		return "Invalid role. Must be one of: owner, admin, member, viewer"
	case "oneof":
		return "Value must be one of the allowed options"
	case "gte":
		return "Value must be greater than or equal to minimum"
	case "lte":
		return "Value must be less than or equal to maximum"
	case "alphanum":
		return "Value must contain only alphanumeric characters"
	case "url":
		return "Invalid URL format"
	default:
		return "Invalid value"
	}
}

func ValidationMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body T
		if err := c.ShouldBindJSON(&body); err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				c.JSON(http.StatusBadRequest, ValidationErrorResponse{
					Error:       "Validation failed",
					Code:        "VALIDATION_ERROR",
					FieldErrors: FormatValidationErrors(validationErrors),
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
				"code":  "INVALID_REQUEST_BODY",
			})
			c.Abort()
			return
		}

		c.Set("validated_body", body)
		c.Next()
	}
}

func GetValidatedBody[T any](c *gin.Context) (T, bool) {
	var zero T
	body, exists := c.Get("validated_body")
	if !exists {
		return zero, false
	}
	typed, ok := body.(T)
	return typed, ok
}
