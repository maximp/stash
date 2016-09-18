package stash

import (
	"bytes"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/maximp/stash/client"
	"github.com/maximp/stash/server"
)

func TestServerComm(t *testing.T) {
	stop := make(chan struct{})

	var buf bytes.Buffer
	cfg := server.Config{
		Logger: log.New(&buf, "", 0),
		Stop:   stop,
	}

	var cmd string
	var arg string
	handler := func(c []byte, a []byte) ([]byte, error) {
		cmd = string(c)
		arg = string(a)
		if cmd == "error" {
			return nil, errors.New("error")
		} else {
			return []byte("ok"), nil
		}
	}

	var err error
	stopped := make(chan struct{})
	go func() {
		err = server.ListenAndServe("", handler, &cfg)
		stopped <- struct{}{}
	}()

	time.Sleep(10 * time.Millisecond)

	if err != nil {
		t.Fatal(err)
	}

	if conn, err := client.Dial("127.0.0.1:7777"); err != nil {
		t.Error(err)
	} else {
		defer conn.Close()
		code, line, err := conn.Cmd("command")
		if code != server.ServerOperationOk || line != "ok" || err != nil || cmd != "command" || arg != "" {
			t.Errorf("error command processing: code=%d, line=%s, err=%v, cmd=%s, arg=%s", code, line, err, cmd, arg)
		}

		code, line, err = conn.Cmd("error")
		if code != server.ServerOperationError || line != "error" || err != nil || cmd != "error" || arg != "" {
			t.Errorf("error command processing: code=%d, line=%s, err=%v, cmd=%s, arg=%s", code, line, err, cmd, arg)
		}
	}

	stop <- struct{}{}
	<-stopped
}
