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
		// contractAddress = "EQB_ryLyj9tdIGuwBOqsxg6bPXeCD55J9GiEP4VJhtVwmz8n"
		contractAddress = "EQAKOF-lITE_xjF8WNuXtV6I9B3vOGEgvEdc2YX9cojyidlZ"
	}

	// pg, err := storage.NewPostgres(pgConn)
	// if err != nil {
	// 	panic(err)
	// }

	mongo, err := storage.NewMongoClient(ctx)
	if err != nil {
		fmt.Println("FAILED TO CONNECT TO MONGO!")
		panic(err)
	}

	n := ton.NewNotifier(http.DefaultClient, "https://seahorse-app-qdt2w.ondigitalocean.app/payments")
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

	// s := handlers.NewService(http.DefaultClient, pg, mongo)
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
