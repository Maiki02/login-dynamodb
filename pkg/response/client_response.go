package response

// PaginatedResponse es una estructura genérica para respuestas paginadas.
type PaginatedResponse struct {
	Docs        interface{} `json:"docs"`
	TotalDocs   int64       `json:"totalDocs"`
	Limit       int64       `json:"limit"`
	TotalPages  int64       `json:"totalPages"`
	Page        int64       `json:"page"`
	HasNextPage bool        `json:"hasNextPage"`
	HasPrevPage bool        `json:"hasPrevPage"`
}
