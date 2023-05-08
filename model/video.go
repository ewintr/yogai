package model

import "github.com/google/uuid"

type Status string

const (
	STATUS_NEW           Status = "new"
	STATUS_NEEDS_SUMMARY Status = "needs_summary"
	STATUS_READY         Status = "ready"
)

type Video struct {
	ID          uuid.UUID
	Status      Status
	YoutubeURL  string
	FeedID      string
	Title       string
	Description string
	Summary     string
}
