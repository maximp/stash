package server

import "log"

const (
	ServerOperationOk    = 200
	ServerOperationError = 300
)

// A Handler type represents server command handler
type Handler func(cmd []byte, arg []byte) (result []byte, err error)

// A Config represents optional server parameters
type Config struct {
	Logger *log.Logger
	Stop   <-chan struct{}
}
