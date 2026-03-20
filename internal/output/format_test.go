package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestFormatterPrintJSON(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	formatter := NewFormatter("json", &out, &errOut)
	if err := formatter.Print(map[string]string{"key": "value"}); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if payload["key"] != "value" {
		t.Fatalf("expected key=value, got %#v", payload)
	}
}

func TestFormatterPrintText(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	formatter := NewFormatter("text", &out, &errOut)
	if err := formatter.Print("hello"); err != nil {
		t.Fatalf("Print returned error: %v", err)
	}

	if strings.TrimSpace(out.String()) != "hello" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestFormatterPrintErrorJSON(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	formatter := NewFormatter("json", &out, &errOut)
	if err := formatter.PrintError(errors.New("boom")); err != nil {
		t.Fatalf("PrintError returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(errOut.Bytes(), &payload); err != nil {
		t.Fatalf("failed to unmarshal error output: %v", err)
	}

	if payload["error"] != "boom" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}
