package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func main() {
	redis_pass := os.Getenv("REDIS_PASSWORD")
	redis_addr := flag.String("redis-address", "localhost:6379", "redis host and port")
	redis_pass = *flag.String("redis-password", redis_pass, "redis password")
	listen_port := flag.String("port", ":8080", "Listening port")
	redis_db := flag.Int("redis-db", 0, "redis db to use")
	flag.Parse()

	// Connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     *redis_addr,
		Password: redis_pass, // no password set
		DB:       *redis_db,  // use default DB
	})

	ctx := context.Background()
	ping_ctx, _ := context.WithTimeout(ctx, 1*time.Second)

	if pong, err := client.Ping(ping_ctx).Result(); err != nil {
		fmt.Println(pong, err)
		os.Exit(1)
	}

	// TODO Pass redis client via context
	handler := func(handler func(w http.ResponseWriter, r *http.Request, client *redis.Client)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, client)
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/getclaps", handler(getClapsHandler)).Methods("GET")
	r.HandleFunc("/update-claps", handler(incClapsHandler)).Methods("POST")
	r.HandleFunc("/", handler(getKudosHandler)).Methods("GET")
	r.HandleFunc("/", handler(incKudosHandler)).Methods("POST")

	http.Handle("/", r)
	err := http.ListenAndServe(*listen_port, nil)
	if err != nil {
		println(err)
	}
}
