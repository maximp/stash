package db

import (
	"strconv"
	"testing"
)

func createDenchDb(b *testing.B) *Database {
	d, err := New(Config{
		QueueLength: 10,
	})
	if err != nil {
		b.Fatal(err)
	}
	return d
}

func BenchmarkSet(b *testing.B) {
	dd := createDenchDb(b)
	defer dd.Close()

	names := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		names = append(names, []byte("name"+strconv.Itoa(b.N)+",value"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dd.Exec(CommandSet, names[i])
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	dd := createDenchDb(b)
	defer dd.Close()

	names := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		names = append(names, []byte("name"+strconv.Itoa(b.N)))
	}

	for i := 0; i < b.N; i++ {
		_, err := dd.Exec(CommandSet, []byte("name"+strconv.Itoa(b.N)+",value"))
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dd.Exec(CommandGet, names[i])
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
	}
}
