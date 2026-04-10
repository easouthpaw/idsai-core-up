package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"idsai-core-up/internal/config"

	"github.com/stretchr/testify/require"
)

func TestNewFromConfigFallsBackToLocalStorage(t *testing.T) {
	dir := t.TempDir()
	store := NewFromConfig(config.Config{
		PublicBaseURL:   "https://demo.example.com",
		LocalStorageDir: dir,
	})

	err := store.PutObject(context.Background(), "avatars/test/user.jpg", "image/jpeg", []byte("demo-image"))
	require.NoError(t, err)
	require.True(t, store.Available())
	require.Equal(t, "https://demo.example.com/media/avatars/test/user.jpg", store.PublicURL("avatars/test/user.jpg"))

	data, err := os.ReadFile(filepath.Join(dir, "avatars", "test", "user.jpg"))
	require.NoError(t, err)
	require.Equal(t, []byte("demo-image"), data)

	objectData, err := store.GetObject(context.Background(), "avatars/test/user.jpg")
	require.NoError(t, err)
	require.Equal(t, []byte("demo-image"), objectData)

	require.NoError(t, store.DeleteObject(context.Background(), "avatars/test/user.jpg"))
	_, err = os.Stat(filepath.Join(dir, "avatars", "test", "user.jpg"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}
