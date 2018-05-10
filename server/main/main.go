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

	var mainServer, devServer *http.Server

	shutdown := make(chan struct{})
	handler := server.New(shutdown)

	if config.DEV {
		mainServer = &http.Server{Addr: ":8080", Handler: handler}
		devServer = &http.Server{Addr: ":8081", Handler: handler}
	} else {
		port := "8080"
		if fromEnv := os.Getenv("PORT"); fromEnv != "" {
			port = fromEnv
		}
		mainServer = &http.Server{Addr: ":" + port, Handler: handler}
	}

	go func() {
		log.Print("Listening on " + mainServer.Addr)
		if err := mainServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if config.DEV {
		go func() {
			log.Print("Listening on " + devServer.Addr)
			if err := devServer.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

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

	if err := mainServer.Shutdown(ctx); err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		log.Println("Main server stopped")
	}

	if config.DEV {
		if err := devServer.Shutdown(ctx); err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			log.Println("Dev server stopped")
		}
	}

}
