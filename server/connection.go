package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"time"
)

// A connection represents single TCP connection to database server
type connection struct {
	*textproto.Conn
	addr   net.Addr
	logger *log.Logger
}

// log prints message to attached or global log interface
func (c *connection) log(v ...interface{}) {
	c.logger.Println(c.addr, ": ", fmt.Sprint(v...))
}

// serve handles client connection
func (c *connection) serve(handler Handler) {
	defer c.Close()

	send := func(code int, result string) {
		if err := c.PrintfLine("%d %s", code, result); err != nil {
			c.log(err)
		}
	}

	for {
		line, err := c.ReadLine()
		if err != nil {
			if err == io.EOF {
				c.log("connection closed")
			} else {
				c.log(err)
			}
			break
		}

		start := time.Now()

		name, arg := parseCommand(line)
		if bytes.Equal(name, []byte("quit")) {
			c.log("quit, connection closed")
			break
		}

		result, err := handler(name, arg)
		if err != nil {
			elapsed := time.Since(start)
			c.log(elapsed, ", ", line, ", ", err)
			send(ServerOperationError, err.Error())
			continue
		}

		if result == nil {
			result = []byte("")
		}

		elapsed := time.Since(start)
		c.log(elapsed, ", ", line)

		send(ServerOperationOk, string(result))
	}
}

// parseCommand parses string and returns name of command and its arg
func parseCommand(str string) (cmd []byte, arg []byte) {
	bb := bytes.TrimSpace([]byte(str))
	fields := bytes.SplitN(bb, []byte{' '}, 2)
	switch len(fields) {
	case 2:
		arg = bytes.TrimSpace(fields[1])
		fallthrough
	case 1:
		cmd = bytes.TrimSpace(fields[0])
	}
	return
}
