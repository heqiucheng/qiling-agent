package service

import "net/http"

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 100
)

type PageRequest struct {
	Page     int
	PageSize int
}

type PageResult[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

func PageRequestFromQuery(r *http.Request) PageRequest {
	query := r.URL.Query()
	page := parsePositiveInt(query.Get("page"), defaultPage)
	pageSize := parsePositiveInt(query.Get("page_size"), defaultPageSize)
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return PageRequest{Page: page, PageSize: pageSize}
}

func NewPageResult[T any](items []T, req PageRequest) PageResult[T] {
	total := len(items)
	start := (req.Page - 1) * req.PageSize
	if start >= total {
		return PageResult[T]{Items: []T{}, Page: req.Page, PageSize: req.PageSize, Total: total}
	}

	end := start + req.PageSize
	if end > total {
		end = total
	}

	return PageResult[T]{Items: items[start:end], Page: req.Page, PageSize: req.PageSize, Total: total}
}

func NewPageResultWithTotal[T any](items []T, req PageRequest, total int) PageResult[T] {
	if items == nil {
		items = []T{}
	}
	return PageResult[T]{Items: items, Page: req.Page, PageSize: req.PageSize, Total: total}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}

	value := 0
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			return fallback
		}
		value = value*10 + int(ch-'0')
	}
	if value <= 0 {
		return fallback
	}
	return value
}
