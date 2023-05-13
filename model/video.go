package model

import "github.com/google/uuid"

type Status string

const (
	StatusNew         Status = "new"
	StatusHasMetadata Status = "has_metadata"
	StatusHasSummary  Status = "has_summary"
	StatusReady       Status = "ready"
)

type Video struct {
	ID          uuid.UUID
	Status      Status
	YoutubeID   string
	FeedID      uuid.UUID
	Title       string
	Description string
	Summary     string
}
