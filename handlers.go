package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	redis "github.com/go-redis/redis/v8"
)

func incURLCount(ctx context.Context, rURL string, conn *redis.Client) string {
	counter, err := conn.Incr(ctx, rURL).Result()
	if err != nil {
		slog.Error("failed to increment counter", "url", rURL, "err", err)
	}
	return strconv.FormatInt(counter, 10)
}

func getURLCount(ctx context.Context, rURL string, conn *redis.Client) string {
	counter, err := conn.Get(ctx, rURL).Result()
	if err == redis.Nil {
		counter = "0"
	} else if err != nil {
		slog.Error("failed to get URL count", "url", rURL, "err", err)
	}
	return counter
}

// Returns the sum of all kudos counters for the domain and its subdomains.
func getDomainCount(ctx context.Context, domain string, conn *redis.Client) string {
	kudos := int64(0)
	var keys []string
	var cursor uint64 = 0
	var err error
	for {
		slog.Debug("scanning", "cursor", cursor, "kudos", kudos)
		// TODO This matches all of:
		// * foo-domain
		// * foo.domain
		// * domain
		// but we only want to match the last two.
		scanRes := conn.Scan(ctx, cursor, "*"+domain+"/*", 100)
		keys, cursor, err = scanRes.Result()
		if err != nil {
			// TODO Set status code to 500
			return "0"
		}
		for _, key := range keys {
			urlObject, err := url.Parse(key)
			if err == nil {
				if !isSubdomainOf(urlObject.Host, domain) {
					continue
				}
			}
			slog.Debug("retrieving url", "key", key)
			urlKudos, err := conn.Get(ctx, key).Result()
			if err != nil {
				slog.Warn("could not retrieve key", "key", key, "err", err)
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
			slog.Error("scan error", "err", err)
		}
		for _, key := range keys {
			slog.Debug("retrieving url", "key", key)
			urlKudos, err := conn.Get(ctx, key).Result()
			if err != nil {
				slog.Warn("could not retrieve key", "key", key, "err", err)
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
	rURL := r.Header.Get("Referrer")
	count := getURLCount(ctx, rURL, conn)
	w.Write([]byte(count))
}

func incClapsHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	rURL := r.Header.Get("Referrer")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if rURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), 400)
		return
	}
	kudos := incURLCount(ctx, rURL, conn)
	w.Write([]byte(kudos))
}

func getKudosHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	var count string
	rURL := r.URL.Query().Get("url")
	if rURL != "" {
		count = getURLCount(ctx, rURL, conn)
	} else {
		domain := r.URL.Query().Get("domain")
		count = getDomainCount(ctx, domain, conn)
	}
	w.Write([]byte(count))
}

func incKudosHandler(w http.ResponseWriter, r *http.Request, conn *redis.Client) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancelFunc()
	rURL := r.URL.Query().Get("url")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if rURL == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), 400)
		return
	}
	kudos := incURLCount(ctx, rURL, conn)
	w.Write([]byte(kudos))
}
