package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()

	data, err := json.Marshal(v)
	require.NoError(t, err)
	return string(data)
}
