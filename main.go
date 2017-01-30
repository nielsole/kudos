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
		Addr:     "localhost:6379",
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
	if rUrl == ""{
		domain := r.URL.Query().Get("domain")
		if domain == ""{
			w.WriteHeader(400)
			w.Write([]byte("Use either 'url' or 'domain' as GET parameters"))
			return
		}
		//else
		if r.Method != "GET"{
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		kudos := uint64(0)
		var keys []string
		var cursor uint64 = 0
		var err error
		for{
			println("Scanning...", cursor, kudos)
			scanRes := conn.Scan(cursor,"*//" + domain + "/*", 100)
			keys, cursor, err = scanRes.Result()
			if err != nil{
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			for _, key := range keys{
				println("Retrieving individual url", key)
				urlKudos, err1 := conn.Get(key).Result()
				if err1 != nil{
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				kudosIncr, _ := strconv.ParseUint(urlKudos, 10, 64)
				kudos += kudosIncr
			}
			if cursor == 0{
				break
			}
		}
		w.Write([]byte(strconv.FormatUint(kudos, 10)))
		return
	}
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