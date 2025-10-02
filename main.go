package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/exp/slog"
	"github.com/mstiehr-dev/k8s-controller/internal/helpers"
)

var (
	port    int
	tlsKey  string
	tlsCert string
)



func main() {
	flag.IntVar(&port, "port", 9093, "Admisson controller port")
	flag.StringVar(&tlsKey, "tls-key", "/etc/webhook/certs/tls.key", "Private key for TLS")
	flag.StringVar(&tlsCert, "tls-crt", "/etc/webhook/certs/tls.crt", "TLS certificate")
	flag.Parse()
	slog.Info("loading certs..")
	certs, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		slog.Error("unable to load certs","error", err)
	}

	http.HandleFunc("/mutate", helpers.Mutate)

	slog.Info("successfully loaded certs. Starting server...", "port", port)
	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{certs},
		},
	}

	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Panic(err)
	}

}
