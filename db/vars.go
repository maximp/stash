package db

import (
	"errors"
	"log"
	"strconv"
)

type Command uint

const (
	CommandNop    Command = iota
	CommandGet    Command = iota
	CommandSet    Command = iota
	CommandPush   Command = iota
	CommandPop    Command = iota
	CommandRemove Command = iota
	CommandTtl    Command = iota
	CommandKeys   Command = iota
)

// ParseCommand resolves command name to Command constant
func ParseCommand(cmd []byte) (Command, error) {
	switch string(cmd) {
	case "nop":
		return CommandNop, nil
	case "get":
		return CommandGet, nil
	case "set":
		return CommandSet, nil
	case "push":
		return CommandPush, nil
	case "pop":
		return CommandPop, nil
	case "remove":
		return CommandRemove, nil
	case "ttl":
		return CommandTtl, nil
	case "keys":
		return CommandKeys, nil
	default:
		return CommandNop, ErrInvalidCommand
	}
}

// String implements fmt.Stringer interface
func (c Command) String() string {
	switch c {
	case CommandNop:
		return "nop"
	case CommandGet:
		return "get"
	case CommandSet:
		return "set"
	case CommandPush:
		return "push"
	case CommandPop:
		return "pop"
	case CommandRemove:
		return "remove"
	case CommandTtl:
		return "ttl"
	case CommandKeys:
		return "keys"
	default:
		return strconv.Itoa(int(c))
	}
}

var (
	ErrInvalidCommand error = errors.New("invalid command name")
	ErrAlreadyClosed  error = errors.New("database already closed")
	ErrInvalidFormat  error = errors.New("invalid command format")
	ErrNotFound       error = errors.New("not found")
	ErrInvalidIndex   error = errors.New("invalid index")
	ErrInvalidType    error = errors.New("invalid type")
	ErrKeyNotFound    error = errors.New("key not found")
)

type Event uint

const (
	EventExpired = iota
)

type EventHandler func(e Event, name []byte)

type Config struct {
	Log         *log.Logger
	QueueLength uint
	Handler     EventHandler
}
