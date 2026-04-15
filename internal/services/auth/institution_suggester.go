package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type InstitutionSuggestionRepository interface {
	FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error)
	SuggestKnownInstitutions(ctx context.Context, tenantID uuid.UUID, educationType, query string, limit int) ([]InstitutionSuggestion, error)
}

type InstitutionSuggestionProvider interface {
	Suggest(ctx context.Context, in InstitutionSuggestRequest) ([]InstitutionSuggestion, error)
}

type InstitutionSuggester struct {
	repo        InstitutionSuggestionRepository
	fallback    InstitutionSuggestionProvider
	localLimit  int
	resultLimit int
}

func NewInstitutionSuggester(repo InstitutionSuggestionRepository, fallback InstitutionSuggestionProvider) *InstitutionSuggester {
	return &InstitutionSuggester{
		repo:        repo,
		fallback:    fallback,
		localLimit:  5,
		resultLimit: 8,
	}
}

func (s *InstitutionSuggester) Suggest(ctx context.Context, in InstitutionSuggestRequest) ([]InstitutionSuggestion, error) {
	query := strings.TrimSpace(in.Query)
	kind := normalizeEducationType(in.Kind)
	if query == "" || kind == "" {
		return nil, ErrInvalidInput
	}

	items := make([]InstitutionSuggestion, 0, s.resultLimit)
	seen := make(map[string]struct{}, s.resultLimit)

	if s.repo != nil {
		tenantCode := normalizeTenantCode(in.TenantCode)
		if tenantCode != "" {
			if tenantID, err := s.repo.FindTenantByCode(ctx, tenantCode); err == nil {
				local, err := s.repo.SuggestKnownInstitutions(ctx, tenantID, kind, query, s.localLimit)
				if err == nil {
					items = appendInstitutionSuggestions(items, seen, local, s.resultLimit)
				}
			}
		}
	}

	if len(items) >= s.resultLimit || s.fallback == nil {
		return items, nil
	}

	remote, err := s.fallback.Suggest(ctx, in)
	if err != nil {
		return items, nil
	}

	items = appendInstitutionSuggestions(items, seen, remote, s.resultLimit)
	return items, nil
}

func appendInstitutionSuggestions(dst []InstitutionSuggestion, seen map[string]struct{}, src []InstitutionSuggestion, limit int) []InstitutionSuggestion {
	for _, item := range src {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name + "|" + strings.TrimSpace(item.Address))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		dst = append(dst, item)
		if len(dst) >= limit {
			break
		}
	}
	return dst
}
