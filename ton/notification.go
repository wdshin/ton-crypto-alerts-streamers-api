package ton

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	DefaultWidgetUri = "https://seahorse-app-qdt2w.ondigitalocean.app/payments"
)

type NotificationError struct {
	Id string
}

func (e NotificationError) Error() string {
	return fmt.Sprintf("Resubmit notification: [id=%s]", e.Id)
}

type NotificationRequest struct {
	Id         string // could be tx hash for example, for more detailed error handling
	Amount     uint64 `json:"amount"`
	Text       string `json:"text"`
	Nickname   string `json:"nickname"`
	StreamerId string `json:"clientId"`
}

type Notifier struct {
	client    *http.Client
	widgetUri string
}

func NewNotifier(client *http.Client, widgetUri string) *Notifier {
	if widgetUri == "" {
		widgetUri = DefaultWidgetUri
	}

	return &Notifier{
		client:    client,
		widgetUri: widgetUri,
	}
}

func (n *Notifier) Send(r NotificationRequest) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}

	resp, err := n.client.Post(
		n.widgetUri,
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return err
	}

	// log.Println("RESPONSE:", resp)
	if resp.StatusCode != http.StatusCreated {
		return NotificationError{r.Id}
	}

	return nil
}
