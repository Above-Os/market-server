package models

type ListResult struct {
	Items      interface{} `json:"items"`
	TotalItems int         `json:"totalItems"`
	TotalCount int64       `json:"totalCount,omitempty"`
}

func NewListResult[T any](items []T) *ListResult {
	return &ListResult{Items: items, TotalItems: len(items)}
}
func NewListResultWithCount[T any](items []T, count int64) *ListResult {
	return &ListResult{Items: items, TotalItems: len(items), TotalCount: count}
}
