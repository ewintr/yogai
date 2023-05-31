package model

import "github.com/google/uuid"

type VideoStatus string

const (
	StatusNew         VideoStatus = "new"
	StatusHasMetadata VideoStatus = "has_metadata"
	StatusHasSummary  VideoStatus = "has_summary"
	StatusReady       VideoStatus = "ready"
)

type YoutubeVideoID string

type YoutubeChannelID string

type Video struct {
	ID                 uuid.UUID
	Status             VideoStatus
	YoutubeID          YoutubeVideoID
	YoutubeChannelID   YoutubeChannelID
	YoutubeTitle       string
	YoutubeDescription string
	YoutubeDuration    string
	YoutubePublishedAt string

	Summary string
}
