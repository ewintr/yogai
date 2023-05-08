package fetch

import (
	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"log"
	"time"
)

type FeedReader interface {
	Unread() ([]*model.Video, error)
	MarkRead(feedID string) error
}

type Fetch struct {
	interval   time.Duration
	videoRepo  storage.VideoRepository
	feedReader FeedReader
	out        chan<- model.Video
}

func NewFetch(videoRepo storage.VideoRepository, feedReader FeedReader, interval time.Duration) *Fetch {
	return &Fetch{
		interval:   interval,
		videoRepo:  videoRepo,
		feedReader: feedReader,
	}
}

func (v *Fetch) Run() {
	ticker := time.NewTicker(v.interval)
	for {
		select {
		case <-ticker.C:
			newVideos, err := v.feedReader.Unread()
			if err != nil {
				log.Println(err)
			}
			for _, video := range newVideos {
				if err := v.videoRepo.Save(video); err != nil {
					log.Println(err)
					continue
				}
				//v.out <- video
				if err := v.feedReader.MarkRead(video.FeedID); err != nil {
					log.Println(err)
				}
			}

		}
	}
}

func (v *Fetch) Out() chan<- model.Video {
	return v.out
}
