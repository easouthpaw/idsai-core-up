package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeContactSender struct {
	lastText string
	err      error
}

func (f *fakeContactSender) SendText(_ context.Context, text string) error {
	f.lastText = text
	return f.err
}

func TestPublicContactHandlerSubmit_SendsTelegramMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sender := &fakeContactSender{}
	h := NewPublicContactHandler(sender, "idsai-test")
	r := gin.New()
	r.POST("/v2/contact", h.Submit)

	body, err := json.Marshal(map[string]string{
		"name":  "Aibolat Yermekbay",
		"phone": "+7 777 123 45 67",
		"email": "aibolat@example.com",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/contact", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, sender.lastText, "Author contact request")
	require.Contains(t, sender.lastText, "Aibolat Yermekbay")
	require.Contains(t, sender.lastText, "+7 777 123 45 67")
	require.Contains(t, sender.lastText, "aibolat@example.com")
}

func TestPublicContactHandlerSubmit_SendsTelegramMessageForUnifiedForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sender := &fakeContactSender{}
	h := NewPublicContactHandler(sender, "idsai-test")
	r := gin.New()
	r.POST("/v2/contact", h.Submit)

	body, err := json.Marshal(map[string]string{
		"contact": "Айболат, +7 777 123 45 67, @easouthpaw",
		"message": "Нужна консультация по внедрению платформы\nи настройке ролей.",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/contact", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, sender.lastText, "Author contact request")
	require.Contains(t, sender.lastText, "Contact: Айболат, +7 777 123 45 67, @easouthpaw")
	require.Contains(t, sender.lastText, "Message: Нужна консультация по внедрению платформы\nи настройке ролей.")
}

func TestPublicContactHandlerSubmit_RejectsInvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPublicContactHandler(&fakeContactSender{}, "idsai-test")
	r := gin.New()
	r.POST("/v2/contact", h.Submit)

	body, err := json.Marshal(map[string]string{
		"name":  "Aibolat",
		"phone": "+7 777 123 45 67",
		"email": "invalid",
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/contact", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, strings.ToLower(w.Body.String()), "email is invalid")
}
