package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

const (
	CONTENT_TYPE_JSON        = "application/json"
	CONTENT_TYPE_HEADER_NAME = "Content-Type"
)

type GreetResponse struct {
	Message string `json:"message"`
}

type QueryResponseList struct {
	Hostname     string          `json:"from"`
	QueryResults []QueryResponse `json:"results"`
}

type QueryResponse struct {
	QueryRequest  map[string]string `json:"request"`
	QueryResponse any               `json:"response"`
}

type HttpResponse struct {
	Response any `json:"response"`
}

func DoHttpRequest(ctx context.Context, url string) any {

	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil,
	)

	if err != nil {
		return fmt.Sprint("Failed to create the request:", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Sprint("Failed to perform the request:", err)
	}

	defer response.Body.Close()

	contentType := response.Header.Get("Content-Type")
	if contentType != CONTENT_TYPE_JSON {

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Sprint("Failed to read response.Body:", err)
		}

		return string(body)
	}

	var object any

	if err := json.NewDecoder(response.Body).Decode(&object); err != nil {
		return fmt.Sprint("Failed to decode response:", err)
	}

	return object
}

func WriteJsonResponse(w http.ResponseWriter, object any) {

	w.Header().Set(CONTENT_TYPE_HEADER_NAME, CONTENT_TYPE_JSON)

	if err := json.NewEncoder(w).Encode(object); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetHostname() string {

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Failed to get hostname:", err)
		hostname = "unknown"
	}

	return hostname
}

func HandleQuery(w http.ResponseWriter, r *http.Request) {

	mutex := sync.Mutex{}
	responses := []QueryResponse{}

	wg := sync.WaitGroup{}

	appendResponse := func(response QueryResponse) {

		mutex.Lock()
		defer mutex.Unlock()

		responses = append(responses, response)
	}

	for key, values := range r.URL.Query() {
		for _, value := range values {

			request := map[string]string{
				key: value,
			}

			wg.Add(1)

			switch key {
			case "url":
				go func() {

					defer wg.Done()

					httpResponse := DoHttpRequest(r.Context(), value)

					appendResponse(QueryResponse{
						QueryRequest:  request,
						QueryResponse: &httpResponse,
					})
				}()

			case "var":
				go func() {

					defer wg.Done()

					appendResponse(QueryResponse{
						QueryRequest:  request,
						QueryResponse: os.Getenv(value),
					})
				}()

			default:
				go func() {

					defer wg.Done()

					appendResponse(QueryResponse{
						QueryRequest: request,
						QueryResponse: fmt.Sprintf(
							"Invalid request, available request: %v",
							[]string{"url", "env"},
						),
					})
				}()
			}
		}
	}

	wg.Wait()

	response := QueryResponseList{
		Hostname:     GetHostname(),
		QueryResults: responses,
	}

	WriteJsonResponse(w, &response)
}

func main() {

	http.HandleFunc("/greet", func(w http.ResponseWriter, r *http.Request) {
		response := GreetResponse{fmt.Sprintf("Hello from server %v!", GetHostname())}
		WriteJsonResponse(w, &response)
	})

	http.HandleFunc("/query", HandleQuery)
	http.Handle("/", http.RedirectHandler("/greet", http.StatusMovedPermanently))

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	address := ":" + port

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal("Failed to listen to address:", err)
	}
}
