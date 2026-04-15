package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/infra/photon"
	"idsai-core-up/internal/services/auth"

	"github.com/gin-gonic/gin"
)

type InstitutionSuggester interface {
	Suggest(ctx context.Context, in auth.InstitutionSuggestRequest) ([]auth.InstitutionSuggestion, error)
}

func (h *AuthHandler) SuggestInstitutions(c *gin.Context) {
	authResponseNoStore(c)

	if h.institutions == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "institution suggestions are unavailable"})
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, dto.SuggestInstitutionsResponse{Items: []dto.InstitutionSuggestionResponse{}})
		return
	}
	if len([]rune(q)) < 2 {
		c.JSON(http.StatusOK, dto.SuggestInstitutionsResponse{Items: []dto.InstitutionSuggestionResponse{}})
		return
	}

	req := auth.InstitutionSuggestRequest{
		TenantCode: tenantCodeFromHeader(c),
		Query:      q,
		Kind:       strings.ToUpper(strings.TrimSpace(c.Query("kind"))),
	}
	if lat, ok := parseOptionalFloat(c.Query("lat")); ok {
		req.Lat = &lat
	}
	if lon, ok := parseOptionalFloat(c.Query("lon")); ok {
		req.Lon = &lon
	}

	items, err := h.institutions.Suggest(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		case errors.Is(err, photon.ErrUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Photon autocomplete is unavailable"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "institution suggestions request failed"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuggestInstitutionsResponse{Items: dto.InstitutionSuggestionResponsesFromService(items)})
}

func parseOptionalFloat(raw string) (float64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, false
	}
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return number, true
}
