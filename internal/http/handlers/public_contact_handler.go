package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode"

	"idsai-core-up/internal/http/dto"

	"github.com/gin-gonic/gin"
)

type contactMessageSender interface {
	SendText(ctx context.Context, text string) error
}

type PublicContactHandler struct {
	sender     contactMessageSender
	serverName string
}

func NewPublicContactHandler(sender contactMessageSender, serverName string) *PublicContactHandler {
	name := strings.TrimSpace(serverName)
	if name == "" {
		name = "idsai"
	}

	return &PublicContactHandler{
		sender:     sender,
		serverName: name,
	}
}

func (h *PublicContactHandler) Submit(c *gin.Context) {
	var req dto.PublicContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	contact := compactContactLine(req.Contact)
	message := compactContactMessage(req.Message)
	if contact != "" || message != "" {
		switch {
		case len(contact) < 3:
			c.JSON(http.StatusBadRequest, gin.H{"error": "contact is required"})
			return
		case len(message) < 5:
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
			return
		case h.sender == nil:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "contact delivery unavailable"})
			return
		}

		text := fmt.Sprintf(
			"📨 Author contact request\nService: %s\nContact: %s\nMessage: %s\nTime: %s",
			h.serverName,
			contact,
			message,
			time.Now().Format(time.RFC3339),
		)

		if err := h.sender.SendText(c.Request.Context(), text); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to deliver message"})
			return
		}

		c.JSON(http.StatusOK, dto.StatusResponse{Status: "sent"})
		return
	}

	name := compactContactLine(req.Name)
	phone := compactContactLine(req.Phone)
	email := compactContactLine(req.Email)

	switch {
	case len(name) < 2:
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	case !validContactPhone(phone):
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone is invalid"})
		return
	case !validContactEmail(email):
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is invalid"})
		return
	case h.sender == nil:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "contact delivery unavailable"})
		return
	}

	text := fmt.Sprintf(
		"📨 Author contact request\nService: %s\nName: %s\nPhone: %s\nEmail: %s\nTime: %s",
		h.serverName,
		name,
		phone,
		email,
		time.Now().Format(time.RFC3339),
	)

	if err := h.sender.SendText(c.Request.Context(), text); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to deliver message"})
		return
	}

	c.JSON(http.StatusOK, dto.StatusResponse{Status: "sent"})
}

func compactContactLine(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	return strings.Join(strings.Fields(v), " ")
}

func compactContactMessage(v string) string {
	v = strings.ReplaceAll(v, "\r\n", "\n")
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	lines := strings.Split(v, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}

	return strings.Join(out, "\n")
}

func validContactEmail(v string) bool {
	if v == "" {
		return false
	}
	_, err := mail.ParseAddress(v)
	return err == nil
}

func validContactPhone(v string) bool {
	if v == "" {
		return false
	}

	digits := 0
	for _, r := range v {
		switch {
		case unicode.IsDigit(r):
			digits++
		case r == '+' || r == '-' || r == '(' || r == ')' || unicode.IsSpace(r):
			continue
		default:
			return false
		}
	}

	return digits >= 7
}
