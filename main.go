package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vladtenlive/ton-donate/handlers"
	"github.com/vladtenlive/ton-donate/storage"
	"github.com/vladtenlive/ton-donate/ton"
	"github.com/vladtenlive/ton-donate/utils"
)

func main() {
	ctx := context.Background()

	utils.ValidateEnvVariables()

	port := os.Getenv("PORT")
	contractAddress := os.Getenv("CONTRACT_ADDRESS")
	notificationUrl := os.Getenv("NOTIFICATION_URL")

	mongo, err := storage.NewMongoClient(ctx)
	if err != nil {
		fmt.Println("FAILED TO CONNECT TO MONGO!")
		panic(err)
	}

	n := ton.NewNotifier(http.DefaultClient, notificationUrl)
	tonConnector, err := ton.New(
		ctx,
		contractAddress,
		nil,
		mongo,
		n,
	)
	if err != nil {
		log.Fatal(err)
	}

	s := handlers.NewService(http.DefaultClient, nil, mongo)

	go tonConnector.Start(ctx, 3*time.Second)

	// APIs
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(middleware.AllowContentType("application/json"))

	r.Use(middleware.Heartbeat("/"))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Group(func(r chi.Router) {
		r.Get("/streamer", s.GetStreamerHandler)
		r.Get("/streamer/{cognitoId}", s.GetStreamerHandler)
		r.Post("/streamer", s.SaveStreamerHandler)
	})
	r.Group(func(r chi.Router) {
		r.Get("/donations", s.GetDonationListHandler)
		r.Post("/donations", s.CreateDonationHandler)
	})
	r.Group(func(r chi.Router) {
		r.Get("/widgets", s.GetWidgetsHandler)
		r.Post("/widgets", s.CreateWidgetHandler)
	})

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
