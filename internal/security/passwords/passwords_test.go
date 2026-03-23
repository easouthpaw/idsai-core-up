package passwords

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerifyArgon2ID(t *testing.T) {
	hash, err := Hash("correct-horse-battery1")
	require.NoError(t, err)
	require.Contains(t, hash, "$argon2id$")

	result, err := Verify(hash, "correct-horse-battery1")
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.False(t, result.NeedsRehash)
}

func TestVerifyBcryptRequestsRehash(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("legacy-password1"), bcrypt.DefaultCost)
	require.NoError(t, err)

	result, err := Verify(string(hash), "legacy-password1")
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.True(t, result.NeedsRehash)
}

func TestValidateRejectsShortPassword(t *testing.T) {
	err := Validate("short")
	require.Error(t, err)
}
