package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type institutionRepoStub struct {
	tenantID    uuid.UUID
	localItems  []InstitutionSuggestion
	localErr    error
}

func (s *institutionRepoStub) FindTenantByCode(ctx context.Context, tenantCode string) (uuid.UUID, error) {
	if s.tenantID == uuid.Nil {
		return uuid.Nil, ErrNotFound
	}
	return s.tenantID, nil
}

func (s *institutionRepoStub) SuggestKnownInstitutions(ctx context.Context, tenantID uuid.UUID, educationType, query string, limit int) ([]InstitutionSuggestion, error) {
	if s.localErr != nil {
		return nil, s.localErr
	}
	return s.localItems, nil
}

type institutionFallbackStub struct {
	items []InstitutionSuggestion
	err   error
}

func (s *institutionFallbackStub) Suggest(ctx context.Context, in InstitutionSuggestRequest) ([]InstitutionSuggestion, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func TestInstitutionSuggesterMergesLocalBeforeFallback(t *testing.T) {
	repo := &institutionRepoStub{
		tenantID: uuid.New(),
		localItems: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderManual,
				ExternalID: "",
				Name:       "Евразийский национальный университет",
				Address:    "Астана",
			},
		},
	}
	fallback := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "r:1",
				Name:       "Евразийский национальный университет",
				Address:    "Астана",
			},
			{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "r:2",
				Name:       "Astana IT University",
				Address:    "Астана",
			},
		},
	}

	suggester := NewInstitutionSuggester(repo, fallback)
	items, err := suggester.Suggest(context.Background(), InstitutionSuggestRequest{
		TenantCode: "CORE",
		Query:      "евраз",
		Kind:       EducationTypeUniversity,
	})
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "Евразийский национальный университет", items[0].Name)
	require.Equal(t, InstitutionProviderManual, items[0].Provider)
	require.Equal(t, "Astana IT University", items[1].Name)
}

func TestInstitutionSuggesterReturnsLocalWhenFallbackFails(t *testing.T) {
	repo := &institutionRepoStub{
		tenantID: uuid.New(),
		localItems: []InstitutionSuggestion{
			{
				Provider: InstitutionProviderManual,
				Name:     "Школа-лицей №17",
				Address:  "Астана",
			},
		},
	}
	fallback := &institutionFallbackStub{err: errors.New("network down")}

	suggester := NewInstitutionSuggester(repo, fallback)
	items, err := suggester.Suggest(context.Background(), InstitutionSuggestRequest{
		TenantCode: "CORE",
		Query:      "лицей",
		Kind:       EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Школа-лицей №17", items[0].Name)
}
