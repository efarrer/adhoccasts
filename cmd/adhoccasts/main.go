package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/efarrer/adhoccasts/filesystem"
	"github.com/efarrer/adhoccasts/server"
)

func main() {
	urlStr := flag.String("url", "http://localhost:8080", "The base url for the podcasts")
	port := flag.Int("port", 8080, "The port to listen on")
	dir := flag.String("dir", "./", "The directory where the adhoc podcasts are stored")
	flag.Parse()

	fs := filesystem.Default()
	serveMux, err := server.NewServeMux(*urlStr, *dir, fs)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Publishing directories under %s as podcasts on %s. Listening on %s\n", *dir, *urlStr, addr)
	err = serveMux.ListenAndServe(addr)
	if err != nil {
		log.Fatal(err)
	}
}
