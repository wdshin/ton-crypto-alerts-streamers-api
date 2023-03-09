package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/vladtenlive/ton-donate/storage"
	parsers "github.com/vladtenlive/ton-donate/utils/parsers"
)

type GetDonationListResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

// ToDo: Confirmed (from transaction), Acked (for sending to notificator)

func (n *Service) GetDonationListHandler(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()

	id := parsers.GetStreamerId(r)
	if id == "" {
		response, _ := json.Marshal(&GetDonationListResponse{"", "Failed to parse streamer id!"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// ToDo: Load all streamer donations based on streamer id.
}

type CreateDonationRequest struct {
	Amount        float64 `json:"amount"`
	From          string  `json:"nickname"`
	WalletAddress string  `json:"wallet_address"`
	StreamerId    string  `json:"streamerId"`
	Message       string  `json:"text"`
	Sign          string  `json:"sign"`
}

func (s *Service) CreateDonationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateDonationRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// ToDo: Check wallet address if exist + streamer id

	if req.Sign == "" {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Check if exist donation with the same sign or error
	donation, err := s.mongoStorage.GetDonationBySign(ctx, req.Sign)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	} else if donation != nil {
		// Donation was saved previously, do not allow flood of donations by same transaction.
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Donation has already been saved."))
		return
	}

	newDonation := storage.Donation{
		TxHash:        "", // we dont know it at this point, only after it's been processed by Ton
		Amount:        req.Amount,
		WalletAddress: req.WalletAddress,
		Sign:          req.Sign,
		From:          req.From,
		StreamerId:    req.StreamerId,
		Message:       req.Message,
		Lt:            "", // datetime of what?
		Verified:      false,
		Acked:         false,
	}

	_, err = s.mongoStorage.CreateDonation(ctx, newDonation)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

type NotificationRequest struct {
	Amount        uint64 `json:"amount"`
	From          string `json:"nickname"`
	WalletAddress string `json:"wallet_address"`
	ClientId      string `json:"clientId"`
	Message       string `json:"text"`
	Sign          string `json:"sign"`
}

func (n *Service) NotificationHandler(w http.ResponseWriter, r *http.Request) {
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

	clientId, err := n.storage.GetStreamerClientIdByWalletAddress(ctx, req.WalletAddress)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	req.ClientId = clientId

	err = n.SendNotification(ctx, req)
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

// func (n *Service) RegisterStreamerHandler(w http.ResponseWriter, r *http.Request) {
// 	ctx := r.Context()

// 	clientId := uuid.New().String()

// 	streamer := storage.Streamer{
// 		WalletAddress: chi.URLParam(r, "wallet_address"),
// 		ClientId:      clientId,
// 	}

// 	err := n.storage.StoreStreamer(ctx, streamer)
// 	if err != nil {
// 		log.Error(err)
// 		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
// 		return
// 	}

// 	w.Write([]byte(fmt.Sprintf(`{"client_id": "%s"}`, clientId)))
// 	w.WriteHeader(http.StatusCreated)
// }

func (n *Service) SendNotification(ctx context.Context, req NotificationRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

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
