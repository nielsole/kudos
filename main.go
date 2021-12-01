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

func get_redis_conn(redis_addr string, redis_pass string, redis_db int) *redis.Client {
	// Connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: redis_pass,
		DB:       redis_db,
	})

	ctx := context.Background()
	ping_ctx, ping_cancel := context.WithTimeout(ctx, 1*time.Second)

	if pong, err := client.Ping(ping_ctx).Result(); err != nil {
		fmt.Println(pong, err)
		os.Exit(1)
	}
	ping_cancel()
	return client
}

func main() {
	redis_pass := os.Getenv("REDIS_PASSWORD")
	redis_addr := flag.String("redis-address", "localhost:6379", "redis host and port")
	redis_pass = *flag.String("redis-password", redis_pass, "redis password")
	listen_port := flag.String("port", ":8080", "Listening port")
	metrics_listen_port := flag.String("metrics-port", ":8081", "Prometheus metrics listen port")
	redis_db := flag.Int("redis-db", 0, "redis db to use")
	flag.Parse()

	client := get_redis_conn(*redis_addr, redis_pass, *redis_db)

	handler := func(handler func(w http.ResponseWriter, r *http.Request, client *redis.Client)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, client)
		}
	}

	mr := mux.NewRouter()
	mr.HandleFunc("/metrics", handler(getMetrics)).Methods("GET")
	go http.ListenAndServe(*metrics_listen_port, nil)

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
