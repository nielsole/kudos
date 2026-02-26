package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

func getRedisConn(redisAddr string, redisPass string, redisDB int) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       redisDB,
	})

	ctx := context.Background()
	pingCtx, pingCancel := context.WithTimeout(ctx, 1*time.Second)

	if _, err := client.Ping(pingCtx).Result(); err != nil {
		slog.Error("failed to connect to Redis", "addr", redisAddr, "err", err)
		os.Exit(1)
	}
	pingCancel()
	return client
}

func main() {
	redisPass := os.Getenv("REDIS_PASSWORD")
	redisAddr := flag.String("redis-address", "localhost:6379", "redis host and port")
	// This inconsistency is introduced by using value based reference for redis_pass
	redisPass = *flag.String("redis-password", redisPass, "redis password")
	listenPort := flag.String("port", ":8080", "Listening port")
	adminListenPort := flag.String("admin-port", ":8081", "Admin listen port; Provides metrics and debugging")
	redisDB := flag.Int("redis-db", 0, "redis db to use")
	flag.Parse()

	client := getRedisConn(*redisAddr, redisPass, *redisDB)

	handler := func(handler func(w http.ResponseWriter, r *http.Request, client *redis.Client)) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, client)
		}
	}

	mr := mux.NewRouter()
	mr.HandleFunc("/metrics", handler(getMetrics)).Methods("GET")
	http.Handle("/", mr)
	// Default ServeMux includes pprof as side-effect of the import.
	go http.ListenAndServe(*adminListenPort, nil)

	r := mux.NewRouter()
	r.HandleFunc("/getclaps", handler(getClapsHandler)).Methods("GET")
	r.HandleFunc("/update-claps", handler(incClapsHandler)).Methods("POST")
	r.HandleFunc("/", handler(getKudosHandler)).Methods("GET")
	r.HandleFunc("/", handler(incKudosHandler)).Methods("POST")

	serveMux := http.NewServeMux()
	serveMux.Handle("/", r)
	server := &http.Server{
		Addr:    *listenPort,
		Handler: serveMux,
	}
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server error", "err", err)
	}
}
