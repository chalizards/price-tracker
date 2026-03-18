package queue

const ScrapeJobsQueue = "scrape_jobs"

type ScrapeJobMessage struct {
	ProductID int `json:"product_id"`
}
