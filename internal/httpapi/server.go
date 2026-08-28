package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"couponbatch/internal/model"
	"couponbatch/internal/service"
)

type Server struct {
	Batches *service.BatchService
	Claims  *service.ClaimService
	Actor   model.Profile
}

func New(batches *service.BatchService, claims *service.ClaimService, actor model.Profile) *Server {
	return &Server{Batches: batches, Claims: claims, Actor: actor}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/api/batches", s.batches)
	mux.HandleFunc("/api/batches/", s.batchByID)
	mux.HandleFunc("/api/claims", s.claim)
	return withJSON(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) batches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		page, err := s.Batches.ListBatches(r.Context(), s.Actor, model.BatchFilter{Query: r.URL.Query().Get("q"), Status: model.BatchStatus(r.URL.Query().Get("status")), Available: r.URL.Query().Get("available") == "true"}, 0)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	case http.MethodPost:
		var batch model.Batch
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			writeError(w, model.ErrInvalidRequest)
			return
		}
		created, err := s.Batches.CreateBatch(r.Context(), s.Actor, batch)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) batchByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/batches/")
	if id == "" {
		writeError(w, model.ErrInvalidRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		batch, err := s.Batches.GetBatch(r.Context(), s.Actor, id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, batch)
	case http.MethodPut:
		candidate, err := decodeUpdate(r)
		if err != nil {
			writeError(w, err)
			return
		}
		candidate.ID = id
		updated, err := s.Batches.UpdateBatch(r.Context(), s.Actor, candidate)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	case http.MethodDelete:
		if err := s.Batches.DeleteBatch(r.Context(), s.Actor, id); err != nil {
			writeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) claim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var request struct {
		BatchID    string `json:"batch_id"`
		CustomerID string `json:"customer_id"`
		Action     string `json:"action"`
		RecordID   string `json:"record_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, model.ErrInvalidRequest)
		return
	}
	if request.Action == "confirm" {
		record, err := s.Claims.Confirm(r.Context(), s.Actor, request.RecordID)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, record)
		return
	}
	if request.Action == "reclaim" {
		record, err := s.Claims.Reclaim(r.Context(), s.Actor, request.RecordID, "api reclaim")
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, record)
		return
	}
	record, err := s.Claims.Claim(r.Context(), s.Actor, request.BatchID, request.CustomerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if err == model.ErrUnauthorized {
		status = http.StatusForbidden
	}
	if err == model.ErrNotFound {
		status = http.StatusNotFound
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func ServerContext(r *http.Request) context.Context { return r.Context() }
