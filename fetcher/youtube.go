package fetcher

import (
	"google.golang.org/api/youtube/v3"
	"strings"
)

type Youtube struct {
	Client *youtube.Service
}

func NewYoutube(client *youtube.Service) *Youtube {
	return &Youtube{Client: client}
}

func (y *Youtube) FetchMetadata(ytIDs []string) (map[string]Metadata, error) {
	call := y.Client.Videos.
		List([]string{"snippet"}).
		Id(strings.Join(ytIDs, ","))

	response, err := call.Do()
	if err != nil {
		return map[string]Metadata{}, err
	}

	mds := make(map[string]Metadata, len(response.Items))
	for _, item := range response.Items {
		mds[item.Id] = Metadata{
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
		}
	}

	return mds, nil
}
