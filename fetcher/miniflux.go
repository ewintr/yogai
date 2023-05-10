package fetcher

import (
	"miniflux.app/client"
	"strings"
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
		return []FeedEntry{}, err
	}

	entries := []FeedEntry{}
	for _, entry := range result.Entries {
		entries = append(entries, FeedEntry{
			EntryID:   entry.ID,
			FeedID:    entry.FeedID,
			YouTubeID: strings.TrimPrefix(entry.URL, "https://www.youtube.com/watch?v="),
		})

		//	ID:          uuid.New(),
		//	Status:      model.STATUS_NEW,
		//	YoutubeURL:  entry.URL,
		//	FeedID:      strconv.Itoa(int(entry.ID)),
		//	Title:       entry.Title,
		//	Description: entry.Content,
		//})
	}

	return entries, nil
}

func (m *Miniflux) MarkRead(entryID int64) error {
	if err := m.client.UpdateEntries([]int64{entryID}, "read"); err != nil {
		return err
	}

	return nil
}
