package fetcher

import (
	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
	"time"
)

type Fetcher struct {
	interval        time.Duration
	videoRepo       storage.VideoRepository
	feedReader      FeedReader
	metadataFetcher MetadataFetcher
	pipeline        chan *model.Video
	needsMetadata   chan *model.Video
	logger          *slog.Logger
}

func NewFetch(videoRepo storage.VideoRepository, feedReader FeedReader, interval time.Duration, metadataFetcher MetadataFetcher, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		interval:        interval,
		videoRepo:       videoRepo,
		feedReader:      feedReader,
		metadataFetcher: metadataFetcher,
		pipeline:        make(chan *model.Video),
		needsMetadata:   make(chan *model.Video),
		logger:          logger,
	}
}

func (f *Fetcher) Run() {
	go f.ReadFeeds()
	go f.MetadataFetcher()

	f.logger.Info("started pipeline")
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
	f.logger.Info("started feed reader")
	ticker := time.NewTicker(f.interval)
	for range ticker.C {
		entries, err := f.feedReader.Unread()
		if err != nil {
			f.logger.Error("failed to fetch unread entries", err)
			continue
		}
		f.logger.Info("fetched unread entries", slog.Int("count", len(entries)))
		if len(entries) == 0 {
			continue
		}

		for _, entry := range entries {
			video := &model.Video{
				ID:        uuid.New(),
				Status:    model.STATUS_NEW,
				YoutubeID: entry.YouTubeID,
				// feed id
			}
			if err := f.videoRepo.Save(video); err != nil {
				f.logger.Error("failed to save video", err)
				continue
			}
			f.pipeline <- video
			if err := f.feedReader.MarkRead(entry.EntryID); err != nil {
				f.logger.Error("failed to mark entry as read", err)
				continue
			}
		}
	}
}

func (f *Fetcher) MetadataFetcher() {
	f.logger.Info("started metadata fetcher")

	buffer := []*model.Video{}
	timeout := time.NewTimer(10 * time.Second)
	fetch := make(chan []*model.Video)

	go func() {
		for videos := range fetch {
			f.logger.Info("fetching metadata", slog.Int("count", len(videos)))
			ids := make([]string, 0, len(videos))
			for _, video := range videos {
				ids = append(ids, video.YoutubeID)
			}
			mds, err := f.metadataFetcher.FetchMetadata(ids)
			if err != nil {
				f.logger.Error("failed to fetch metadata", err)
				continue
			}
			for _, video := range videos {
				video.Title = mds[video.YoutubeID].Title
				video.Description = mds[video.YoutubeID].Description

				if err := f.videoRepo.Save(video); err != nil {
					f.logger.Error("failed to save video", err)
					continue
				}
			}
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
