package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
)

func main() {

	id := rand.Int()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		_, err := fmt.Fprintf(w, "Hello from server %v!", id)

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
		log.Fatal(err)
	}
}
