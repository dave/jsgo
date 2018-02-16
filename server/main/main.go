package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dave/jsgo/config"

	"context"
	"os/signal"
	"syscall"

	"github.com/dave/jsgo/server"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	shutdown := make(chan struct{})
	handler := server.New(shutdown)
	s := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

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

	// Signal to all the compile handlers that the server wants to shut down
	close(shutdown)

	ctx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout)
	defer cancel()

	// Wait for all compile jobs to be cancelled
	handler.Waitgroup.Wait()

	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Server stopped")
	}

}
