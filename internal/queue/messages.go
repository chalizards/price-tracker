package queue

const ScrapeJobsQueue = "scrape_jobs"

type ScrapeJobMessage struct {
	StoreID int `json:"store_id"`
}
