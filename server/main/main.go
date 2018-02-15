package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server"

	"context"
	"os/signal"
	"syscall"

	"golang.org/x/net/websocket"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	h := &handler{
		mux: http.NewServeMux(),
	}
	s := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	h.mux.Handle("/", http.HandlerFunc(server.PageHandler))
	h.mux.Handle("/_ws/", websocket.Handler(server.SocketHandler))
	h.mux.Handle("/favicon.ico", http.HandlerFunc(server.IconHandler))
	h.mux.Handle("/compile.css", http.HandlerFunc(server.CssHandler))
	h.mux.Handle("/_ah/health", http.HandlerFunc(server.HealthCheckHandler))

	go func() {
		log.Print("Listening on port " + port)

		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Set up graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Server stopped")
	}

}

type handler struct {
	mux *http.ServeMux
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}
