package webhooklisten

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var hookListenLog = logging.Logger("hooklisten")

type WebHook struct {
	Header http.Header
	Body   []byte
}

type WebHookHandler struct {
	hookEvents chan *WebHook
}

func NewWebHookHandler(hookEvents chan *WebHook) *WebHookHandler {
	return &WebHookHandler{hookEvents: hookEvents}
}
func (h *WebHookHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c, err := websocket.Accept(w, req, nil)
	if err != nil {
		hookListenLog.Errorf("unable to accept header %v %v ", req.Header, err)
		if c != nil {
			_ = c.Close(websocket.StatusInternalError, "unable to accept")
		}
		return
	}
	defer func() {
		_ = c.Close(websocket.StatusInternalError, "falling")
	}()

	ctx := req.Context()

	hookListenLog.Infof("recevie connect %s", req.RemoteAddr)

	tm := time.NewTicker(time.Second * 60)
	defer tm.Stop()

	ctx = c.CloseRead(ctx)
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	for {
		select {
		case val := <-h.hookEvents:
			err = wsjson.Write(ctx, c, val)
			if err != nil {
				hookListenLog.Errorf("send val to %s fail %v", req.RemoteAddr, err)
				return
			}
		case <-tm.C:
			err := c.Ping(ctx)
			if err != nil {
				hookListenLog.Errorf("ping fail close connection %v", err)
				return
			}
		}
	}
}

func WaitForWebHookEvent(ctx context.Context, remoteURL string) (chan *WebHook, error) {
	webhookCh := make(chan *WebHook, 20)

	listen := func(c *websocket.Conn) error {
		tm := time.NewTicker(time.Second * 10)
		defer tm.Stop()

		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("exit by cancel")
			default:
				webHook := &WebHook{}
				err := wsjson.Read(ctx, c, webHook)
				if err != nil {
					return err
				}
				webhookCh <- webHook
			}
		}
	}

	dial := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext:           dial.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 5)
				c, _, err := websocket.Dial(ctx, remoteURL, &websocket.DialOptions{
					HTTPClient: httpClient,
				})
				if err != nil {
					hookListenLog.Errorf("dial %s fail %v", remoteURL, err)
					continue
				}

				c.SetReadLimit(1 << 32)
				defer c.Close(websocket.StatusInternalError, "falling") //nolint

				err = listen(c)
				if err != nil {
					hookListenLog.Errorf("listen fail wait to restart %v", err)
				}
			}
		}
	}()

	return webhookCh, nil
}
