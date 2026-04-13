package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ronexlemon/blockscan/internal/api/handlers"
	"github.com/ronexlemon/blockscan/internal/storage"
)


func NewRouter(repo *storage.Repository)http.Handler{
	r :=chi.NewRouter()
    h:= handlers.NewHandler(repo)
	r.Get("/health",handlers.HealthHandler)
	r.Route("/blocks", func(r chi.Router) {
		r.Get("/", h.LatestBlocksHandler)
		r.Get("/latest", h.GetLatestBlockHandler)
    r.Get("/{number}",  h.GetBlockHandler)
    r.Get("/{number}/transactions", h.TransactionsByBlock)
	})
	r.Route("/tx", func(r chi.Router) {
		r.Get("/",h.LatestTxHandler)
		r.Get("/{address}", h.GetAddressTransactions)
	})
	return r

}