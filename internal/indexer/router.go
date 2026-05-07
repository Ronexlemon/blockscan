package indexer

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func NewIndexerRouter()http.Handler{
	r :=chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	return r

}