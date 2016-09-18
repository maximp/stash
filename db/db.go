package db

import (
	"bytes"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type result struct {
	value []byte
	err   error
}

var (
	resultOk            = result{[]byte("Ok"), nil}
	resultNotFound      = result{nil, ErrNotFound}
	resultInvalidFormat = result{nil, ErrInvalidFormat}
	resultInvalidIndex  = result{nil, ErrInvalidIndex}
	resultInvalidType   = result{nil, ErrInvalidType}
	resultKeyNotFound   = result{nil, ErrKeyNotFound}
)

type task struct {
	cmd Command
	arg []byte
	ret chan result
}

type key string

type value interface {
	get() result
	set(k []byte) result
	getKey(k []byte) result
	setKey(k []byte, nv []byte) result // nv = nil - remove key
	pop() result
	push(k []byte) result
	empty() bool
}

// A Database type implements in-memory cache engine
type Database struct {
	queue   chan task
	started bool
	closing bool
	log     *log.Logger
	m       map[key]value
	t       map[key]*time.Timer
	e       EventHandler
}

// New creates new Database instance
func New(cfg Config) (*Database, error) {
	d := &Database{
		queue:   make(chan task, cfg.QueueLength),
		started: false,
		closing: false,
		log:     cfg.Log,
		m:       make(map[key]value, 1024),
		t:       make(map[key]*time.Timer, 1024),
		e:       cfg.Handler,
	}

	if d.log == nil {
		d.log = log.New(ioutil.Discard, "", 0)
	}

	go d.run()

	return d, nil
}

// Close closes in-memory cache database
func (d *Database) Close() error {
	d.closing = true

	if d.queue == nil {
		return ErrAlreadyClosed
	}

	close(d.queue)
	d.queue = nil

	return nil
}

// Exec executes single command
func (d *Database) Exec(cmd Command, arg []byte) ([]byte, error) {

	if d.closing {
		return nil, ErrAlreadyClosed
	}

	if !d.started {
		return nil, ErrNotStarted
	}

	// create channel to get result
	ret := make(chan result)

	// create and send task
	d.queue <- task{cmd: cmd, arg: arg, ret: ret}

	// wait for result
	result := <-ret

	// return result
	return result.value, result.err
}

func (d *Database) run() {
	d.started = true
	for {
		t, ok := <-d.queue
		if !ok {
			break
		}
		switch t.cmd {
		case CommandNop:
			t.ret <- resultOk
		case CommandGet:
			t.ret <- d.get(t.arg)
		case CommandSet:
			t.ret <- d.set(t.arg)
		case CommandPush:
			t.ret <- d.push(t.arg)
		case CommandPop:
			t.ret <- d.pop(t.arg)
		case CommandRemove:
			t.ret <- d.remove(t.arg)
		case CommandTTL:
			t.ret <- d.ttl(t.arg)
		case CommandKeys:
			t.ret <- d.keys(t.arg)
		default:
			t.ret <- result{nil, ErrInvalidCommand}
		}
	}

	for _, t := range d.t {
		t.Stop()
	}

	d.started = false
}

func parseArg(arg []byte) [][]byte {
	result := make([][]byte, 0, 3)

	slash := false
	start := 0
	for i, b := range arg {
		switch b {
		case '\\':
			slash = true
		case ',':
			if !slash {
				result = append(result, bytes.TrimSpace(arg[start:i]))
				start = i + 1
			}
			fallthrough
		default:
			slash = false
		}
	}

	result = append(result, bytes.TrimSpace(arg[start:]))

	return result
}

func (d *Database) get(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		if v, ok := d.m[key(args[0])]; ok {
			return v.get()
		}

		return resultNotFound

	case 2:
		if v, ok := d.m[key(args[0])]; ok {
			return v.getKey(args[1])
		}

		return resultNotFound

	default:
		return result{nil, ErrInvalidFormat}
	}
}

func (d *Database) set(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)

	switch len(args) {
	case 2:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			return v.set(args[1])
		}

		d.m[k] = str(args[1])

		return resultOk

	case 3:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			return v.setKey(args[1], args[2])
		}

		v := make(dict)
		v.setKey(args[1], args[2])
		d.m[k] = v

		return resultOk

	default:
		return resultInvalidFormat
	}
}

func (d *Database) push(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 2:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			return v.push(args[1])
		}

		v := new(list)
		v.push(args[1])
		d.m[k] = v

		return resultOk

	default:
		return resultInvalidFormat
	}
}

func (d *Database) pop(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			r := v.pop()
			if v.empty() {
				delete(d.m, k)
				if t, ok := d.t[k]; ok {
					t.Stop()
					delete(d.t, k)
				}
			}
			return r
		}
		return resultNotFound

	default:
		return resultInvalidFormat
	}
}

func (d *Database) remove(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		k := key(args[0])
		if _, ok := d.m[k]; ok {
			delete(d.m, k)
			if t, ok := d.t[k]; ok {
				t.Stop()
				delete(d.t, k)
			}
			return resultOk
		}

		return resultNotFound

	case 2:
		k := key(args[0])
		if v, ok := d.m[k]; ok {
			return v.setKey(args[1], nil)
		}

		return resultNotFound

	default:
		return resultInvalidFormat
	}
}

func (d *Database) ttl(arg []byte) result {
	if len(arg) == 0 {
		return resultInvalidFormat
	}

	args := parseArg(arg)
	switch len(args) {
	case 2:
		timeout, err := strconv.ParseUint(string(args[1]), 10, 64)
		if err != nil {
			return result{nil, err}
		}

		k := key(args[0])
		if _, ok := d.m[k]; !ok {
			return resultNotFound
		}

		duration := time.Millisecond * time.Duration(timeout)
		if t, ok := d.t[k]; ok {
			if !t.Stop() {
				<-t.C
			}
			t.Reset(duration)
		} else {
			t := time.AfterFunc(duration, func() {
				k := args[0]
				d.Exec(CommandRemove, k)
				if d.e != nil {
					d.e(EventExpired, k)
				}
			})
			d.t[k] = t
		}

		return resultOk

	default:
		return resultInvalidFormat
	}
}

func (d *Database) keys(arg []byte) result {
	if arg == nil || bytes.Equal(arg, []byte("")) {
		var r []byte
		var first = false
		for k := range d.m {
			if first {
				r = append(r, ',')
			} else {
				first = true
			}
			r = append(r, k...)
		}
		return result{r, nil}
	}

	args := parseArg(arg)
	switch len(args) {
	case 1:
		v, ok := d.m[key(args[0])]
		if !ok {
			return resultNotFound
		}

		dv, ok := v.(dict)
		if !ok {
			return resultNotFound
		}

		var r []byte
		var first = false
		for k := range dv {
			if first {
				r = append(r, ',')
			} else {
				first = true
			}
			r = append(r, k...)
		}

		return result{r, nil}

	default:
		return resultInvalidFormat
	}
}
