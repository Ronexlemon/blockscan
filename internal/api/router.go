package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ronexlemon/blockscan/internal/api/handlers"
)


func NewRouter()http.Handler{
	r :=chi.NewRouter()

	r.Get("/health",handlers.HealthHandler)
	r.Route("/blocks", func(r chi.Router) {
		r.Get("/", handlers.LatestBlockHandler)
		r.Get("/{number}", handlers.GetBlockByNumber)
	})
	r.Route("/tx", func(r chi.Router) {
		r.Get("/",handlers.LatestTxHandler)
		r.Get("/{hash}", handlers.GetTransaction)
	})
	return r

}