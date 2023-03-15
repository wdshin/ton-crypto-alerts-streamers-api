package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vladtenlive/ton-donate/storage"
	parsers "github.com/vladtenlive/ton-donate/utils/parsers"
)

type GetWidgetListResponse struct {
	Data  *[]GetWidgetListModel `json:"data"`
	Error string                `json:"error"`
}

type GetWidgetListModel struct {
	Type          string `json:"type,omitempty"`
	AmountGoal    uint64 `json:"amount_goal,omitempty"`
	AmountCurrent uint64 `json:"amount_current,omitempty"`
	IsActive      bool   `json:"isActive,omitempty"`
}

func (s *Service) GetWidgetsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	streamerId := parsers.GetStreamerId(r, s.auth)
	if streamerId == "" {
		response, _ := json.Marshal(&GetWidgetListResponse{nil, "Failed to parse streamer id."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	widgets, err := s.mongoStorage.GetWidgets(ctx, streamerId)
	if err != nil {
		response, _ := json.Marshal(&GetWidgetListResponse{nil, "Failed to load streamer widgets info."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	widgetsModel := make([]GetWidgetListModel, 0)
	for _, widget := range *widgets {
		widgetsModel = append(widgetsModel, GetWidgetListModel{
			Type:          widget.Type,
			AmountGoal:    widget.AmountGoal,
			AmountCurrent: widget.AmountCurrent,
			IsActive:      widget.IsActive})
	}
	response, _ := json.Marshal(&GetWidgetListResponse{&widgetsModel, ""})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type CreateWidgetRequest struct {
	Type          string `json:"type,omitempty"`
	AmountGoal    uint64 `json:"amount_goal,omitempty"`
	AmountCurrent uint64 `json:"amount_current,omitempty"`
}

type CreateWidgetResponse struct {
	Data  *CreateWidgetResponseModel `json:"data"`
	Error string                     `json:"error"`
}

type CreateWidgetResponseModel struct {
	WidgetId string `json:"widgetId,omitempty"`
}

func (s *Service) CreateWidgetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	streamerId := parsers.GetStreamerId(r, s.auth)
	if streamerId == "" {
		response, _ := json.Marshal(&CreateWidgetResponse{nil, "Failed to parse streamer id."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	var payload CreateWidgetRequest
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		response, _ := json.Marshal(&CreateWidgetResponse{nil, "Failed to parse widget payload."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	// ToDo: Create streamer donations widget info based on streamer id.
	widget := storage.Widget{
		StreamerId:    streamerId,
		Type:          payload.Type,
		AmountGoal:    payload.AmountGoal,
		AmountCurrent: payload.AmountCurrent,
		IsActive:      true, // ToDo: create active widget selection
	}
	result, err := s.mongoStorage.CreateWidget(ctx, widget)
	if err != nil {
		response, _ := json.Marshal(&CreateWidgetResponse{nil, "Failed to load streamer widgets info."})

		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		return
	}

	response, _ := json.Marshal(&CreateWidgetResponse{&CreateWidgetResponseModel{fmt.Sprint(result.InsertedID)}, ""})

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
