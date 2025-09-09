package rasl

import (
	"context"
	"crypto/sha256"
	"hash"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/hyphacoop/go-dasl/cid"
	"lukechampine.com/blake3"
)

// RedirectHandler takes RASL requests and redirects them based on the map.
//
// redirects is a map of CIDs to redirect URLs or paths.
// Redirects can be relative, but should at least start from the root ("/...")
// for safety.
//
// The redirects map can't be safely modified after being passed to this function.
func RedirectHandler(redirects map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/.well-known/rasl/bafkr") {
			http.NotFound(w, r)
			return
		}
		cidStr := r.URL.Path[len("/.well-known/rasl/"):]
		// Could do further validation on the CID but tbh it's up to the caller to provide
		// valid CID strings
		redir, ok := redirects[cidStr]
		if !ok {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, redir, http.StatusTemporaryRedirect)
	})
}

// FuncHandler takes RASL requests and returns the data provided by a custom function.
// You can use this to easily create more complex CID -> data mappings, for example retrieving
// from some custom storage system.
//
// The function can return (nil, nil) to indicate the CID is not available and this is expected.
// This will result in status code 404.
// Otherwise, errors are returned over HTTP with status code 500.
//
// If the returned io.Reader also implements Close, it will be closed after all data is sent.
// This makes returning files and other Closers safe, although note errors when closing
// cannot be handled.
//
// It is up to the caller to validate that the data returned actually matches the hash digest
// of the CID it's in response to.
func FuncHandler(f func(cid.Cid) (io.Reader, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/.well-known/rasl/bafkr") {
			http.NotFound(w, r)
			return
		}
		c, err := cid.NewCidFromString(r.URL.Path[len("/.well-known/rasl/"):])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		reader, err := f(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if reader == nil {
			http.NotFound(w, r)
			return
		}

		// Spec requires forcing this to prevent smuggling extra info
		w.Header().Add("Content-Type", "application/octet-stream")

		if _, err := io.Copy(w, reader); err != nil {
			// Unfortunately some data might already be written at this point
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if closer, ok := reader.(io.Closer); ok {
			closer.Close()
		}
	})
}

// CidDirectoryHandler takes RASL requests and opens a file on disk with the relevant CID as its name.
// This is just a wrapper around the existing http.FileServer, with correct path and CID checking.
//
// This is useful if you already have a storage system that is just files with CID names
// stored in a single directory.
//
// File contents are not validated.
//
// CIDs in requests are validated, so if this directory contains other files without CID names,
// these will not be retrieved. Directories are not descended into.
func CidDirectoryHandler(dir string) http.Handler {
	dirHandler := http.StripPrefix("/.well-known/rasl/", http.FileServer(http.Dir(dir)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/.well-known/rasl/bafkr") {
			http.NotFound(w, r)
			return
		}
		cidStr := r.URL.Path[len("/.well-known/rasl/"):]
		// Validate CID
		if _, err := cid.NewCidFromString(cidStr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Add("Content-Type", "application/octet-stream")
		dirHandler.ServeHTTP(w, r)
	})
}

// DirectoryHandler recursively hashes all the files in a directory so they can be requested by CID over RASL.
// If the hash digest matches, that file will be opened and written back to the client.
//
// Directory hashing only occurs once when this function is called, so editing,
// renaming, moving, or deleting those files will
// result in errors. Altering the directory while this function is hashing it can cause errors.
//
// Files are read in parallel. An error is returned if there is any issue reading files
// in the initial hashing period.
//
// Set hashBlake3 to true to also hash every file with BLAKE3, not just SHA-256.
func DirectoryHandler(dir string, hashBlake3 bool) (http.Handler, error) {
	// Initial hashing
	// Have worker pool iterate over path channel
	// Inspired by: https://github.com/makew0rld/merkdir/blob/f69ec2d2218689a423d548f56aabd8514ec49591/commands.go#L34
	// Which I give myself permission to reuse under the license in this repo

	var wg sync.WaitGroup
	errCh := make(chan error)
	pathCh := make(chan string)

	type ret struct {
		digestBlake3 []byte
		digestSha256 []byte
		path         string
	}
	retCh := make(chan ret)
	ctx, cancel := context.WithCancel(context.Background())

	// Launch workers, waiting for paths
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Cancelled
					return
				case path, ok := <-pathCh:
					if !ok {
						// No more paths
						return
					}

					f, err := os.Open(filepath.Join(dir, path))
					if err != nil {
						errCh <- err
						return
					}
					// Close f manually so that files aren't left open while the loop runs
					// Otherwise the max open file limit will be hit for large dirs on at
					// least some OSes like macOS.

					// Calculate data
					var w io.Writer
					hasherSha256 := sha256.New()
					var hasherBlake3 hash.Hash
					if hashBlake3 {
						hasherBlake3 = blake3.New(cid.HashLength, nil)
						w = io.MultiWriter(hasherSha256, hasherBlake3)
					} else {
						w = hasherSha256
					}
					_, err = io.Copy(w, f)
					f.Close()
					if err != nil {
						errCh <- err
						return
					}

					// Return data
					r := ret{
						digestSha256: hasherSha256.Sum(nil),
						path:         path,
					}
					if hashBlake3 {
						r.digestBlake3 = hasherBlake3.Sum(nil)
					}
					retCh <- r
				}
			}
		}()
	}

	// Distribute file paths to workers
	go func() {
		fs.WalkDir(os.DirFS(dir), ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				errCh <- err
				return err // Stop
			}
			if d.IsDir() {
				return nil
			}
			if d.Type() != 0 {
				// Some sort of special file
				return nil
			}
			pathCh <- path
			return nil
		})
		close(pathCh)
	}()

	// Signal when all workers are done with no errors
	go func() {
		wg.Wait()
		errCh <- nil
	}()

	// Process worker results
	cidPaths := make(map[string]string)
outer:
	for {
		select {
		case err := <-errCh:
			if err == nil {
				// All workers done without errors
				break outer
			} else {
				// One worker had an error, stop all of them
				cancel()
				return nil, err
			}
		case r := <-retCh:
			c, _ := cid.NewCidFromInfo(cid.CodecRaw, cid.HashTypeSha256, r.digestSha256)
			cidPaths[c.String()] = r.path
			if hashBlake3 {
				c, _ := cid.NewCidFromInfo(cid.CodecRaw, cid.HashTypeBlake3, r.digestBlake3)
				cidPaths[c.String()] = r.path
			}
		}
	}

	// Create handler
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/.well-known/rasl/bafkr") {
			http.NotFound(w, r)
			return
		}
		cidStr := r.URL.Path[len("/.well-known/rasl/"):]

		path, ok := cidPaths[cidStr]
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Add("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, filepath.Join(dir, path))
	}), nil
}
