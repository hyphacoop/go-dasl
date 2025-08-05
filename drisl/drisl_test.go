package drisl

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

type daslTestCase struct {
	Type string
	Data string
	Tags []string
	Name string
}

func TestDaslJson(t *testing.T) {
	// Parse all the JSON files and run all the relevant tests as subtests
	// Kind of like table-driven testing, but on the fly
	// https://go.dev/wiki/TableDrivenTests

	err := filepath.WalkDir("dasl-testing/fixtures/cbor/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var tests []*daslTestCase
		if err := json.Unmarshal(b, &tests); err != nil {
			return err
		}
		runTests(t, tests)
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func runTests(t *testing.T, tests []*daslTestCase) {
	for _, test := range tests {
		if !slices.Contains(test.Tags, "basic") && !slices.Contains(test.Tags, "dag-cbor") {
			continue
		}
		testData, err := hex.DecodeString(test.Data)
		if err != nil {
			panic(fmt.Errorf("failed to decode hex: %s", test.Data))
		}

		switch test.Type {
		case "roundtrip":
			t.Run(test.Name, func(t *testing.T) {
				var v any
				if err := Unmarshal(testData, &v); err != nil {
					t.Errorf("Unmarshal error: %v", err)
					return
				}
				b, err := Marshal(v)
				if err != nil {
					t.Errorf("Marshal error: %v", err)
					return
				}
				if !bytes.Equal(testData, b) {
					t.Errorf("got %x, want %x", b, testData)
				}
			})
			t.Run(test.Name+"-Valid", func(t *testing.T) {
				if !Valid(testData) {
					t.Errorf("got false, want true")
				}
			})
		case "invalid_in":
			t.Run(test.Name, func(t *testing.T) {
				var v any
				if err := Unmarshal(testData, &v); err == nil {
					t.Error("Unmarshal didn't raise an error")
				}
			})
			t.Run(test.Name+"-Valid", func(t *testing.T) {
				if Valid(testData) {
					t.Errorf("got true, want false")
				}
			})
		case "invalid_out":
			t.Run(test.Name, func(t *testing.T) {
				// Decode data with neutral CBOR decoder, then confirm it cannot be
				// encoded by the strict DRISL encoder.
				var v any
				if err := cbor.Unmarshal(testData, &v); err != nil {
					t.Errorf("cbor.Unmarshal failed unexpectedly: %v", err)
					return
				}
				if _, err := Marshal(v); err == nil {
					t.Error("Marshal didn't raise an error")
				}
			})
			t.Run(test.Name+"-Valid", func(t *testing.T) {
				if Valid(testData) {
					t.Errorf("got true, want false")
				}
			})
		default:
			panic(fmt.Errorf("unknown test type '%s'", test.Type))
		}
	}
}
