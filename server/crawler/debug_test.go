// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asciimoo/hister/server/model"
	"github.com/asciimoo/hister/server/testutil"
)

func TestDebugHandlerJSON(t *testing.T) {
	testutil.InitModel(t)

	_, err := model.CreateNamedCrawlJobWithURLs(
		"debug-job", "https://example.com/", `{}`, "test crawl",
		[]string{"https://example.com/a", "https://example.com/b"},
	)
	if err != nil {
		t.Fatalf("CreateNamedCrawlJobWithURLs() error: %v", err)
	}

	handler := DebugHandler()
	req := httptest.NewRequest(http.MethodGet, "/debug/crawler", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Jobs []struct {
			ID           string `json:"id"`
			StartURL     string `json:"start_url"`
			Label        string `json:"label"`
			Status       string `json:"status"`
			PagesFetched int64  `json:"pages_fetched"`
			URLCounts    struct {
				Pending int64 `json:"pending"`
				Done    int64 `json:"done"`
			} `json:"url_counts"`
		} `json:"jobs"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Jobs) != 1 {
		t.Fatalf("jobs count = %d, want 1", len(resp.Jobs))
	}
	job := resp.Jobs[0]
	if job.ID != "debug-job" {
		t.Fatalf("job.ID = %q, want %q", job.ID, "debug-job")
	}
	if job.Label != "test crawl" {
		t.Fatalf("job.Label = %q, want %q", job.Label, "test crawl")
	}
	if job.Status != model.CrawlJobRunning {
		t.Fatalf("job.Status = %q, want %q", job.Status, model.CrawlJobRunning)
	}
	if job.PagesFetched != 0 {
		t.Fatalf("job.PagesFetched = %d, want 0", job.PagesFetched)
	}
	if job.URLCounts.Pending != 2 {
		t.Fatalf("url_counts.pending = %d, want 2", job.URLCounts.Pending)
	}
}

func TestDebugHandlerEmptyDB(t *testing.T) {
	testutil.InitModel(t)

	handler := DebugHandler()
	req := httptest.NewRequest(http.MethodGet, "/debug/crawler", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Jobs []any `json:"jobs"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Jobs) != 0 {
		t.Fatalf("jobs count = %d, want 0", len(resp.Jobs))
	}
}
