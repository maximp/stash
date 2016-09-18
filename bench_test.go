package stash

import (
	"strconv"
	"testing"
	"time"

	"github.com/maximp/stash/client"
	"github.com/maximp/stash/db"
	"github.com/maximp/stash/server"
)

func BenchmarkClientSetGet(b *testing.B) {
	d, err := db.New(db.Config{
		QueueLength: 10,
	})
	if err != nil {
		b.Fatal(err)
	}

	handler := func(cmd []byte, arg []byte) ([]byte, error) {
		c, err := db.ParseCommand(cmd)
		if err != nil {
			return nil, err
		}
		return d.Exec(c, arg)
	}

	stop := make(chan struct{})
	cfg := server.Config{
		Stop: stop,
	}

	stopped := make(chan struct{})
	go func() {
		server.ListenAndServe("", handler, &cfg)
		stopped <- struct{}{}
	}()

	setCommands := make([]string, 0, b.N)
	getCommands := make([]string, 0, b.N)
	for i := 0; i < b.N; i++ {
		n := strconv.Itoa(b.N)
		setCommands = append(setCommands, "set name"+n+",some value "+n)
		getCommands = append(getCommands, "get name"+n)
	}

	time.Sleep(10 * time.Millisecond)

	conn, err := client.Dial("127.0.0.1:7777")
	if err != nil {
		b.Error(err)
		stop <- struct{}{}
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := conn.Cmd(setCommands[i])
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
		_, _, err = conn.Cmd(getCommands[i])
		if err != nil {
			b.Fatalf("set str, test failed %v", err)
		}
	}
	b.StopTimer()

	stop <- struct{}{}
}
