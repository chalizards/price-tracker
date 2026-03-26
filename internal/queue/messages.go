package queue

const ScrapeJobsQueue = "scrape_jobs"

type ScrapeJobMessage struct {
	OfferID int `json:"offer_id"`
}
