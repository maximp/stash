package client

import (
	"errors"
	"io"
	"net/textproto"
	"strings"
)

// A Client represents client connection to cache engine
type Client struct {
	conn *textproto.Conn
}

// Dial connects to the given address textproto.Dial
// and then returns a new Client for the connection.
func Dial(addr string) (*Client, error) {
	if conn, err := textproto.Dial("tcp", addr); err != nil {
		return nil, err
	} else {
		return &Client{conn}, nil
	}
}

// Close implements io.Closer interface and closes client connection
func (c Client) Close() error {
	return c.conn.Close()
}

// Cmd sends given command to server and waits for reply
func (c Client) Cmd(str string) (code int, line string, err error) {

	// convert to single line
	str = encode(str)

	// send command to remote
	if _, err = c.conn.Cmd("%s", str); err != nil {
		return
	}

	// read reply
	code, line, err = c.conn.ReadCodeLine(0)
	if err == io.EOF {
		err = errors.New("connection closed")
		return
	} else if err != nil {
		return
	}

	// convert message from single line to multiline
	line = decode(line)

	return
}

var (
	crlf_encoder = strings.NewReplacer("\n", "\\n", "\r", "\\r")
	crlf_decoder = strings.NewReplacer("\\n", "\n", "\\r", "\r")
)

// encode converts CR to \r, LF to \n
func encode(s string) string {
	return crlf_encoder.Replace(s)
}

// decode converts \n to LF, \r to CR
func decode(s string) string {
	return crlf_decoder.Replace(s)
}