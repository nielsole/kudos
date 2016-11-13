package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"gopkg.in/redis.v5"
	"strconv"
)

func createRedisConnection() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	return client
}

func handler1(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	rUrl := r.URL.Query().Get("url")
	if r.Method == "GET" {
		//counter.count++
		counter, err := conn.Get(rUrl).Result()
		if err == redis.Nil {
			counter = "0"
		} else if err != nil {
			println("Something went wrong while retrieving the current value", counter, err)
		}
		w.Write([]byte(counter))
	} else {
		counter, err := conn.Incr(rUrl).Result()
		if err != nil {
			println(err)
		}
		w.Write([]byte(strconv.FormatInt(counter, 10)))
	}
}

func main() {
	conn := createRedisConnection()
	handler := func(w http.ResponseWriter, r *http.Request) {
		handler1(w, r, conn)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil{
		println(err)
	}

}
