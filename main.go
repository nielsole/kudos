package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
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
	// This inconsistency is introduced by using value based reference for redis_pass
	redis_pass = *flag.String("redis-password", redis_pass, "redis password")
	listen_port := flag.String("port", ":8080", "Listening port")
	admin_listen_port := flag.String("admin-port", ":8081", "Admin listen port; Provides metrics and debugging")
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
	http.Handle("/", mr)
	// Default ServeMux includes pprof as side-effect of the import.
	go http.ListenAndServe(*admin_listen_port, nil)

	r := mux.NewRouter()
	r.HandleFunc("/getclaps", handler(getClapsHandler)).Methods("GET")
	r.HandleFunc("/update-claps", handler(incClapsHandler)).Methods("POST")
	r.HandleFunc("/", handler(getKudosHandler)).Methods("GET")
	r.HandleFunc("/", handler(incKudosHandler)).Methods("POST")

	serveMux := http.NewServeMux()
	serveMux.Handle("/", r)
	server := &http.Server{
		Addr:    *listen_port,
		Handler: serveMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		println(err)
	}
}
