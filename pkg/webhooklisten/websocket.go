package webhooklisten

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log/v2"
)

var hookListenLog = logging.Logger("hooklisten")

type WebHook struct {
	Header http.Header
	Body   []byte
}

type WebHookHandler struct {
	hookEvents chan *WebHook
	upgrader   websocket.Upgrader
}

func NewWebHookHandler(hookEvents chan *WebHook) *WebHookHandler {
	return &WebHookHandler{
		hookEvents: hookEvents,
		upgrader:   websocket.Upgrader{},
	}
}
func (h *WebHookHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c, err := h.upgrader.Upgrade(w, req, nil)
	if err != nil {
		hookListenLog.Errorf("unable to accept header %v %v ", req.Header, err)
		return
	}
	defer c.Close()
	log := hookListenLog.With("target", req.RemoteAddr)
	log.Info("recevie connect")

	tm := time.NewTicker(time.Second * 30)
	defer tm.Stop()
	for {
		select {
		case val := <-h.hookEvents:
			if err := c.WriteJSON(val); err != nil {
				log.Errorf("send val to %s failed: %v", req.RemoteAddr, err)
				return
			}
			log.Info("send hookevent ok")
		case <-tm.C:
			if err := c.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(time.Second*10)); err != nil {
				log.Errorf("[ping] failed : %v", err)
				return
			}
			// log.Debug("ping ok")
		}
	}
}
func formatWsUrl(urlStr string) (*url.URL, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
		break
	default:
		return nil, fmt.Errorf("unknow scheme: %v", u.Scheme)
	}
	return u, nil
}
func WaitForWebHookEvent(ctx context.Context, remoteURL string) (chan *WebHook, error) {
	u, err := formatWsUrl(remoteURL)
	if err != nil {
		return nil, err
	}
	targetUrl := u.String()
	hookListenLog.Infof("web-hook-proxy-url: %v", targetUrl)

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
				err := c.ReadJSON(webHook)
				if err != nil {
					return err
				}
				webhookCh <- webHook
			}
		}
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 5)
				c, _, err := websocket.DefaultDialer.Dial(targetUrl, nil)
				if err != nil {
					hookListenLog.Errorf("dial %s fail %v", targetUrl, err)
					continue
				}
				defer c.Close()
				c.SetReadLimit(1 << 32)

				if err := listen(c); err != nil {
					hookListenLog.Errorf("listen fail wait to restart : %v", err)
				}
			}
		}
	}()

	return webhookCh, nil
}
