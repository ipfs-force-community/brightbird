package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ipfs-force-community/brightbird/hookforward/webhooklisten"
	logging "github.com/ipfs/go-log/v2"
)

var mainLog = logging.Logger("main")

var (
	// Listening web server options
	listenAddress      = flag.String("listen", "", "Specify an address to accept HTTP requests, e.g. \":14334\"")
	tlsListenAddress   = flag.String("tls-listen", "", "Specify an address to accept HTTPS requests, e.g. \":14334\"")
	tlsCertificateFile = flag.String("tls-cert", "proxy.crt", "Path to the TLS certificate chain to use")
	tlsPrivateKeyFile  = flag.String("tls-key", "proxy.key", "Path to the private key for the TLS certificate")
)

func usage() {
	logging.SetAllLoggers(logging.LevelDebug)
	fmt.Fprintln(os.Stderr, "Receives git webhooks, keeps a local mirror of the repo up-to-date, then forwards the webhook to another server.")
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func startListening(handler http.Handler, address, tlsAddress, tlsCertFile, tlsKeyFile string) {
	isRunning := false
	if *listenAddress != "" {
		go serveHTTP(address, handler)
		isRunning = true
	}
	if *tlsListenAddress != "" {
		go serveTLS(tlsAddress, tlsCertFile, tlsKeyFile, handler)
		isRunning = true
	}
	if !isRunning {
		log.Fatal("Quitting as neither HTTP nor TLS were enabled")
	}
}

func main() {
	// Get the command line options
	flag.Usage = usage
	flag.Parse()

	hookEvents := make(chan *webhooklisten.WebHook, 20)
	// Start the listening web server
	handler, err := NewHandler(hookEvents)
	if err != nil {
		mainLog.Errorln("Invalid config:", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	mux.Handle("/ws", webhooklisten.NewWebHookHandler(hookEvents))
	startListening(mux, *listenAddress, *tlsListenAddress, *tlsCertificateFile, *tlsPrivateKeyFile)

	// Wait for our eventual death
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	mainLog.Errorln("Shutting down...")
}
