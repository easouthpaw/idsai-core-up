package storage

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type localStorage struct {
	baseDir       string
	publicBaseURL string
}

func newLocalStorage(baseDir, publicBaseURL string) ObjectStorage {
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		return nopStorage{}
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nopStorage{}
	}
	return &localStorage{
		baseDir:       baseDir,
		publicBaseURL: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
	}
}

func (s *localStorage) PutObject(ctx context.Context, key, contentType string, body []byte) error {
	_ = ctx
	_ = contentType

	resolvedPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(resolvedPath), ".upload-*")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(body); err != nil {
		_ = tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, resolvedPath)
}

func (s *localStorage) GetObject(ctx context.Context, key string) ([]byte, error) {
	_ = ctx

	resolvedPath, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(resolvedPath)
}

func (s *localStorage) DeleteObject(ctx context.Context, key string) error {
	_ = ctx

	resolvedPath, err := s.resolvePath(key)
	if err != nil {
		return nil
	}
	if err := os.Remove(resolvedPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *localStorage) PublicURL(key string) string {
	key, err := normalizeObjectKey(key)
	if err != nil {
		return ""
	}
	if s.publicBaseURL == "" {
		return "/media/" + key
	}
	return s.publicBaseURL + "/media/" + key
}

func (s *localStorage) Available() bool {
	return true
}

func (s *localStorage) resolvePath(key string) (string, error) {
	key, err := normalizeObjectKey(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, filepath.FromSlash(key)), nil
}

func normalizeObjectKey(key string) (string, error) {
	key = strings.ReplaceAll(strings.TrimSpace(key), "\\", "/")
	key = strings.TrimPrefix(path.Clean("/"+key), "/")
	if key == "" || key == "." {
		return "", fmt.Errorf("empty object key")
	}
	return key, nil
}
