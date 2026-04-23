package dto

type PublicContactRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Contact string `json:"contact"`
	Message string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type HealthStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
