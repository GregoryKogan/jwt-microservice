package ping

import (
	"log/slog"
	"net/http"
)

type PingHandler interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type PingHandlerImpl struct{}

func NewPingHandler() PingHandler {
	return &PingHandlerImpl{}
}

func (h *PingHandlerImpl) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("Invalid method for ping", "method", r.Method, "path", r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	slog.Debug("Ping request received", "remote_addr", r.RemoteAddr)
	w.Write([]byte("pong"))
}
