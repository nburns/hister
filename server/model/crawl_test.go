package model_test

import (
	"slices"
	"testing"

	"github.com/asciimoo/hister/server/model"
	"github.com/asciimoo/hister/server/testutil"
)

func TestCreateNamedCrawlJobWithURLs(t *testing.T) {
	testutil.InitModel(t)
	urls := []string{
		"https://example.com/one",
		"https://example.com/two",
		"https://example.com/one",
	}

	jobID, err := model.CreateNamedCrawlJobWithURLs(
		"urls.txt", urls[0], `{"NoDepth":true}`, "reference", urls,
	)
	if err != nil {
		t.Fatalf("CreateNamedCrawlJobWithURLs() error: %v", err)
	}
	if jobID != "urls.txt" {
		t.Fatalf("job ID = %q, want %q", jobID, "urls.txt")
	}

	job, err := model.GetCrawlJob(jobID)
	if err != nil {
		t.Fatalf("GetCrawlJob() error: %v", err)
	}
	if job == nil {
		t.Fatal("GetCrawlJob() returned nil")
	}
	if job.StartURL != urls[0] {
		t.Fatalf("start URL = %q, want %q", job.StartURL, urls[0])
	}
	if job.Label != "reference" {
		t.Fatalf("label = %q, want %q", job.Label, "reference")
	}

	var queued []string
	if err := model.ForEachCrawlURL(jobID, func(_ string, _ int, rawURL string) error {
		queued = append(queued, rawURL)
		return nil
	}); err != nil {
		t.Fatalf("ForEachCrawlURL() error: %v", err)
	}
	wantQueued := []string{urls[0], urls[1]}
	if !slices.Equal(queued, wantQueued) {
		t.Fatalf("queued URLs = %q, want %q", queued, wantQueued)
	}

	secondJobID, err := model.CreateNamedCrawlJobWithURLs(
		"urls.txt", urls[0], `{"NoDepth":true}`, "", urls[:1],
	)
	if err != nil {
		t.Fatalf("second CreateNamedCrawlJobWithURLs() error: %v", err)
	}
	if secondJobID != "urls.txt-2" {
		t.Fatalf("second job ID = %q, want %q", secondJobID, "urls.txt-2")
	}
}

func TestCreateNamedCrawlJobWithURLsRejectsEmptyQueue(t *testing.T) {
	testutil.InitModel(t)

	if _, err := model.CreateNamedCrawlJobWithURLs("urls.txt", "", `{}`, "", nil); err == nil {
		t.Fatal("CreateNamedCrawlJobWithURLs() expected an error")
	}
	jobs, err := model.ListCrawlJobs()
	if err != nil {
		t.Fatalf("ListCrawlJobs() error: %v", err)
	}
	if len(jobs) != 0 {
		t.Fatalf("job count = %d, want 0", len(jobs))
	}
}

func TestSetURLResult(t *testing.T) {
	testutil.InitModel(t)

	jobID, err := model.CreateNamedCrawlJobWithURLs(
		"set-result-job", "https://example.com/", `{}`, "test", []string{"https://example.com/"},
	)
	if err != nil {
		t.Fatalf("CreateNamedCrawlJobWithURLs() error: %v", err)
	}

	cur, err := model.NextPendingCrawlURL(jobID)
	if err != nil || cur == nil {
		t.Fatalf("NextPendingCrawlURL() error: %v, cur: %v", err, cur)
	}

	if err := model.SetURLResult(cur.ID, model.CrawlURLDone, 200, 1234, ""); err != nil {
		t.Fatalf("SetURLResult() error: %v", err)
	}

	stats, err := model.GetCrawlJobStats(jobID)
	if err != nil {
		t.Fatalf("GetCrawlJobStats() error: %v", err)
	}
	if stats.Done != 1 {
		t.Fatalf("Done = %d, want 1", stats.Done)
	}
	if stats.Count2xx != 1 {
		t.Fatalf("Count2xx = %d, want 1", stats.Count2xx)
	}
}

