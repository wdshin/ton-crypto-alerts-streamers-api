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

	pg, err := storage.NewPostgres(os.Getenv("PG_CONN"))
	if err != nil {
		panic(err)
	}

	tonConnector, err := ton.New(
		ctx,
		"EQB_ryLyj9tdIGuwBOqsxg6bPXeCD55J9GiEP4VJhtVwmz8n",
		pg,
	)
	if err != nil {
		log.Fatal(err)
	}

	notifiactionService := notification.NewService(http.DefaultClient, pg)

	go tonConnector.Start(ctx, 5*time.Second)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := notifiactionService.Start(port); err != nil {
		log.Fatal(err)
	}
}
