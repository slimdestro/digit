package handler

import (
	"digit/internal/model"
	repository "digit/internal/repo"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type BookHandler struct {
	repo repository.BookRepository
}

func NewBookHandler(r repository.BookRepository) *BookHandler {
	return &BookHandler{repo: r}
}

func (h *BookHandler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.repo.List()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, books, http.StatusOK)
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var req model.CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Author) == "" {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}
	id, err := h.repo.Create(req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]uint64{"id": id}, http.StatusCreated)
}

func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	book, err := h.repo.Get(id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if book == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, book, http.StatusOK)
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	var req model.UpdateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Author) == "" {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}
	err := h.repo.Update(id, req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookHandler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := getIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	err := h.repo.Delete(id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func getIDFromPath(path string) (uint64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 {
		return 0, false
	}
	id, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
