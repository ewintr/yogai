package storage

import (
	"context"

	"go-mod.ewintr.nl/yogai/model"
)

type FeedRelRepository interface {
	Save(feed *model.Feed) error
	FindByStatus(statuses ...model.FeedStatus) ([]*model.Feed, error)
}

type VideoRelRepository interface {
	Save(video *model.Video) error
	FindByStatus(statuses ...model.VideoStatus) ([]*model.Video, error)
}

type VideoVecRepository interface {
	Save(ctx context.Context, video *model.Video) error
}
