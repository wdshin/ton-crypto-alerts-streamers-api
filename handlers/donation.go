package handlers

import (
	"encoding/json"
	"net/http"

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
	Amount        uint64 `json:"amount"`
	From          string `json:"nickname"`
	WalletAddress string `json:"wallet_address"`
	StreamerId    string `json:"streamerId"`
	Message       string `json:"text"`
	Sign          string `json:"sign"`
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
		From:          req.From,
		StreamerId:    req.StreamerId,
		WalletAddress: req.WalletAddress,
		Amount:        req.Amount,
		Message:       req.Message,
		Sign:          req.Sign,
		TxHash:        "", // we dont know it at this point, only after it's been processed by Ton
		Lt:            0,  // we dont know it at this point, only after it's been processed by Ton
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
