package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/gommon/log"
	"github.com/vladtenlive/ton-donate/pkg/storage"
	parsers "github.com/vladtenlive/ton-donate/pkg/utils/parsers"
)

type GetDonationListResponse struct {
	Data  *[]GetDonationListModel `json:"data"`
	Error string                  `json:"error"`
}

type GetDonationListModel struct {
	From    string `json:"nickname,omitempty" bson:"nickname,omitempty"`
	Message string `json:"text,omitempty" bson:"message,omitempty"`
	Amount  uint64 `json:"amount,omitempty" bson:"amount,omitempty"`
}

// ToDo: Confirmed (from transaction), Acked (for sending to notificator)

func (s *Service) GetDonationListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	streamerId := parsers.GetStreamerId(r, s.auth)
	if streamerId == "" {
		response, _ := json.Marshal(&GetDonationListResponse{nil, "Failed to parse streamer id."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	donations, err := s.mongoStorage.GetStreamerDonations(ctx, streamerId)
	if err != nil {
		response, _ := json.Marshal(&GetDonationListResponse{nil, "Failed to load streamer donations."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	donationsModel := make([]GetDonationListModel, 0)
	for _, donation := range *donations {
		donationsModel = append(donationsModel, GetDonationListModel{
			From:    donation.From,
			Message: donation.Message,
			Amount:  donation.Amount})
	}
	response, _ := json.Marshal(&GetDonationListResponse{&donationsModel, ""})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type CreateDonationRequest struct {
	Amount        uint64 `json:"amount"`
	From          string `json:"nickname"`
	WalletAddress string `json:"wallet_address"`
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

	if req.Sign == "" {
		log.Error("Wrong sign: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Wrong sign: " + err.Error()))
		return
	}

	if req.WalletAddress == "" {
		log.Error("Please provide wallet address: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Please provide wallet address: " + err.Error()))
		return
	}

	streamer, err := s.mongoStorage.GetStreamerByWalletAddress(ctx, req.WalletAddress)
	if err != nil {
		log.Error("Streamer with current wallet address does not exist: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Streamer with current wallet address does not exist: " + err.Error()))
		return
	} else if streamer == nil {
		log.Error("Streamer with current wallet address does not exist: ", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Streamer with current wallet address does not exist: " + err.Error()))
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
		StreamerId:    streamer.StreamerId,
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
