package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpx "idsai-core-up/internal/http"

	"github.com/stretchr/testify/require"
)

func TestSwaggerRoute_Available(t *testing.T) {
	r := httpx.NewRouter(nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}
