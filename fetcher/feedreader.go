package fetcher

import "ewintr.nl/yogai/model"

type FeedEntry struct {
	EntryID          int64
	FeedID           int64
	YoutubeChannelID string
	YoutubeID        string
}

type ChannelReader interface {
	Search(channelID model.YoutubeChannelID, pageToken string) ([]model.YoutubeVideoID, string, error)
}

type FeedReader interface {
	Unread() ([]FeedEntry, error)
	MarkRead(feedID int64) error
}
