package learning

import "testing"

func TestPageResultNormalizesInputsAndCalculatesTotalPages(t *testing.T) {
	query := normalizePage(PageQuery{Page: 0, PerPage: 1000, Search: "  course  "})
	if query.Page != 1 || query.PerPage != 25 || query.Search != "course" {
		t.Fatalf("normalized query = %+v", query)
	}
	result := pageResult([]string{"item"}, PageQuery{Page: 2, PerPage: 10}, 21)
	if result.Page != 2 || result.PerPage != 10 || result.Total != 21 || result.TotalPages != 3 {
		t.Fatalf("page result = %+v", result)
	}
}
