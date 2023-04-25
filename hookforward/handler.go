package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hunjixin/brightbird/hookforward/webhooklisten"
	logging "github.com/ipfs/go-log/v2"
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
	webhookLog.Infof("Incoming webhook from %s %s %s", req.RemoteAddr, req.Method, req.URL)

	data, err := io.ReadAll(req.Body)
	if err != nil {
		webhookLog.Errorf("read body fail %s", err)
		return
	}

	fmt.Println(string(data))

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
