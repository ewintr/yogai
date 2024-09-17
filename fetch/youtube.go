package fetch

import (
	"strings"

	"go-mod.ewintr.nl/yogai/model"
	"google.golang.org/api/youtube/v3"
)

type Youtube struct {
	Client *youtube.Service
}

func NewYoutube(client *youtube.Service) *Youtube {
	return &Youtube{Client: client}
}

func (y *Youtube) Search(channelID model.YoutubeChannelID, pageToken string) ([]model.YoutubeVideoID, string, error) {
	call := y.Client.Search.
		List([]string{"id"}).
		MaxResults(50).
		Type("video").
		Order("date").
		ChannelId(string(channelID))

	if pageToken != "" {
		call.PageToken(pageToken)
	}

	response, err := call.Do()
	if err != nil {
		return []model.YoutubeVideoID{}, "", err
	}

	ids := make([]model.YoutubeVideoID, len(response.Items))
	for i, item := range response.Items {
		ids[i] = model.YoutubeVideoID(item.Id.VideoId)
	}

	return ids, response.NextPageToken, nil
}

func (y *Youtube) FetchMetadata(ytIDs []model.YoutubeVideoID) (map[model.YoutubeVideoID]Metadata, error) {
	strIDs := make([]string, len(ytIDs))
	for i, id := range ytIDs {
		strIDs[i] = string(id)
	}
	call := y.Client.Videos.
		List([]string{"snippet,contentDetails"}).
		Id(strings.Join(strIDs, ","))

	response, err := call.Do()
	if err != nil {
		return map[model.YoutubeVideoID]Metadata{}, err
	}

	mds := make(map[model.YoutubeVideoID]Metadata, len(response.Items))
	for _, item := range response.Items {
		if item.Snippet == nil {
			continue
		}
		md := Metadata{
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			PublishedAt: item.Snippet.PublishedAt,
		}

		if item.ContentDetails != nil {
			md.Duration = item.ContentDetails.Duration
		}

		mds[model.YoutubeVideoID(item.Id)] = md
	}

	return mds, nil
}
