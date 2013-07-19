package main

import (
	"flag"
	"github.com/discordianfish/wizbal/lb"
	"log"
)

var (
	listen = flag.String("listen", ":8080", "Port to listen on.")
	host   = flag.String("host", ".app.example.com", "Strip that portion from host header.")
	domain = flag.String("domain", "srv.example.com", "Resolve striped host header relative to this domain.")
)

func init() {
	flag.Parse()
}

func main() {
	wizbal := lb.New(*listen, *host, *domain)
	log.Printf("Started")
	if err := wizbal.ListenAndServe(); err != nil {
		log.Fatalf("Listening failed: %s", err)
	}
}
