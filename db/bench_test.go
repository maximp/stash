package db

import (
	"strconv"
	"testing"
	"time"
)

func createBenchDb(b *testing.B) *Database {
	d, err := New(Config{
		QueueLength: 10,
	})
	if err != nil {
		b.Fatal(err)
	}
	return d
}

func BenchmarkDbSet(b *testing.B) {
	dd := createBenchDb(b)
	defer dd.Close()

	names := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		names = append(names, []byte("name"+strconv.Itoa(b.N)+",value"))
	}

	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := dd.Exec(CommandSet, names[i])
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
	}
	b.StopTimer()
}

func BenchmarkDbGet(b *testing.B) {
	dd := createBenchDb(b)
	defer dd.Close()

	names := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		names = append(names, []byte("name"+strconv.Itoa(b.N)))
	}

	time.Sleep(10 * time.Millisecond)

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
	b.StopTimer()
}
