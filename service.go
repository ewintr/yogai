package main

import (
	"database/sql"
	"ewintr.nl/yogai/feed"
	"ewintr.nl/yogai/storage"
	"fmt"
	"os"
)

func main() {
	pgInfo := struct {
		Host     string
		Port     string
		User     string
		Password string
		Database string
	}{
		Host:     getParam("POSTGRES_HOST", "localhost"),
		Port:     getParam("POSTGRES_PORT", "5432"),
		User:     getParam("POSTGRES_USER", "yogai"),
		Password: getParam("POSTGRES_PASSWORD", "yogai"),
		Database: getParam("POSTGRES_DB", "yogai"),
	}
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgInfo.Host, pgInfo.Port, pgInfo.User, pgInfo.Password, pgInfo.Database))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	_, err = storage.NewPostgres(db)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mlxInfo := struct {
		Endpoint string
		ApiKey   string
	}{
		Endpoint: getParam("MINIFLUX_ENDPOINT", "http://localhost/v1"),
		ApiKey:   getParam("MINIFLUX_APIKEY", ""),
	}
	mflx := feed.NewMiniflux(mlxInfo.Endpoint, mlxInfo.ApiKey)
	_, err = mflx.Unread()
	fmt.Println(err)
}

func getParam(param, def string) string {
	if val, ok := os.LookupEnv(param); ok {
		return val
	}
	return def
}
