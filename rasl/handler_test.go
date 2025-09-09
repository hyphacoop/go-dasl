package rasl_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/rasl"
)

func TestRedirectHandler_Success(t *testing.T) {
	testCid := cid.HashBytes([]byte("test data"))
	redirects := map[string]string{
		testCid.String(): "/redirect/target",
	}

	handler := rasl.RedirectHandler(redirects)

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Fatalf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/redirect/target" {
		t.Fatalf("Expected Location '/redirect/target', got '%s'", location)
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	testCid := cid.HashBytes([]byte("test data"))
	redirects := map[string]string{} // Empty map

	handler := rasl.RedirectHandler(redirects)

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestFuncHandler_Success(t *testing.T) {
	testData := []byte("hello world")
	testCid := cid.HashBytes(testData)

	handler := rasl.FuncHandler(func(c cid.Cid) (io.Reader, error) {
		if c.String() == testCid.String() {
			return bytes.NewReader(testData), nil
		}
		return nil, nil // Not found
	})

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Fatalf("Expected Content-Type 'application/octet-stream', got '%s'", contentType)
	}

	body := w.Body.Bytes()
	if !bytes.Equal(body, testData) {
		t.Fatalf("Expected body %q, got %q", testData, body)
	}
}

func TestFuncHandler_NotFound(t *testing.T) {
	testCid := cid.HashBytes([]byte("test data"))

	handler := rasl.FuncHandler(func(c cid.Cid) (io.Reader, error) {
		return nil, nil // Always not found
	})

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCidDirectoryHandler_Success(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testData := []byte("hello world")
	testCid := cid.HashBytes(testData)

	filePath := filepath.Join(tmpDir, testCid.String())
	if err := os.WriteFile(filePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	handler := rasl.CidDirectoryHandler(tmpDir)

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Fatalf("Expected Content-Type 'application/octet-stream', got '%s'", contentType)
	}

	body := w.Body.Bytes()
	if !bytes.Equal(body, testData) {
		t.Fatalf("Expected body %q, got %q", testData, body)
	}
}

func TestCidDirectoryHandler_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testCid := cid.HashBytes([]byte("nonexistent"))

	handler := rasl.CidDirectoryHandler(tmpDir)

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDirectoryHandler_Success(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	testData := []byte("hello world")

	filePath := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(filePath, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	handler, err := rasl.DirectoryHandler(tmpDir, false) // SHA-256 only
	if err != nil {
		t.Fatalf("Failed to create DirectoryHandler: %v", err)
	}

	// Calculate expected CID for the test data
	expectedCid := cid.HashBytes(testData)

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+expectedCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Fatalf("Expected Content-Type 'application/octet-stream', got '%s'", contentType)
	}

	body := w.Body.Bytes()
	if !bytes.Equal(body, testData) {
		t.Fatalf("Expected body %q, got %q", testData, body)
	}
}

func TestDirectoryHandler_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testCid := cid.HashBytes([]byte("nonexistent"))

	handler, err := rasl.DirectoryHandler(tmpDir, false)
	if err != nil {
		t.Fatalf("Failed to create DirectoryHandler: %v", err)
	}

	req := httptest.NewRequest("GET", "/.well-known/rasl/"+testCid.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
