package cli

import (
	"net/http"
	"testing"
)

func TestEnsureHTTPSuccess(t *testing.T) {
	if err := EnsureHTTPSuccess(&http.Response{StatusCode: http.StatusOK}, nil); err != nil {
		t.Fatalf("expected 200 to succeed, got %v", err)
	}

	if err := EnsureHTTPSuccess(&http.Response{StatusCode: http.StatusCreated}, nil); err != nil {
		t.Fatalf("expected 201 to succeed, got %v", err)
	}

	if err := EnsureHTTPSuccess(&http.Response{StatusCode: http.StatusMovedPermanently}, nil); err == nil {
		t.Fatal("expected 301 to fail")
	}

	if err := EnsureHTTPSuccess(&http.Response{StatusCode: http.StatusBadRequest}, nil); err == nil {
		t.Fatal("expected 400 to fail")
	}
}
