package upstreambench

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hyphacoop/go-dasl/cid"
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
	em := encMode()
	bytes, err := em.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(bytes)))
	for b.Loop() {
		em.Marshal(v)
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
	em := encMode()
	marshaled, err := em.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(marshaled)))
	// Benchmark
	dm := decMode()
	for b.Loop() {
		var v any
		dm.Unmarshal(marshaled, &v)
	}
}

func BenchmarkMarshalTwitterCid(b *testing.B) {
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	// Add in CIDs
	cid := cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")
	for _, status := range v.(map[string]any)["statuses"].([]any) {
		status.(map[string]any)["cid"] = cid
	}
	em := encMode()
	bytes, err := em.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(bytes)))
	for b.Loop() {
		em.Marshal(v)
	}
}

func BenchmarkUnmarshalTwitterCid(b *testing.B) {
	// Load
	data, err := os.ReadFile("testdata/twitter.json")
	if err != nil {
		panic(err)
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	// Add in CIDs
	cid := cid.MustNewCidFromString("bafkreifn5yxi7nkftsn46b6x26grda57ict7md2xuvfbsgkiahe2e7vnq4")
	for _, status := range v.(map[string]any)["statuses"].([]any) {
		status.(map[string]any)["cid"] = cid
	}
	// Marshal and measure
	em := encMode()
	marshaled, err := em.Marshal(v)
	if err != nil {
		panic(err)
	}
	b.SetBytes(int64(len(marshaled)))
	// Benchmark
	dm := decMode()
	for b.Loop() {
		var v any
		dm.Unmarshal(marshaled, &v)
	}
}