func TestCrawlJobIncrements(t *testing.T) {
	testutil.InitModel(t)

	jobID, err := model.CreateNamedCrawlJobWithURLs(
		"inc-job", "https://example.com/", `{}`, "test", []string{"https://example.com/"},
	)
	if err != nil {
		t.Fatalf("CreateNamedCrawlJobWithURLs() error: %v", err)
	}

	if err := model.IncrementCrawlJobPages(jobID, 500); err != nil {
		t.Fatalf("IncrementCrawlJobPages() error: %v", err)
	}
	if err := model.IncrementCrawlJobPages(jobID, 250); err != nil {
		t.Fatalf("IncrementCrawlJobPages() error: %v", err)
	}
	if err := model.IncrementCrawlJobRetries(jobID, 2); err != nil {
		t.Fatalf("IncrementCrawlJobRetries() error: %v", err)
	}
	if err := model.IncrementCrawlJobBreakerTrips(jobID, 1); err != nil {
		t.Fatalf("IncrementCrawlJobBreakerTrips() error: %v", err)
	}
	if err := model.IncrementCrawlJobRobotsDenials(jobID, 3); err != nil {
		t.Fatalf("IncrementCrawlJobRobotsDenials() error: %v", err)
	}
	if err := model.IncrementCrawlJobBudgetStops(jobID, 1); err != nil {
		t.Fatalf("IncrementCrawlJobBudgetStops() error: %v", err)
	}

	job, err := model.GetCrawlJob(jobID)
	if err != nil || job == nil {
		t.Fatalf("GetCrawlJob() error: %v", err)
	}
	if job.PagesFetched != 2 {
		t.Fatalf("PagesFetched = %d, want 2", job.PagesFetched)
	}
	if job.BytesFetched != 750 {
		t.Fatalf("BytesFetched = %d, want 750", job.BytesFetched)
	}
	if job.Retries != 2 {
		t.Fatalf("Retries = %d, want 2", job.Retries)
	}
	if job.BreakerTrips != 1 {
		t.Fatalf("BreakerTrips = %d, want 1", job.BreakerTrips)
	}
	if job.RobotsDenials != 3 {
		t.Fatalf("RobotsDenials = %d, want 3", job.RobotsDenials)
	}
	if job.BudgetStops != 1 {
		t.Fatalf("BudgetStops = %d, want 1", job.BudgetStops)
	}
}

func TestGetCrawlJobStatsHTTPBreakdown(t *testing.T) {
	testutil.InitModel(t)

	jobID, err := model.CreateNamedCrawlJobWithURLs(
		"http-stats-job", "https://example.com/", `{}`, "test", []string{
			"https://example.com/a",
			"https://example.com/b",
			"https://example.com/c",
			"https://example.com/d",
		},
	)
	if err != nil {
		t.Fatalf("CreateNamedCrawlJobWithURLs() error: %v", err)
	}

	// Set different HTTP statuses on each URL.
	codes := []int{200, 301, 404, 500}
	for _, code := range codes {
		cur, err := model.NextPendingCrawlURL(jobID)
		if err != nil || cur == nil {
			t.Fatalf("NextPendingCrawlURL() error: %v cur: %v", err, cur)
		}
		if err := model.SetURLResult(cur.ID, model.CrawlURLDone, code, 100, ""); err != nil {
			t.Fatalf("SetURLResult(%d) error: %v", code, err)
		}
	}

	stats, err := model.GetCrawlJobStats(jobID)
	if err != nil {
		t.Fatalf("GetCrawlJobStats() error: %v", err)
	}
	if stats.Count2xx != 1 {
		t.Fatalf("Count2xx = %d, want 1", stats.Count2xx)
	}
	if stats.Count3xx != 1 {
		t.Fatalf("Count3xx = %d, want 1", stats.Count3xx)
	}
	if stats.Count4xx != 1 {
		t.Fatalf("Count4xx = %d, want 1", stats.Count4xx)
	}
	if stats.Count5xx != 1 {
		t.Fatalf("Count5xx = %d, want 1", stats.Count5xx)
	}
}
