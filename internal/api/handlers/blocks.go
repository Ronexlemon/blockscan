package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ronexlemon/blockscan/internal/storage"
)

func (h *Handler) LatestBlocksHandler(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	result, err := h.Repo.GetLatestBlocks(r.Context(), storage.Pagination{
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetBlockHandler(w http.ResponseWriter, r *http.Request) {
	blockStr := chi.URLParam(r, "number")
	blockNumber, err := strconv.ParseUint(blockStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid block number")
		return
	}

	block, err := h.Repo.GetBlockByNumber(r.Context(), blockNumber)
	if err != nil {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}

	writeJSON(w, http.StatusOK, block)
}

func (h *Handler) GetLatestBlockHandler(w http.ResponseWriter, r *http.Request) {
	block, err := h.Repo.GetLatestBlock(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, block)
}