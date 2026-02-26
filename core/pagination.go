package core

// PageQuery holds pagination parameters parsed from query string.
type PageQuery struct {
	Page  int
	Limit int
}

// Page is the generic paginated response container.
type Page[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}

// NewPage constructs a Page from a slice of data and pagination parameters.
func NewPage[T any](data []T, total, page, limit int) Page[T] {
	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}
	return Page[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// ParsePagination parses ?page= and ?limit= from the query string.
// Defaults: page=1, limit=20. Maximum limit: 100.
func (c *Ctx) ParsePagination() PageQuery {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return PageQuery{Page: page, Limit: limit}
}
