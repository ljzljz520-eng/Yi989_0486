package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"couponbatch/internal/permission"
	"couponbatch/internal/service"
	"couponbatch/internal/store"
)

func TestHTTPHealthAndInvalidBatch(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "api.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	actor := permission.DefaultAdmin("admin")
	handler := New(service.NewBatchService(s), service.NewClaimService(s), actor).Handler()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusOK {
		t.Fatal(response.Code)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/batches", bytes.NewBufferString("bad")))
	if response.Code != http.StatusBadRequest {
		t.Fatal(response.Code)
	}
}
