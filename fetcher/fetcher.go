package fetcher

import (
	"time"

	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

type Fetcher struct {
	interval        time.Duration
	feedRepo        storage.FeedRepository
	videoRepo       storage.VideoRepository
	feedReader      FeedReader
	channelReader   ChannelReader
	metadataFetcher MetadataFetcher
	summaryFetcher  SummaryFetcher
	feedPipeline    chan *model.Feed
	videoPipeline   chan *model.Video
	needsMetadata   chan *model.Video
	needsSummary    chan *model.Video
	logger          *slog.Logger
}

func NewFetch(feedRepo storage.FeedRepository, videoRepo storage.VideoRepository, channelReader ChannelReader, feedReader FeedReader, interval time.Duration, metadataFetcher MetadataFetcher, summaryFetcher SummaryFetcher, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		interval:        interval,
		feedRepo:        feedRepo,
		videoRepo:       videoRepo,
		channelReader:   channelReader,
		feedReader:      feedReader,
		metadataFetcher: metadataFetcher,
		summaryFetcher:  summaryFetcher,
		feedPipeline:    make(chan *model.Feed, 10),
		videoPipeline:   make(chan *model.Video, 10),
		needsMetadata:   make(chan *model.Video, 10),
		needsSummary:    make(chan *model.Video, 10),
		logger:          logger,
	}
}

func (f *Fetcher) Run() {
	go f.FetchHistoricalVideos()
	go f.FindNewFeeds()

	go f.ReadFeeds()
	go f.MetadataFetcher()
	go f.SummaryFetcher()
	go f.FindUnprocessed()

	f.logger.Info("started videoPipeline")
	for {
		select {
		case video := <-f.videoPipeline:
			switch video.Status {
			case model.StatusNew:
				f.needsMetadata <- video
			case model.StatusHasMetadata:
				f.needsSummary <- video
			case model.StatusHasSummary:
				video.Status = model.StatusReady
				f.logger.Info("video is ready", slog.String("id", video.ID.String()))

			}
			if err := f.videoRepo.Save(video); err != nil {
				f.logger.Error("failed to save video", err)
				continue
			}
		}

	}
}

func (f *Fetcher) FindNewFeeds() {
	f.logger.Info("looking for new feeds")
	feeds, err := f.feedRepo.FindByStatus(model.FeedStatusNew)
	if err != nil {
		f.logger.Error("failed to fetch feeds", err)
		return
	}
	for _, feed := range feeds {
		f.feedPipeline <- feed
	}
}

func (f *Fetcher) FetchHistoricalVideos() {
	f.logger.Info("started historical video fetcher")

	for feed := range f.feedPipeline {
		f.logger.Info("fetching historical videos", slog.String("channelid", string(feed.YoutubeChannelID)))
		token := ""
		for {
			token = f.FetchHistoricalVideoPage(feed.YoutubeChannelID, token)
			if token == "" {
				break
			}
		}
		feed.Status = model.FeedStatusReady
		if err := f.feedRepo.Save(feed); err != nil {
			f.logger.Error("failed to save feed", err)
			continue
		}
	}
}

func (f *Fetcher) FetchHistoricalVideoPage(channelID model.YoutubeChannelID, pageToken string) string {
	f.logger.Info("fetching historical video page", slog.String("channelid", string(channelID)), slog.String("pagetoken", pageToken))
	ytIDs, pageToken, err := f.channelReader.Search(channelID, pageToken)
	if err != nil {
		f.logger.Error("failed to fetch channel", err)
		return ""
	}
	for _, ytID := range ytIDs {
		video := &model.Video{
			ID:               uuid.New(),
			Status:           model.StatusNew,
			YoutubeID:        ytID,
			YoutubeChannelID: channelID,
		}
		if err := f.videoRepo.Save(video); err != nil {
			f.logger.Error("failed to save video", err)
			continue
		}
		f.videoPipeline <- video
	}

	f.logger.Info("fetched historical video page", slog.String("channelid", string(channelID)), slog.String("pagetoken", pageToken), slog.Int("count", len(ytIDs)))
	return pageToken
}

func (f *Fetcher) FindUnprocessed() {
	f.logger.Info("looking for unprocessed videos")
	videos, err := f.videoRepo.FindByStatus(model.StatusNew, model.StatusHasMetadata)
	if err != nil {
		f.logger.Error("failed to fetch unprocessed videos", err)
		return
	}
	f.logger.Info("found unprocessed videos", slog.Int("count", len(videos)))
	for _, video := range videos {
		f.videoPipeline <- video
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
				ID:               uuid.New(),
				Status:           model.StatusNew,
				YoutubeID:        model.YoutubeVideoID(entry.YoutubeID),
				YoutubeChannelID: model.YoutubeChannelID(entry.YoutubeChannelID),
			}
			if err := f.videoRepo.Save(video); err != nil {
				f.logger.Error("failed to save video", err)
				continue
			}
			f.videoPipeline <- video
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
			ids := make([]model.YoutubeVideoID, 0, len(videos))
			for _, video := range videos {
				ids = append(ids, video.YoutubeID)
			}
			mds, err := f.metadataFetcher.FetchMetadata(ids)
			if err != nil {
				f.logger.Error("failed to fetch metadata", err)
				continue
			}
			for _, video := range videos {
				md := mds[video.YoutubeID]
				video.YoutubeTitle = md.Title
				video.YoutubeDescription = md.Description
				video.YoutubeDuration = md.Duration
				video.YoutubePublishedAt = md.PublishedAt
				video.Status = model.StatusHasMetadata

				if err := f.videoRepo.Save(video); err != nil {
					f.logger.Error("failed to save video", err)
					continue
				}
			}
			f.logger.Info("fetched metadata", slog.Int("count", len(videos)))
		}
	}()

	for {
		select {
		case video := <-f.needsMetadata:
			timeout.Reset(10 * time.Second)
			buffer = append(buffer, video)
			if len(buffer) >= 50 {
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

func (f *Fetcher) SummaryFetcher() {
	for {
		select {
		case video := <-f.needsSummary:
			f.logger.Info("fetching summary", slog.String("id", video.ID.String()))
			if err := f.summaryFetcher.FetchSummary(video); err != nil {
				f.logger.Error("failed to fetch summary", err)
				continue
			}
			video.Status = model.StatusHasSummary
			f.logger.Info("fetched summary", slog.String("id", video.ID.String()))
			f.videoPipeline <- video
		}
	}
}
