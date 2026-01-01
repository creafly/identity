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

		validate.RegisterValidation("uuid", validateUUID)
		validate.RegisterValidation("tenant_role", validateTenantRole)
		validate.RegisterValidation("password", validatePassword)
		validate.RegisterValidation("username", validateUsername)

		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			v.RegisterValidation("uuid", validateUUID)
			v.RegisterValidation("tenant_role", validateTenantRole)
			v.RegisterValidation("password", validatePassword)
			v.RegisterValidation("username", validateUsername)
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

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasNumber = true
		}
	}
	return hasUpper && hasLower && hasNumber
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
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
		return "Password must be at least 8 characters with uppercase, lowercase and number"
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
				c.JSON(http.StatusBadRequest, gin.H{
					"error":  "Validation failed",
					"errors": FormatValidationErrors(validationErrors),
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request body",
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
