package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/ronexlemon/blockscan/internal/api"
	"github.com/ronexlemon/blockscan/internal/storage"
)



func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("[WARN] no .env file found, using system env")
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    db := storage.NewDataBaseConnection()
    repo := storage.Repository{Db: db.DB}
    router := api.NewRouter(&repo)

    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    go func() {
        <-ctx.Done()
        log.Println("[INFO] shutting down API server...")
        server.Shutdown(context.Background())
    }()

    go func() {
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
        <-quit
        log.Println("[INFO] received shutdown signal")
        cancel()
    }()

    log.Println("API running on :8080")
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}