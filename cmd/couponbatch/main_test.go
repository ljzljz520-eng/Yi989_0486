package main

import (
	"context"
	"net/http"
	"testing"
)

func TestShutdown(t *testing.T) {
	server := &http.Server{}
	if err := shutdown(context.Background(), server); err != nil {
		t.Fatal(err)
	}
}
