package fetcher

type Metadata struct {
	Title       string
	Description string
}

type MetadataFetcher interface {
	FetchMetadata([]string) (map[string]Metadata, error)
}
