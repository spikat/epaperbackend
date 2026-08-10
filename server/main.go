package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/httpx"
	"github.com/jonathanribas/epaperbackend/pkg/registry"
	"github.com/jonathanribas/epaperbackend/server/debug"

	_ "github.com/jonathanribas/epaperbackend/services/weather"
)

func main() {
	debugFlag := flag.Bool("debug", false, "enable debug UI server")
	flag.Parse()

	cfg := config.Load()
	if *debugFlag {
		cfg.Debug = true
	}

	if err := registry.Bootstrap(cfg); err != nil {
		log.Fatalf("bootstrap services: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /", handleRoot)
	if err := registry.RegisterRoutes(mux); err != nil {
		log.Fatalf("register routes: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	mainServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("main API listening on :%d", cfg.Port)
		if err := mainServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("main server: %v", err)
		}
	}()

	if cfg.Debug {
		dbg := debug.New(cfg)
		go func() {
			log.Printf("debug UI listening on :%d", cfg.DebugPort)
			if err := dbg.ListenAndServe(ctx); err != nil && err != http.ErrServerClosed {
				log.Fatalf("debug server: %v", err)
			}
		}()
	}

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = mainServer.Shutdown(shutdownCtx)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"services": registry.Health(r.Context()),
	})
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"name":     "epaperbackend",
		"services": registry.Infos(),
	})
}
