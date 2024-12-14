package ping

import "net/http"

type PingHandler interface {
	Ping(w http.ResponseWriter, r *http.Request)
}

type PingHandlerImpl struct{}

func NewPingHandler() PingHandler {
	return &PingHandlerImpl{}
}

func (h *PingHandlerImpl) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("pong"))
}
