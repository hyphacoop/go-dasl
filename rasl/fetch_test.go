package rasl_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/rasl"
)

func TestFetch_Success(t *testing.T) {
	// Test data that matches the CID
	testData := []byte("hello world")
	testCid := cid.HashBytes(testData)

	// Create HTTPS test server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/.well-known/rasl/%s", testCid.String())
		if r.URL.Path != expectedPath {
			http.Error(w, "not found", 404)
			return
		}
		w.WriteHeader(200)
		w.Write(testData)
	}))
	defer server.Close()

	// Extract hostname from server URL (remove https://)
	serverHost := server.URL[8:] // Remove "https://" prefix

	// Create RASL URL with the test server as a hint
	raslURL := &rasl.URL{
		Cid:   testCid,
		Hints: []string{serverHost},
	}

	// Fetch data
	reader, err := raslURL.FetchWithClient(server.Client())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	defer reader.Close()

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Reading data failed: %v", err)
	}

	// Verify data matches
	if !bytes.Equal(data, testData) {
		t.Fatalf("Data mismatch: got %q, want %q", data, testData)
	}
}

func TestFetch_AllHintsFail(t *testing.T) {
	testData := []byte("hello world")
	testCid := cid.HashBytes(testData)

	// Create RASL URL with non-existent hints
	raslURL := &rasl.URL{
		Cid:   testCid,
		Hints: []string{"nonexistent1.example", "nonexistent2.example"},
	}

	// Fetch should fail
	_, err := raslURL.Fetch()
	if err != rasl.ErrAllHintsFailed {
		t.Fatalf("Expected ErrAllHintsFailed, got: %v", err)
	}
}

func TestFetch_ParallelHints(t *testing.T) {
	// Test data that matches the CID
	testData := []byte("hello world")
	testCid := cid.HashBytes(testData)

	// Create HTTPS test server (working hint)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/.well-known/rasl/%s", testCid.String())
		if r.URL.Path != expectedPath {
			http.Error(w, "not found", 404)
			return
		}
		w.WriteHeader(200)
		w.Write(testData)
	}))
	defer server.Close()

	// Extract hostname from server URL (remove https://)
	serverHost := server.URL[8:] // Remove "https://" prefix

	// Create RASL URL with mix of working and non-existent hints
	raslURL := &rasl.URL{
		Cid:   testCid,
		Hints: []string{"nonexistent1.example", serverHost, "nonexistent2.example"},
	}

	// Fetch data - should succeed with the working hint
	reader, err := raslURL.FetchWithClient(server.Client())
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}
	defer reader.Close()

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Reading data failed: %v", err)
	}

	// Verify data matches
	if !bytes.Equal(data, testData) {
		t.Fatalf("Data mismatch: got %q, want %q", data, testData)
	}
}

