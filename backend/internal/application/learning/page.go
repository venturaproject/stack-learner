package learning

import "strings"

type PageQuery struct {
	Page    int
	PerPage int
	Search  string
}

type PageResult[T any] struct {
	Items      []T            `json:"-"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	Total      int64          `json:"total"`
	TotalPages int            `json:"total_pages"`
	Summary    *CourseSummary `json:"summary,omitempty"`
}

type CourseSummary struct {
	CourseCount    int64 `json:"course_count"`
	PublishedCount int64 `json:"published_count"`
	ModuleCount    int64 `json:"module_count"`
}

func normalizePage(q PageQuery) PageQuery {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 25
	}
	maxInt := int(^uint(0) >> 1)
	if q.Page > maxInt/q.PerPage {
		q.Page = maxInt / q.PerPage
	}
	q.Search = strings.TrimSpace(q.Search)
	return q
}

func pageResult[T any](items []T, q PageQuery, total int64) PageResult[T] {
	q = normalizePage(q)
	pages := int((total + int64(q.PerPage) - 1) / int64(q.PerPage))
	if pages < 1 {
		pages = 1
	}
	return PageResult[T]{Items: items, Page: q.Page, PerPage: q.PerPage, Total: total, TotalPages: pages}
}
