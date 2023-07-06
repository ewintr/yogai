package fetch

import (
	"strings"

	"miniflux.app/client"
)

type Entry struct {
	MinifluxEntryID string
	MinifluxFeedID  string
	MinifluxURL     string
	Title           string
	Description     string
}

type MinifluxInfo struct {
	Endpoint string
	ApiKey   string
}

type Miniflux struct {
	client *client.Client
}

func NewMiniflux(mflInfo MinifluxInfo) *Miniflux {
	return &Miniflux{
		client: client.New(mflInfo.Endpoint, mflInfo.ApiKey),
	}
}

func (m *Miniflux) Unread() ([]FeedEntry, error) {
	result, err := m.client.Entries(&client.Filter{Status: "unread"})
	if err != nil {
		return nil, err
	}

	entries := make([]FeedEntry, 0, len(result.Entries))
	for _, entry := range result.Entries {
		entries = append(entries, FeedEntry{
			EntryID:          entry.ID,
			FeedID:           entry.FeedID,
			YoutubeChannelID: strings.TrimPrefix(entry.Feed.FeedURL, "https://www.youtube.com/feeds/videos.xml?channel_id="),
			YoutubeID:        strings.TrimPrefix(entry.URL, "https://www.youtube.com/watch?v="),
		})
	}

	return entries, nil
}

func (m *Miniflux) MarkRead(entryID int64) error {
	if err := m.client.UpdateEntries([]int64{entryID}, "read"); err != nil {
		return err
	}

	return nil
}
