// SPDX-License-Identifier: AGPL-3.0-or-later

package crawler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/asciimoo/hister/server/model"
)

type debugURLCounts struct {
	Pending    int64 `json:"pending"`
	InProgress int64 `json:"in_progress"`
	Done       int64 `json:"done"`
	Failed     int64 `json:"failed"`
	Skipped    int64 `json:"skipped"`
	Count2xx   int64 `json:"count_2xx"`
	Count3xx   int64 `json:"count_3xx"`
	Count4xx   int64 `json:"count_4xx"`
	Count5xx   int64 `json:"count_5xx"`
}

type debugJob struct {
	ID            string         `json:"id"`
	StartURL      string         `json:"start_url"`
	Label         string         `json:"label"`
	Status        string         `json:"status"`
	PagesFetched  int64          `json:"pages_fetched"`
	BytesFetched  int64          `json:"bytes_fetched"`
	Retries       int64          `json:"retries"`
	BreakerTrips  int64          `json:"breaker_trips"`
	RobotsDenials int64          `json:"robots_denials"`
	BudgetStops   int64          `json:"budget_stops"`
	URLCounts     debugURLCounts `json:"url_counts"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type debugResponse struct {
	Jobs []debugJob `json:"jobs"`
}

// DebugHandler returns an HTTP handler that emits a JSON object with all crawl
// jobs and their DB-persisted stats. It is intended to be mounted at /debug/crawler.
func DebugHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobs, err := model.ListCrawlJobs()
		if err != nil {
			log.Error().Err(err).Msg("debug/crawler: failed to list crawl jobs")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		out := make([]debugJob, 0, len(jobs))
		for _, job := range jobs {
			stats, err := model.GetCrawlJobStats(job.ID)
			if err != nil {
				log.Warn().Err(err).Str("job_id", job.ID).Msg("debug/crawler: failed to get stats")
				continue
			}
			out = append(out, debugJob{
				ID:            job.ID,
				StartURL:      job.StartURL,
				Label:         job.Label,
				Status:        job.Status,
				PagesFetched:  job.PagesFetched,
				BytesFetched:  job.BytesFetched,
				Retries:       job.Retries,
				BreakerTrips:  job.BreakerTrips,
				RobotsDenials: job.RobotsDenials,
				BudgetStops:   job.BudgetStops,
				URLCounts: debugURLCounts{
					Pending:    stats.Pending,
					InProgress: stats.InProgress,
					Done:       stats.Done,
					Failed:     stats.Failed,
					Skipped:    stats.Skipped,
					Count2xx:   stats.Count2xx,
					Count3xx:   stats.Count3xx,
					Count4xx:   stats.Count4xx,
					Count5xx:   stats.Count5xx,
				},
				CreatedAt: job.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: job.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(debugResponse{Jobs: out})
	}
}
