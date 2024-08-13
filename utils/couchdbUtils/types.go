package couchdbUtils

import "github.com/couchbase/gocb/v2"

type Pagination struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	PageSize    int `json:"page_size"`
	TotalCount  int `json:"total_count"`
}

type ResponseListWithPagination[T any] struct {
	Data       T          `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Update request struct
type Update struct {
	ElementID  string           `json:"id"`
	Field      string           `json:"field"`
	NewData    interface{}      `json:"newData"`
	Collection *gocb.Collection `json:"collection"`
}
