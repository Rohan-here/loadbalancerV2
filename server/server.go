package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "high")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func registerWithLoadBalancer(loadBalancerURL, serverURL string) error {
	data := map[string]string{
		"url": serverURL,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(loadBalancerURL+"/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register with load balancer: %s", resp.Status)
	}

	fmt.Println("Registered with load balancer")
	return nil
}

func main() {
	portPtr := flag.String("port", "8081", "Port for the server to run on")
	flag.Parse()

	port := ":" + *portPtr
	serverURL := "http://localhost" + port
	loadBalancerURL := "http://localhost:8080"

	err := registerWithLoadBalancer(loadBalancerURL, serverURL)
	if err != nil {
		log.Fatalf("Error registering with load balancer: %v", err)
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/ping", pingHandler)

	fmt.Printf("Server started on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
