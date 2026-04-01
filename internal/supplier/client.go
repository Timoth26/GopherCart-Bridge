package supplier

import (
	"log/slog"
	"net/http"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	log        *slog.Logger
}

func NewClient(httpClient *http.Client, baseURL string, log *slog.Logger) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		log:        log,
	}
}
