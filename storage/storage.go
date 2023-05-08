package storage

import (
	"ewintr.nl/yogai/model"
)

type VideoRepository interface {
	Save(video *model.Video) error
}
