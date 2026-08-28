package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"couponbatch/internal/config"
	"couponbatch/internal/httpapi"
	"couponbatch/internal/permission"
	"couponbatch/internal/service"
	"couponbatch/internal/store"
)

func main() {
	cfg := config.Load()
	s, err := store.Open(cfg.DataPath)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	actor := permission.DefaultAdmin("operator")
	if err := s.SaveProfile(actor); err != nil {
		log.Fatal(err)
	}
	server := httpapi.New(service.NewBatchService(s), service.NewClaimService(s), actor)
	if os.Getenv("COUPONBATCH_CHECK") == "1" {
		fmt.Println("coupon batch service ready at", cfg.ListenAddress())
		return
	}
	log.Printf("coupon batch service listening on %s", cfg.ListenAddress())
	if err := http.ListenAndServe(cfg.ListenAddress(), server.Handler()); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdown(ctx context.Context, server *http.Server) error { return server.Shutdown(ctx) }
