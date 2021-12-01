package main

import "testing"

func assert(t *testing.T, input bool, expected bool) {
	if input != expected {
		t.Error("Assertion failed")
	}
}
func TestIsSubdomainOf(t *testing.T) {
	assert(t, isSubdomainOf("foo.example.com", "example.com"), true)
	assert(t, isSubdomainOf("foo-example.com", "example.com"), false)
	assert(t, isSubdomainOf("foo-example.com", ""), false)
	assert(t, isSubdomainOf("foo.example.com", ".example.com"), false)
	assert(t, isSubdomainOf("example.com", "example.com"), true)
	assert(t, isSubdomainOf("localhost", "localhost"), true)
}
