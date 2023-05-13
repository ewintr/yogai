package main

import (
	"context"
	"ewintr.nl/yogai/fetcher"
	"ewintr.nl/yogai/storage"
	"golang.org/x/exp/slog"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr))
	postgres, err := storage.NewPostgres(storage.PostgresInfo{
		Host:     getParam("POSTGRES_HOST", "localhost"),
		Port:     getParam("POSTGRES_PORT", "5432"),
		User:     getParam("POSTGRES_USER", "yogai"),
		Password: getParam("POSTGRES_PASSWORD", "yogai"),
		Database: getParam("POSTGRES_DB", "yogai"),
	})
	if err != nil {
		logger.Error("unable to connect to postgres", err)
		os.Exit(1)
	}
	videoRepo := storage.NewPostgresVideoRepository(postgres)

	mflx := fetcher.NewMiniflux(fetcher.MinifluxInfo{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost/v1"),
		ApiKey:   getParam("MINIFLUX_APIKEY", ""),
	})

	fetchInterval, err := time.ParseDuration(getParam("FETCH_INTERVAL", "1m"))
	if err != nil {
		logger.Error("unable to parse fetch interval", err)
		os.Exit(1)
	}

	ytClient, err := youtube.NewService(ctx, option.WithAPIKey(getParam("YOUTUBE_API_KEY", "")))
	if err != nil {
		logger.Error("unable to create youtube service", err)
		os.Exit(1)
	}
	yt := fetcher.NewYoutube(ytClient)

	openAIClient := fetcher.NewOpenAI(getParam("OPENAI_API_KEY", ""))

	fetcher := fetcher.NewFetch(videoRepo, mflx, fetchInterval, yt, openAIClient, logger)
	go fetcher.Run()
	logger.Info("service started")

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)
	<-done

	logger.Info("service stopped")
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
