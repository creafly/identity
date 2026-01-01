package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Locale string

const (
	LocaleEnUS Locale = "en-US"
	LocaleRuRU Locale = "ru-RU"
)

var DefaultLocale = LocaleEnUS

var SupportedLocales = []Locale{LocaleEnUS, LocaleRuRU}

type Messages struct {
	Errors        ErrorMessages         `json:"errors"`
	Auth          AuthMessages          `json:"auth"`
	Tenant        TenantMessages        `json:"tenant"`
	Role          RoleMessages          `json:"role"`
	Claim         ClaimMessages         `json:"claim"`
	TOTP          TOTPMessages          `json:"totp"`
	User          UserMessages          `json:"user"`
	PasswordReset PasswordResetMessages `json:"passwordReset"`
}

type ErrorMessages struct {
	InvalidRequest      string `json:"invalidRequest"`
	Unauthorized        string `json:"unauthorized"`
	Forbidden           string `json:"forbidden"`
	NotFound            string `json:"notFound"`
	InternalError       string `json:"internalError"`
	ValidationFailed    string `json:"validationFailed"`
	UserNotFound        string `json:"userNotFound"`
	UserAlreadyExists   string `json:"userAlreadyExists"`
	UsernameAlreadyUsed string `json:"usernameAlreadyUsed"`
	UserBlocked         string `json:"userBlocked"`
	InvalidCredentials  string `json:"invalidCredentials"`
	TokenExpired        string `json:"tokenExpired"`
	InvalidToken        string `json:"invalidToken"`
}

type AuthMessages struct {
	RegisterSuccess string `json:"registerSuccess"`
	LoginSuccess    string `json:"loginSuccess"`
	LogoutSuccess   string `json:"logoutSuccess"`
	PasswordChanged string `json:"passwordChanged"`
	TokenRefreshed  string `json:"tokenRefreshed"`
}

type TenantMessages struct {
	Created          string `json:"created"`
	Updated          string `json:"updated"`
	Deleted          string `json:"deleted"`
	NotFound         string `json:"notFound"`
	AlreadyExists    string `json:"alreadyExists"`
	InvalidSlug      string `json:"invalidSlug"`
	MemberAdded      string `json:"memberAdded"`
	MemberRemoved    string `json:"memberRemoved"`
	InvitationSent   string `json:"invitationSent"`
	InvitationFailed string `json:"invitationFailed"`
	AlreadyMember    string `json:"alreadyMember"`
	CallbackSuccess  string `json:"callbackSuccess"`
}

type RoleMessages struct {
	Created       string `json:"created"`
	Updated       string `json:"updated"`
	Deleted       string `json:"deleted"`
	NotFound      string `json:"notFound"`
	AlreadyExists string `json:"alreadyExists"`
	Assigned      string `json:"assigned"`
	Unassigned    string `json:"unassigned"`
}

type ClaimMessages struct {
	Created       string `json:"created"`
	Deleted       string `json:"deleted"`
	NotFound      string `json:"notFound"`
	AlreadyExists string `json:"alreadyExists"`
	Assigned      string `json:"assigned"`
	Unassigned    string `json:"unassigned"`
}

type TOTPMessages struct {
	SetupSuccess         string `json:"setupSuccess"`
	EnabledSuccess       string `json:"enabledSuccess"`
	DisabledSuccess      string `json:"disabledSuccess"`
	InvalidCode          string `json:"invalidCode"`
	AlreadyEnabled       string `json:"alreadyEnabled"`
	NotEnabled           string `json:"notEnabled"`
	NotSetup             string `json:"notSetup"`
	VerificationRequired string `json:"verificationRequired"`
}

type UserMessages struct {
	Blocked   string `json:"blocked"`
	Unblocked string `json:"unblocked"`
}

type PasswordResetMessages struct {
	RequestSent   string `json:"requestSent"`
	PasswordReset string `json:"passwordReset"`
	TokenNotFound string `json:"tokenNotFound"`
	TokenExpired  string `json:"tokenExpired"`
}

var (
	localeCache = make(map[Locale]*Messages)
	cacheMutex  sync.RWMutex
)

func IsValidLocale(locale string) bool {
	for _, l := range SupportedLocales {
		if string(l) == locale {
			return true
		}
	}
	return false
}

