package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	MinLength = 10

	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

var (
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters", MinLength)
	ErrPasswordBlank    = errors.New("password must not be blank")
)

type VerifyResult struct {
	Valid       bool
	NeedsRehash bool
}

func Validate(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrPasswordBlank
	}
	if len(password) < MinLength {
		return ErrPasswordTooShort
	}
	return nil
}

func Hash(password string) (string, error) {
	if err := Validate(password); err != nil {
		return "", err
	}

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	sum := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

func Verify(encodedHash, password string) (VerifyResult, error) {
	switch {
	case strings.HasPrefix(encodedHash, "$argon2id$"):
		return verifyArgon2ID(encodedHash, password)
	case strings.HasPrefix(encodedHash, "$2a$"), strings.HasPrefix(encodedHash, "$2b$"), strings.HasPrefix(encodedHash, "$2y$"):
		err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return VerifyResult{}, nil
			}
			return VerifyResult{}, err
		}
		return VerifyResult{Valid: true, NeedsRehash: true}, nil
	default:
		return VerifyResult{}, nil
	}
}

func verifyArgon2ID(encodedHash, password string) (VerifyResult, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return VerifyResult{}, nil
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return VerifyResult{}, nil
	}
	if version != argon2.Version {
		return VerifyResult{}, nil
	}

	var memory uint32
	var timeCost uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return VerifyResult{}, nil
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return VerifyResult{}, nil
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return VerifyResult{}, nil
	}

	sum := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(expected)))
	if subtle.ConstantTimeCompare(sum, expected) != 1 {
		return VerifyResult{}, nil
	}

	needsRehash := memory != argonMemory || timeCost != argonTime || threads != argonThreads || len(expected) != int(argonKeyLen)
	return VerifyResult{Valid: true, NeedsRehash: needsRehash}, nil
}
