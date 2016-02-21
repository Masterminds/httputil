package httputil

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestResolver(t *testing.T) {
	r := NewResolver([]string{
		"* /loomings",
		"THE /carpet/bag",
		"* /carpet/bag",
		"THE /spouter/inn",
		"THE /counterpane/**",
		"* /breakfast/*",
		"* /nantucket/*/*/*",
		"* /nantucket/*/*",
		"* /enter/**",
		"**",
	})

	testResolver(t, r)
}

func testResolver(t *testing.T, r *Resolver) {

	tests := map[string]string{
		"GET /loomings":                         "* /loomings",
		"THE /carpet/bag":                       "THE /carpet/bag",
		"POST /carpet/bag":                      "* /carpet/bag",
		"POST /breakfast/123":                   "* /breakfast/*",
		"HEAD /nantucket/1/2/3":                 "* /nantucket/*/*/*",
		"PUT /nantucket/1/2":                    "* /nantucket/*/*",
		"THE /counterpane/1/2/3/4":              "THE /counterpane/**",
		"THIS /SHOULD/match/ANYTHING":           "**",
		"GET /enter/ahab/to/him/stubb/the/pipe": "* /enter/**",
	}

	for give, expect := range tests {
		got, err := r.ResolvePath(give)
		if err != nil {
			t.Errorf("Expected %q for %q, got error %s", expect, give, err)
			continue
		}
		if got != expect {
			t.Errorf("Expected %q, got %q", expect, got)
		}
	}

	u, err := url.Parse("/nantucket/1/2/3")
	if err != nil {
		panic(err)
	}
	req := &http.Request{Method: "GET", URL: u}
	got, err := r.Resolve(req)
	if err != nil {
		t.Errorf("Failed request resolver: %s.", err)
	} else if got != "* /nantucket/*/*/*" {
		t.Errorf("Unexpected route match: %q", got)
	}
}

func TestConcurrentResolver(t *testing.T) {
	// This test is a canary for breaking concurrency assumptions.
	r := NewResolver([]string{
		"* /loomings",
		"THE /carpet/bag",
		"* /carpet/bag",
		"THE /spouter/inn",
		"THE /counterpane/**",
		"* /breakfast/*",
		"* /nantucket/*/*/*",
		"* /nantucket/*/*",
		"* /enter/**",
		"**",
	})
	go testResolver(t, r)
	go testResolver(t, r)
	testResolver(t, r)
	time.Sleep(15 * time.Millisecond)
}
