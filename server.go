package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v9"
	"github.com/gorilla/mux"
	"os"
)

var ctx = context.Background()
var redisClient *redis.Client

func initRedis() {
	redisHost := os.Getenv("REDIS_HOST") // Get from environment variables
	redisPort := os.Getenv("REDIS_PORT")

	if redisHost == "" {
		redisHost = "localhost" // Fallback for local testing
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // No password by default
		DB:       0,  // Default Redis DB
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}
	fmt.Println("Connected to Redis at", redisAddr)
}

func setKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	err := redisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		http.Error(w, "Failed to set key", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Key %s set successfully", key)
}

func getKey(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	value, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to get key", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Value: %s", value)
}

func main() {
	initRedis()

	r := mux.NewRouter()
	r.HandleFunc("/set", setKey).Methods("POST")
	r.HandleFunc("/get/{key}", getKey).Methods("GET")

	apiPort := os.Getenv("API_PORT") // Get port from environment
	if apiPort == "" {
		apiPort = "9091" // Default fallback
	}

	fmt.Println("Server is running on port", apiPort)
	log.Fatal(http.ListenAndServe(":"+apiPort, r))
}

