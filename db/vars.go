package db

import (
	"errors"
	"log"
	"strconv"
)

// A Command represents engine command code passed into Database.Exec function
type Command uint

// Command constants
const (
	CommandNop    Command = iota
	CommandGet    Command = iota
	CommandSet    Command = iota
	CommandPush   Command = iota
	CommandPop    Command = iota
	CommandRemove Command = iota
	CommandTTL    Command = iota
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
		return CommandTTL, nil
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
	case CommandTTL:
		return "ttl"
	case CommandKeys:
		return "keys"
	default:
		return strconv.Itoa(int(c))
	}
}

// Errors returned by database
var (
	ErrInvalidCommand = errors.New("invalid command name")
	ErrAlreadyClosed  = errors.New("database already closed")
	ErrInvalidFormat  = errors.New("invalid command format")
	ErrNotFound       = errors.New("not found")
	ErrInvalidIndex   = errors.New("invalid index")
	ErrInvalidType    = errors.New("invalid type")
	ErrKeyNotFound    = errors.New("key not found")
)

// An Event represents event code passed into user-defined event handler
type Event uint

// Event constants
const (
	EventExpired = iota
)

// EventHandler declares type of user-defined event handler for database engine events
type EventHandler func(e Event, name []byte)

// Config contains user-defined parameters to initialize engine
type Config struct {
	Log         *log.Logger
	QueueLength uint
	Handler     EventHandler
}
