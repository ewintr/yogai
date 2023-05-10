package fetcher

type FeedEntry struct {
	EntryID   int64
	FeedID    int64
	YouTubeID string
}

type FeedReader interface {
	Unread() ([]FeedEntry, error)
	MarkRead(feedID int64) error
}
