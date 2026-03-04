package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPage(t *testing.T) {
	page := NewPage([]int{1, 2, 3}, 7, 2, 3)
	if page.Total != 7 {
		t.Fatalf("Total = %d, want 7", page.Total)
	}
	if page.Page != 2 {
		t.Fatalf("Page = %d, want 2", page.Page)
	}
	if page.Limit != 3 {
		t.Fatalf("Limit = %d, want 3", page.Limit)
	}
	if page.TotalPages != 3 {
		t.Fatalf("TotalPages = %d, want 3", page.TotalPages)
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{name: "defaults", query: "", wantPage: 1, wantLimit: 20},
		{name: "custom values", query: "?page=3&limit=50", wantPage: 3, wantLimit: 50},
		{name: "page lower bound", query: "?page=0&limit=10", wantPage: 1, wantLimit: 10},
		{name: "limit lower bound", query: "?page=1&limit=-2", wantPage: 1, wantLimit: 20},
		{name: "limit upper bound", query: "?page=1&limit=500", wantPage: 1, wantLimit: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got PageQuery
			app := newHTTPXTestApp("GET", "/page", func(c *Ctx) error {
				got = c.ParsePagination()
				return c.NoContent()
			})
			resp, err := app.Test(httptest.NewRequest("GET", "/page"+tt.query, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}
			if got.Page != tt.wantPage || got.Limit != tt.wantLimit {
				t.Fatalf("got (page=%d,limit=%d), want (page=%d,limit=%d)", got.Page, got.Limit, tt.wantPage, tt.wantLimit)
			}
		})
	}
}
