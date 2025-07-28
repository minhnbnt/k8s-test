package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Failed to get hostname:", err)
		hostname = "unknown"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		_, err := fmt.Fprintf(w, "Hello from server %v!", hostname)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	address := ":" + port

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal("Failed to listen to address:", err)
	}
}
