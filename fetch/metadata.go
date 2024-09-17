package fetch

import "go-mod.ewintr.nl/yogai/model"

type Metadata struct {
	Title       string
	Description string
	Duration    string
	PublishedAt string
}

type MetadataFetcher interface {
	FetchMetadata([]model.YoutubeVideoID) (map[model.YoutubeVideoID]Metadata, error)
}
