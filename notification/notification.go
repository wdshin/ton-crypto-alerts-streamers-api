package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/labstack/gommon/log"
	"github.com/vladtenlive/ton-donate/storage"
)

type Service struct {
	client  *http.Client
	storage storage.Storage
}

func NewService(client *http.Client, storage storage.Storage) *Service {
	return &Service{
		client:  client,
		storage: storage,
	}
}

func (n *Service) Start(port string) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/alert", n.Handler)

	return http.ListenAndServe(port, r)
}

type NotificationRequest struct {
	Amount   uint64 `json:"amount"`
	From     string `json:"nickname"`
	ClientId string `json:"clientId"`
	Message  string `json:"text"`
	Sign     string `json:"sign"`
}

func (n *Service) Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req NotificationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	tx, err := n.storage.GetTransactionBySign(ctx, req.Sign)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if tx == nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if tx.Acked {
		log.Error("notification already sent")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if tx.Amount != req.Amount {
		log.Error("amount not equal")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if time.Since(tx.CreatedAt.UTC()) > 30*time.Minute {
		log.Error("transaction expired")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// TODO: check wallet address from id

	err = n.Send(ctx, req)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = n.storage.AckTransaction(ctx, tx.TxHash)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type AlertRequest struct {
}

func (n *Service) Send(ctx context.Context, req NotificationRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	resp, err := n.client.Post(
		"https://seahorse-app-qdt2w.ondigitalocean.app/payments",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return errors.New("resubmit donate")
	}

	return nil
}
