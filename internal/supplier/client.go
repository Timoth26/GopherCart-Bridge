package supplier

import (
	"log/slog"
	"net/http"
	"time"
)

type Base struct {
	HTTP *http.Client
	Log  *slog.Logger
}

func NewBase(timeout time.Duration, log *slog.Logger) Base {
	return Base{
		HTTP: &http.Client{Timeout: timeout},
		Log:  log,
	}
}
