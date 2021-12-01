package main

import (
	"context"
	"testing"
)

func TestRedisIntegration(t *testing.T) {
	client := get_redis_conn("127.0.0.1:6379", "", 0)
	if urlCount := getUrlCount(context.Background(), "test", client); urlCount != "0" {
		t.Error("Unexpected urlcount")
	}
	if count := getDomainCount(context.Background(), "test.example.com", client); count != "0" {
		t.Error("Unexpected domain count")
	}
	if count := incUrlCount(context.Background(), "http://test.example.com/", client); count != "1" {
		t.Error("Unexpected url count")
	}
	if urlCount := getUrlCount(context.Background(), "http://test.example.com/", client); urlCount != "1" {
		t.Error("Unexpected urlcount", urlCount)
	}
	if urlCount := incUrlCount(context.Background(), "https://bar.example.com/foo", client); urlCount != "1" {
		t.Error("Unexpected urlcount", urlCount)
	}
	if count := getDomainCount(context.Background(), "test.example.com", client); count != "1" {
		t.Error("Unexpected domain count", count)
	}
	if count := getDomainCount(context.Background(), "example.com", client); count != "2" {
		t.Error("Unexpected domain count", count)
	}
	// Note the `á` in example
	if count := getDomainCount(context.Background(), "exámple.com", client); count != "0" {
		t.Error("Unexpected domain count", count)
	}
	if count := getDomainCount(context.Background(), "", client); count != "0" {
		t.Error("Unexpected domain count", count)
	}
}
