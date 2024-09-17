package fetch

import "go-mod.ewintr.nl/yogai/model"

type SummaryFetcher interface {
	FetchSummary(video *model.Video) error
}
