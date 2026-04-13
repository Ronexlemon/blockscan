package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ronexlemon/blockscan/internal/storage"
)



type Handler struct {
	Repo *storage.Repository
}

func NewHandler(repo *storage.Repository) *Handler {
	return &Handler{Repo: repo}
}


func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) LatestTxHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := h.Repo.GetLatestTransactions(r.Context(), storage.Pagination{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) TransactionsByBlock(w http.ResponseWriter, r *http.Request) {
	blockStr := chi.URLParam(r, "number")
	if blockStr == "" {
		writeError(w, http.StatusBadRequest, "block number is required")
		return
	}

	blockNumber, err := strconv.ParseUint(blockStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block number")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := h.Repo.GetBlockTransactionsPaginated(r.Context(), blockNumber, storage.Pagination{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}





func (h *Handler) GetAddressTransactions(w http.ResponseWriter, r *http.Request) {
	address := chi.URLParam(r, "address")
	if address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := h.Repo.GetTransactionsByAddress(r.Context(), address, storage.Pagination{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

