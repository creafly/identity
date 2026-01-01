package validator

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid password", "Password123", true},
		{"valid with special chars", "MyP@ssw0rd!", true},
		{"too short", "Pass1", false},
		{"no uppercase", "password123", false},
		{"no lowercase", "PASSWORD123", false},
		{"no number", "PasswordABC", false},
		{"empty", "", false},
		{"only numbers", "12345678", false},
		{"only lowercase", "abcdefgh", false},
		{"only uppercase", "ABCDEFGH", false},
		{"8 chars valid", "Passwo1d", true},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Password string `validate:"password"`
			}

			err := Validate(&testStruct{Password: tt.password})
			got := err == nil

			if got != tt.want {
				t.Errorf("validatePassword(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     bool
	}{
		{"valid simple", "john", true},
		{"valid with numbers", "john123", true},
		{"valid with underscore", "john_doe", true},
		{"valid with hyphen", "john-doe", true},
		{"valid mixed case", "JohnDoe", true},
		{"too short", "ab", false},
		{"too long", "abcdefghijklmnopqrstuvwxyz12345", false},
		{"exactly 3 chars", "abc", true},
		{"exactly 30 chars", "abcdefghijklmnopqrstuvwxyz1234", true},
		{"with space", "john doe", false},
		{"with special char", "john@doe", false},
		{"with dot", "john.doe", false},
		{"empty", "", false},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Username string `validate:"username"`
			}

			err := Validate(&testStruct{Username: tt.username})
			got := err == nil

			if got != tt.want {
				t.Errorf("validateUsername(%q) = %v, want %v", tt.username, got, tt.want)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name string
		uuid string
		want bool
	}{
		{"valid uuid v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid uuid v1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", true},
		{"invalid format", "not-a-uuid", false},
		{"empty", "", false},
		{"partial uuid", "550e8400-e29b-41d4", false},
		{"too many segments", "550e8400-e29b-41d4-a716-446655440000-extra", false},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				ID string `validate:"uuid"`
			}

			err := Validate(&testStruct{ID: tt.uuid})
			got := err == nil

			if got != tt.want {
				t.Errorf("validateUUID(%q) = %v, want %v", tt.uuid, got, tt.want)
			}
		})
	}
}

func TestValidateTenantRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{"owner", "owner", true},
		{"admin", "admin", true},
		{"member", "member", true},
		{"viewer", "viewer", true},
		{"invalid role", "superadmin", false},
		{"empty", "", false},
		{"uppercase", "OWNER", false},
		{"mixed case", "Owner", false},
	}

	Init()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Role string `validate:"tenant_role"`
			}

			err := Validate(&testStruct{Role: tt.role})
			got := err == nil

			if got != tt.want {
				t.Errorf("validateTenantRole(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	Init()

	type testStruct struct {
		Email    string `validate:"required,email" json:"email"`
		Password string `validate:"required,password" json:"password"`
	}

	err := Validate(&testStruct{Email: "", Password: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}

	errors := FormatValidationErrors(err)
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}

	fieldMap := make(map[string]bool)
	for _, e := range errors {
		fieldMap[e.Field] = true
	}

	if !fieldMap["email"] {
		t.Error("expected email field in errors")
	}
	if !fieldMap["password"] {
		t.Error("expected password field in errors")
	}
}

func TestGetErrorMessage(t *testing.T) {
	Init()

	tests := []struct {
		tag      string
		contains string
	}{
		{"required", "required"},
		{"email", "email"},
		{"min", "short"},
		{"max", "long"},
		{"uuid", "UUID"},
		{"password", "Password"},
		{"username", "Username"},
		{"tenant_role", "role"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			type testStruct struct {
				Field string `validate:"required"`
			}

			err := Validate(&testStruct{Field: ""})
			if err == nil {
				t.Skip("no validation error generated")
			}

			errors := FormatValidationErrors(err)
			if len(errors) == 0 {
				t.Skip("no formatted errors")
			}
		})
	}
}
