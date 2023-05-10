package model

import "github.com/google/uuid"

type Status string

const (
	STATUS_NEW   Status = "new"
	STATUS_READY Status = "ready"
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
