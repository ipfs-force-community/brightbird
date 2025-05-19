package main

import (
	"io"
	"net/http"

	"github.com/google/go-github/v51/github"
	logging "github.com/ipfs/go-log/v2"

	"github.com/ipfs-force-community/brightbird/pkg/webhooklisten"
)

var webhookLog = logging.Logger("webhook")

type Handler struct {
	hookEvents chan *webhooklisten.WebHook
}

func NewHandler(hookEvents chan *webhooklisten.WebHook) (h *Handler, err error) {
	h = &Handler{
		hookEvents: hookEvents,
	}
	return
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Log request
	webhookLog.With(
		"remote", req.RemoteAddr,
		"method", req.Method,
		"url", req.URL,
		"event", req.Header.Get(github.EventTypeHeader),
	).Infof("Incoming webhook")

	data, err := io.ReadAll(req.Body)
	if err != nil {
		webhookLog.Errorf("read body failed %s", err)
		return
	}
	webhookLog.Debug(string(data))

	webhook := &webhooklisten.WebHook{
		Header: req.Header,
		Body:   data,
	}
	select {
	case h.hookEvents <- webhook:
	default:
		webhookLog.Infof("hook event channel is full")
	}
}
