package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/vladtenlive/ton-donate/notification"
	"github.com/vladtenlive/ton-donate/storage"
	"github.com/vladtenlive/ton-donate/ton"
)

func main() {
	ctx := context.Background()

	pgConn := os.Getenv("PG_CONN")
	if pgConn == "" {
		pgConn = "postgres://postgres:mysecretpassword@localhost:5432/postgres?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	if contractAddress == "" {
		contractAddress = "EQB_ryLyj9tdIGuwBOqsxg6bPXeCD55J9GiEP4VJhtVwmz8n"
	}

	pg, err := storage.NewPostgres(pgConn)
	if err != nil {
		panic(err)
	}

	tonConnector, err := ton.New(
		ctx,
		contractAddress,
		pg,
	)
	if err != nil {
		log.Fatal(err)
	}

	notifiactionService := notification.NewService(http.DefaultClient, pg)

	go tonConnector.Start(ctx, 5*time.Second)

	if err := notifiactionService.Start(port); err != nil {
		log.Fatal(err)
	}
}
