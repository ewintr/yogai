package fetch

import (
	"ewintr.nl/yogai/model"
	"github.com/google/uuid"
	"miniflux.app/client"
	"strconv"
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

func (m *Miniflux) Unread() ([]*model.Video, error) {
	result, err := m.client.Entries(&client.Filter{Status: "unread"})
	if err != nil {
		return []*model.Video{}, err
	}

	videos := []*model.Video{}
	for _, entry := range result.Entries {
		videos = append(videos, &model.Video{
			ID:          uuid.New(),
			Status:      model.STATUS_NEW,
			YoutubeURL:  entry.URL,
			FeedID:      strconv.Itoa(int(entry.ID)),
			Title:       entry.Title,
			Description: entry.Content,
		})
	}

	return videos, nil
}

func (m *Miniflux) MarkRead(entryID string) error {
	id, err := strconv.ParseInt(entryID, 10, 64)
	if err != nil {
		return err
	}

	if err := m.client.UpdateEntries([]int64{id}, "read"); err != nil {
		return err
	}

	return nil
}
