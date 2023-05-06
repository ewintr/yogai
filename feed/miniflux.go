package feed

import (
	"fmt"
	"miniflux.app/client"
)

type Miniflux struct {
	client *client.Client
}

func NewMiniflux(url, apiKey string) *Miniflux {
	return &Miniflux{
		client: client.New(url, apiKey),
	}
}

func (m *Miniflux) Feeds() error {

	feeds, err := m.client.Feeds()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(feeds)

	return nil
}

type Entry struct {
	ChannelID int
	Title     string
	URL       string
}

func (m *Miniflux) Unread() ([]Entry, error) {
	result, err := m.client.Entries(&client.Filter{Status: "unread"})
	if err != nil {
		return []Entry{}, err
	}

	for _, entry := range result.Entries {
		fmt.Println(entry.ID, entry.Title, entry.URL)
	}

	return []Entry{}, nil
}
