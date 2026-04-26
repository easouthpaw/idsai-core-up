package auth

import "context"

type InstitutionAutocompleteProvider struct {
	schoolCatalog    InstitutionSuggestionProvider
	universityCatalog InstitutionSuggestionProvider
	fallback         InstitutionSuggestionProvider
}

func NewInstitutionAutocompleteProvider(schoolCatalog, universityCatalog, fallback InstitutionSuggestionProvider) *InstitutionAutocompleteProvider {
	return &InstitutionAutocompleteProvider{
		schoolCatalog:    schoolCatalog,
		universityCatalog: universityCatalog,
		fallback:         fallback,
	}
}

func (p *InstitutionAutocompleteProvider) Suggest(ctx context.Context, in InstitutionSuggestRequest) ([]InstitutionSuggestion, error) {
	kind := normalizeEducationType(in.Kind)

	var catalogErr error

	if kind == EducationTypeSchool && p.schoolCatalog != nil {
		items, err := p.schoolCatalog.Suggest(ctx, in)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		catalogErr = err
	}

	if kind == EducationTypeUniversity && p.universityCatalog != nil {
		items, err := p.universityCatalog.Suggest(ctx, in)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		if err != nil && catalogErr == nil {
			catalogErr = err
		}
	}

	if p.fallback != nil {
		items, err := p.fallback.Suggest(ctx, in)
		if err == nil {
			return items, nil
		}
		if catalogErr == nil {
			return nil, err
		}
	}

	if catalogErr != nil {
		return nil, catalogErr
	}
	return nil, nil
}
