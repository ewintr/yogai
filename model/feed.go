package model

import "github.com/google/uuid"

type FeedStatus string

const (
	FeedStatusNew   FeedStatus = "new"
	FeedStatusReady FeedStatus = "ready"
)

type Feed struct {
	ID               uuid.UUID
	Status           FeedStatus
	Title            string
	YoutubeChannelID YoutubeChannelID
}
