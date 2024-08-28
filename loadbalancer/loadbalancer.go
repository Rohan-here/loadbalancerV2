package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

type LoadBalancer struct {
	mu      sync.Mutex
	servers []string
	index   int
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		servers: make([]string, 0),
		index:   0,
	}
}

func (lb *LoadBalancer) RegisterServer(serverURL string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, serverURL)
	fmt.Printf("Registered server: %s\n", serverURL)
}

func (lb *LoadBalancer) GetNextServer() (string, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.servers) == 0 {
		return "", fmt.Errorf("no servers available")
	}

	server := lb.servers[lb.index]
	lb.index = (lb.index + 1) % len(lb.servers)
	return server, nil
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server, err := lb.GetNextServer()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	resp, err := http.Get(server + r.URL.Path)
	if err != nil {
		http.Error(w, "server error", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read response from server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (lb *LoadBalancer) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	lb.RegisterServer(data.URL)
	w.WriteHeader(http.StatusOK)
}

func main() {
	lb := NewLoadBalancer()

	http.HandleFunc("/register", lb.RegisterHandler)
	http.Handle("/", lb)

	fmt.Println("Load balancer started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
