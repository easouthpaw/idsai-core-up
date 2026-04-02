package auth

import "errors"

const (
	StatusActive   = "ACTIVE"
	StatusPending  = "PENDING"
	StatusDisabled = "DISABLED"

	TokenPurposeEmailVerification = "EMAIL_VERIFICATION"
	TokenPurposePasswordReset     = "PASSWORD_RESET"
	TokenPurposeEmailChange       = "EMAIL_CHANGE"

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
	ErrEmailInUse                = errors.New("email already in use")
	ErrDepartmentNotFound        = errors.New("department not found")
	ErrGroupNotFound             = errors.New("group not found")
	ErrGroupMismatch             = errors.New("group does not belong to department")
	ErrGroupUnchanged            = errors.New("requested group equals current group")
	ErrPendingGroupRequestExists = errors.New("pending group change request already exists")
	ErrGroupRequestNotFound      = errors.New("group change request not found")
	ErrGroupRequestReviewed      = errors.New("group change request already reviewed")
	ErrForbidden                 = errors.New("forbidden")
	ErrTokenExpired              = errors.New("token expired")
	ErrTokenInvalid              = errors.New("token invalid")
	ErrSessionExpired            = errors.New("session expired")
	ErrSessionInvalid            = errors.New("session invalid")
	ErrInvalidCurrentPassword    = errors.New("invalid current password")
	ErrNoPendingEmail            = errors.New("no pending email")
	ErrStorageUnavailable        = errors.New("storage unavailable")
	ErrPasswordResetUnavailable  = errors.New("password reset unavailable")
)
