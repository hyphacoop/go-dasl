package drisl_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hyphacoop/go-dasl/drisl"
)

func BenchmarkMarshalTwitter(b *testing.B) {
	// Load
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	// Measure
	bytes, err := drisl.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(bytes)))
	// Benchmark
	for b.Loop() {
		drisl.Marshal(v)
	}
}

func BenchmarkUnmarshalTwitter(b *testing.B) {
	// Load
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	// Marshal and measure
	marshaled, err := drisl.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(marshaled)))
	// Benchmark
	for b.Loop() {
		var v any
		drisl.Unmarshal(marshaled, &v)
	}
}
