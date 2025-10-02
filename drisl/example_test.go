package drisl_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hyphacoop/go-dasl/cid"
	"github.com/hyphacoop/go-dasl/drisl"
)

func ExampleMarshal() {
	type Data struct {
		Name      string    `cbor:"name"`
		Count     int       `cbor:"count"`
		Timestamp time.Time `cbor:"timestamp"`
		ID        cid.Cid   `cbor:"id"`
	}

	// Create a CID for some data
	id, _ := drisl.CidForValue(map[string]string{"hello": "world"})

	data := Data{
		Name:      "example",
		Count:     42,
		Timestamp: time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC),
		ID:        id,
	}

	bytes, err := drisl.Marshal(data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", bytes)
	// Output:
	// a4626964d82a58250001711220785197229dc8bb1152945da58e2348f7e279eeded06cc2ca736d0e879858b501646e616d65676578616d706c6565636f756e74182a6974696d657374616d7074323032332d30362d31355431343a33303a34355a
}

func ExampleUnmarshal() {
	type Data struct {
		Name      string    `cbor:"name"`
		Count     int       `cbor:"count"`
		Timestamp time.Time `cbor:"timestamp"`
		ID        cid.Cid   `cbor:"id"`
	}

	// Sample DRISL bytes
	bytes, _ := hex.DecodeString("a4626964d82a58250001711220785197229dc8bb1152945da58e2348f7e279eeded06cc2ca736d0e879858b501646e616d65676578616d706c6565636f756e74182a6974696d657374616d7074323032332d30362d31355431343a33303a34355a")

	var data Data
	err := drisl.Unmarshal(bytes, &data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", data)
	// Output:
	// {Name:example Count:42 Timestamp:2023-06-15 14:30:45 +0000 UTC ID:bafyreidykglsfhoixmivffc5uwhcgshx4j465xwqntbmu43nb2dzqwfvae}
}

func ExampleCidForValue() {
	// Calculate CID for simple data
	data := map[string]any{
		"name":  "Alice",
		"age":   30,
		"admin": true,
	}

	id, err := drisl.CidForValue(data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", id)
	// Output:
	// bafyreihlticva4wkngdttc46hdnldewyxl7amaifb3e2ghipxv5auu3pcm
}

func ExampleNewEncoder() {
	var buf bytes.Buffer
	encoder := drisl.NewEncoder(&buf)

	// Encode multiple values to the buffer
	err := encoder.Encode("hello")
	if err != nil {
		panic(err)
	}

	err = encoder.Encode(42)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", buf.Bytes())
	// Output:
	// 6568656c6c6f182a
}

func ExampleNewDecoder() {
	// DRISL bytes containing two values: "hello" and 42
	data := []byte{0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x18, 0x2a}
	reader := bytes.NewReader(data)
	decoder := drisl.NewDecoder(reader)

	var str string
	err := decoder.Decode(&str)
	if err != nil {
		panic(err)
	}

	var num int
	err = decoder.Decode(&num)
	if err != nil {
		panic(err)
	}

	fmt.Printf("String: %s, Number: %d\n", str, num)
	// Output:
	// String: hello, Number: 42
}

func ExampleDecOptions_DecMode() {
	// Create decoder with custom options
	opts := drisl.DecOptions{
		MaxArrayElements: 100,
		NoFloats:         true,
	}

	decoder, err := opts.DecMode()
	if err != nil {
		panic(err)
	}

	// This will succeed
	var data []int
	err = decoder.Unmarshal([]byte{0x83, 0x01, 0x02, 0x03}, &data)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", data)
	// Output:
	// [1 2 3]
}

func ExampleEncOptions_EncMode() {
	// Create encoder with custom time mode
	opts := drisl.EncOptions{
		Time: drisl.TimeUnix,
	}

	encoder, err := opts.EncMode()
	if err != nil {
		panic(err)
	}

	timestamp := time.Unix(1609459200, 0).UTC() // 2021-01-01 00:00:00 UTC
	bytes, err := encoder.Marshal(timestamp)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%x\n", bytes)
	// Output:
	// 1a5fee6600
}
