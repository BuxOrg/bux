package datastore

// QueryParams object to use when limiting and sorting database query results
type QueryParams struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	OrderByField  string `json:"order_by_field"`
	SortDirection string `json:"sort_direction"`
}
