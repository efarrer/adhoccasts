package handler

import (
	"fmt"
	"net/http"
	"time"
)

type Logger struct{}

func NewLogger() Logger {
	return Logger{}
}

func (l Logger) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s Serving %s\n", time.Now().Format(time.DateTime), r.URL.String())
		handler.ServeHTTP(w, r)
	})
}
