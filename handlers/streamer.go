package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/gommon/log"
	"github.com/vladtenlive/ton-donate/storage"
	parsers "github.com/vladtenlive/ton-donate/utils/parsers"
)

type GetStreamerResponse struct {
	Data  *GetStreamerModel `json:"data"`
	Error string            `json:"error"`
}

type GetStreamerModel struct {
	StreamerId    string `json:"streamerId,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty"`
}

func (s *Service) GetStreamerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cognitoId := parsers.GetCognitoId(r, s.auth)
	if cognitoId == "" {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Failed to parse cognito id!"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	streamer, err := s.mongoStorage.GetStreamerByCognitoId(ctx, cognitoId)
	if err != nil {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Streamer with such id does not exist!"})

		w.WriteHeader(http.StatusNotFound)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(
		&GetStreamerResponse{
			&GetStreamerModel{
				StreamerId:    streamer.StreamerId,
				WalletAddress: streamer.WalletAddress}, ""})
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type SaveStreamerRequest struct {
	// StreamerId    string `json:"streamerId,omitempty"`
	// CognitoId     string `json:"cognito_id,omitempty"`
	WalletAddress string `json:"wallet_address,omitempty"`
}

type RegisterStreamerResponse struct {
	Data  SaveStreamerModel `json:"data"`
	Error string            `json:"error"`
}

type SaveStreamerModel struct {
	StreamerId string `json:"streamerId,omitempty"`
}

func (s *Service) SaveStreamerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload SaveStreamerRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Failed to parse new streamer payload."})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	streamerId := parsers.GetStreamerId(r, s.auth)
	if streamerId == "" {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Failed to parse cognito id!"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// Check if another streamer registered such wallet, will allow to update to the same if streamer is the same
	foundStreamer, err := s.mongoStorage.GetStreamerByWalletAddress(ctx, payload.WalletAddress)
	if err != nil {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Failed to verify streamer's wallet address."})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	} else if foundStreamer != nil && streamerId != foundStreamer.StreamerId {
		response, _ := json.Marshal(&GetStreamerResponse{nil, "Streamer with this wallet address has been already registered."})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// var streamerId string
	// if payload.StreamerId == "" {
	// 	streamerId = uuid.New().String()
	// } else {
	// 	streamerId = payload.StreamerId
	// }

	streamer := storage.Streamer{
		WalletAddress: payload.WalletAddress,
		StreamerId:    streamerId,
		CognitoId:     streamerId, // ToDo: get from cognito, but they are actually same guid.
	}

	_, err = s.mongoStorage.SaveStreamer(ctx, streamer)
	if err != nil {
		log.Error(err)

		response, _ := json.Marshal(&RegisterStreamerResponse{Error: "Failed to parse streamer id!"})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(&RegisterStreamerResponse{SaveStreamerModel{streamerId}, ""})

	w.Write(response)
	w.WriteHeader(http.StatusCreated)
}
