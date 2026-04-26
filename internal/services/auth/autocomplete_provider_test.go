package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstitutionAutocompleteProviderPrefersSchoolCatalog(t *testing.T) {
	catalog := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderKZSchools,
				ExternalID: "egov:1",
				Name:       "Школа-лицей №17",
				Address:    "Астана",
			},
		},
	}
	fallback := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "photon:1",
				Name:       "Школа-лицей №17",
				Address:    "Астана",
			},
		},
	}

	provider := NewInstitutionAutocompleteProvider(catalog, nil, fallback)
	items, err := provider.Suggest(context.Background(), InstitutionSuggestRequest{
		Query: "лицей 17",
		Kind:  EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, InstitutionProviderKZSchools, items[0].Provider)
}

func TestInstitutionAutocompleteProviderFallsBackWhenSchoolCatalogHasNoMatches(t *testing.T) {
	catalog := &institutionFallbackStub{}
	fallback := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "photon:1",
				Name:       "Школа-лицей №17",
				Address:    "Астана",
			},
		},
	}

	provider := NewInstitutionAutocompleteProvider(catalog, nil, fallback)
	items, err := provider.Suggest(context.Background(), InstitutionSuggestRequest{
		Query: "лицей 17",
		Kind:  EducationTypeSchool,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, InstitutionProviderPhoton, items[0].Provider)
}

func TestInstitutionAutocompleteProviderIgnoresSchoolCatalogForUniversities(t *testing.T) {
	catalog := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider: InstitutionProviderKZSchools,
				Name:     "Школа-лицей №17",
			},
		},
	}
	fallback := &institutionFallbackStub{
		items: []InstitutionSuggestion{
			{
				Provider:   InstitutionProviderPhoton,
				ExternalID: "photon:2",
				Name:       "Astana IT University",
				Address:    "Астана",
			},
		},
	}

	provider := NewInstitutionAutocompleteProvider(catalog, nil, fallback)
	items, err := provider.Suggest(context.Background(), InstitutionSuggestRequest{
		Query: "astana",
		Kind:  EducationTypeUniversity,
	})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "Astana IT University", items[0].Name)
	require.Equal(t, InstitutionProviderPhoton, items[0].Provider)
}
