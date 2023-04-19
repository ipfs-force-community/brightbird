package main

import (
	"log"
	"net/http"
)

type Handler struct {
	mirrorPaths []string
}

func NewHandler(mirrorPaths []string) (h *Handler, err error) {
	h = &Handler{
		mirrorPaths: mirrorPaths,
	}
	return
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Log request
	log.Printf("Incoming webhook from %s %s %s", req.RemoteAddr, req.Method, req.URL)
	for _, mirrorPath := range h.mirrorPaths {
		mirrorPath_copy := mirrorPath
		go func() {
			req, err := http.NewRequest(http.MethodPost, mirrorPath_copy, req.Body)
			if err != nil {
				log.Printf("forward %s hook fail %v", mirrorPath_copy, err)
			}
			_, err = http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("send hook request %s fail %v", mirrorPath_copy, err)
			}
		}()
	}

}
