package auth

import "strings"

const (
	InstitutionProviderManual    = "manual"
	InstitutionProviderPhoton    = "photon"
	InstitutionProviderKZSchools = "kz_schools"
)

type InstitutionSelection struct {
	Provider   string
	ExternalID string
	Name       string
	Address    string
}

type InstitutionSuggestRequest struct {
	TenantCode string
	Query      string
	Kind       string
	Lat        *float64
	Lon        *float64
}

type InstitutionSuggestion struct {
	Provider   string
	ExternalID string
	Name       string
	Address    string
}

func normalizeInstitutionSelection(in InstitutionSelection) InstitutionSelection {
	out := InstitutionSelection{
		Provider:   strings.ToLower(strings.TrimSpace(in.Provider)),
		ExternalID: strings.TrimSpace(in.ExternalID),
		Name:       strings.TrimSpace(in.Name),
		Address:    strings.TrimSpace(in.Address),
	}
	if out.Name == "" {
		return InstitutionSelection{}
	}
	if out.Provider == "" {
		out.Provider = InstitutionProviderManual
	}
	return out
}
