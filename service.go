package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"ewintr.nl/yogai/fetch"
	"ewintr.nl/yogai/handler"
	"ewintr.nl/yogai/process"
	"ewintr.nl/yogai/storage"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slog"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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
	videoRelRepo := storage.NewPostgresVideoRepository(postgres)
	feedRelRepo := storage.NewPostgresFeedRepository(postgres)

	mflxClient := fetch.NewMiniflux(fetch.MinifluxInfo{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost/v1"),
		ApiKey:   getParam("MINIFLUX_APIKEY", ""),
	})

	fetchInterval, err := time.ParseDuration(getParam("FETCH_INTERVAL", "1m"))
	if err != nil {
		logger.Error("unable to parse fetch interval", err)
		os.Exit(1)
	}

	yt, err := youtube.NewService(ctx, option.WithAPIKey(getParam("YOUTUBE_API_KEY", "")))
	if err != nil {
		logger.Error("unable to create youtube service", err)
		os.Exit(1)
	}
	ytClient := fetch.NewYoutube(yt)

	openaiKey := getParam("OPENAI_API_KEY", "")
	openAIClient := openai.NewClient(openaiKey)

	wvResetSchema := getParam("WEAVIATE_RESET_SCHEMA", "false") == "true"
	wvClient, err := storage.NewWeaviate(getParam("WEAVIATE_HOST", ""), getParam("WEAVIATE_API_KEY", ""), openaiKey)
	if err != nil {
		logger.Error("unable to create weaviate client", err)
		os.Exit(1)
	}
	if wvResetSchema {
		logger.Info("resetting weaviate schema")
		if err := wvClient.ResetSchema(); err != nil {
			logger.Error("unable to reset weaviate schema", err)
			os.Exit(1)
		}
	}

	fetcher := fetch.NewFetch(feedRelRepo, videoRelRepo, ytClient, mflxClient, fetchInterval, ytClient, logger)
	go fetcher.Run()
	logger.Info("fetch service started")

	procs := process.NewProcessors(openAIClient)
	for i := 0; i < 4; i++ {
		go process.NewPipeline(fetcher.Out(), procs, videoRelRepo, wvClient, logger.With(slog.Int("pipeline", i))).Run()
	}
	logger.Info("processing service started")

	port, err := strconv.Atoi(getParam("API_PORT", "8080"))
	if err != nil {
		logger.Error("invalid port", err)
		os.Exit(1)
	}
	go http.ListenAndServe(fmt.Sprintf(":%d", port), handler.NewServer(videoRelRepo, logger))
	logger.Info("http server started")

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
