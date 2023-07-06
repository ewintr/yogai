package process

import (
	"context"

	"ewintr.nl/yogai/model"
	"ewintr.nl/yogai/storage"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
)

type VideoProcessor interface {
	Name() string
	Do(ctx context.Context, video *model.Video) error
}

type Processors struct {
	procs map[string]VideoProcessor
}

func NewProcessors(openAIClient *openai.Client) *Processors {
	return &Processors{
		procs: map[string]VideoProcessor{
			"summarizer": NewOpenAISummarizer(openAIClient),
		},
	}
}

func (p *Processors) Next(video *model.Video) VideoProcessor {
	if video.Summary == "" {
		return p.procs["summarizer"]
	}

	return nil
}

type Pipeline struct {
	in         chan *model.Video
	procs      *Processors
	logger     *slog.Logger
	relStorage storage.VideoRelRepository
	vecStorage storage.VideoVecRepository
}

func NewPipeline(in chan *model.Video, processors *Processors, relDB storage.VideoRelRepository, vecDB storage.VideoVecRepository, logger *slog.Logger) *Pipeline {
	return &Pipeline{
		in:         in,
		procs:      processors,
		relStorage: relDB,
		vecStorage: vecDB,
		logger:     logger,
	}
}

func (p *Pipeline) Run() {
	ctx := context.Background()
	for video := range p.in {
		p.Process(ctx, video)
	}
}

func (p *Pipeline) Process(ctx context.Context, video *model.Video) {
	p.logger.Info("processing video", slog.String("video", string(video.YoutubeID)))
	for {
		next := p.procs.Next(video)
		if next == nil {
			p.logger.Info("no more processors for video", slog.String("video", string(video.YoutubeID)))
			return
		}

		p.logger.Info("processing video", slog.String("video", string(video.YoutubeID)), slog.String("processor", next.Name()))
		if err := next.Do(context.Background(), video); err != nil {
			p.logger.Error("failed to process video", slog.String("video", string(video.YoutubeID)), slog.String("processor", next.Name()), slog.String("error", err.Error()))
			return
		}
		if err := p.relStorage.Save(video); err != nil {
			p.logger.Error("failed to save video in rel db", slog.String("video", string(video.YoutubeID)), slog.String("error", err.Error()))
			return
		}
		if err := p.vecStorage.Save(ctx, video); err != nil {
			p.logger.Error("failed to save video in rel db", slog.String("video", string(video.YoutubeID)), slog.String("error", err.Error()))
			return
		}
	}
}
