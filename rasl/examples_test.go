package rasl_test

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/rasl"
)

// ExampleParse demonstrates how to parse a RASL URL string.
func ExampleParse() {
	// Parse a RASL URL with hints
	raslURL, err := rasl.Parse("rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=berjon.com&hint=bsky.app")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CID: %s\n", raslURL.Cid)
	fmt.Printf("Hints: %v\n", raslURL.Hints)
	fmt.Printf("Path: %s\n", raslURL.Path)

	// Parse a RASL URL without hints
	simple, err := rasl.Parse("rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Simple CID: %s\n", simple.Cid)
	fmt.Printf("Has hints: %t\n", len(simple.Hints) > 0)

	// Output:
	// CID: bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// Hints: [berjon.com bsky.app]
	// Path: /
	// Simple CID: bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4
	// Has hints: false
}

// ExampleURL_String demonstrates how to convert a RASL URL struct back to a string.
func ExampleURL_String() {
	// Create a RASL URL manually
	c, _ := cid.NewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")
	raslURL := &rasl.URL{
		Cid:   c,
		Hints: []string{"example.com", "backup.example.org:8080"},
		Path:  "/",
	}

	fmt.Println(raslURL.String())

	// Output:
	// rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4/?hint=example.com&hint=backup.example.org%3A8080
}

// ExampleURL_Fetch demonstrates how to fetch content using a RASL URL.
func ExampleURL_Fetch() {
	// This example shows the basic usage pattern
	// Note: This won't work without actual RASL servers running
	raslURL, err := rasl.Parse("rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4?hint=example.com")
	if err != nil {
		log.Fatal(err)
	}

	reader, err := raslURL.Fetch()
	if err != nil {
		// In a real scenario, you'd handle specific errors like rasl.ErrAllHintsFailed
		fmt.Printf("Failed to fetch: %v\n", err)
		return
	}
	defer reader.Close()

	// Read the content
	data, err := io.ReadAll(reader)
	if err != nil {
		// Handle rasl.ErrCidValidation and other errors
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	fmt.Printf("Fetched %d bytes\n", len(data))
}

// ExampleRedirectHandler demonstrates creating a HTTP handler for RASL redirects.
func ExampleRedirectHandler() {
	// Create a map of CIDs to redirect URLs
	redirects := map[string]string{
		"bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4": "/static/data.txt",
		"bafkreigdvqhvgm6rh5hcp6pdjdwfzc7d3r4kx5kqu4d4lsmjqzjz4o4afe": "https://cdn.example.com/image.png",
	}

	// Create the handler
	handler := rasl.RedirectHandler(redirects)

	// Use in a HTTP server
	http.Handle("/.well-known/rasl/", handler)
	http.ListenAndServe(":8080", nil)
}

// ExampleFuncHandler demonstrates creating a custom RASL handler with a function.
func ExampleFuncHandler() {
	// Create a custom function to serve content
	contentFunc := func(c cid.Cid) (io.Reader, error) {
		// In a real implementation, you'd look up the CID in your storage system
		cidStr := c.String()

		// Example: serve different content based on CID
		switch cidStr {
		case "bafkreif2nhn3zj2cqomadl7t2cegn5ithxjfplwoyg7k62abhaahkequie":
			return strings.NewReader("Hello, RASL!"), nil
		case "bafkreifwqwdlolr2kwsrogr2m6x5gvnccbawlhke5zo352f5uomylbnr5m":
			return strings.NewReader("Some text content"), nil
		default:
			// Return nil, nil for 404
			return nil, nil
		}
	}

	// Create the handler
	handler := rasl.FuncHandler(contentFunc)

	// Use in a HTTP server
	http.Handle("/.well-known/rasl/", handler)
	http.ListenAndServe(":8080", nil)
}

// ExampleCidDirectoryHandler demonstrates serving files from a directory by CID name.
func ExampleCidDirectoryHandler() {
	// Serve files from /storage/ directory
	// Files should be named with their CID (e.g., "bafkrei...")
	handler := rasl.CidDirectoryHandler("/storage")

	// Use in a HTTP server
	http.Handle("/.well-known/rasl/", handler)
	http.ListenAndServe(":8080", nil)
}

// ExampleDirectoryHandler demonstrates automatically hashing directory contents for RASL serving.
func ExampleDirectoryHandler() {
	// Create handler that hashes all files in the directory
	// Files can be requested by their computed CID
	handler, err := rasl.DirectoryHandler("/storage", true) // true = also hash with BLAKE3
	if err != nil {
		log.Fatal(err)
	}

	// Use in a HTTP server
	http.Handle("/.well-known/rasl/", handler)
	http.ListenAndServe(":8080", nil)
}

// ExampleURL_FetchWithClient demonstrates using a custom HTTP client for fetching.
func ExampleURL_FetchWithClient() {
	raslURL, err := rasl.Parse("rasl://bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4?hint=example.com")
	if err != nil {
		log.Fatal(err)
	}

	// Create a custom HTTP client with timeout
	client := &http.Client{
		Timeout: 30, // 30 second timeout
	}

	reader, err := raslURL.FetchWithClient(client)
	if err != nil {
		fmt.Printf("Failed to fetch with custom client: %v\n", err)
		return
	}
	defer reader.Close()

	// Output would depend on whether the server is actually available
}
