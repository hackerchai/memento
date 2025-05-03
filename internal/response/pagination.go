package response

// PaginationRequest defines common pagination query parameters.
type PaginationRequest struct {
	Page    int `query:"page" validate:"omitempty,min=1"`             // Page number (default: 1)
	PerPage int `query:"per_page" validate:"omitempty,min=1,max=100"` // Items per page (default: 10, max: 100)
}

// GetOffset calculates the offset based on page and per_page.
func (p *PaginationRequest) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.GetLimit()
}

// GetLimit returns the number of items per page, with default.
func (p *PaginationRequest) GetLimit() int {
	if p.PerPage < 1 {
		p.PerPage = 10 // Default limit
	}
	if p.PerPage > 100 {
		p.PerPage = 100 // Max limit
	}
	return p.PerPage
}

// PaginationResponse defines the structure for paginated responses.
type PaginationResponse struct {
	Total int         `json:"total"` // Total number of items
	Page  int         `json:"page"`  // Current page number
	Data  interface{} `json:"data"`  // Data for the current page
}
