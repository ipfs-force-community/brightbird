package main

import (
	"log"
	"net/http"
)

func serveHTTP(address string, handler http.Handler) {
	log.Println("Listening for HTTP requests at", address)
	err := http.ListenAndServe(address, handler)
	if err != nil {
		log.Fatal(err)
	}
}

func serveTLS(address, certFile, keyFile string, handler http.Handler) {
	log.Println("Listening for TLS requests at ", address)
	err := http.ListenAndServeTLS(address, certFile, keyFile, handler)
	if err != nil {
		log.Fatal(err)
	}
}
