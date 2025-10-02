package rasl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"

	"github.com/hyphacoop/go-dasl/cid"
)

var (
	ErrAllHintsFailed = errors.New("go-dasl/rasl: all hints failed")
	ErrCidValidation  = errors.New("go-dasl/rasl: data doesn't match CID")
)

// Fetch retrieves the content if possible.
//
// All the provided hints are attempted in parallel, and the first successful response is used.
// There is no overall timeout, although there may be underlying handshake timeouts, etc.
// If all hints fail, ErrAllHintsFailed is returned.
// It is not safe to modify the hints slice while Fetch is running.
//
// The data is streamed back. CID validation is performed, but can only be confirmed once all the
// data has been read out. If the CID doesn't match the data, ErrCidValidation will be returned on the
// last Read call instead of io.EOF.
//
// Close the reader to clean up the network connection.
func (ru *URL) Fetch() (io.ReadCloser, error) {
	return ru.FetchWithClient(http.DefaultClient)
}

// FetchWithClient is the same as Fetch(), but allows setting a custom http.Client.
// This can be used to set an overall timeout, make requests with cookies,
// allow custom certificates, etc.
func (ru *URL) FetchWithClient(client *http.Client) (io.ReadCloser, error) {
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}
	if len(ru.Hints) == 0 {
		return nil, ErrAllHintsFailed
	}

	// Collect request results
	type ret struct {
		resp *http.Response
		err  error
	}
	numReqs := len(ru.Hints)
	retCh := make(chan ret, numReqs)

	cidStr := ru.Cid.String()
	cancelers := make([]context.CancelFunc, numReqs)
	for i, hint := range ru.Hints {
		ctx, cancel := context.WithCancel(context.Background())
		cancelers[i] = cancel
		go func() {
			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://%s/.well-known/rasl/%s", hint, cidStr), nil)
			if err != nil {
				retCh <- ret{nil, err}
				return
			}
			resp, err := client.Do(req)
			retCh <- ret{resp, err}
		}()
	}

	i := 0
	var body io.ReadCloser
	for r := range retCh {
		if r.err == nil && r.resp.StatusCode == 200 {
			// One hint succeeded, continue with this one only
			for j := range cancelers {
				if i == j {
					continue
				}
				cancelers[j]()
			}
			body = r.resp.Body
			i++
			break
		} else if r.resp != nil {
			// Clean up resources
			r.resp.Body.Close()
		}
		i++
		if i == numReqs {
			// All requests processed, nothing worked
			return nil, ErrAllHintsFailed
		}
	}

	// Clean up the other goroutines and the network requests
	if i < numReqs {
		go func() {
			for r := range retCh {
				if r.resp != nil {
					r.resp.Body.Close()
				}
				i++
				if i == numReqs {
					// Nothing else will come out of that channel
					return
				}
			}
		}()
	}

	// Validate CID while letting user read
	return &verifyReader{
		cid:    ru.Cid,
		rc:     body,
		hasher: ru.Cid.Hasher(),
	}, nil
}

type verifyReader struct {
	cid    cid.Cid
	rc     io.ReadCloser
	hasher hash.Hash
}

func (vr *verifyReader) Read(p []byte) (n int, err error) {
	n, err = vr.rc.Read(p)
	if n > 0 {
		// A copy is required since Write is allowed to modify the input
		// Only copy the bytes actually read
		pCopy := make([]byte, n)
		copy(pCopy, p[:n])
		vr.hasher.Write(pCopy)
	}
	if err == io.EOF {
		// All bytes have been read
		// Check hash and report
		dgst := vr.cid.Digest()
		if !bytes.Equal(dgst[:], vr.hasher.Sum(nil)) {
			return 0, ErrCidValidation
		}
	}
	return
}

func (vr *verifyReader) Close() error {
	return vr.rc.Close()
}
