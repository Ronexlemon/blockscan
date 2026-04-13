package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Transaction struct {
	Hash   string `json:"hash"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}
func LatestTxHandler(w http.ResponseWriter, r *http.Request) {
	txs := []Transaction{
		{
			Hash:   "0xabc",
			From:   "0x1223",
			To:     "0x345555",
			Amount: 4,
		},
	}
	json.NewEncoder(w).Encode(txs)
}

func GetTransaction(w http.ResponseWriter, r *http.Request){
	txhash := chi.URLParam(r,"hash")

	result:=Transaction{
		Hash: txhash,
		Amount: 10,
		From: "0x1234566666666666699999283476455464746745655",
		To: "0x5459dfr738675835474562385689497692135326646",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}