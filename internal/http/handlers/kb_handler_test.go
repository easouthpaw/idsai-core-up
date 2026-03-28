package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKBHandlerRequireEditor_AllowsProfessorAndAdmin(t *testing.T) {
	t.Run("professor", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("isProfessor", true)

		h := &KBHandler{}
		require.True(t, h.requireEditor(c))
	})

	t.Run("admin", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("isAdmin", true)

		h := &KBHandler{}
		require.True(t, h.requireEditor(c))
	})
}

func TestKBHandlerContextIdentityUsesMiddlewareKeys(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	tenantID := uuid.New()
	userID := uuid.New()
	c.Set("tenantID", tenantID)
	c.Set("userID", userID)

	h := &KBHandler{}
	require.Equal(t, tenantID, h.tenantID(c))
	require.Equal(t, userID, h.userID(c))
}

