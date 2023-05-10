package main

import (
	"ewintr.nl/yogai/fetcher"
	"ewintr.nl/yogai/storage"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	postgres, err := storage.NewPostgres(storage.PostgresInfo{
		Host:     getParam("POSTGRES_HOST", "localhost"),
		Port:     getParam("POSTGRES_PORT", "5432"),
		User:     getParam("POSTGRES_USER", "yogai"),
		Password: getParam("POSTGRES_PASSWORD", "yogai"),
		Database: getParam("POSTGRES_DB", "yogai"),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	videoRepo := storage.NewPostgresVideoRepository(postgres)

	mflx := fetcher.NewMiniflux(fetcher.MinifluxInfo{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost/v1"),
		ApiKey:   getParam("MINIFLUX_APIKEY", ""),
	})

	fetchInterval, err := time.ParseDuration(getParam("FETCH_INTERVAL", "1m"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fetcher := fetcher.NewFetch(videoRepo, mflx, fetchInterval)
	go fetcher.Run()

	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt)
	<-done

	//fmt.Println(err)
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
