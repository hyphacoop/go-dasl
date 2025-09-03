package jsonbench

import (
	"encoding/json"
	"os"
	"testing"
)

func BenchmarkMarshalTwitter(b *testing.B) {
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(bytes)))
	for b.Loop() {
		json.Marshal(v)
	}
}

func BenchmarkUnmarshalTwitter(b *testing.B) {
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(data)))
	for b.Loop() {
		var v any
		json.Unmarshal(data, &v)
	}
}
