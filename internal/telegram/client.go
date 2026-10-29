package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client handles interaction with the Telegram Bot API.
type Client interface {
	SendMessage(ctx context.Context, botToken string, chatID int64, text string) error
	GetFileURL(ctx context.Context, botToken, fileID string) (string, error)
}

type client struct {
	httpClient *http.Client
}

// NewClient creates a new Telegram Client.
func NewClient() Client {
	return &client{
		httpClient: &http.Client{},
	}
}

func (c *client) SendMessage(ctx context.Context, botToken string, chatID int64, text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	return nil
}

func (c *client) GetFileURL(ctx context.Context, botToken, fileID string) (string, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", botToken, fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("telegram API error: status %d", resp.StatusCode)
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if !result.Ok || result.Result.FilePath == "" {
		return "", fmt.Errorf("failed to get file path from telegram")
	}

	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", botToken, result.Result.FilePath), nil
}
