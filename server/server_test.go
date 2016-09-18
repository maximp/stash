package server

import (
	"bytes"
	"errors"
	"log"
	"net/textproto"
	"testing"
	"time"
)

func TestServerCreateError(t *testing.T) {
	if err := ListenAndServe("", nil, nil); err.Error() != "invalid server handler" {
		t.Error(err)
	}
}

func TestServerListen(t *testing.T) {
	stop := make(chan struct{})

	var buf bytes.Buffer
	cfg := Config{
		Logger: log.New(&buf, "", 0),
		Stop:   stop,
	}

	handler := func(cmd []byte, arg []byte) ([]byte, error) {
		return nil, nil
	}

	var err error
	stopped := make(chan struct{})
	go func() {
		err = ListenAndServe("", handler, &cfg)
		stopped <- struct{}{}
	}()

	time.Sleep(10 * time.Millisecond)

	if str := buf.String(); str != "listening  [::]:7777\n" {
		t.Errorf("listening: %s", str)
	}
	buf.Reset()

	if err != nil {
		t.Fatal(err)
	} else {
		stop <- struct{}{}
	}

	<-stopped

	if str := buf.String(); str != "received stop signal...\n" {
		t.Errorf("stop: %s", str)
	}
}

func TestServerComm(t *testing.T) {
	stop := make(chan struct{})

	var buf bytes.Buffer
	cfg := Config{
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
		}
		return []byte("ok"), nil
	}

	var err error
	stopped := make(chan struct{})
	go func() {
		err = ListenAndServe("", handler, &cfg)
		stopped <- struct{}{}
	}()

	time.Sleep(10 * time.Millisecond)

	if err != nil {
		t.Fatal(err)
	}

	if conn, err := textproto.Dial("tcp", "127.0.0.1:7777"); err != nil {
		t.Error(err)
	} else {
		defer conn.Close()

		if _, err := conn.Cmd("command"); err != nil {
			t.Error(err)
		}

		code, line, err := conn.ReadCodeLine(0)

		if code != ServerOperationOk || line != "ok" || err != nil || cmd != "command" || arg != "" {
			t.Errorf("error command processing: code=%d, line=%s, err=%v, cmd=%s, arg=%s", code, line, err, cmd, arg)
		}

		if _, err := conn.Cmd("error"); err != nil {
			t.Error(err)
		}

		code, line, err = conn.ReadCodeLine(0)

		if code != ServerOperationError || line != "error" || err != nil || cmd != "error" || arg != "" {
			t.Errorf("error command processing: code=%d, line=%s, err=%v, cmd=%s, arg=%s", code, line, err, cmd, arg)
		}
	}

	stop <- struct{}{}
	<-stopped
}
