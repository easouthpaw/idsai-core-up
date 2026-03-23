package auth

import "errors"

const (
	StatusActive   = "ACTIVE"
	StatusPending  = "PENDING"
	StatusDisabled = "DISABLED"

	TokenPurposeEmailVerification = "EMAIL_VERIFICATION"
	TokenPurposePasswordReset     = "PASSWORD_RESET"

	TokenIssuer             = "idsai-core-up"
	AccessCookieName        = "idsai_access"
	RefreshCookieName       = "idsai_refresh"
	PasswordResetCookieName = "idsai_password_reset"
)

var (
	ErrInvalidInput              = errors.New("invalid input")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrEmailVerificationRequired = errors.New("email verification required")
	ErrTooManyAttempts           = errors.New("too many attempts")
	ErrNotFound                  = errors.New("not found")
	ErrUserExists                = errors.New("user already exists")
	ErrTokenExpired              = errors.New("token expired")
	ErrTokenInvalid              = errors.New("token invalid")
	ErrSessionExpired            = errors.New("session expired")
	ErrSessionInvalid            = errors.New("session invalid")
)
