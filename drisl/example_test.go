package drisl_test

import (
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
	id, _ := drisl.CalculateCidForValue(map[string]string{"hello": "world"})

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
