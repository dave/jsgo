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

	var mainServer, dev1Server, dev2Server, dev3Server *http.Server

	shutdown := make(chan struct{})
	handler := server.New(shutdown)

	if config.DEV {
		mainServer = &http.Server{Addr: ":8080", Handler: handler}
		dev1Server = &http.Server{Addr: ":8081", Handler: handler}
		dev2Server = &http.Server{Addr: ":8082", Handler: handler}
		dev3Server = &http.Server{Addr: ":8083", Handler: handler}
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
			log.Print("Listening on " + dev1Server.Addr)
			if err := dev1Server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()

		go func() {
			log.Print("Listening on " + dev2Server.Addr)
			if err := dev2Server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()

		go func() {
			log.Print("Listening on " + dev3Server.Addr)
			if err := dev3Server.ListenAndServe(); err != http.ErrServerClosed {
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
		if err := dev1Server.Shutdown(ctx); err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			log.Println("Dev 1 server stopped")
		}

		if err := dev2Server.Shutdown(ctx); err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			log.Println("Dev 2 server stopped")
		}

		if err := dev3Server.Shutdown(ctx); err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			log.Println("Dev 3 server stopped")
		}
	}

}
