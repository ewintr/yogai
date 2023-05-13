package fetcher

import "ewintr.nl/yogai/model"

type SummaryFetcher interface {
	FetchSummary(video *model.Video) error
}
