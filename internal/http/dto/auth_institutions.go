package dto

import "idsai-core-up/internal/services/auth"

type SuggestInstitutionsResponse struct {
	Items []InstitutionSuggestionResponse `json:"items"`
}

type InstitutionSuggestionResponse struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	Name       string `json:"name"`
	Address    string `json:"address,omitempty"`
}

func InstitutionSuggestionResponsesFromService(items []auth.InstitutionSuggestion) []InstitutionSuggestionResponse {
	if items == nil {
		return nil
	}
	out := make([]InstitutionSuggestionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, InstitutionSuggestionResponse{
			Provider:   item.Provider,
			ExternalID: item.ExternalID,
			Name:       item.Name,
			Address:    item.Address,
		})
	}
	return out
}
