package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jefjesuswt/fleetings-tracker/internal/parser"
)

type Client struct {
	WebhookURL string
	HTTPClient *http.Client
}

func NewClient(webhookUrl string) *Client {
	return &Client{
		WebhookURL: webhookUrl,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Send(reminder parser.Reminder, phaseTitle string) error {

	mensaje := fmt.Sprintf("%s\n\n📝 %s\n\n📂 _%s_", phaseTitle, reminder.Content, reminder.File)

	payload := NotificationPayload{
		Text: mensaje,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error serializando payload del webhook: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("error creando request del webhook: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red enviando webhook: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("el webhook respondió con status inesperado: %d", res.StatusCode)
	}

	return nil
}
