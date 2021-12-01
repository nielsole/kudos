package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	redis "github.com/go-redis/redis/v8"
)

func incUrlCount(ctx context.Context, rUrl string, conn *redis.Client) string {
	counter, err := conn.Incr(ctx, rUrl).Result()
	if err != nil {
		println(err)
	}
	return strconv.FormatInt(counter, 10)
}

func getUrlCount(ctx context.Context, rUrl string, conn *redis.Client) string {
	counter, err := conn.Get(ctx, rUrl).Result()
	if err == redis.Nil {
		counter = "0"
	} else if err != nil {
		println("Something went wrong while retrieving the current value", counter, err)
	}
	return counter
}

func getDomainCount(ctx context.Context, domain string, conn *redis.Client) string {
	kudos := int64(0)
	var keys []string
	var cursor uint64 = 0
	var err error
	for {
		println("Scanning...", cursor, kudos)
		// TODO this matches any domain with the same prefix.
		// Of course we only want actual domains to match, so we would need to parse the keys coming from redis before counting them.
		scanRes := conn.Scan(ctx, cursor, "*"+domain+"/*", 100)
		keys, cursor, err = scanRes.Result()
		if err != nil {
			// TODO Set status code to 500
			return "0"
		}
		for _, key := range keys {
			println("Retrieving individual url", key)
			urlKudos, error_url_redis := conn.Get(ctx, key).Result()
			if error_url_redis != nil {
				println("Could not retrieve key: ", key)
			}
			kudosIncr, _ := strconv.ParseInt(urlKudos, 10, 64)
			kudos += kudosIncr
		}
		if cursor == 0 {
			break
		}
	}
	return strconv.FormatInt(kudos, 10)
}

func getMetrics(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10000*time.Millisecond)
	defer cancelFunc()
	var keys []string
	var cursor uint64 = 0
	var err error
	scanRes := conn.Scan(ctx, cursor, "*", 100)
	keys, cursor, err = scanRes.Result()
	for {
		if err != nil {
			println("Error during scanning: ", err)
		}
		for _, key := range keys {
			println("Retrieving individual url", key)
			urlKudos, error_url_redis := conn.Get(ctx, key).Result()
			if error_url_redis != nil {
				println("Could not retrieve key: ", key)
			}
			u, err := url.Parse(key)
			domain := ""
			if err == nil {
				domain = u.Host
			}
			kudos, _ := strconv.ParseInt(urlKudos, 10, 64)
			w.Write([]byte(fmt.Sprintf("kudos{url=\"%s\",domain=\"%s\"} %d\n", strings.ReplaceAll(key, "\"", "\\\""), domain, kudos)))
		}
		if cursor == 0 {
			break
		}
		scanRes := conn.Scan(ctx, cursor, "*", 100)
		keys, cursor, err = scanRes.Result()
	}
}

func getClapsHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	rUrl := r.Header.Get("Referrer")
	count := getUrlCount(ctx, rUrl, conn)
	w.Write([]byte(count))
}

func incClapsHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	rUrl := r.Header.Get("Referrer")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if rUrl == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), 400)
		return
	}
	kudos := incUrlCount(ctx, rUrl, conn)
	w.Write([]byte(kudos))
}

func getKudosHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var count string
	rUrl := r.URL.Query().Get("url")
	if rUrl != "" {
		count = getUrlCount(ctx, rUrl, conn)
	} else {
		domain := r.URL.Query().Get("domain")
		count = getDomainCount(ctx, domain, conn)
	}
	w.Write([]byte(count))
}

func incKudosHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	rUrl := r.URL.Query().Get("url")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if rUrl == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), 400)
		return
	}
	kudos := incUrlCount(ctx, rUrl, conn)
	w.Write([]byte(kudos))
}
