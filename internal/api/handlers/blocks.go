package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)


func LatestBlockHandler(w http.ResponseWriter, r *http.Request){
	blocks := []string{"block1","block2"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(blocks)
}

func GetBlockByNumber(w http.ResponseWriter, r *http.Request){
	number := chi.URLParam(r,"number")

	block := map[string]string{
		"block_number":number,
	}

	json.NewEncoder(w).Encode(block)
}