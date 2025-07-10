package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodPost, "/generate", app.GenerateHandler)
	router.HandlerFunc(http.MethodPost, "/generate-stream", app.GenerateStreamHandler)
	return router
}