func ParseLocale(locale string) Locale {
	if IsValidLocale(locale) {
		return Locale(locale)
	}
	return DefaultLocale
}

func GetMessages(locale Locale) *Messages {
	cacheMutex.RLock()
	if messages, ok := localeCache[locale]; ok {
		cacheMutex.RUnlock()
		return messages
	}
	cacheMutex.RUnlock()

	messages, err := loadLocale(locale)
	if err != nil {
		if locale != DefaultLocale {
			return GetMessages(DefaultLocale)
		}
		return getHardcodedFallback()
	}

	cacheMutex.Lock()
	localeCache[locale] = messages
	cacheMutex.Unlock()

	return messages
}

func loadLocale(locale Locale) (*Messages, error) {
	filePath := filepath.Join("resources", string(locale)+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var messages Messages
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return &messages, nil
}

func getHardcodedFallback() *Messages {
	return &Messages{
		Errors: ErrorMessages{
			InvalidRequest:      "Invalid request",
			Unauthorized:        "Unauthorized",
			Forbidden:           "Forbidden",
			NotFound:            "Not found",
			InternalError:       "Internal server error",
			ValidationFailed:    "Validation failed",
			UserNotFound:        "User not found",
			UserAlreadyExists:   "User already exists",
			UsernameAlreadyUsed: "Username is already taken",
			UserBlocked:         "Your account has been blocked",
			InvalidCredentials:  "Invalid credentials",
			TokenExpired:        "Token expired",
			InvalidToken:        "Invalid token",
		},
		Auth: AuthMessages{
			RegisterSuccess: "Registration successful",
			LoginSuccess:    "Login successful",
			LogoutSuccess:   "Logout successful",
			PasswordChanged: "Password changed successfully",
			TokenRefreshed:  "Token refreshed successfully",
		},
		Tenant: TenantMessages{
			Created:          "Tenant created successfully",
			Updated:          "Tenant updated successfully",
			Deleted:          "Tenant deleted successfully",
			NotFound:         "Tenant not found",
			AlreadyExists:    "Tenant with this slug already exists",
			InvalidSlug:      "Invalid slug format",
			MemberAdded:      "Member added to tenant",
			MemberRemoved:    "Member removed from tenant",
			InvitationSent:   "Invitation sent successfully",
			InvitationFailed: "Failed to send invitation",
			AlreadyMember:    "User is already a member of this tenant",
			CallbackSuccess:  "Member added via invitation callback",
		},
		Role: RoleMessages{
			Created:       "Role created successfully",
			Updated:       "Role updated successfully",
			Deleted:       "Role deleted successfully",
			NotFound:      "Role not found",
			AlreadyExists: "Role with this name already exists",
			Assigned:      "Role assigned successfully",
			Unassigned:    "Role unassigned successfully",
		},
		Claim: ClaimMessages{
			Created:       "Claim created successfully",
			Deleted:       "Claim deleted successfully",
			NotFound:      "Claim not found",
			AlreadyExists: "Claim already exists",
			Assigned:      "Claim assigned successfully",
			Unassigned:    "Claim unassigned successfully",
		},
		TOTP: TOTPMessages{
			SetupSuccess:         "Two-factor authentication setup initiated",
			EnabledSuccess:       "Two-factor authentication enabled successfully",
			DisabledSuccess:      "Two-factor authentication disabled successfully",
			InvalidCode:          "Invalid verification code",
			AlreadyEnabled:       "Two-factor authentication is already enabled",
			NotEnabled:           "Two-factor authentication is not enabled",
			NotSetup:             "Two-factor authentication has not been set up",
			VerificationRequired: "Two-factor authentication required",
		},
		User: UserMessages{
			Blocked:   "User blocked successfully",
			Unblocked: "User unblocked successfully",
		},
		PasswordReset: PasswordResetMessages{
			RequestSent:   "If an account with that email exists, a password reset link has been sent",
			PasswordReset: "Password has been reset successfully",
			TokenNotFound: "Invalid or expired password reset token",
			TokenExpired:  "Password reset token has expired",
		},
	}
}

func PreloadLocales() {
	for _, locale := range SupportedLocales {
		GetMessages(locale)
	}
}

func ClearCache() {
	cacheMutex.Lock()
	localeCache = make(map[Locale]*Messages)
	cacheMutex.Unlock()
}
