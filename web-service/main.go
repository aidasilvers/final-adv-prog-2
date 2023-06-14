package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prongbang/grpc-microservice/web-service/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	authAddress = "localhost:50051"
)

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	// Set up a connection to the server
	authConn, authErr := grpc.Dial(authAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if authErr != nil {
		log.Fatalf("did not connect: %v", authErr)
	}
	defer func(authConn *grpc.ClientConn) { _ = authConn.Close() }(authConn)
	authClient := auth.NewAuthClient(authConn)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Login handler
	mux.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var data UserCredentials
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		resp, err := authClient.Login(ctx, &auth.LoginRequest{Username: data.Username, Password: data.Password})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Printf("Server listening on :8000 port")
	log.Fatal(http.ListenAndServe(":8000", mux))
}
