package http

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// TODO: register handlers.
	return mux
}
