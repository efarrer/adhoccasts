package handler

import (
	"log"
	"net/http"
)

func httpError(w http.ResponseWriter, status int, msg string, err error) {
	log.Printf("Http error %d %s:%s\n", status, msg, err)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}
