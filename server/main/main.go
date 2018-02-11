package main

import (
	"log"
	"net/http"
	"os"

	"github.com/dave/jsgo/server"
	"golang.org/x/net/websocket"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	http.HandleFunc("/", server.Handler)
	http.Handle("/_ws/", websocket.Handler(server.SocketHandler))
	http.HandleFunc("/favicon.ico", server.FaviconHandler)
	http.HandleFunc("/_ah/health", server.HealthCheckHandler)
	log.Print("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
