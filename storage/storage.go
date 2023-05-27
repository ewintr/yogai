package storage

import (
	"ewintr.nl/yogai/model"
)

type FeedRepository interface {
	Save(feed *model.Feed) error
	FindByStatus(statuses ...model.FeedStatus) ([]*model.Feed, error)
}

type VideoRepository interface {
	Save(video *model.Video) error
	FindByStatus(statuses ...model.VideoStatus) ([]*model.Video, error)
}
