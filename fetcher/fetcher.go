package fetcher

import (
	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

type Fetcher struct {
	interval      time.Duration
	videoRepo     storage.VideoRepository
	feedReader    FeedReader
	pipeline      chan *model.Video
	needsMetadata chan *model.Video
}

func NewFetch(videoRepo storage.VideoRepository, feedReader FeedReader, interval time.Duration) *Fetcher {
	return &Fetcher{
		interval:      interval,
		videoRepo:     videoRepo,
		feedReader:    feedReader,
		pipeline:      make(chan *model.Video),
		needsMetadata: make(chan *model.Video),
	}
}

func (f *Fetcher) Run() {
	go f.ReadFeeds()
	go f.MetadataFetcher()

	for {
		select {
		case video := <-f.pipeline:
			switch video.Status {
			case model.STATUS_NEW:
				f.needsMetadata <- video
			}
		}
	}
}

func (f *Fetcher) ReadFeeds() {
	ticker := time.NewTicker(f.interval)
	for range ticker.C {
		entries, err := f.feedReader.Unread()
		if err != nil {
			log.Println(err)
		}
		for _, entry := range entries {
			video := &model.Video{
				ID:        uuid.New(),
				Status:    model.STATUS_NEW,
				YoutubeID: entry.YouTubeID,
				// feed id
			}
			if err := f.videoRepo.Save(video); err != nil {
				log.Println(err)
				continue
			}
			f.pipeline <- video
			if err := f.feedReader.MarkRead(entry.EntryID); err != nil {
				log.Println(err)
			}
		}
	}
}

func (f *Fetcher) MetadataFetcher() {
	buffer := []*model.Video{}
	timeout := time.NewTimer(10 * time.Second)
	fetch := make(chan []*model.Video)

	go func() {
		for videos := range fetch {
			fmt.Println("MD Fetching metadata")
			fmt.Printf("%d videos to fetch\n", len(videos))
		}
	}()

	for {
		select {
		case video := <-f.needsMetadata:
			timeout.Reset(10 * time.Second)
			buffer = append(buffer, video)
			if len(buffer) >= 10 {
				batch := make([]*model.Video, len(buffer))
				copy(batch, buffer)
				fetch <- batch
				buffer = []*model.Video{}
			}
		case <-timeout.C:
			if len(buffer) == 0 {
				continue
			}
			batch := make([]*model.Video, len(buffer))
			copy(batch, buffer)
			fetch <- batch
			buffer = []*model.Video{}
		}
	}
}
